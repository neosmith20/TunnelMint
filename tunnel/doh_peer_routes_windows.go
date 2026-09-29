//go:build windows

/* SPDX-License-Identifier: MIT
 *
 * Copyright (C) 2026 WireHush contributors. All Rights Reserved.
 */

package tunnel

import (
	"errors"
	"fmt"
	"net/netip"
	"sort"
	"sync"
	"unsafe"

	"golang.org/x/sys/windows"

	"golang.zx2c4.com/wireguard/windows/conf"
	"golang.zx2c4.com/wireguard/windows/driver"
)

// selectDoHPeer returns the only peer that can safely own every endpoint
// address that is not already covered by a peer AllowedIPs route. If the
// existing configuration is already cryptokey-routable, no peer is needed.
func selectDoHPeer(config *conf.Config, addresses []netip.Addr) (int, bool, error) {
	if config == nil || len(config.Peers) == 0 {
		return -1, false, errors.New("encrypted DNS requires a configured WireGuard peer")
	}
	selected := -1
	needsRoute := false
	for _, address := range addresses {
		matches := make([]int, 0, len(config.Peers))
		for peerIndex, peer := range config.Peers {
			for _, allowed := range peer.AllowedIPs {
				if allowed.Contains(address) {
					matches = append(matches, peerIndex)
					break
				}
			}
		}
		if len(matches) == 0 {
			needsRoute = true
			if len(config.Peers) != 1 {
				return -1, true, fmt.Errorf("cannot safely assign DoH endpoint %s to an unambiguous WireGuard peer", address)
			}
			matches = []int{0}
		}
		if selected == -1 {
			selected = matches[0]
		}
		if needsRoute && (len(matches) != 1 || matches[0] != selected) {
			return -1, true, fmt.Errorf("DoH endpoint %s maps to an ambiguous WireGuard peer", address)
		}
	}
	return selected, needsRoute, nil
}

type dohPeerRouteManager struct {
	adapter          *driver.Adapter
	config           *conf.Config
	setConfiguration func(*driver.Adapter, *driver.Interface, uint32) error
	mu               sync.Mutex
	active           map[netip.Prefix]struct{}
	peerIndex        int
}

func newDoHPeerRouteManager(adapter *driver.Adapter, config *conf.Config) *dohPeerRouteManager {
	return &dohPeerRouteManager{
		adapter:   adapter,
		config:    config,
		active:    make(map[netip.Prefix]struct{}),
		peerIndex: -1,
		setConfiguration: func(adapter *driver.Adapter, interfaze *driver.Interface, size uint32) error {
			return adapter.SetConfiguration(interfaze, size)
		},
	}
}

func (m *dohPeerRouteManager) Add(prefix netip.Prefix) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if !prefix.IsValid() {
		return false, errors.New("invalid DoH endpoint route prefix")
	}
	if m.coveredByConfig(prefix.Addr()) {
		return false, nil
	}
	next := clonePrefixes(m.active)
	next[prefix] = struct{}{}
	peerIndex, err := m.apply(next)
	if err != nil {
		return false, err
	}
	m.active = next
	if peerIndex >= 0 {
		m.peerIndex = peerIndex
	}
	return true, nil
}

func (m *dohPeerRouteManager) Delete(prefix netip.Prefix) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.active[prefix]; !ok {
		return nil
	}
	next := clonePrefixes(m.active)
	delete(next, prefix)
	peerIndex, err := m.apply(next)
	if err != nil {
		return err
	}
	m.active = next
	if len(next) == 0 {
		m.peerIndex = -1
	} else if peerIndex >= 0 {
		m.peerIndex = peerIndex
	}
	return nil
}

func (m *dohPeerRouteManager) coveredByConfig(address netip.Addr) bool {
	if m.config == nil {
		return false
	}
	for _, peer := range m.config.Peers {
		for _, allowed := range peer.AllowedIPs {
			if allowed.Contains(address) {
				return true
			}
		}
	}
	return false
}

func (m *dohPeerRouteManager) apply(active map[netip.Prefix]struct{}) (int, error) {
	if len(active) == 0 && m.peerIndex < 0 {
		return -1, nil
	}
	if m.config == nil || len(m.config.Peers) == 0 {
		return -1, errors.New("encrypted DNS requires a configured WireGuard peer")
	}
	peerIndex := m.peerIndex
	needsRoute := len(active) > 0
	if needsRoute {
		addresses := make([]netip.Addr, 0, len(active))
		for prefix := range active {
			addresses = append(addresses, prefix.Addr())
		}
		selected, selectedNeedsRoute, err := selectDoHPeer(m.config, addresses)
		if err != nil {
			return -1, err
		}
		if !selectedNeedsRoute {
			return -1, nil
		}
		if peerIndex >= 0 && selected != peerIndex {
			return -1, fmt.Errorf("DoH endpoint peer changed from %d to %d", peerIndex, selected)
		}
		peerIndex = selected
	}
	allowed := append([]netip.Prefix(nil), m.config.Peers[peerIndex].AllowedIPs...)
	activePrefixes := make([]netip.Prefix, 0, len(active))
	for prefix := range active {
		activePrefixes = append(activePrefixes, prefix)
	}
	sort.Slice(activePrefixes, func(i, j int) bool { return activePrefixes[i].String() < activePrefixes[j].String() })
	for _, prefix := range activePrefixes {
		found := false
		for _, existing := range allowed {
			if existing == prefix {
				found = true
				break
			}
		}
		if !found {
			allowed = append(allowed, prefix)
		}
	}
	peer := &driver.Peer{
		Flags:           driver.PeerHasPublicKey | driver.PeerUpdateOnly | driver.PeerReplaceAllowedIPs,
		PublicKey:       m.config.Peers[peerIndex].PublicKey,
		AllowedIPsCount: uint32(len(allowed)),
	}
	builder := new(driver.ConfigBuilder)
	builder.Preallocate(uint32(unsafe.Sizeof(driver.Interface{}) + unsafe.Sizeof(driver.Peer{}) + uintptr(len(allowed))*unsafe.Sizeof(driver.AllowedIP{})))
	builder.AppendInterface(&driver.Interface{PeerCount: 1})
	builder.AppendPeer(peer)
	for _, prefix := range allowed {
		address := &driver.AllowedIP{Cidr: uint8(prefix.Bits())}
		copy(address.Address[:], prefix.Addr().AsSlice())
		if prefix.Addr().Is4() {
			address.AddressFamily = windows.AF_INET
		} else {
			address.AddressFamily = windows.AF_INET6
		}
		builder.AppendAllowedIP(address)
	}
	interfaze, size := builder.Interface()
	if m.setConfiguration == nil {
		return -1, errors.New("WireGuard peer configuration updater is unavailable")
	}
	if err := m.setConfiguration(m.adapter, interfaze, size); err != nil {
		return -1, fmt.Errorf("update WireGuard peer AllowedIPs for DoH endpoint: %w", err)
	}
	return peerIndex, nil
}

func clonePrefixes(prefixes map[netip.Prefix]struct{}) map[netip.Prefix]struct{} {
	clone := make(map[netip.Prefix]struct{}, len(prefixes))
	for prefix := range prefixes {
		clone[prefix] = struct{}{}
	}
	return clone
}
