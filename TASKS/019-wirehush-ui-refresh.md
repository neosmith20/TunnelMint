# Task019 – Implement WireHush UI Mockup And Dark Mode

## Goal

Implement the approved WireHush UI redesign shown in the owner-provided mockup and make dark mode a first-class supported presentation. This task intentionally interrupts the current release-candidate pass because the existing UI is not visually acceptable as the final WireHush release.

## Branch

Use:

`codex/wirehush-ui-refresh`

Base this work on current `main` after the merged WireHush rename.

## Owner Mockup

The approved visual reference is on the Windows build/test VM at:

`C:\TunnelMint-Test\UI-Mockup\UI-Mockup.png`

Treat that image as the primary visual target for composition, spacing, hierarchy, navigation, card layout, dark presentation, accent usage, and overall product feel.

This is a mockup, not a requirement to copy every literal placeholder value or invent new backend features that do not exist.

## Required Visual Direction

WireHush should look like its own product, not like the official WireGuard Windows client with different branding.

Implement the overall structure shown in the mockup:

- Compact branded header with WireHush logo/name and subtitle.
- Left `Connections` rail with tunnel list and clear connected/disconnected states.
- Bottom-left connection actions such as Add / Import / Delete.
- Main selected-tunnel header with tunnel name, state, connection duration where available, and Connect/Disconnect action.
- Clear summary information near the top for important connection information.
- Distinct detail sections/tabs for relevant areas such as Overview, Network, DNS, Peer, and Allowed IPs where supported by existing data.
- Card/section based hierarchy instead of the old WireGuard-style giant Interface/Peer boxes.
- Dedicated encrypted-DNS presentation so WireHush's DoH behavior is obvious when active.
- Clean empty-state layout when no tunnel is selected.
- Approved WireHush icon/logo/assets.

Do not add fake data or unsupported backend capabilities merely to mimic the picture.

## Dark Mode

Dark mode is mandatory for the release UI.

Implement a polished dark presentation matching the mockup's direction:

- Dark charcoal/navy surfaces rather than pure black.
- WireHush cyan/blue branding used as restrained accents.
- Green only for healthy/connected/success state.
- High-contrast readable text and controls.
- Native-looking Windows behavior, not a web dashboard pasted into a desktop window.
- No neon/cyberpunk effects, giant gradients, excessive glow, or gamer-RGB styling.

If practical with the current UI framework, support system light/dark appearance or a simple appearance setting. At minimum, the final Windows app must have a real usable dark mode and must not regress existing functionality.

## Functional Guardrails

This is a UI implementation task, not a networking redesign.

Do not change established networking behavior unless required to fix a concrete regression discovered while wiring the UI.

Preserve:

- Tunnel import/edit/connect/disconnect behavior.
- Plain DNS behavior.
- Encrypted DNS / DoH behavior.
- Fail-closed behavior.
- Bootstrap resolver family matching.
- Settings behavior.
- WireGuard coexistence.
- Existing upgrade/migration compatibility identifiers documented in `docs/wirehush-identity-migration.md`.

Do not add Android work, new protocols, telemetry, accounts, cloud features, or unrelated features.

## Mockup Interpretation Notes

The mockup is the visual target, but do not blindly implement misleading labels.

For example:

- Only show connection duration if the application already has or can reliably derive it.
- Only show traffic graphs if existing runtime statistics can support them cleanly without broad new infrastructure. If a graph is too invasive for this pass, preserve the card layout and present existing transfer statistics cleanly instead.
- Do not expose full private keys or sensitive tunnel data.
- Do not show a plaintext fallback resolver as active fallback for encrypted DNS; WireHush's established behavior is fail-closed after encrypted DNS activation. Any DNS card wording must accurately describe the actual implementation.
- Technical values such as URLs, endpoint names, IPs, key material, and protocol names should retain technically correct capitalization.

## Wording / UI Polish

Keep the owner-requested professional capitalization and wording pass intact.

Use clear labels such as:

- Connections
- Add Tunnel
- Import Tunnel(s)
- Delete
- Connection
- Network
- DNS
- Peer
- Allowed IPs
- Latest Handshake
- Persistent Keepalive
- Connect
- Disconnect
- Settings
- Log

Avoid placeholder-looking empty space and awkward detached branding.

## Acceptance

At minimum verify on the Windows build/test VM:

1. Main WireHush window materially matches the approved mockup's composition and hierarchy.
2. Dark mode is usable throughout the main workflow, including tunnel list, selected tunnel details, Settings, dialogs, and menus owned by WireHush where practical.
3. WireHush branding/logo/icon is crisp and correctly placed.
4. Empty state looks intentional and professional.
5. Selected disconnected tunnel state looks correct.
6. Selected connected tunnel state looks correct.
7. Connect/Disconnect still works.
8. Add/import/edit/delete tunnel actions still work.
9. Settings remain accessible and bootstrap resolver controls still function.
10. Plain DNS and DoH behavior are not regressed by UI work.
11. No secret/private key material is newly exposed by the redesigned details view.
12. UI no longer strongly resembles the stock WireGuard Windows client at a glance.

Use automated tests where useful, but visual owner review is required before calling this task complete.

## Workflow

1. Work from `codex/wirehush-ui-refresh`.
2. Use `C:\TunnelMint-Test\UI-Mockup\UI-Mockup.png` as the visual reference.
3. Build/test locally on the Windows build/test VM first.
4. Present the implemented UI to the owner for visual approval before final merge.
5. Fix concrete visual/function issues from that review.
6. Commit and push the branch.
7. Open a PR against `main`.
8. Run GitHub Windows validation.
9. Report the final commit SHA, PR number, validation results, and any deliberate differences from the mockup.

## Relationship To Task018

Task018 release-candidate validation is paused until this UI task is merged. After Task019 merges, restart/rebase the RC pass on the new `main` so the release candidate validates the actual final UI rather than the discarded interim UI.

## Out Of Scope

- Android
- New VPN protocols
- New DNS transports
- New account/cloud systems
- Telemetry
- Broad networking refactors
- Public release publishing
- GitHub Pages design
