//go:build windows

/* SPDX-License-Identifier: MIT
 *
 * Copyright (C) 2026 TunnelMint contributors. All Rights Reserved.
 */

package tunnel

import (
	"errors"
	"net/netip"
	"reflect"
	"testing"

	"golang.org/x/sys/windows"

	"golang.zx2c4.com/wireguard/windows/conf"
	"golang.zx2c4.com/wireguard/windows/driver"
)

func newTestDoHPeerRouteManager(config *conf.Config) (*dohPeerRouteManager, *[][]netip.Prefix) {
	manager := newDoHPeerRouteManager(nil, config)
	var updates [][]netip.Prefix
	manager.setConfiguration = func(_ *driver.Adapter, interfaze *driver.Interface, _ uint32) error {
		peer := interfaze.FirstPeer()
		allowed := make([]netip.Prefix, 0, peer.AllowedIPsCount)
		for i := uint32(0); i < peer.AllowedIPsCount; i++ {
			item := peer.FirstAllowedIP()
			for step := uint32(0); step < i; step++ {
				item = item.NextAllowedIP()
			}
			var address netip.Addr
			if item.AddressFamily == windows.AF_INET {
				var bytes [4]byte
				copy(bytes[:], item.Address[:4])
				address = netip.AddrFrom4(bytes)
			} else {
				address, _ = netip.AddrFromSlice(item.Address[:])
			}
			allowed = append(allowed, netip.PrefixFrom(address, int(item.Cidr)))
		}
		updates = append(updates, allowed)
		return nil
	}
	return manager, &updates
}

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

func TestDoHPeerRouteManagerRestoresOriginalAllowedIPs(t *testing.T) {
	original := []netip.Prefix{netip.MustParsePrefix("10.0.0.0/8")}
	manager, updates := newTestDoHPeerRouteManager(&conf.Config{Peers: []conf.Peer{{AllowedIPs: original}}})
	temporary := netip.MustParsePrefix("198.51.100.9/32")
	if owned, err := manager.Add(temporary); err != nil || !owned {
		t.Fatalf("Add() = owned %v err %v", owned, err)
	}
	if err := manager.Delete(temporary); err != nil {
		t.Fatal(err)
	}
	if err := manager.Delete(temporary); err != nil {
		t.Fatal(err)
	}
	want := [][]netip.Prefix{append(append([]netip.Prefix(nil), original...), temporary), original}
	if !reflect.DeepEqual(*updates, want) {
		t.Fatalf("driver updates = %v, want %v", *updates, want)
	}
}

func TestDoHPeerRouteManagerReplacesTemporaryEndpoint(t *testing.T) {
	original := []netip.Prefix{netip.MustParsePrefix("10.0.0.0/8")}
	manager, updates := newTestDoHPeerRouteManager(&conf.Config{Peers: []conf.Peer{{AllowedIPs: original}}})
	oldPrefix := netip.MustParsePrefix("198.51.100.9/32")
	newPrefix := netip.MustParsePrefix("203.0.113.9/32")
	if _, err := manager.Add(oldPrefix); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.Add(newPrefix); err != nil {
		t.Fatal(err)
	}
	if err := manager.Delete(oldPrefix); err != nil {
		t.Fatal(err)
	}
	if got := (*updates)[len(*updates)-1]; !reflect.DeepEqual(got, []netip.Prefix{original[0], newPrefix}) {
		t.Fatalf("replacement update = %v", got)
	}
}

func TestDoHPeerRouteManagerFailurePreservesLiveState(t *testing.T) {
	original := []netip.Prefix{netip.MustParsePrefix("10.0.0.0/8")}
	manager, _ := newTestDoHPeerRouteManager(&conf.Config{Peers: []conf.Peer{{AllowedIPs: original}}})
	temporary := netip.MustParsePrefix("198.51.100.9/32")
	updateErr := errors.New("injected driver update failure")
	manager.setConfiguration = func(*driver.Adapter, *driver.Interface, uint32) error { return updateErr }
	if _, err := manager.Add(temporary); !errors.Is(err, updateErr) {
		t.Fatalf("Add() error = %v, want injected failure", err)
	}
	if len(manager.active) != 0 || manager.peerIndex != -1 {
		t.Fatalf("failed add changed manager state: active=%v peer=%d", manager.active, manager.peerIndex)
	}
	manager.setConfiguration = func(_ *driver.Adapter, _ *driver.Interface, _ uint32) error { return nil }
	if _, err := manager.Add(temporary); err != nil {
		t.Fatal(err)
	}
	manager.setConfiguration = func(*driver.Adapter, *driver.Interface, uint32) error { return updateErr }
	if err := manager.Delete(temporary); !errors.Is(err, updateErr) {
		t.Fatalf("Delete() error = %v, want injected failure", err)
	}
	if len(manager.active) != 1 || manager.peerIndex != 0 {
		t.Fatalf("failed delete changed manager state: active=%v peer=%d", manager.active, manager.peerIndex)
	}
}
