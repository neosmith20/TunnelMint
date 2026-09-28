/* SPDX-License-Identifier: MIT
 *
 * Copyright (C) 2026 TunnelMint contributors. All Rights Reserved.
 */

// Package product contains the names owned by TunnelMint's Windows runtime.
// WireGuard protocol and driver identifiers stay in their upstream packages;
// these values identify the application and its persistent resources.
package product

const (
	Name = "TunnelMint"

	ManagerServiceName         = "TunnelMintManager"
	ManagerServiceDisplayName  = "TunnelMint Manager"
	TunnelServicePrefix        = "TunnelMintTunnel$"
	TunnelServiceDisplayPrefix = "TunnelMint Tunnel: "

	DataDirectoryName = "TunnelMint"
	AdminRegistryKey  = `Software\TunnelMint`

	ManagerWindowClass = "TunnelMint UI - Manage Tunnels"
	ManagerWindowTitle = "TunnelMint"

	UserAgentName = "TunnelMint"
)
