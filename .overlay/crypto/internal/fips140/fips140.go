/* SPDX-License-Identifier: MIT
 *
 * Copyright (C) 2019-2026 WireGuard LLC. All Rights Reserved.
 */

package fips140

const (
	Enabled = false
	debug   = false
)

func Supported() error {
	panic("")
}

func Name() string {
	panic("")
}

func Version() string {
	panic("")
}

// CAST matches the disabled-FIPS behavior of Go 1.27.1: self-tests are not
// run when FIPS is disabled.
func CAST(string, func() error) {}

// PCT matches the disabled-FIPS behavior of Go 1.27.1: self-tests are not
// run when FIPS is disabled.
func PCT(string, func() error) {}

func RecordApproved() {}

func RecordNonApproved() {}

func ResetServiceIndicator() {
	panic("")
}

func ServiceIndicator() bool {
	panic("")
}
