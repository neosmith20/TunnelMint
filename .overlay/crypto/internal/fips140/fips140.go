/* SPDX-License-Identifier: MIT
 *
 * Copyright (C) 2019-2026 WireGuard LLC. All Rights Reserved.
 */

package fips140

import "errors"

const (
	Enabled = false
	debug   = false
)

func Supported() error {
	return errors.New("FIPS 140-3 mode is unavailable in this build")
}

func Name() string {
	return "Go Cryptographic Module"
}

func Version() string {
	// crypto/tls checks this value even when FIPS mode is disabled.
	return "latest"
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
}

func ServiceIndicator() bool {
	return false
}
