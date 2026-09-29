/* SPDX-License-Identifier: MIT
 *
 * Copyright (C) 2026 WireHush contributors. All Rights Reserved.
 */

package tunnel

import (
	"net/netip"
	"testing"
)

func TestDoHRouteOwnershipPreservesCreatedRouteOnReapply(t *testing.T) {
	prefix := netip.MustParsePrefix("192.0.2.7/32")
	ownership := newDoHRouteOwnership()
	ownership.markCreated(prefix)
	if !ownership.markExisting(prefix) {
		t.Fatal("reapplying a WireHush-created route lost ownership")
	}
}

func TestDoHRouteOwnershipKeepsPreexistingRouteUnowned(t *testing.T) {
	prefix := netip.MustParsePrefix("192.0.2.8/32")
	ownership := newDoHRouteOwnership()
	if ownership.markExisting(prefix) {
		t.Fatal("pre-existing route became WireHush-owned")
	}
	ownership.forget(prefix)
}

func TestDoHRouteOwnershipRecreatedRouteIsOwned(t *testing.T) {
	prefix := netip.MustParsePrefix("192.0.2.9/32")
	ownership := newDoHRouteOwnership()
	ownership.markCreated(prefix)
	ownership.forget(prefix)
	ownership.markCreated(prefix)
	if !ownership.existing(prefix) {
		t.Fatal("recreated WireHush route was not recorded as owned")
	}
}
