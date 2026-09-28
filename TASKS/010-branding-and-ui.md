# Task 010 — TunnelMint Branding and User-Facing UI

## Goal

Turn the development client into a clearly identifiable TunnelMint application while preserving the simple tunnel workflow and avoiding a needless redesign.

## Product identity

Update TunnelMint-owned user-facing identity as appropriate:

- application/window titles;
- executable/product naming where safe;
- About dialog;
- tray/menu strings;
- installer-facing product strings that are not handled later in Task 011;
- logs/diagnostics headings;
- temporary development icon/assets.

Do not remove required upstream copyright/license notices from source or distributed notices. WireGuard may be named where necessary as the underlying tunnel technology and for attribution, but TunnelMint must not present itself as the upstream WireGuard application or reuse upstream branding as TunnelMint's own identity.

Use a simple original TunnelMint development icon/visual treatment that can be replaced later. Do not copy another project's logo or artwork.

## Main UI behavior

Keep the main workflow simple:

1. Import/add tunnel.
2. Select tunnel.
3. Activate/deactivate.
4. See connection status.
5. Edit configuration.
6. Open Settings when desired.

Do not add dashboard clutter.

## DNS status

For the selected/active tunnel, expose concise DNS information such as:

- `DNS: Plain` for ordinary IP DNS;
- `DNS: Encrypted (DoH)` for HTTPS DNS;
- resolver hostname/endpoint identity where useful;
- readiness/error state when encrypted DNS is initializing or failed.

Do not display secrets, private keys, full query contents, or unnecessary diagnostic noise in the normal UI.

## Diagnostics

Provide enough operator-facing status to distinguish failures in:

- tunnel activation;
- bootstrap resolution;
- DoH TLS/HTTP transport;
- local DNS proxy;
- route/leak protection.

Keep detailed logs available without dumping them into the main screen.

## Compatibility

- Existing tunnel import/edit behavior must remain usable.
- `DNS = <IP>` and `DNS = https://...` must serialize correctly.
- TunnelMint and the installed upstream client must remain visually and operationally distinct.

## Verification and report

Run UI-relevant tests, parser/config tests, and `cmd /c build.bat`.

Launch the development build on the VM and verify the visible application is clearly TunnelMint and does not hand off to the upstream app.

Create `docs/task-010-report.md` covering product strings/assets changed, UI behavior, DNS status states, launch verification, tests, build result, and remaining unverified items.

Commit, push, and continue to Task 011.