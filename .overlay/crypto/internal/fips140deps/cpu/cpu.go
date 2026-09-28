/* SPDX-License-Identifier: MIT
 *
 * Copyright (C) 2019-2026 WireGuard LLC. All Rights Reserved.
 */

// This compatibility overlay keeps the upstream WireGuard SHA-512 dispatch
// helpers while matching the Go 1.27 fips140deps/cpu API.
package cpu

import (
	"internal/cpu"
	"internal/goarch"
)

const (
	BigEndian = goarch.BigEndian
	AMD64     = goarch.IsAmd64 == 1
	ARM64     = goarch.IsArm64 == 1
	PPC64     = goarch.IsPpc64 == 1
	PPC64le   = goarch.IsPpc64le == 1
)

var (
	ARM64HasAES     = cpu.ARM64.HasAES
	ARM64HasPMULL   = cpu.ARM64.HasPMULL
	ARM64HasSHA2    = cpu.ARM64.HasSHA2
	ARM64HasSHA512  = cpu.ARM64.HasSHA512
	ARM64HasSHA3    = cpu.ARM64.HasSHA3
	X86HasADX       = cpu.X86.HasADX
	X86HasAES       = cpu.X86.HasAES
	X86HasAVX       = cpu.X86.HasAVX
	X86HasAVX2      = cpu.X86.HasAVX2
	X86HasBMI2      = cpu.X86.HasBMI2
	X86HasPCLMULQDQ = cpu.X86.HasPCLMULQDQ
	X86HasSHA       = cpu.X86.HasSHA
	X86HasSSE41     = cpu.X86.HasSSE41
	X86HasSSSE3     = cpu.X86.HasSSSE3
)

func HasSHA512AVX2() bool {
	return cpu.X86.HasAVX && cpu.X86.HasAVX2 && cpu.X86.HasBMI2
}

func HasSHA512ARM64() bool {
	return cpu.ARM64.HasSHA512
}
