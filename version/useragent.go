/* SPDX-License-Identifier: MIT
 *
 * Copyright (C) 2019-2026 WireGuard LLC. All Rights Reserved.
 */

package version

import (
	"fmt"
	"runtime"

	"golang.zx2c4.com/wireguard/windows/product"
)

func Arch() string {
	switch runtime.GOARCH {
	case "arm", "arm64", "amd64":
		return runtime.GOARCH
	case "386":
		return "x86"
	default:
		panic("Unrecognized GOARCH")
	}
}

func UserAgent() string {
	return fmt.Sprintf("%s/%s (%s; %s)", product.UserAgentName, Number, OsName(), Arch())
}
