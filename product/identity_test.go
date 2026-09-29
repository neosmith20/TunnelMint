/* SPDX-License-Identifier: MIT
 *
 * Copyright (C) 2026 WireHush contributors. All Rights Reserved.
 */

package product

import "testing"

func TestRuntimeIdentityIsWireHushOwned(t *testing.T) {
	if Name != "WireHush" {
		t.Fatalf("unexpected product name %q", Name)
	}
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

func TestLegacyPersistentIdentityIsExplicit(t *testing.T) {
	if ManagerServiceName != LegacyManagerServiceName || TunnelServicePrefix != LegacyTunnelServicePrefix {
		t.Fatal("service compatibility identifiers changed")
	}
	if DataDirectoryName != LegacyDataDirectoryName || AdminRegistryKey != LegacyAdminRegistryKey {
		t.Fatal("data compatibility identifiers changed")
	}
}
