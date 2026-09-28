/* SPDX-License-Identifier: MIT
 *
 * Copyright (C) 2019-2026 WireGuard LLC. All Rights Reserved.
 */

package conf

import (
	"fmt"
	"log"
	"net/netip"
	"time"
	"unsafe"

	"golang.zx2c4.com/wireguard/windows/tunnel/winipcfg"

	"golang.org/x/sys/windows"
	"golang.zx2c4.com/wireguard/windows/services"
)

func resolveHostname(name string) (resolvedIPString string, err error) {
	maxTries := 10
	if services.StartedAtBoot() {
		maxTries *= 3
	}
	for i := 0; i < maxTries; i++ {
		if i > 0 {
			time.Sleep(time.Second * 4)
		}
		resolvedIPString, err = resolveHostnameOnce(name)
		if err == nil {
			return
		}
		if err == windows.WSATRY_AGAIN {
			log.Printf("Temporary DNS error when resolving %s, so sleeping for 4 seconds", name)
			continue
		}
		if err == windows.WSAHOST_NOT_FOUND && services.StartedAtBoot() {
			log.Printf("Host not found when resolving %s at boot time, so sleeping for 4 seconds", name)
			continue
		}
		return
	}
	return
}

func resolveHostnameOnce(name string) (resolvedIPString string, err error) {
	hints := windows.AddrinfoW{
		Family:   windows.AF_UNSPEC,
		Socktype: windows.SOCK_DGRAM,
		Protocol: windows.IPPROTO_IP,
	}
	var result *windows.AddrinfoW
	name16, err := windows.UTF16PtrFromString(name)
	if err != nil {
		return
	}
	err = windows.GetAddrInfoW(name16, nil, &hints, &result)
	if err != nil {
		return
	}
	if result == nil {
		err = windows.WSAHOST_NOT_FOUND
		return
	}
	defer windows.FreeAddrInfoW(result)
	var v6 netip.Addr
	for ; result != nil; result = result.Next {
		if result.Family != windows.AF_INET && result.Family != windows.AF_INET6 {
			continue
		}
		addr := (*winipcfg.RawSockaddrInet)(unsafe.Pointer(result.Addr)).Addr()
		if addr.Is4() {
			return addr.String(), nil
		} else if !v6.IsValid() && addr.Is6() {
			v6 = addr
		}
	}
	if v6.IsValid() {
		return v6.String(), nil
	}
	err = windows.WSAHOST_NOT_FOUND
	return
}

func (config *Config) ResolveEndpoints() error {
	return config.ResolveEndpointsWith(func(host string) ([]netip.Addr, error) {
		resolved, err := resolveHostname(host)
		if err != nil {
			return nil, err
		}
		address, err := netip.ParseAddr(resolved)
		if err != nil {
			return nil, err
		}
		return []netip.Addr{address}, nil
	})
}

// ResolveEndpointsWith resolves peer endpoint hostnames through an explicit
// resolver. IP literals bypass the callback and endpoint ports are preserved.
func (config *Config) ResolveEndpointsWith(resolve func(string) ([]netip.Addr, error)) error {
	if resolve == nil {
		return fmt.Errorf("endpoint resolver is unavailable")
	}
	for i := range config.Peers {
		if config.Peers[i].Endpoint.IsEmpty() {
			continue
		}
		if _, err := netip.ParseAddr(config.Peers[i].Endpoint.Host); err == nil {
			continue
		}
		addresses, err := resolve(config.Peers[i].Endpoint.Host)
		if err != nil {
			return fmt.Errorf("resolve peer endpoint %q: %w", config.Peers[i].Endpoint.Host, err)
		}
		var resolved netip.Addr
		for _, address := range addresses {
			if address.IsValid() && !address.IsUnspecified() && !address.IsMulticast() {
				resolved = address
				break
			}
		}
		if !resolved.IsValid() {
			return fmt.Errorf("resolve peer endpoint %q: no usable addresses", config.Peers[i].Endpoint.Host)
		}
		config.Peers[i].Endpoint.Host = resolved.String()
	}
	return nil
}
