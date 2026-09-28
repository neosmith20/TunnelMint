//go:build windows

/* SPDX-License-Identifier: MIT
 *
 * Copyright (C) 2026 TunnelMint contributors. All Rights Reserved.
 */

package tunnel

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/netip"
	"path/filepath"
	"time"

	"golang.org/x/sys/windows"

	"golang.zx2c4.com/wireguard/windows/bootstrap"
	"golang.zx2c4.com/wireguard/windows/conf"
	"golang.zx2c4.com/wireguard/windows/dnsproxy"
	"golang.zx2c4.com/wireguard/windows/doh"
	"golang.zx2c4.com/wireguard/windows/dohruntime"
	"golang.zx2c4.com/wireguard/windows/driver"
	"golang.zx2c4.com/wireguard/windows/tunnel/firewall"
	"golang.zx2c4.com/wireguard/windows/tunnel/winipcfg"
)

const dohRuntimeTimeout = 10 * time.Second

func encryptedDNSConfigured(config *conf.Config) bool {
	return config != nil && len(config.Interface.DNSOverHTTPS) > 0
}

func bootstrapResolversForConfig(config *conf.Config, resolvers []netip.Addr) []netip.Addr {
	var allowIPv4, allowIPv6 bool
	for _, family := range configuredInterfaceFamilies(config) {
		switch family {
		case windows.AF_INET:
			allowIPv4 = true
		case windows.AF_INET6:
			allowIPv6 = true
		}
	}
	if !allowIPv4 && !allowIPv6 {
		// Preserve the existing configuration error path when the tunnel has no
		// usable family information yet.
		return append([]netip.Addr(nil), resolvers...)
	}
	return bootstrap.ResolversForFamilies(resolvers, allowIPv4, allowIPv6)
}

func activateEncryptedDNS(ctx context.Context, config *conf.Config, luid winipcfg.LUID, adapter *driver.Adapter, configuredResolvers []netip.Addr) (*dohruntime.Session, error) {
	if !encryptedDNSConfigured(config) {
		return nil, nil
	}
	if len(config.Interface.DNSOverHTTPS) != 1 {
		return nil, errors.New("exactly one encrypted DNS endpoint is supported per tunnel")
	}
	endpoint := config.Interface.DNSOverHTTPS[0]
	if len(configuredResolvers) == 0 {
		return nil, errors.New("at least one bootstrap resolver is required")
	}
	resolver := bootstrap.NewResolver(configuredResolvers)
	resolver.Timeout = dohRuntimeTimeout
	peerEndpoints := make([]netip.Addr, 0, len(config.Peers))
	for _, peer := range config.Peers {
		if address, err := netip.ParseAddr(peer.Endpoint.Host); err == nil {
			peerEndpoints = append(peerEndpoints, address)
		}
	}
	peerRoutes := newDoHPeerRouteManager(adapter, config)
	ownedWindowsRoutes := newDoHRouteOwnership()
	var client *doh.Client
	hooks := dohruntime.Hooks{
		Bootstrap: func(ctx context.Context, endpoint string) ([]netip.Addr, error) {
			return resolver.ResolveEndpoint(ctx, endpoint)
		},
		AddRoute: func(prefix netip.Prefix) (bool, error) {
			peerOwned, err := peerRoutes.Add(prefix)
			if err != nil {
				return false, err
			}
			nextHop := netip.IPv4Unspecified()
			if prefix.Addr().Is6() {
				nextHop = netip.IPv6Unspecified()
			}
			if _, err := luid.Route(prefix, nextHop); err == nil {
				ownedWindowsRoutes.markExisting(prefix)
				return peerOwned, nil
			}
			if err := luid.AddRoute(prefix, nextHop, 0); err != nil {
				_ = peerRoutes.Delete(prefix)
				if err == windows.ERROR_OBJECT_ALREADY_EXISTS {
					ownedWindowsRoutes.markExisting(prefix)
					return peerOwned, nil
				}
				return false, err
			}
			ownedWindowsRoutes.markCreated(prefix)
			return true, nil
		},
		DelRoute: func(prefix netip.Prefix) error {
			nextHop := netip.IPv4Unspecified()
			if prefix.Addr().Is6() {
				nextHop = netip.IPv6Unspecified()
			}
			windowsOwned := ownedWindowsRoutes.existing(prefix)
			if windowsOwned {
				if err := luid.DeleteRoute(prefix, nextHop); err != nil && err != windows.ERROR_NOT_FOUND {
					return err
				}
			}
			if err := peerRoutes.Delete(prefix); err != nil {
				if windowsOwned {
					if restoreErr := luid.AddRoute(prefix, nextHop, 0); restoreErr != nil && restoreErr != windows.ERROR_OBJECT_ALREADY_EXISTS {
						return errors.Join(err, fmt.Errorf("restore DoH host route after peer cleanup failure: %w", restoreErr))
					}
				}
				return err
			}
			ownedWindowsRoutes.forget(prefix)
			return nil
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
		ReplaceProxy: func(_ context.Context, _ string, _ []netip.Addr, existing io.Closer) error {
			if client == nil {
				return errors.New("DoH transport was not verified")
			}
			proxy, ok := existing.(*dnsproxy.Proxy)
			if !ok {
				return errors.New("existing DNS proxy does not support transport replacement")
			}
			return proxy.SetClient(client)
		},
		SetDNS: func() (func() error, error) {
			previous, err := luid.DNS()
			if err != nil {
				return nil, err
			}
			setLoopbackDNS := func() error {
				return luid.SetDNS(windows.AF_INET, []netip.Addr{netip.MustParseAddr("127.0.0.1")}, config.Interface.DNSSearch)
			}
			if err := setLoopbackDNS(); err != nil {
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
		ReapplyDNS: func() error {
			return luid.SetDNS(windows.AF_INET, []netip.Addr{netip.MustParseAddr("127.0.0.1")}, config.Interface.DNSSearch)
		},
		Finalize: func() error {
			exceptions := []netip.Addr{netip.MustParseAddr("127.0.0.1"), netip.MustParseAddr("::1")}
			return firewall.ReconfigureDNS(uint64(luid), shouldNotRestrictFirewall(config), exceptions)
		},
	}
	return dohruntime.Activate(ctx, dohruntime.Config{Endpoint: endpoint, PeerEndpointAddress: peerEndpoints}, hooks)
}

func configuredBootstrapResolvers() ([]netip.Addr, error) {
	root, err := conf.RootDirectory(true)
	if err != nil {
		return nil, errors.New("locate TunnelMint data directory: " + err.Error())
	}
	settings, err := bootstrap.Load(filepath.Join(root, "bootstrap-dns.json"))
	if err != nil {
		return nil, errors.New("load bootstrap DNS settings: " + err.Error())
	}
	resolvers, err := settings.EnabledResolvers()
	if err != nil {
		return nil, errors.New("validate bootstrap DNS settings: " + err.Error())
	}
	return resolvers, nil
}

func dohProbeQuery() []byte {
	return []byte{0, 0, 1, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 1}
}
