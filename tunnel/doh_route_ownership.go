/* SPDX-License-Identifier: MIT
 *
 * Copyright (C) 2026 TunnelMint contributors. All Rights Reserved.
 */

package tunnel

import "net/netip"

// dohRouteOwnership remembers which host routes TunnelMint created. A route
// that was already present must stay unowned, while reapplying a route that
// TunnelMint created must preserve that ownership across recovery.
type dohRouteOwnership struct {
	routes map[netip.Prefix]bool
}

func newDoHRouteOwnership() *dohRouteOwnership {
	return &dohRouteOwnership{routes: make(map[netip.Prefix]bool)}
}

func (o *dohRouteOwnership) existing(prefix netip.Prefix) bool {
	return o.routes[prefix]
}

func (o *dohRouteOwnership) markExisting(prefix netip.Prefix) bool {
	owned := o.existing(prefix)
	o.routes[prefix] = owned
	return owned
}

func (o *dohRouteOwnership) markCreated(prefix netip.Prefix) {
	o.routes[prefix] = true
}

func (o *dohRouteOwnership) forget(prefix netip.Prefix) {
	delete(o.routes, prefix)
}
