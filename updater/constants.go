/* SPDX-License-Identifier: MIT
 *
 * Copyright (C) 2019-2026 WireGuard LLC. All Rights Reserved.
 */

package updater

const (
	// updatesEnabled remains false until TunnelMint has its own signed release
	// infrastructure. The updater package stays isolated for that future work.
	updatesEnabled         = false
	releasePublicKeyBase64 = "RWRNqGKtBXftKTKPpBPGDMe8jHLnFQ0EdRy8Wg0apV6vTDFLAODD83G4"
	updateServerHost       = "updates.invalid"
	updateServerPort       = 443
	updateServerUseHttps   = true
	latestVersionPath      = "/releases/latest.sig"
	msiPath                = "/releases/%s"
	msiArchPrefix          = "tunnelmint-%s-"
	msiSuffix              = ".msi"
)
