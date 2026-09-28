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
	"golang.zx2c4.com/wireguard/windows/services"
	"golang.zx2c4.com/wireguard/windows/tunnel/winipcfg"
)

func TestReinitializeAdapterStopsBeforeBringUpWhenConfigurationFails(t *testing.T) {
	wantErr := errors.New("configuration failed")
	bringUp := false
	serviceError, err := reinitializeAdapter(func() error { return wantErr }, func() error {
		bringUp = true
		return nil
	})
	if !errors.Is(err, wantErr) || serviceError != services.ErrorDeviceSetConfig {
		t.Fatalf("result = (%v, %v), want configuration error", serviceError, err)
	}
	if bringUp {
		t.Fatal("adapter bring-up ran after configuration failed")
	}
}

func TestReinitializeAdapterStopsBeforeRecoveryWhenBringUpFails(t *testing.T) {
	wantErr := errors.New("bring-up failed")
	serviceError, err := reinitializeAdapter(func() error { return nil }, func() error { return wantErr })
	if !errors.Is(err, wantErr) || serviceError != services.ErrorDeviceBringUp {
		t.Fatalf("result = (%v, %v), want bring-up error", serviceError, err)
	}
}

func TestWaitForInitialConfigurationRequiresEveryConfiguredFamily(t *testing.T) {
	iw := &interfaceWatcher{
		started: make(chan winipcfg.AddressFamily, 2),
		errors:  make(chan interfaceWatcherError, 1),
	}
	iw.started <- windows.AF_INET6
	iw.started <- windows.AF_INET
	serviceError, err := iw.WaitForInitialConfiguration([]winipcfg.AddressFamily{windows.AF_INET, windows.AF_INET6})
	if err != nil || serviceError != services.ErrorSuccess {
		t.Fatalf("WaitForInitialConfiguration() = (%v, %v)", serviceError, err)
	}
}

func TestConfiguredInterfaceFamilies(t *testing.T) {
	config := &conf.Config{
		Interface: conf.Interface{Addresses: []netip.Prefix{netip.MustParsePrefix("192.0.2.2/24")}},
		Peers:     []conf.Peer{{AllowedIPs: []netip.Prefix{netip.MustParsePrefix("2001:db8::/64")}}},
	}
	if got, want := configuredInterfaceFamilies(config), []winipcfg.AddressFamily{windows.AF_INET, windows.AF_INET6}; !reflect.DeepEqual(got, want) {
		t.Fatalf("configuredInterfaceFamilies() = %v, want %v", got, want)
	}
}
