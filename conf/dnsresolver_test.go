//go:build windows

/* SPDX-License-Identifier: MIT
 *
 * Copyright (C) 2026 TunnelMint contributors. All Rights Reserved.
 */

package conf

import (
	"errors"
	"net/netip"
	"testing"
)

func TestResolveEndpointsWithUsesExplicitResolverAndPreservesPort(t *testing.T) {
	config := &Config{Peers: []Peer{
		{Endpoint: Endpoint{Host: "peer.example", Port: 51820}},
		{Endpoint: Endpoint{Host: "192.0.2.8", Port: 51821}},
	}}
	var calls []string
	err := config.ResolveEndpointsWith(func(host string) ([]netip.Addr, error) {
		calls = append(calls, host)
		return []netip.Addr{netip.MustParseAddr("198.51.100.8")}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(calls) != 1 || calls[0] != "peer.example" {
		t.Fatalf("resolver calls = %v", calls)
	}
	if got := config.Peers[0].Endpoint; got.Host != "198.51.100.8" || got.Port != 51820 {
		t.Fatalf("resolved endpoint = %#v", got)
	}
	if got := config.Peers[1].Endpoint; got.Host != "192.0.2.8" || got.Port != 51821 {
		t.Fatalf("literal endpoint changed = %#v", got)
	}
}

func TestResolveEndpointsWithFailsClosed(t *testing.T) {
	config := &Config{Peers: []Peer{{Endpoint: Endpoint{Host: "peer.example", Port: 51820}}}}
	wantErr := errors.New("bootstrap unavailable")
	err := config.ResolveEndpointsWith(func(string) ([]netip.Addr, error) { return nil, wantErr })
	if !errors.Is(err, wantErr) {
		t.Fatalf("error = %v, want %v", err, wantErr)
	}
	if config.Peers[0].Endpoint.Host != "peer.example" {
		t.Fatal("failed resolution changed endpoint")
	}
}
