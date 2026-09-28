/* SPDX-License-Identifier: MIT
 *
 * Copyright (C) 2026 TunnelMint contributors. All Rights Reserved.
 */

package bootstrap

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"strings"
	"testing"
	"time"

	"golang.zx2c4.com/wireguard/windows/doh"
)

func TestDefaultResolvers(t *testing.T) {
	want := []string{"1.1.1.1", "1.0.0.1", "9.9.9.9", "149.112.112.112", "8.8.8.8", "8.8.4.4"}
	got := DefaultResolvers()
	if len(got) != len(want) {
		t.Fatalf("default resolver count = %d, want %d", len(got), len(want))
	}
	for i, resolver := range got {
		if resolver.String() != want[i] {
			t.Errorf("default resolver %d = %s, want %s", i, resolver, want[i])
		}
	}
}

func TestResolveEndpointOrderedFailoverAndFamilies(t *testing.T) {
	first := netip.MustParseAddr("192.0.2.1")
	second := netip.MustParseAddr("192.0.2.2")
	ipv4 := netip.MustParseAddr("198.51.100.10")
	ipv6 := netip.MustParseAddr("2001:db8::10")
	var calls []netip.Addr
	resolver := NewResolver([]netip.Addr{first, second})
	resolver.Lookup = func(_ context.Context, server netip.Addr, _ string) ([]netip.Addr, error) {
		calls = append(calls, server)
		if server == first {
			return nil, errors.New("resolver unavailable")
		}
		return []netip.Addr{ipv4, ipv6}, nil
	}
	addresses, err := resolver.ResolveEndpoint(context.Background(), "https://dns.example.test/dns-query")
	if err != nil {
		t.Fatal(err)
	}
	if len(calls) != 2 || calls[0] != first || calls[1] != second {
		t.Fatalf("resolver call order = %v", calls)
	}
	if len(addresses) != 2 || addresses[0] != ipv4 || addresses[1] != ipv6 {
		t.Fatalf("addresses = %v", addresses)
	}
}

func TestResolveEndpointIPLiteralBypass(t *testing.T) {
	called := false
	resolver := NewResolver(nil)
	resolver.Lookup = func(context.Context, netip.Addr, string) ([]netip.Addr, error) {
		called = true
		return nil, errors.New("must not be called")
	}
	addresses, err := resolver.ResolveEndpoint(context.Background(), "https://192.0.2.10/dns-query")
	if err != nil {
		t.Fatal(err)
	}
	if called || len(addresses) != 1 || addresses[0].String() != "192.0.2.10" {
		t.Fatalf("literal bypass returned %v, called=%v", addresses, called)
	}
}

func TestResolveEndpointFailsClosedWithoutSystemFallback(t *testing.T) {
	called := 0
	resolver := NewResolver([]netip.Addr{netip.MustParseAddr("192.0.2.1"), netip.MustParseAddr("192.0.2.2")})
	resolver.Lookup = func(context.Context, netip.Addr, string) ([]netip.Addr, error) {
		called++
		return nil, errors.New("no response")
	}
	if _, err := resolver.ResolveEndpoint(context.Background(), "https://dns.example.test/dns-query"); err == nil {
		t.Fatal("expected bootstrap failure")
	}
	if called != 2 {
		t.Fatalf("lookup calls = %d, want 2", called)
	}
}

func TestResolveEndpointTimeout(t *testing.T) {
	resolver := NewResolver([]netip.Addr{netip.MustParseAddr("192.0.2.1")})
	resolver.Timeout = 20 * time.Millisecond
	resolver.Lookup = func(ctx context.Context, _ netip.Addr, _ string) ([]netip.Addr, error) {
		<-ctx.Done()
		return nil, ctx.Err()
	}
	_, err := resolver.ResolveEndpoint(context.Background(), "https://dns.example.test/dns-query")
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("timeout error = %v", err)
	}
}

func TestResolveEndpointTimeoutThenLaterResolverSucceeds(t *testing.T) {
	first := netip.MustParseAddr("192.0.2.1")
	second := netip.MustParseAddr("192.0.2.2")
	resolver := NewResolver([]netip.Addr{first, second})
	resolver.Timeout = 80 * time.Millisecond
	var calls []netip.Addr
	resolver.Lookup = func(ctx context.Context, server netip.Addr, _ string) ([]netip.Addr, error) {
		calls = append(calls, server)
		if server == first {
			<-ctx.Done()
			return nil, ctx.Err()
		}
		return []netip.Addr{netip.MustParseAddr("198.51.100.20")}, nil
	}
	addresses, err := resolver.ResolveEndpoint(context.Background(), "https://dns.example.test/dns-query")
	if err != nil {
		t.Fatal(err)
	}
	if len(calls) != 2 || calls[0] != first || calls[1] != second {
		t.Fatalf("resolver call order = %v", calls)
	}
	if len(addresses) != 1 || addresses[0].String() != "198.51.100.20" {
		t.Fatalf("addresses = %v", addresses)
	}
}

func TestEndpointCacheHitAndInvalidation(t *testing.T) {
	resolver := NewResolver([]netip.Addr{netip.MustParseAddr("192.0.2.1")})
	resolver.Cache = NewCache(time.Minute, 4)
	calls := 0
	resolver.Lookup = func(context.Context, netip.Addr, string) ([]netip.Addr, error) {
		calls++
		return []netip.Addr{netip.MustParseAddr("198.51.100.10")}, nil
	}
	endpoint := "https://dns.example.test/dns-query"
	if _, err := resolver.ResolveEndpoint(context.Background(), endpoint); err != nil {
		t.Fatal(err)
	}
	if _, err := resolver.ResolveEndpoint(context.Background(), endpoint); err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Fatalf("lookup calls after cache hit = %d, want 1", calls)
	}
	resolver.InvalidateEndpoint(endpoint)
	if _, err := resolver.ResolveEndpoint(context.Background(), endpoint); err != nil {
		t.Fatal(err)
	}
	if calls != 2 {
		t.Fatalf("lookup calls after invalidation = %d, want 2", calls)
	}
}

func TestBootstrappedDialPreservesTLSHostnameAndPath(t *testing.T) {
	const path = "/dns-query/provider"
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.Host, "example.com:") {
			t.Errorf("Host = %q, want example.com with port", r.Host)
		}
		if r.URL.Path != path {
			t.Errorf("path = %q, want %q", r.URL.Path, path)
		}
		w.Header().Set("Content-Type", "application/dns-message")
		_, _ = w.Write([]byte{0x12, 0x34, 0x81, 0x80})
	}))
	defer server.Close()
	endpoint := strings.Replace(server.URL, "https://127.0.0.1", "https://example.com", 1) + path
	resolver := NewResolver([]netip.Addr{netip.MustParseAddr("192.0.2.1")})
	resolver.Lookup = func(context.Context, netip.Addr, string) ([]netip.Addr, error) {
		return []netip.Addr{netip.MustParseAddr("127.0.0.1")}, nil
	}
	addresses, err := resolver.ResolveEndpoint(context.Background(), endpoint)
	if err != nil {
		t.Fatal(err)
	}
	httpClient := doh.NewHTTPClientForAddresses(addresses, server.Client())
	dohClient, err := doh.NewClient(endpoint, doh.Options{HTTPClient: httpClient})
	if err != nil {
		t.Fatal(err)
	}
	response, err := dohClient.Query(context.Background(), []byte{0x12, 0x34})
	if err != nil {
		t.Fatal(err)
	}
	if len(response) != 4 {
		t.Fatalf("response length = %d", len(response))
	}
}
