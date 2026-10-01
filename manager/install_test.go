/* SPDX-License-Identifier: MIT
 *
 * Copyright (C) 2026 WireHush contributors. All Rights Reserved.
 */

package manager

import (
	"testing"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc/mgr"

	"golang.zx2c4.com/wireguard/windows/product"
)

func TestManagerServiceConfigUsesDemandStart(t *testing.T) {
	config := managerServiceConfig()
	if config.StartType != mgr.StartManual {
		t.Fatalf("manager StartType = %d, want %d", config.StartType, mgr.StartManual)
	}
	if config.StartType == mgr.StartAutomatic {
		t.Fatal("manager StartType must not be automatic")
	}
	if config.DisplayName != product.ManagerServiceDisplayName {
		t.Fatalf("manager DisplayName = %q, want %q", config.DisplayName, product.ManagerServiceDisplayName)
	}
	if config.ServiceType != windows.SERVICE_WIN32_OWN_PROCESS {
		t.Fatalf("manager ServiceType = %d, want %d", config.ServiceType, windows.SERVICE_WIN32_OWN_PROCESS)
	}
	if config.ErrorControl != mgr.ErrorNormal {
		t.Fatalf("manager ErrorControl = %d, want %d", config.ErrorControl, mgr.ErrorNormal)
	}
}
