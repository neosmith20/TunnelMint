# Task 005 — Windows Runtime Identity and Coexistence

## Implemented

TunnelMint now uses application-owned runtime identities from `product/identity.go`:

- manager service `TunnelMintManager` with display name `TunnelMint Manager`;
- tunnel services with the `TunnelMintTunnel$` prefix and `TunnelMint Tunnel: ` display prefix;
- protected data, configuration, and log storage under `Program Files\TunnelMint\Data`;
- administrator settings under `HKLM\Software\TunnelMint`;
- UI window class and title owned by TunnelMint, so launching the executable cannot find or raise an upstream WireGuard window;
- `TunnelMint/<version> (...)` user-agent and TunnelMint version-resource/manifest identity.
- main window, tray, about, and update-disabled messages identify the application as TunnelMint.

The installer custom action service and registry identifiers were updated to the same TunnelMint names. Anonymous inherited pipes remain the manager/UI IPC mechanism; there is no global named pipe or mutex to collide with upstream WireGuard. Each manager service instance owns its inherited pipe set.

## Updater behavior

Automatic update checking now reports the disabled state once and never opens an update session. The manual `/update` command and manager update RPC also return a disabled result. The upstream updater package remains isolated and buildable for a future TunnelMint-signed implementation, but its network entry point is guarded by `updatesEnabled = false` and uses no upstream update host at runtime.

WireGuard protocol, WireGuardNT driver, adapter type, and upstream source notices remain unchanged because they identify the tunnel technology or preserve licensing; they are not TunnelMint application state.

## Verification

- Focused tests: `go test -vet=off ./product ./doh ./bootstrap` — passed.
- Windows build: `cmd /c build.bat` — passed for x86, amd64, and arm64.
- Static identity audit found no TunnelMint runtime use of `WireGuardManager`, `WireGuardTunnel$`, `Software\\WireGuard`, or `Program Files\\WireGuard`.

The disposable VM used for this task does not have a normal upstream WireGuard installation available, so live service coexistence, install/uninstall, and UI handoff checks remain unverified. No tunnel credentials or private keys were used.

