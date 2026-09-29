# Task 018 — WireHush Release-Candidate Evidence

This document records the final short release-candidate gate. It does not replace the detailed Task 014 owner-acceptance history.

## Automated Checks

- [x] Current-main amd64 `wirehush.exe` built with the production FIPS overlay.
- [x] Built `wirehush.exe /update` exited normally and reported only that WireHush updates are disabled.
- [x] Focused bootstrap, DNS, DoH, product, tunnel, firewall, manager, UI, and configuration-parser tests passed.
- [x] Current-main amd64 MSI linked as `wirehush-amd64-0.1.0.msi`.
- [x] Package inspection found `LICENSE`, `WIREGUARD-COPYING`, `GPL-2.0.txt`, and `THIRD_PARTY_NOTICES.md`.
- [x] Active-product branding audit found `TunnelMint` only in documented compatibility identifiers, the transitional repository URL, and Task 014 test-result locations.
- [x] Project-owned required notice uses WireHush; WireGuard and GPL attribution remain packaged.

Local WiX ICE validation could not access the Windows Installer service from the restricted Codex account. The MSI was linked with validation suppressed for local inspection; GitHub Windows validation remains the reproducible build/test gate.

## Elevated Release-Candidate Checks

Pending owner execution from Administrator PowerShell using `scripts/task018-release-candidate.ps1`:

- [x] Upgrade passed on the owner VM: one TunnelMint Development installation upgraded to WireHush Development (MSI exit 0); `wirehush.exe /update` exited 0; WireHush Manager was Running; a WireHush UI window and Start Menu shortcut were present.
- [x] Official WireGuard remained unchanged by upgrade: `wireguard.exe` remained present at version 1.1.1 with the same length; no WireGuard Manager service was added or changed.
- [x] The existing owner tunnel was activated from WireHush without recording its configuration. The elevated live-state check found one running tunnel service and one up adapter with four tunnel routes; its traffic probe connected and recorded tunnel-adapter traffic in both directions.
- [x] Plain-DNS disconnect restoration passed: after user-initiated disconnect, the tunnel service and adapter were absent, tunnel routes were gone, and ordinary TCP connectivity and DNS resolution succeeded.
- [x] Known-good DoH acceptance passed on the owner VM. Owner observed a fresh resolver request for the configured client identifier. The elevated capture found the WireHush adapter DNS loopback-only, a successful baseline DNS response over the configured endpoint, and zero observed UDP/TCP port 53 events outside loopback. After the owner revoked the endpoint/client identifier, a fresh DNS response was not observed and port 53 remained unused; no system-resolver fallback was observed. A subsequent disconnect removed the tunnel service, adapter, and all six tunnel routes, while ordinary TCP and DNS recovered.
- [x] Uninstall/reinstall lifecycle passed: WireHush uninstall and reinstall each exited 0, and the reinstalled product started with its executable, manager service, UI window, and Start Menu shortcut present. Official WireGuard remained present at version 1.1.1 with the same executable length and no WireGuard Manager service.

## Signing State

The newly built `wirehush.exe` and MSI are NotSigned. No signing certificate is configured in this repository build environment; an unsigned public build and SmartScreen behavior remain an owner release decision unless a signing certificate is supplied. The pre-existing official `wireguard.exe` signature remains Valid and identifies WireGuard LLC.

## Repository-Rename Handoff

The public GitHub repository remains `neosmith20/TunnelMint` until the owner performs the repository rename. The source uses that URL only as a transitional repository URL. After the owner renames it to `WireHush`, local remotes and public links require an explicit follow-up verification.
