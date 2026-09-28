//go:build windows

/* SPDX-License-Identifier: MIT
 *
 * Copyright (C) 2026 TunnelMint contributors. All Rights Reserved.
 */

package tunnel

import (
	"net/netip"
	"testing"

	"golang.zx2c4.com/wireguard/windows/conf"
)

func TestSelectDoHPeerRequiresUnambiguousSplitTunnelPath(t *testing.T) {
	config := &conf.Config{Peers: []conf.Peer{
		{AllowedIPs: []netip.Prefix{netip.MustParsePrefix("10.0.0.0/8")}},
		{AllowedIPs: []netip.Prefix{netip.MustParsePrefix("192.0.2.0/24")}},
	}}
	if _, _, err := selectDoHPeer(config, []netip.Addr{netip.MustParseAddr("198.51.100.5")}); err == nil {
		t.Fatal("unowned split-tunnel endpoint was assigned without an unambiguous peer")
	}
	peer, needsRoute, err := selectDoHPeer(&conf.Config{Peers: []conf.Peer{{AllowedIPs: []netip.Prefix{netip.MustParsePrefix("10.0.0.0/8")}}}}, []netip.Addr{netip.MustParseAddr("198.51.100.5")})
	if err != nil || peer != 0 || !needsRoute {
		t.Fatalf("single-peer endpoint selection = peer %d needsRoute %v err %v", peer, needsRoute, err)
	}
}

func TestSelectDoHPeerLeavesFullTunnelCryptokeyRoutingAlone(t *testing.T) {
	config := &conf.Config{Peers: []conf.Peer{{AllowedIPs: []netip.Prefix{netip.MustParsePrefix("0.0.0.0/0")}}}}
	peer, needsRoute, err := selectDoHPeer(config, []netip.Addr{netip.MustParseAddr("198.51.100.5")})
	if err != nil || peer != 0 || needsRoute {
		t.Fatalf("full-tunnel endpoint selection = peer %d needsRoute %v err %v", peer, needsRoute, err)
	}
}
