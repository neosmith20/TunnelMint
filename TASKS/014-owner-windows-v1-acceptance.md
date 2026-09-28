# Task 014 — Owner Windows v1 Acceptance Checklist

## Purpose

This is the real-machine acceptance checklist to run after Task 013 hardening is merged/reviewed. These checks require an elevated disposable Windows VM, real tunnel configuration/peer access, and owner observation. Coding agents must not mark a check passed based only on unit tests, mocks, static review, or build success.

Record exact build/commit tested and evidence/results in `docs/task-014-owner-acceptance.md` when performed.

## Install / product isolation

- [ ] Install the amd64 MSI elevated on a clean Windows 11 VM.
- [ ] Confirm TunnelMint launches normally after install.
- [ ] Confirm `TunnelMintManager` installs/runs and no upstream manager/service is reused.
- [ ] Confirm TunnelMint data/config paths are under TunnelMint-owned locations.
- [ ] Install or retain official WireGuard side-by-side.
- [ ] Confirm launching TunnelMint never raises/attaches to the upstream WireGuard UI.
- [ ] Import the same tunnel configuration into both clients and confirm their services/adapters do not collide.
- [ ] Confirm distinct adapter GUID/device identity for the same tunnel input.
- [ ] Uninstall TunnelMint and confirm upstream WireGuard remains intact.
- [ ] Reinstall TunnelMint and confirm clean behavior.
- [ ] Exercise upgrade from one TunnelMint development build to a newer build when two package versions are available.

## Plain DNS tunnel

- [ ] Import a normal tunnel using `DNS = <IP>`.
- [ ] Connect and confirm a real WireGuard handshake.
- [ ] Confirm normal tunneled traffic works.
- [ ] Confirm configured plain DNS works exactly as expected.
- [ ] Disconnect and confirm routes/DNS return to the pre-tunnel state.

## DoH — full tunnel

- [ ] Use a tunnel with `AllowedIPs = 0.0.0.0/0` and IPv6 equivalent where supported.
- [ ] Configure `DNS = https://...` with a real trusted DoH endpoint/path.
- [ ] Connect and confirm real handshake/traffic.
- [ ] Confirm bootstrap occurs only as needed to locate the DoH hostname.
- [ ] Confirm subsequent DNS queries use the DoH endpoint over HTTPS.
- [ ] Confirm DoH traffic traverses the active WireGuard tunnel.
- [ ] Capture packets and verify no unintended plaintext DNS escapes after bootstrap.
- [ ] Confirm endpoint path/client identifier is preserved exactly.

## DoH — split tunnel

- [ ] Use a split tunnel whose original `AllowedIPs` do not already contain the DoH endpoint IP.
- [ ] Connect successfully after Task 013 runtime AllowedIPs fix.
- [ ] Confirm DoH endpoint traffic is cryptokey-routed through the intended WireGuard peer.
- [ ] Confirm unrelated Internet traffic preserves the split-tunnel behavior.
- [ ] Confirm runtime-added endpoint host prefixes disappear on disconnect.
- [ ] Re-resolve/refresh endpoint address where practical and confirm the old runtime host prefix is removed only after the replacement is usable.

## Bootstrap Settings

- [ ] Open Settings as a normal non-admin desktop user.
- [ ] Enable/disable built-in bootstrap resolvers and confirm settings persist.
- [ ] Reorder resolvers and confirm order persists.
- [ ] Add a custom resolver and confirm it persists.
- [ ] Disable every built-in resolver, leave only the custom resolver enabled, and confirm encrypted-DNS tunnel activation succeeds.
- [ ] Confirm disabled built-ins are not used during packet capture.
- [ ] Restore defaults and confirm expected six built-ins/order return.
- [ ] Confirm invalid/multicast/unspecified/duplicate entries are rejected.

## Failure-closed / leak testing

- [ ] Point DoH at a hostname that cannot bootstrap and confirm activation fails without normal DNS fallback.
- [ ] Use an endpoint with an invalid/untrusted TLS certificate and confirm activation/query fails closed.
- [ ] Use a valid hostname with an invalid DoH path/status response and confirm failure is surfaced.
- [ ] Make the DoH endpoint unreachable after activation and confirm DNS does not silently fall back to plaintext.
- [ ] During activation/finalization, capture DNS traffic and look specifically for any plaintext-DNS leak during WFP policy transition.
- [ ] Disconnect while DNS queries are active and confirm cleanup/restoration.
- [ ] Force-stop the tunnel/service where safe and confirm dynamic firewall state/routes do not remain stuck.

## Network/lifecycle transitions

- [ ] Reboot Windows with TunnelMint installed and confirm normal startup behavior.
- [ ] Reboot with a tunnel configured/active according to supported behavior and inspect recovery.
- [ ] Sleep and resume with a plain-DNS tunnel.
- [ ] Sleep and resume with a DoH tunnel.
- [ ] Switch Ethernet -> Wi-Fi while connected.
- [ ] Switch Wi-Fi -> Ethernet while connected.
- [ ] Disconnect/reconnect the physical network while connected.
- [ ] Confirm DNS, endpoint routes, and firewall state remain correct after transitions.

## IPv6

- [ ] Test a tunnel with IPv6 tunnel addressing/AllowedIPs.
- [ ] Test a DoH endpoint resolving to IPv6.
- [ ] Verify endpoint routing/cryptokey routing over IPv6.
- [ ] Verify no IPv6 DNS leak.
- [ ] Verify cleanup/restoration after disconnect.

## Installer/package review

- [ ] Confirm installed package includes TunnelMint license and required third-party notices.
- [ ] Confirm Start Menu/product naming says TunnelMint, not upstream product branding.
- [ ] Confirm uninstall removes only TunnelMint-owned files/services/adapters/settings intended for removal.
- [ ] Confirm uninstall does not remove or damage official WireGuard.
- [ ] Confirm unsigned development build warnings are expected until signing infrastructure exists.

## Acceptance outcome

Windows v1 may be considered beta-ready only after all release-blocking failures discovered here are fixed and retested. Any failed item should become a focused bug/task with reproduction evidence before public release.

Android work remains deferred until Windows behavior is owner-accepted and stable.
