# Task017 – Complete TunnelMint → WireHush Identity Rename And Release Polish

## Goal

Perform the coordinated **TunnelMint → WireHush** identity migration now that the UI differentiation and bootstrap-family work has merged. This is a release-polish task, not a feature-expansion task.

Base this work on current `main` after PR #11 (`Differentiate UI and scope bootstrap by family`).

## Branch

Use:

`codex/wirehush-rename-release-polish`

## Approved Brand Assets

On the Windows build/test VM, the approved WireHush brand kit is available at:

`C:\TunnelMint-Test\WireHush_Brand_Kit\`

Use only the production assets needed by the application. Do **not** commit the entire ZIP/reference folder.

## Required Rename Scope

Change current product identity from TunnelMint to WireHush across the active product surface, including where applicable:

- Application title and visible UI branding
- App icon/logo/tray icon/About branding
- `tunnelmint.exe` → `wirehush.exe`
- Installer product name and user-facing installer strings
- Start Menu shortcuts
- Installed application naming
- Service names where safe and deliberate
- Program Files / application data paths where safe and deliberate
- IPC names / window classes / updater identifiers / adapter namespace if they are product-brand coupled and can be migrated safely
- Build artifact names
- GitHub Actions artifact/package names
- Tests and test fixtures that refer to the active product identity
- Scripts that refer to active product paths/names
- Current README/support/security/contribution-facing product wording
- Current screenshots/assets/branding references
- User-facing errors, dialogs, Settings text, tray labels, and About text

Historical acceptance/task reports may retain `TunnelMint` where necessary to accurately describe historical evidence. Do not rewrite history merely to eliminate the string.

## Persistence / Upgrade Safety

Do **not** blindly rename persistent identifiers.

Before changing any of the following, determine whether preserving the existing value is required for clean upgrade/coexistence behavior:

- MSI UpgradeCode / ProductCode strategy
- Persistent GUIDs
- Adapter GUID namespace
- Service identity
- Data/config paths
- Registry locations
- IPC identifiers
- Other persisted installation/application identifiers

Where a rename would break upgrade, uninstall, coexistence, or existing configuration discovery, implement an explicit migration/compatibility path or preserve the identifier and document why.

Official WireGuard must remain untouched and able to coexist with WireHush.

## Branding / UI

Use the approved WireHush logo and icon from the brand kit.

Do not use the WireGuard dragon/logo as WireHush branding.

Keep the UI native Windows, restrained, simple, professional, and visibly distinct from the official WireGuard Windows client.

## Wording / Capitalization Sweep

Perform the final owner-requested wording polish at the same time.

User-facing labels should use professional Title Case / first-letter capitalization consistently where appropriate. Review buttons, menus, tooltips, dialogs, Settings, errors, About text, installer text, Start Menu text, and tray text.

Examples of desired wording style:

- `Import Tunnel(s) From File…`
- `Add Empty Tunnel…`
- `Export All Tunnels To Zip…`
- `Restore Defaults`
- `Move Up`
- `Move Down`
- `Enable`
- `Disable`
- `Latest Handshake`
- `Persistent Keepalive`
- `DNS Servers`
- `Add Bootstrap Resolver`

Do not mechanically Title Case technical values, URLs, filenames, DNS names, protocol names, logs, or data where capitalization would be incorrect.

## Legal / Attribution Guardrails

This task changes product identity; it does **not** erase upstream attribution.

Preserve all required third-party notices and source headers, including WireGuard copyright/license notices and GPL-marked imported material.

Update TunnelMint-specific project-owned notices to WireHush where appropriate, but do not alter third-party license text or attribution merely for branding consistency.

Use the WireGuard name descriptively only where appropriate and do not imply affiliation or endorsement.

Do not make claims such as “official WireGuard client.”

Do not claim legal clearance or non-infringement.

## Functional Guardrails

This task must **not** change established network behavior except where a rename/migration requires it.

Specifically preserve:

- Plain DNS behavior
- Encrypted DNS / DoH behavior
- Fail-closed DNS behavior
- No system/plaintext DNS fallback after encrypted DNS activation
- Bootstrap resolver family matching
- TLS certificate validation
- Redirect rejection behavior
- Tunnel lifecycle and cleanup
- WireGuard coexistence

Do not add Android work or unrelated features.

## Validation

Build/test on the Windows build/test VM first.

At minimum verify:

1. Clean install presents WireHush identity.
2. Executable/package/artifact names use WireHush where intended.
3. WireHush Manager/service and official WireGuard can coexist without collision.
4. Existing tunnel import/connect/disconnect still works.
5. Plain DNS tunnel still works.
6. DoH tunnel still works and remains fail-closed.
7. Bootstrap family behavior is unchanged from the merged implementation.
8. Uninstall removes WireHush-owned components only.
9. Reinstall/upgrade behavior is sane and any retained legacy identifier/path is intentional and documented.
10. No active current-product UI/docs/build artifacts accidentally remain branded TunnelMint, except intentionally preserved historical material or compatibility identifiers.
11. Required third-party notices remain present in packaged output.

Prefer automated validation wherever practical. Do not create another giant manual owner test matrix unless a concrete blocker is found.

## GitHub Workflow

1. Start from current `main`.
2. Implement and test on the build/test VM.
3. Commit the complete rename/polish work to `codex/wirehush-rename-release-polish`.
4. Push the branch.
5. Open a PR against `main`.
6. Let GitHub Windows validation reproduce the build/tests.
7. Report exact commit SHA, PR number, validation results, deliberate retained legacy identifiers, and any remaining release blockers.

## Out Of Scope

- Android
- New VPN protocols
- New DNS transports
- New account/cloud systems
- New telemetry
- Feature expansion unrelated to the rename/release polish
- Public release publishing itself

The next step after this task is a short WireHush release-candidate regression and final release/legal packaging review.
