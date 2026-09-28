/* SPDX-License-Identifier: MIT
 *
 * Copyright (C) 2026 TunnelMint contributors. All Rights Reserved.
 */

package product

import "testing"

func TestRuntimeIdentityIsTunnelMintOwned(t *testing.T) {
	if ManagerServiceName == "WireGuardManager" {
		t.Fatal("manager service name collides with upstream WireGuard")
	}
	if TunnelServicePrefix == "WireGuardTunnel$" {
		t.Fatal("tunnel service prefix collides with upstream WireGuard")
	}
	if DataDirectoryName == "WireGuard" || AdminRegistryKey == `Software\WireGuard` {
		t.Fatal("persistent storage identity collides with upstream WireGuard")
	}
	if ManagerWindowClass == "WireGuard UI - Manage Tunnels" {
		t.Fatal("window identity collides with upstream WireGuard")
	}
}
