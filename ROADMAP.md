# TunnelMint Roadmap

TunnelMint is intentionally starting small. The goal is to produce a clean Windows client first, prove the encrypted DNS design, and only then expand to Android.

## Phase 0 — Baseline

- [x] Import the tunnel foundation while preserving all required upstream notices and licensing.
- [x] Build the untouched Windows baseline successfully.
- [ ] Confirm tunnel import, connect, disconnect, handshake, and traffic behavior.
- [x] Document a reproducible build procedure.
- [ ] Establish TunnelMint development branding so test builds are clearly identifiable.

**Exit criteria:** A reproducible TunnelMint development build behaves correctly before any new networking behavior is introduced.

## Phase 1 — DNS Configuration

- [x] Preserve normal `DNS = <IP>` behavior.
- [x] Recognize `DNS = https://...` as a DNS-over-HTTPS endpoint.
- [x] Preserve encrypted DNS endpoint paths and identifiers exactly.
- [x] Add parsing tests for plain DNS and encrypted DNS configurations.

**Exit criteria:** TunnelMint can safely distinguish normal DNS from DoH without breaking standard configurations.

## Phase 2 — Encrypted DNS Engine

- [x] Implement DoH request and response handling.
- [x] Validate TLS certificates normally and securely.
- [ ] Implement automatic hostname bootstrap.
- [ ] Ship multiple sensible bootstrap resolvers for reliability.
- [ ] Ensure bootstrap is used only to locate the encrypted DNS endpoint.
- [ ] Cache endpoint information where safe and useful.

**Exit criteria:** A configured DoH endpoint can be reached reliably and DNS queries can be resolved through it.

## Phase 3 — Tunnel Integration

- [ ] Route encrypted DNS traffic through the active tunnel.
- [ ] Integrate the encrypted resolver with Windows DNS behavior.
- [ ] Prevent DNS leaks when encrypted DNS is selected.
- [ ] Do not silently fall back to plain DNS.
- [ ] Restore DNS and networking state cleanly when the tunnel disconnects.
- [ ] Test full-tunnel and split-tunnel configurations.
- [ ] Test IPv4 and IPv6 behavior.

**Exit criteria:** `DNS = https://...` works end-to-end through the tunnel without unintended DNS fallback or leakage.

## Phase 4 — Settings

Add a simple Settings area for bootstrap DNS.

- [ ] Show built-in bootstrap resolvers.
- [ ] Allow resolvers to be enabled or disabled.
- [ ] Allow resolvers to be reordered.
- [ ] Allow custom bootstrap resolvers.
- [ ] Allow restoring defaults.
- [ ] Keep normal users out of advanced configuration unless they choose to open it.

**Exit criteria:** Default behavior requires no setup, while users retain control over bootstrap resolver choices.

## Phase 5 — TunnelMint UI

- [ ] Keep tunnel import, activation, status, and editing simple.
- [ ] Make the interface visually distinct and clearly TunnelMint.
- [ ] Display whether DNS is plain or encrypted.
- [ ] Show basic tunnel, handshake, traffic, and DNS status.
- [ ] Keep the interface uncluttered.

**Exit criteria:** The client is simple enough to install, import a tunnel, and connect without documentation.

## Phase 6 — Windows Release

- [ ] Build a normal Windows installer.
- [ ] Verify install, upgrade, uninstall, reboot, sleep, and wake behavior.
- [ ] Test network changes between Ethernet and Wi-Fi.
- [ ] Test tunnel and encrypted DNS failure scenarios.
- [ ] Verify no private keys, credentials, or test configurations are included in release artifacts.
- [ ] Prepare release notes and known limitations.

**Exit criteria:** TunnelMint can be installed and used as normal end-user Windows software.

## Phase 7 — Android

Android development begins only after the Windows behavior and configuration model are stable.

- [ ] Reuse the same TunnelMint configuration behavior where practical.
- [ ] Support standard tunnel import.
- [ ] Support `DNS = <IP>` and `DNS = https://...` consistently with Windows.
- [ ] Add automatic configurable bootstrap DNS.
- [ ] Add encrypted DNS leak protection.
- [ ] Package as a normal installable Android application.

## Not in Initial Scope

The first releases are not intended to become a general-purpose network security suite.

Unless explicitly added later, initial scope excludes:

- Ad blocking
- DNS filtering
- IDS or IPS features
- Mesh networking
- Commercial VPN accounts or subscriptions
- Tracking or analytics
- Unrelated network-management features

**KISS: make the common path simple, automatic, and reliable.**
