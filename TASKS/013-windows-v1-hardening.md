# Task 013 — Windows v1 Hardening

## Goal

Harden the merged Windows v1 implementation before owner/manual acceptance. This task is about fixing correctness, leak-resistance, coexistence, normal-user Settings behavior, packaging/licensing accuracy, and regression protection. Do not add unrelated features.

Start from current `main` after PR #3 (`Complete TunnelMint Windows v1 implementation`) and create a focused branch, preferably:

`codex/windows-v1-hardening`

Before changing anything, read `AGENTS.md`, `ROADMAP.md`, `KNOWN_LIMITATIONS.md`, `THIRD_PARTY_NOTICES.md`, and `docs/task-012-report.md`.

Work through every item below. Do not stop after the first fix. Push checkpoints as useful, but keep the hardening work on this branch until all code items are complete and tested. Do not merge to `main` yourself.

---

## P0 — Release blockers

### 1. Make split-tunnel DoH actually routable by WireGuard

Current behavior adds Windows `/32` or `/128` host routes for bootstrapped DoH endpoint addresses, but Windows routing alone is insufficient when the selected WireGuard peer does not already own those destination prefixes in its runtime `AllowedIPs`.

Fix the runtime so an encrypted-DNS endpoint is cryptokey-routable through the correct WireGuard peer in split-tunnel configurations.

Requirements:

- Full-tunnel behavior must remain unchanged.
- For split tunnel, add only the minimum DoH endpoint host prefixes needed at runtime.
- Do not permanently rewrite the user's configuration file merely to make the runtime work.
- Preserve the user's original configured `AllowedIPs` when the tunnel stops.
- Endpoint refresh/re-resolution must update the runtime peer prefixes safely.
- Never steal/assign a DoH endpoint prefix to an ambiguous or incorrect peer.
- If TunnelMint cannot determine a safe peer/path, activation must fail closed with a useful error instead of routing DoH outside the tunnel.
- Add tests proving both full-tunnel and split-tunnel behavior, including cleanup and endpoint replacement.

Likely areas: `tunnel/service.go`, `tunnel/doh_integration_windows.go`, driver configuration/runtime peer update code, `dohruntime`.

### 2. Remove the DNS leak window during WFP reconfiguration

Current `firewall.ReconfigureDNS` disables the existing dynamic WFP session before installing the replacement session. That creates a non-zero interval with no TunnelMint firewall policy.

Replace this with make-before-break or an equivalent transactional design.

Requirements:

- At no point during encrypted-DNS activation/finalization may ordinary plaintext DNS be temporarily reopened because one WFP session was removed before its replacement became active.
- Prefer updating/installing replacement policy while the old restrictive policy remains active, then retire the old policy only after the new policy is successfully committed.
- Failure must leave a restrictive/fail-closed policy in place.
- Do not broaden the policy to solve this.
- Add focused failure-injection tests for replacement setup/commit/cleanup ordering.

Likely area: `tunnel/firewall/blocker.go` and supporting WFP/session code.

### 3. Fix Settings persistence for a normal non-admin UI

The Settings UI currently reads/writes `Program Files\\TunnelMint\\Data\\bootstrap-dns.json` directly while the data root ACL grants full access only to SYSTEM/Administrators. This must work for the normal unelevated TunnelMint UI without weakening the protected tunnel/data ACL.

Requirements:

- Keep the protected machine-wide TunnelMint data/config directory ACL restrictive.
- Do not grant normal users blanket write access to the protected data root.
- Move bootstrap Settings read/write through the existing elevated manager/service IPC, or another narrowly scoped privileged mechanism consistent with the current architecture.
- Validate the full Settings payload again in the privileged side before saving.
- Use atomic persistence.
- UI errors must be useful and must not expose secrets.
- Settings remain machine-wide unless there is a compelling existing product convention to the contrary.
- Add tests for authorization/validation/serialization and the UI-to-manager handoff where practical.

Likely areas: `ui/settingspage.go`, manager IPC/service code, `bootstrap/settings.go`, `conf/path_windows.go` only if needed. Do not weaken the ACL as the fix.

### 4. Give TunnelMint its own deterministic adapter GUID namespace

`tunnel/deterministicguid.go` still uses the upstream WireGuard deterministic/fixed GUID labels. TunnelMint and an installed upstream WireGuard client importing the same tunnel/config must not derive the same adapter GUID.

Requirements:

- Use TunnelMint-owned deterministic GUID label(s)/namespace for TunnelMint-created adapters.
- Preserve deterministic behavior within TunnelMint across reconnect/reboot.
- Do not modify WireGuard protocol/keys/crypto.
- Add regression tests showing identical tunnel input under the upstream label and TunnelMint label cannot produce the same deterministic GUID.
- Preserve required upstream copyright/license notices.

Likely area: `tunnel/deterministicguid.go`.

---

## P1 — Correctness and resilience

### 5. Fix DoH endpoint refresh without rebinding a second proxy to port 53

`dohruntime.Session.Refresh` currently prepares a replacement proxy before closing the old proxy. The real proxy binds TCP and UDP on the same loopback `:53`, so a second live instance cannot reliably bind the same address/port.

Refactor refresh so endpoint address changes can be applied without trying to own `127.0.0.1:53` twice.

Preferred direction: keep the existing loopback listener and atomically replace/swap its upstream verified DoH client/transport, or implement another design that does not create an avoidable DNS outage or bind collision.

Requirements:

- Existing proxy remains available until replacement upstream transport is verified.
- Do not expose plaintext fallback.
- Do not create a window where no DNS listener exists unless the session is intentionally failing closed.
- Clean stale endpoint routes only after the replacement path is ready.
- Add an integration-style local test that would fail if two real proxies tried to bind the same loopback port.

Likely areas: `dohruntime/lifecycle.go`, `dnsproxy/proxy.go`, `tunnel/doh_integration_windows.go`.

### 6. Improve bootstrap resolver failover timeout semantics

The current resolver uses one shared timeout context across the entire ordered resolver list. A dead first resolver can consume the whole deadline and prevent later configured resolvers from being attempted.

Requirements:

- Give each resolver a bounded attempt while also enforcing a reasonable overall bootstrap deadline.
- Preserve configured ordering.
- Cancellation from the caller must still stop immediately.
- Avoid making startup painfully slow if several resolvers are dead.
- Add deterministic tests where resolver 1 times out and resolver 2 succeeds.

Likely area: `bootstrap/resolver.go`.

### 7. Align bootstrap firewall exceptions with actual Settings

The initial encrypted-DNS firewall path currently appends `bootstrap.DefaultResolvers()` rather than the actual enabled resolver list loaded from `bootstrap-dns.json`.

Requirements:

- The temporary bootstrap DNS exception must be derived from the exact resolver addresses TunnelMint is about to use, including custom resolvers and disabled built-ins.
- A disabled built-in resolver should not remain explicitly allowed merely because it is a default.
- A custom-only resolver configuration must work.
- Keep the exception scoped to the bootstrap window and then remove it when encrypted DNS is ready.
- Add tests for custom-only, reordered, and disabled-default configurations.

Likely areas: `tunnel/addressconfig.go`, `tunnel/service.go`, `tunnel/doh_integration_windows.go`, firewall setup sequencing.

---

## P2 — Packaging, docs, license inventory, and CI

### 8. Correct README/project status and stale links

Update `README.md` to match the current project state.

At minimum:

- Remove the stale statement that the Windows client/encrypted DNS functionality has not been implemented.
- Replace the deleted `CONTRIBUTOR_COPYRIGHT_ASSIGNMENT.md` link with `CONTRIBUTOR_LICENSE_AGREEMENT.md`.
- Describe the build as alpha/development software awaiting real Windows acceptance rather than implying production readiness.
- Do not claim real tunnel/leak/sleep/wake/install verification that has not happened.

### 9. Make third-party license inventory accurate

`THIRD_PARTY_NOTICES.md` currently describes the imported Windows foundation broadly as MIT, but individual upstream files in the imported tree have their own SPDX/license markers, including GPL-2.0-marked installer sources such as `installer/wireguard.wxs` and `installer/fetcher/fetcher.c`.

Requirements:

- Audit the imported/distributed upstream material sufficiently to avoid stating that every imported file is MIT when file-level licensing says otherwise.
- Preserve all upstream SPDX/copyright headers.
- Update `THIRD_PARTY_NOTICES.md` with an accurate, conservative summary of applicable upstream/file-level licensing.
- Do not attempt to relicense third-party source as TunnelMint-owned PolyForm code.
- If a license conclusion is uncertain, document the uncertainty rather than inventing an answer.

This is engineering/documentation work, not legal advice; do not silently delete upstream components just to make the notice simpler.

### 10. Include required notices with distributed Windows packages

The current MSI primarily installs the executables. Ensure distributed packages include or otherwise reliably accompany required license/notice material.

At minimum include appropriate copies of:

- TunnelMint `LICENSE`
- `WIREGUARD-COPYING`
- `THIRD_PARTY_NOTICES.md`

If additional notice/license files are required by the audit above, include those too.

Requirements:

- Installed location should make notices discoverable without cluttering the primary UI.
- Installer uninstall/upgrade behavior must handle these files normally.
- Preserve separation from upstream WireGuard's installation.

Likely area: `installer/wireguard.wxs` and release docs.

### 11. Add Windows CI regression protection

`main` currently has no CI status checks. Add a GitHub Actions Windows workflow suitable for pull requests and main-branch pushes.

Requirements:

- Run focused TunnelMint tests at minimum for `bootstrap`, `dnsproxy`, `doh`, `dohruntime`, `product`, and the focused DNS config tests.
- Build the Windows client using the project's supported/reproducible build path.
- If full installer building is practical and deterministic in GitHub Actions, include it; otherwise document why it remains a separate release build.
- Do not upload private configs, secrets, or credentials.
- Cache only safe build dependencies if useful.
- A failed focused test/build must fail the workflow.

Do not weaken tests just to make CI green.

### 12. Update roadmap/limitations accurately

After the fixes are complete:

- Update `ROADMAP.md` only for items genuinely proven by automated/static work.
- Keep real-machine acceptance items unchecked until owner testing happens.
- Update `KNOWN_LIMITATIONS.md` to remove issues actually fixed and retain those still requiring manual validation.
- Create `docs/task-013-report.md` with all verification details and any remaining risks.

---

## Required verification

Run focused tests for every affected package and add regression tests for each bug fixed above.

At minimum run:

`go test -vet=off ./bootstrap ./dnsproxy ./doh ./dohruntime ./product`

and the focused config/DNS tests already used by Task 012.

Run the normal Windows build:

`cmd /c build.bat`

Run the installer build if the installer/license changes permit it in the environment:

`cmd /c installer\\build.bat`

If an environment-bound test cannot run, state exactly why. Do not classify an unexecuted real-Windows scenario as passed.

Static/security checks must include:

- no production TLS verification bypass;
- no plaintext DoH fallback;
- no embedded credentials/private tunnel configs;
- no unintended broad firewall exception;
- TunnelMint-owned services/data/adapter GUID namespace remain separate from upstream client identity;
- required third-party notices remain present.

## Completion report

Create `docs/task-013-report.md` containing:

- branch name and commit SHA(s);
- each finding above and the exact fix used;
- files/packages changed;
- split-tunnel peer/AllowedIPs design and cleanup behavior;
- WFP replacement design and failure behavior;
- Settings IPC/privilege design;
- adapter GUID namespace change;
- proxy refresh design;
- bootstrap timeout/failover behavior;
- firewall/settings resolver alignment;
- licensing/installer notice changes;
- CI workflow summary;
- tests and build results;
- items still requiring owner/manual Windows testing.

## Completion criteria

Task 013 is complete only when all P0 and P1 defects above are fixed with regression coverage, packaging/docs/CI work is complete, the normal Windows build passes, and the branch is pushed for review.

Do **not** mark the Windows v1 release as production-ready and do **not** mark owner/manual acceptance items as complete. Do not begin Android work.
