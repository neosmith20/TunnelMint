/* SPDX-License-Identifier: MIT
 *
 * Copyright (C) 2026 TunnelMint contributors. All Rights Reserved.
 */

// Package dohruntime coordinates the encrypted-DNS activation stages that
// surround a running tunnel. Platform-specific hooks own the actual route,
// proxy, and Windows DNS operations.
package dohruntime

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/netip"
	"sync"
)

type Stage uint8

const (
	StageCreated Stage = iota
	StageBootstrapped
	StageRouted
	StageVerified
	StageProxyStarted
	StageDNSConfigured
	StageReady
)

type Config struct {
	Endpoint            string
	PeerEndpointAddress []netip.Addr
}

type Hooks struct {
	Bootstrap  func(context.Context, string) ([]netip.Addr, error)
	AddRoute   func(netip.Prefix) (owned bool, err error)
	DelRoute   func(netip.Prefix) error
	Verify     func(context.Context, string, []netip.Addr) error
	StartProxy func(context.Context, string, []netip.Addr) (io.Closer, error)
	SetDNS     func() (restore func() error, err error)
}

type Session struct {
	mu       sync.Mutex
	stage    Stage
	routes   []netip.Prefix
	delRoute func(netip.Prefix) error
	proxy    io.Closer
	restore  func() error
	hooks    Hooks
	config   Config
	closed   bool
	lastErr  error
}

func Activate(ctx context.Context, config Config, hooks Hooks) (*Session, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if config.Endpoint == "" {
		return nil, errors.New("encrypted DNS endpoint is required")
	}
	if hooks.Bootstrap == nil || hooks.AddRoute == nil || hooks.DelRoute == nil || hooks.Verify == nil || hooks.StartProxy == nil || hooks.SetDNS == nil {
		return nil, errors.New("encrypted DNS lifecycle hooks are incomplete")
	}
	s := &Session{stage: StageCreated, delRoute: hooks.DelRoute, hooks: hooks, config: config}
	rollback := func(err error) (*Session, error) {
		s.lastErr = err
		_ = s.Close()
		return nil, err
	}
	candidates, err := hooks.Bootstrap(ctx, config.Endpoint)
	if err != nil {
		return rollback(fmt.Errorf("bootstrap DoH endpoint: %w", err))
	}
	candidates, err = selectRouteCandidates(candidates, config.PeerEndpointAddress)
	if err != nil {
		return rollback(err)
	}
	s.stage = StageBootstrapped
	for _, address := range candidates {
		prefix := netip.PrefixFrom(address, address.BitLen())
		owned, err := hooks.AddRoute(prefix)
		if err != nil {
			return rollback(fmt.Errorf("add DoH host route %s: %w", prefix, err))
		}
		if owned {
			s.routes = append(s.routes, prefix)
		}
	}
	s.stage = StageRouted
	if err := hooks.Verify(ctx, config.Endpoint, candidates); err != nil {
		return rollback(fmt.Errorf("verify DoH transport: %w", err))
	}
	s.stage = StageVerified
	s.proxy, err = hooks.StartProxy(ctx, config.Endpoint, candidates)
	if err != nil {
		return rollback(fmt.Errorf("start local DNS proxy: %w", err))
	}
	s.stage = StageProxyStarted
	s.restore, err = hooks.SetDNS()
	if err != nil {
		return rollback(fmt.Errorf("configure Windows DNS: %w", err))
	}
	s.stage = StageDNSConfigured
	s.stage = StageReady
	return s, nil
}

func (s *Session) Stage() Stage {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.stage
}

func (s *Session) Routes() []netip.Prefix {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]netip.Prefix(nil), s.routes...)
}

func (s *Session) LastError() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.lastErr
}

// Refresh verifies a new ordered endpoint candidate set and atomically moves
// the proxy and TunnelMint-owned host routes to it. Existing DNS settings stay
// in place while the replacement is prepared.
func (s *Session) Refresh(ctx context.Context, candidates []netip.Addr) error {
	s.mu.Lock()
	if s.closed || s.stage != StageReady {
		s.mu.Unlock()
		return errors.New("encrypted DNS session is not ready")
	}
	hooks, endpoint, peers, oldRoutes, oldProxy := s.hooks, s.config.Endpoint, s.config.PeerEndpointAddress, append([]netip.Prefix(nil), s.routes...), s.proxy
	s.mu.Unlock()
	selected, err := selectRouteCandidates(candidates, peers)
	if err != nil {
		return err
	}
	oldSet := make(map[netip.Prefix]bool, len(oldRoutes))
	for _, route := range oldRoutes {
		oldSet[route] = true
	}
	newRoutes := make([]netip.Prefix, 0, len(selected))
	addedRoutes := make([]netip.Prefix, 0, len(selected))
	for _, address := range selected {
		prefix := netip.PrefixFrom(address, address.BitLen())
		if oldSet[prefix] {
			newRoutes = append(newRoutes, prefix)
			delete(oldSet, prefix)
			continue
		}
		owned, err := hooks.AddRoute(prefix)
		if err != nil {
			for _, added := range addedRoutes {
				_ = hooks.DelRoute(added)
			}
			return fmt.Errorf("add refreshed DoH host route %s: %w", prefix, err)
		}
		if owned {
			newRoutes = append(newRoutes, prefix)
			addedRoutes = append(addedRoutes, prefix)
		}
	}
	if err := hooks.Verify(ctx, endpoint, selected); err != nil {
		for _, route := range newRoutes {
			if !containsPrefix(oldRoutes, route) {
				_ = hooks.DelRoute(route)
			}
		}
		return fmt.Errorf("verify refreshed DoH transport: %w", err)
	}
	newProxy, err := hooks.StartProxy(ctx, endpoint, selected)
	if err != nil {
		for _, route := range newRoutes {
			if !containsPrefix(oldRoutes, route) {
				_ = hooks.DelRoute(route)
			}
		}
		return fmt.Errorf("start refreshed local DNS proxy: %w", err)
	}
	for route := range oldSet {
		if err := hooks.DelRoute(route); err != nil {
			_ = newProxy.Close()
			return fmt.Errorf("remove stale DoH host route %s: %w", route, err)
		}
	}
	if oldProxy != nil {
		_ = oldProxy.Close()
	}
	s.mu.Lock()
	s.routes, s.proxy = newRoutes, newProxy
	s.mu.Unlock()
	return nil
}

func containsPrefix(routes []netip.Prefix, want netip.Prefix) bool {
	for _, route := range routes {
		if route == want {
			return true
		}
	}
	return false
}

func (s *Session) Close() error {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return nil
	}
	s.closed = true
	restore, proxy, delRoute, routes := s.restore, s.proxy, s.delRoute, append([]netip.Prefix(nil), s.routes...)
	s.restore, s.proxy, s.delRoute, s.routes = nil, nil, nil, nil
	s.mu.Unlock()

	var firstErr error
	if restore != nil {
		if err := restore(); err != nil {
			firstErr = err
		}
	}
	if proxy != nil {
		if err := proxy.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	for i := len(routes) - 1; i >= 0; i-- {
		if err := delRoute(routes[i]); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func selectRouteCandidates(candidates, peerEndpoints []netip.Addr) ([]netip.Addr, error) {
	unsafe := make(map[netip.Addr]bool, len(peerEndpoints))
	for _, address := range peerEndpoints {
		if address.IsValid() {
			unsafe[address] = true
		}
	}
	seen := make(map[netip.Addr]bool, len(candidates))
	selected := make([]netip.Addr, 0, len(candidates))
	for _, address := range candidates {
		if !address.IsValid() || address.IsUnspecified() || address.IsMulticast() {
			continue
		}
		if unsafe[address] {
			continue
		}
		if !seen[address] {
			seen[address] = true
			selected = append(selected, address)
		}
	}
	if len(selected) == 0 {
		return nil, errors.New("no safe DoH endpoint address can be routed through the tunnel")
	}
	return selected, nil
}
