/* SPDX-License-Identifier: MIT
 *
 * Copyright (C) 2019-2026 WireGuard LLC. All Rights Reserved.
 */

package conf

import (
	"sync"

	"golang.org/x/sys/windows/registry"

	"golang.zx2c4.com/wireguard/windows/product"
)

const adminRegKey = product.AdminRegistryKey

var (
	adminKey     registry.Key
	adminKeyOnce sync.Once
	adminKeyErr  error
)

func openAdminKey() (registry.Key, error) {
	adminKeyOnce.Do(func() {
		adminKey, adminKeyErr = registry.OpenKey(registry.LOCAL_MACHINE, adminRegKey, registry.QUERY_VALUE|registry.WOW64_64KEY)
	})
	return adminKey, adminKeyErr
}

func AdminBool(name string) bool {
	key, err := openAdminKey()
	if err != nil {
		return false
	}
	val, _, err := key.GetIntegerValue(name)
	if err != nil {
		return false
	}
	return val != 0
}
