//go:build windows

/* SPDX-License-Identifier: MIT
 *
 * Copyright (C) 2026 TunnelMint contributors. All Rights Reserved.
 */

package tunnel

import (
	"net/netip"
	"reflect"
	"testing"
)

func TestEncryptedDNSBootstrapExceptionsUseConfiguredResolvers(t *testing.T) {
	tests := []struct {
		name      string
		resolvers []netip.Addr
	}{
		{name: "custom-only", resolvers: []netip.Addr{netip.MustParseAddr("192.0.2.53")}},
		{name: "reordered", resolvers: []netip.Addr{netip.MustParseAddr("9.9.9.9"), netip.MustParseAddr("1.1.1.1")}},
		{name: "disabled-defaults", resolvers: []netip.Addr{netip.MustParseAddr("203.0.113.53")}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := encryptedDNSBootstrapExceptions(test.resolvers)
			want := append(append([]netip.Addr(nil), test.resolvers...), netip.MustParseAddr("127.0.0.1"), netip.MustParseAddr("::1"))
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("exceptions = %v, want %v", got, want)
			}
		})
	}
}
