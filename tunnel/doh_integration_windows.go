//go:build windows

/* SPDX-License-Identifier: MIT
 *
 * Copyright (C) 2026 TunnelMint contributors. All Rights Reserved.
 */

package tunnel

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/netip"
	"time"

	"golang.org/x/sys/windows"

	"golang.zx2c4.com/wireguard/windows/bootstrap"
	"golang.zx2c4.com/wireguard/windows/conf"
	"golang.zx2c4.com/wireguard/windows/dnsproxy"
	"golang.zx2c4.com/wireguard/windows/doh"
	"golang.zx2c4.com/wireguard/windows/dohruntime"
	"golang.zx2c4.com/wireguard/windows/tunnel/firewall"
	"golang.zx2c4.com/wireguard/windows/tunnel/winipcfg"
)

const dohRuntimeTimeout = 10 * time.Second

func encryptedDNSConfigured(config *conf.Config) bool {
	return config != nil && len(config.Interface.DNSOverHTTPS) > 0
}

func activateEncryptedDNS(ctx context.Context, config *conf.Config, luid winipcfg.LUID) (*dohruntime.Session, error) {
	if !encryptedDNSConfigured(config) {
		return nil, nil
	}
	if len(config.Interface.DNSOverHTTPS) != 1 {
		return nil, errors.New("exactly one encrypted DNS endpoint is supported per tunnel")
	}
	endpoint := config.Interface.DNSOverHTTPS[0]
	resolver := bootstrap.NewResolver(nil)
	resolver.Timeout = dohRuntimeTimeout
	peerEndpoints := make([]netip.Addr, 0, len(config.Peers))
	for _, peer := range config.Peers {
		if address, err := netip.ParseAddr(peer.Endpoint.Host); err == nil {
			peerEndpoints = append(peerEndpoints, address)
		}
	}
	var client *doh.Client
	hooks := dohruntime.Hooks{
		Bootstrap: func(ctx context.Context, endpoint string) ([]netip.Addr, error) {
			return resolver.ResolveEndpoint(ctx, endpoint)
		},
		AddRoute: func(prefix netip.Prefix) (bool, error) {
			nextHop := netip.IPv4Unspecified()
			if prefix.Addr().Is6() {
				nextHop = netip.IPv6Unspecified()
			}
			if _, err := luid.Route(prefix, nextHop); err == nil {
				return false, nil
			}
			if err := luid.AddRoute(prefix, nextHop, 0); err != nil {
				if err == windows.ERROR_OBJECT_ALREADY_EXISTS {
					return false, nil
				}
				return false, err
			}
			return true, nil
		},
		DelRoute: func(prefix netip.Prefix) error {
			nextHop := netip.IPv4Unspecified()
			if prefix.Addr().Is6() {
				nextHop = netip.IPv6Unspecified()
			}
			err := luid.DeleteRoute(prefix, nextHop)
			if err == windows.ERROR_NOT_FOUND {
				return nil
			}
			return err
		},
		Verify: func(ctx context.Context, endpoint string, addresses []netip.Addr) error {
			httpClient := doh.NewHTTPClientForAddresses(addresses, http.DefaultClient)
			var err error
			client, err = doh.NewClient(endpoint, doh.Options{HTTPClient: httpClient, Timeout: dohRuntimeTimeout, MaxResponseSize: doh.DefaultMaxResponseSize})
			if err != nil {
				return err
			}
			_, err = client.Query(ctx, dohProbeQuery())
			return err
		},
		StartProxy: func(_ context.Context, _ string, _ []netip.Addr) (io.Closer, error) {
			if client == nil {
				return nil, errors.New("DoH transport was not verified")
			}
			proxy, err := dnsproxy.New(dnsproxy.Options{Address: dnsproxy.DefaultAddress, Client: client, QueryTimeout: dohRuntimeTimeout})
			if err != nil {
				return nil, err
			}
			if err := proxy.Start(); err != nil {
				return nil, err
			}
			return proxy, nil
		},
		SetDNS: func() (func() error, error) {
			previous, err := luid.DNS()
			if err != nil {
				return nil, err
			}
			if err := luid.SetDNS(windows.AF_INET, []netip.Addr{netip.MustParseAddr("127.0.0.1")}, config.Interface.DNSSearch); err != nil {
				return nil, err
			}
			return func() error {
				var v4 []netip.Addr
				for _, address := range previous {
					if address.Is4() {
						v4 = append(v4, address)
					}
				}
				return luid.SetDNS(windows.AF_INET, v4, config.Interface.DNSSearch)
			}, nil
		},
		Finalize: func() error {
			exceptions := []netip.Addr{netip.MustParseAddr("127.0.0.1"), netip.MustParseAddr("::1")}
			return firewall.ReconfigureDNS(uint64(luid), shouldNotRestrictFirewall(config), exceptions)
		},
	}
	return dohruntime.Activate(ctx, dohruntime.Config{Endpoint: endpoint, PeerEndpointAddress: peerEndpoints}, hooks)
}

func dohProbeQuery() []byte {
	return []byte{0, 0, 1, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 1}
}
