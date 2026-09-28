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

func CAST(name string, test func() error) {
	if err := test(); err != nil {
		panic(name + ": " + err.Error())
	}
}

func PCT(name string, test func() error) {
	if err := test(); err != nil {
		panic(name + ": " + err.Error())
	}
}

func RecordApproved() {}

func RecordNonApproved() {}

func ResetServiceIndicator() {
	panic("")
}

func ServiceIndicator() bool {
	panic("")
}
