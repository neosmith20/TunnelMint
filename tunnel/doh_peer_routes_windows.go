//go:build windows

/* SPDX-License-Identifier: MIT
 *
 * Copyright (C) 2026 TunnelMint contributors. All Rights Reserved.
 */

package tunnel

import (
	"errors"
	"fmt"
	"net/netip"
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
	adapter *driver.Adapter
	config  *conf.Config
	mu      sync.Mutex
	active  map[netip.Prefix]struct{}
}

func newDoHPeerRouteManager(adapter *driver.Adapter, config *conf.Config) *dohPeerRouteManager {
	return &dohPeerRouteManager{adapter: adapter, config: config, active: make(map[netip.Prefix]struct{})}
}

func (m *dohPeerRouteManager) Add(prefix netip.Prefix) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.coveredByConfig(prefix.Addr()) {
		return false, nil
	}
	m.active[prefix] = struct{}{}
	if err := m.apply(); err != nil {
		delete(m.active, prefix)
		return false, err
	}
	return true, nil
}

func (m *dohPeerRouteManager) Delete(prefix netip.Prefix) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.active[prefix]; !ok {
		return nil
	}
	delete(m.active, prefix)
	return m.apply()
}

func (m *dohPeerRouteManager) coveredByConfig(address netip.Addr) bool {
	for _, peer := range m.config.Peers {
		for _, allowed := range peer.AllowedIPs {
			if allowed.Contains(address) {
				return true
			}
		}
	}
	return false
}

func (m *dohPeerRouteManager) apply() error {
	addresses := make([]netip.Addr, 0, len(m.active))
	for prefix := range m.active {
		addresses = append(addresses, prefix.Addr())
	}
	peerIndex, needsRoute, err := selectDoHPeer(m.config, addresses)
	if err != nil {
		return err
	}
	if !needsRoute {
		return nil
	}
	allowed := append([]netip.Prefix(nil), m.config.Peers[peerIndex].AllowedIPs...)
	for prefix := range m.active {
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
	if err := m.adapter.SetConfiguration(interfaze, size); err != nil {
		return fmt.Errorf("update WireGuard peer AllowedIPs for DoH endpoint: %w", err)
	}
	return nil
}
