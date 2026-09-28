/* SPDX-License-Identifier: MIT
 *
 * Copyright (C) 2026 TunnelMint contributors. All Rights Reserved.
 */

package tunnel

import (
	"errors"
	"testing"

	"golang.zx2c4.com/wireguard/windows/services"
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
