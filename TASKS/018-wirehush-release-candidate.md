# Task018 – WireHush Release-Candidate Regression And Final Release Gate

## Goal

Take the freshly merged WireHush identity build and perform the **short final Windows release-candidate regression** plus the last release/legal packaging gate. This is not a feature task.

Base: current `main` at `1f2f7ea0a9c90ef36b64fe5423f3c28f71196df6` (merged PR #13).

Branch:

`codex/wirehush-release-candidate`

## Scope

Validate the actual merged WireHush product as a release candidate. Fix only concrete regressions/blockers discovered during this pass.

## Required RC Checks

1. Build amd64 `wirehush.exe` and MSI on the Windows build/test VM.
2. Verify startup/UI presents WireHush branding and approved icon/logo.
3. Verify upgrade from the previously installed TunnelMint development build to WireHush using the retained compatibility identifiers.
4. Verify official WireGuard remains installed/functional and is not modified by WireHush install/upgrade/uninstall.
5. Import/use the existing owner test tunnel without exposing secrets.
6. Plain DNS: connect, traffic works, DNS works, disconnect restores prior DNS.
7. DoH: connect using the known-good configured endpoint, confirm DNS works through DoH, and confirm no system/plaintext fallback after readiness.
8. Fail-closed: one concise break test only; breaking the active DoH path must cause DNS failure rather than plaintext/system fallback.
9. Bootstrap family behavior: rely on automated coverage for IPv4-only, IPv6-only, and dual-stack family matching unless a concrete regression requires manual proof.
10. Disconnect/cleanup leaves no stale DNS/firewall/route state owned by WireHush.
11. Uninstall removes WireHush-owned components only; official WireGuard remains untouched.
12. Reinstall after uninstall succeeds.
13. Packaged output contains required `LICENSE`, `WIREGUARD-COPYING`, `GPL-2.0.txt`, and `THIRD_PARTY_NOTICES.md` (or their deliberate current equivalents).
14. Search active product UI/docs/build outputs for unintended `TunnelMint` branding. Remaining occurrences are acceptable only when historical evidence or deliberate compatibility identifiers documented in `docs/wirehush-identity-migration.md`.

## Final Release / Legal Packaging Gate

Review, but do not rewrite third-party license text:

- Project-owned licensing / Required Notice names WireHush where appropriate.
- WireGuard copyright/license attribution remains intact.
- GPL-marked imported installer/fetcher material retains its required notices/source licensing treatment.
- `THIRD_PARTY_NOTICES.md`, `TRADEMARKS.md`, `PRIVACY.md`, `SECURITY.md`, `SUPPORT.md`, README, and installer-facing claims match actual verified behavior.
- No wording implies WireHush is an official WireGuard product or endorsed by WireGuard LLC.
- No claim promises behavior not actually verified, especially encrypted DNS, fallback/leak behavior, telemetry, or privacy claims.
- Record any licensing/trademark question needing owner/counsel review instead of inventing legal certainty.

## Code Signing

Determine and report the current Windows signing state. Do not block engineering on obtaining a certificate unless one is already available, but clearly identify an unsigned public build / SmartScreen warning as a release decision for the owner.

## Repository Name Handoff

The GitHub repository is still named `TunnelMint`. Do not fake a repository rename in source.

After this RC passes, the owner will rename the GitHub repository to `WireHush` from repository Settings. Then verify/update local remotes and any final repository URLs where needed. Treat the old GitHub repository URL as transitional, not the final public identity.

## Testing Style

Keep this SHORT. Use automated tests first. Do not recreate the giant Task014 owner matrix.

Ask for manual owner interaction only where elevation or real installed networking genuinely requires it.

Do not expose tunnel keys, client IDs, credentials, or sensitive packet contents in logs, commits, screenshots, or PR text.

## Completion

When complete:

1. Commit only necessary RC fixes/evidence/docs to `codex/wirehush-release-candidate`.
2. Push the branch and open a PR against `main` if source/docs changed. If no changes are required, report that explicitly and do not create a meaningless PR.
3. Let GitHub Windows validation run on any changed head.
4. Report:
   - RC commit SHA
   - pass/fail for the checks above
   - any concrete blockers
   - signing state
   - any deliberate legacy identifiers still present
   - whether it is ready for repository rename + release automation/publishing

## Out Of Scope

- Android
- New features
- New VPN/DNS protocols
- Broad refactors
- Public release publishing itself
- GitHub Pages site design

The next step after a clean Task018 is: rename the GitHub repository to `WireHush`, then set up/finalize GitHub Release automation, release assets/checksums, and the simple Pages/download site.
