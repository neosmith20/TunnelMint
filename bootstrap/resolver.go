/* SPDX-License-Identifier: MIT
 *
 * Copyright (C) 2026 WireHush contributors. All Rights Reserved.
 */

// Package bootstrap resolves DoH endpoint hostnames without using the system
// resolver path that encrypted DNS mode will later replace.
package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/netip"
	"net/url"
	"strings"
	"sync"
	"time"
)

const (
	DefaultTimeout    = 5 * time.Second
	DefaultCacheTTL   = 5 * time.Minute
	DefaultCacheLimit = 64
)

var defaultResolvers = []netip.Addr{
	netip.MustParseAddr("1.1.1.1"),
	netip.MustParseAddr("1.0.0.1"),
	netip.MustParseAddr("8.8.8.8"),
	netip.MustParseAddr("8.8.4.4"),
	netip.MustParseAddr("4.2.2.1"),
	netip.MustParseAddr("4.2.2.2"),
	netip.MustParseAddr("2606:4700:4700::1111"),
	netip.MustParseAddr("2606:4700:4700::1001"),
	netip.MustParseAddr("2001:4860:4860::8888"),
	netip.MustParseAddr("2001:4860:4860::8844"),
	netip.MustParseAddr("2620:fe::11"),
	netip.MustParseAddr("2620:fe::fe:11"),
}

// DefaultResolvers returns the built-in ordered bootstrap resolver list.
func DefaultResolvers() []netip.Addr {
	return append([]netip.Addr(nil), defaultResolvers...)
}

// ResolversForFamilies keeps enabled bootstrap resolvers whose address family
// can carry the encrypted-DNS path. Callers select the families from the
// tunnel configuration; this function never substitutes another family.
func ResolversForFamilies(resolvers []netip.Addr, allowIPv4, allowIPv6 bool) []netip.Addr {
	filtered := make([]netip.Addr, 0, len(resolvers))
	for _, resolver := range resolvers {
		if (resolver.Is4() && allowIPv4) || (resolver.Is6() && allowIPv6) {
			filtered = append(filtered, resolver)
		}
	}
	return filtered
}

// LookupFunc resolves host through one explicitly selected bootstrap resolver.
// It is injectable so tests and later platform integration can avoid public DNS.
type LookupFunc func(context.Context, netip.Addr, string) ([]netip.Addr, error)

type cacheEntry struct {
	addresses []netip.Addr
	expires   time.Time
	used      time.Time
}

// Cache is a small bounded endpoint-address cache. Entries expire after ttl;
// callers should call Invalidate or Clear after tunnel/network changes.
type Cache struct {
	mu         sync.Mutex
	entries    map[string]cacheEntry
	ttl        time.Duration
	maxEntries int
	now        func() time.Time
}

func NewCache(ttl time.Duration, maxEntries int) *Cache {
	if ttl <= 0 {
		ttl = DefaultCacheTTL
	}
	if maxEntries <= 0 {
		maxEntries = DefaultCacheLimit
	}
	return &Cache{entries: make(map[string]cacheEntry), ttl: ttl, maxEntries: maxEntries, now: time.Now}
}

func (c *Cache) Get(key string) ([]netip.Addr, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	entry, ok := c.entries[key]
	if !ok {
		return nil, false
	}
	now := c.now()
	if !now.Before(entry.expires) {
		delete(c.entries, key)
		return nil, false
	}
	entry.used = now
	c.entries[key] = entry
	return append([]netip.Addr(nil), entry.addresses...), true
}

func (c *Cache) Put(key string, addresses []netip.Addr) {
	c.mu.Lock()
	defer c.mu.Unlock()
	now := c.now()
	if _, ok := c.entries[key]; !ok && len(c.entries) >= c.maxEntries {
		var oldestKey string
		var oldest time.Time
		for candidate, entry := range c.entries {
			if oldestKey == "" || entry.used.Before(oldest) {
				oldestKey, oldest = candidate, entry.used
			}
		}
		delete(c.entries, oldestKey)
	}
	c.entries[key] = cacheEntry{
		addresses: append([]netip.Addr(nil), addresses...),
		expires:   now.Add(c.ttl),
		used:      now,
	}
}

func (c *Cache) Invalidate(key string) {
	c.mu.Lock()
	delete(c.entries, key)
	c.mu.Unlock()
}

func (c *Cache) Clear() {
	c.mu.Lock()
	c.entries = make(map[string]cacheEntry)
	c.mu.Unlock()
}

// Resolver resolves DoH endpoint hostnames through ordered IP bootstrap
// resolvers. Lookup is intentionally explicit so no system resolver fallback
// can occur when encrypted DNS is selected.
type Resolver struct {
	Resolvers []netip.Addr
	Timeout   time.Duration
	Cache     *Cache
	Lookup    LookupFunc
}

func NewResolver(resolvers []netip.Addr) *Resolver {
	if len(resolvers) == 0 {
		resolvers = DefaultResolvers()
	}
	return &Resolver{Resolvers: append([]netip.Addr(nil), resolvers...), Timeout: DefaultTimeout, Cache: NewCache(DefaultCacheTTL, DefaultCacheLimit)}
}

func validateEndpoint(endpoint string) (*url.URL, error) {
	parsed, err := url.ParseRequestURI(endpoint)
	if err != nil || !strings.EqualFold(parsed.Scheme, "https") || parsed.Hostname() == "" {
		return nil, fmt.Errorf("invalid HTTPS DoH endpoint %q", endpoint)
	}
	return parsed, nil
}

// ResolveEndpoint returns ordered IP candidates for endpoint. IP-literal
// endpoints bypass bootstrap and are returned directly.
func (r *Resolver) ResolveEndpoint(ctx context.Context, endpoint string) ([]netip.Addr, error) {
	parsed, err := validateEndpoint(endpoint)
	if err != nil {
		return nil, err
	}
	return r.ResolveHost(ctx, parsed.Hostname())
}

// ResolveHost returns ordered IP candidates for a hostname through the
// configured bootstrap resolvers. It is used before a tunnel exists when a
// peer endpoint must avoid the ordinary system resolver.
func (r *Resolver) ResolveHost(ctx context.Context, host string) ([]netip.Addr, error) {
	host = strings.TrimSpace(host)
	if host == "" {
		return nil, errors.New("bootstrap hostname is empty")
	}
	if address, err := netip.ParseAddr(host); err == nil {
		return []netip.Addr{address}, nil
	}
	cacheKey := strings.ToLower(host)
	if r.Cache != nil {
		if addresses, ok := r.Cache.Get(cacheKey); ok {
			return addresses, nil
		}
	}
	timeout := r.Timeout
	if timeout <= 0 {
		timeout = DefaultTimeout
	}
	resolveCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	lookup := r.Lookup
	if lookup == nil {
		lookup = lookupWithResolver
	}
	var failures []string
	var lastLookupErr error
	deadline, hasDeadline := resolveCtx.Deadline()
	for index, resolver := range r.Resolvers {
		if !resolver.IsValid() {
			failures = append(failures, "invalid resolver address")
			continue
		}
		attemptCtx := resolveCtx
		var attemptCancel context.CancelFunc
		if hasDeadline {
			remaining := len(r.Resolvers) - index
			attemptBudget := time.Until(deadline) / time.Duration(remaining)
			if attemptBudget <= 0 {
				break
			}
			// Keep each dead resolver from consuming the complete overall
			// deadline, while still allowing the final resolver the remaining
			// time in the bootstrap budget.
			attemptCtx, attemptCancel = context.WithTimeout(resolveCtx, attemptBudget)
		}
		addresses, lookupErr := lookup(attemptCtx, resolver, host)
		if attemptCancel != nil {
			attemptCancel()
		}
		if lookupErr != nil {
			lastLookupErr = lookupErr
			failures = append(failures, fmt.Sprintf("%s: %v", resolver, lookupErr))
			if resolveCtx.Err() != nil {
				break
			}
			continue
		}
		addresses = addressesForFamily(addresses, resolver)
		if len(addresses) == 0 {
			failures = append(failures, fmt.Sprintf("%s: no addresses", resolver))
			continue
		}
		if r.Cache != nil {
			r.Cache.Put(cacheKey, addresses)
		}
		return addresses, nil
	}
	if resolveCtx.Err() != nil {
		return nil, fmt.Errorf("bootstrap resolution for %q failed: %w", host, resolveCtx.Err())
	}
	if errors.Is(lastLookupErr, context.DeadlineExceeded) || errors.Is(lastLookupErr, context.Canceled) {
		return nil, fmt.Errorf("bootstrap resolution for %q failed: %w", host, lastLookupErr)
	}
	return nil, fmt.Errorf("bootstrap resolution for %q failed: %s", host, strings.Join(failures, "; "))
}

func addressesForFamily(addresses []netip.Addr, resolver netip.Addr) []netip.Addr {
	filtered := make([]netip.Addr, 0, len(addresses))
	for _, address := range addresses {
		if (resolver.Is4() && address.Is4()) || (resolver.Is6() && address.Is6()) {
			filtered = append(filtered, address)
		}
	}
	return uniqueAddresses(filtered)
}

// InvalidateEndpoint removes a cached hostname after a tunnel/network change.
func (r *Resolver) InvalidateEndpoint(endpoint string) {
	if r.Cache == nil {
		return
	}
	parsed, err := validateEndpoint(endpoint)
	if err == nil {
		r.Cache.Invalidate(strings.ToLower(parsed.Hostname()))
	}
}

// InvalidateAll removes all cached endpoint addresses.
func (r *Resolver) InvalidateAll() {
	if r.Cache != nil {
		r.Cache.Clear()
	}
}

func uniqueAddresses(addresses []netip.Addr) []netip.Addr {
	seen := make(map[netip.Addr]struct{}, len(addresses))
	unique := make([]netip.Addr, 0, len(addresses))
	for _, address := range addresses {
		if !address.IsValid() {
			continue
		}
		if _, ok := seen[address]; ok {
			continue
		}
		seen[address] = struct{}{}
		unique = append(unique, address)
	}
	return unique
}

func lookupWithResolver(ctx context.Context, resolver netip.Addr, host string) ([]netip.Addr, error) {
	resolverConfig := &net.Resolver{
		PreferGo:     true,
		StrictErrors: true,
		Dial: func(ctx context.Context, network, _ string) (net.Conn, error) {
			return (&net.Dialer{}).DialContext(ctx, network, net.JoinHostPort(resolver.String(), "53"))
		},
	}
	addresses, err := resolverConfig.LookupNetIP(ctx, "ip", host)
	if err != nil {
		return nil, err
	}
	return uniqueAddresses(addresses), nil
}
