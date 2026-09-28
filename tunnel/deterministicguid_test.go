/* SPDX-License-Identifier: MIT
 *
 * Copyright (C) 2026 TunnelMint contributors. All Rights Reserved.
 */

package tunnel

import (
	"testing"

	"golang.zx2c4.com/wireguard/windows/conf"
)

func TestDeterministicGUIDUsesTunnelMintNamespace(t *testing.T) {
	c := &conf.Config{Name: "same-name", Interface: conf.Interface{PrivateKey: *conf.NewPrivateKey()}}
	tunnelMint := deterministicGUIDWithLabel(c, deterministicGUIDLabel)
	upstream := deterministicGUIDWithLabel(c, upstreamDeterministicGUIDLabel)
	if *tunnelMint == *upstream {
		t.Fatal("TunnelMint deterministic GUID collides with upstream namespace")
	}
	if *tunnelMint != *deterministicGUID(c) {
		t.Fatal("deterministicGUID did not use the TunnelMint namespace")
	}
}
