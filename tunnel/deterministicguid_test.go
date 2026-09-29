/* SPDX-License-Identifier: MIT
 *
 * Copyright (C) 2026 WireHush contributors. All Rights Reserved.
 */

package tunnel

import (
	"testing"

	"golang.zx2c4.com/wireguard/windows/conf"
)

func TestDeterministicGUIDKeepsLegacyNamespace(t *testing.T) {
	c := &conf.Config{Name: "same-name", Interface: conf.Interface{PrivateKey: *conf.NewPrivateKey()}}
	legacy := deterministicGUIDWithLabel(c, deterministicGUIDLabel)
	upstream := deterministicGUIDWithLabel(c, upstreamDeterministicGUIDLabel)
	if *legacy == *upstream {
		t.Fatal("legacy deterministic GUID collides with upstream namespace")
	}
	if *legacy != *deterministicGUID(c) {
		t.Fatal("deterministicGUID did not preserve the legacy adapter namespace")
	}
}
