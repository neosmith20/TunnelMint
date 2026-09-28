# Task 015 — Pre-Acceptance Recovery Hardening

## Goal

Fix the remaining lifecycle/recovery issues found after Task 013 before owner/manual Windows acceptance begins. This task is intentionally narrow: cleanup correctness, refresh transactional behavior, adapter/network recovery, CI coverage, and encrypted-DNS startup resolution policy.

Start from current `main` after PR #4 (`Harden Windows v1 before owner acceptance`) and create a focused branch, preferably:

`codex/pre-acceptance-recovery-hardening`

Before changing anything, read:

- `AGENTS.md`
- `ROADMAP.md`
- `KNOWN_LIMITATIONS.md`
- `TASKS/014-owner-windows-v1-acceptance.md`
- `docs/task-013-report.md`

Do not add unrelated features. Do not begin Android work. Do not mark Task 014 manual checks as passed.

---

## 1. Correct runtime DoH peer `AllowedIPs` cleanup and failure bookkeeping

The Task 013 `dohPeerRouteManager` dynamically adds endpoint `/32` or `/128` prefixes to the selected WireGuard peer, but cleanup is incomplete.

Current problem:

- When the last temporary prefix is removed, `apply()` can return without sending the original configured `AllowedIPs` back to WireGuardNT.
- `Delete()` mutates the in-memory `active` set before the driver configuration update is known to have succeeded. A failed `SetConfiguration()` can therefore leave internal bookkeeping inconsistent with the actual adapter state.

Requirements:

- Removing the final temporary DoH prefix must restore the selected peer to exactly its original configured `AllowedIPs`.
- Add/delete operations must be transactional from the manager's point of view: if the driver update fails, in-memory ownership/bookkeeping must still reflect the actual live state.
- Do not modify the persisted user configuration.
- Full-tunnel configurations that need no temporary prefix must remain unchanged.
- Cleanup must be idempotent.
- Add focused tests for:
  - single temporary prefix add then delete;
  - multiple temporary prefixes with one-by-one deletion;
  - final-prefix deletion restoring original `AllowedIPs`;
  - injected driver/update failure preserving manager state;
  - endpoint replacement old -> new without retaining stale prefixes.

Likely areas:

- `tunnel/doh_peer_routes_windows.go`
- `tunnel/doh_peer_routes_test.go`

---

## 2. Make `dohruntime.Session.Refresh` transactionally consistent

Current refresh flow can partially commit:

1. add new routes;
2. verify new DoH transport;
3. swap proxy client;
4. remove stale old routes;
5. update `s.routes`.

If stale-route removal fails after the proxy has already been swapped, the method returns while the live proxy/routes and `s.routes` can disagree.

Requirements:

- Define and implement a clear transactional refresh state model.
- On success, the proxy transport, route ownership, runtime peer `AllowedIPs`, and `s.routes` must all represent the same endpoint candidate set.
- On failure, either:
  - fully roll back to the old verified working state; or
  - commit the new state completely and report any non-critical stale cleanup separately without corrupting ownership/accounting.
- Never return with `s.routes` knowingly describing a different ownership set from the live state.
- Preserve the existing listener; do not rebind a second `127.0.0.1:53` proxy.
- Add failure-injection tests for:
  - new route add failure;
  - new transport verification failure;
  - proxy swap failure;
  - stale route deletion failure;
  - successful old -> new refresh;
  - repeated refresh to the same endpoints.

Likely areas:

- `dohruntime/lifecycle.go`
- `dohruntime/lifecycle_test.go`
- `dnsproxy/proxy.go` only if required

---

## 3. Recover encrypted-DNS runtime after adapter/network reinitialization

The inherited `interfaceWatcher` can reapply the original configured Windows routes/DNS and, when the adapter is brought back up, can reapply the original WireGuard driver configuration.

That can invalidate live encrypted-DNS state by removing one or more of:

- dynamic DoH Windows host routes;
- runtime peer DoH `AllowedIPs`;
- loopback Windows DNS (`127.0.0.1` / intended TunnelMint DNS state);
- other runtime assumptions owned by the encrypted-DNS session.

This is directly relevant to Task 014 reboot/sleep/wake and Ethernet/Wi-Fi transition testing.

Requirements:

- Introduce an explicit recovery/reapply path between the interface watcher and encrypted-DNS runtime.
- After an adapter/interface reinitialization, a tunnel using `DNS = https://...` must either:
  - restore all required DoH runtime state and return to ready; or
  - fail closed and stop/fail the tunnel with a useful logged error.
- Do not silently continue with Windows DNS reset to a non-TunnelMint resolver while reporting the DoH session as ready.
- Do not silently continue after the runtime peer endpoint prefix has been lost.
- Plain-DNS tunnels must retain upstream behavior.
- Avoid races between watcher callbacks, refresh, and shutdown.
- Recovery should be idempotent and safe if both IPv4 and IPv6 watcher events fire.
- Add unit/integration-style tests around the recovery coordination where possible. Real sleep/wake and adapter transitions remain Task 014 owner tests.

Likely areas:

- `tunnel/interfacewatcher.go`
- `tunnel/service.go`
- `tunnel/doh_integration_windows.go`
- `dohruntime/lifecycle.go`

Prefer a small explicit callback/interface rather than global hidden state.

---

## 4. Expand Windows CI to run the actual hardening tests

The current GitHub Actions workflow does not run several packages containing the most security-sensitive Task 013 changes.

Update `.github/workflows/windows.yml` so CI runs the supported focused tests for at least:

- `./bootstrap`
- `./dnsproxy`
- `./doh`
- `./dohruntime`
- `./product`
- `./tunnel`
- `./tunnel/firewall`
- `./manager`
- `./ui`
- existing focused `./conf` parser tests

Requirements:

- Keep the Windows build step.
- Do not make CI depend on privileged adapter/service/DPAPI tests that cannot reliably run on GitHub-hosted runners.
- If a package contains mixed privileged and unprivileged tests, select the deterministic subset rather than simply dropping that package from CI.
- Ensure the new Task 015 tests are executed by CI.
- Document any intentionally excluded tests in workflow comments or `KNOWN_LIMITATIONS.md`.

---

## 5. Eliminate or explicitly control pre-tunnel peer-hostname DNS when encrypted DNS is selected

Current startup calls the inherited `config.ResolveEndpoints()`, which uses the Windows system resolver (`GetAddrInfoW`) before TunnelMint enables its firewall. For a configuration such as:

```ini
[Peer]
Endpoint = vpn.example.com:51820

[Interface]
DNS = https://dns.example.com/dns-query
```

TunnelMint can therefore emit a normal system-DNS lookup for `vpn.example.com` before the tunnel exists.

For encrypted-DNS mode, that undermines a simple "no unintended plaintext DNS" story and complicates Task 014 packet-leak acceptance.

Preferred behavior:

- When `DNSOverHTTPS` is configured, resolve WireGuard peer endpoint hostnames through TunnelMint's configured bootstrap resolver path instead of the ordinary system resolver.
- IP-literal peer endpoints must bypass DNS normally.
- Plain-DNS tunnels should preserve existing upstream endpoint resolution behavior unless there is a strong reason to change it.
- Use the same ordered enabled bootstrap Settings that encrypted-DNS endpoint bootstrap uses.
- No system-resolver fallback in encrypted-DNS mode.
- If peer endpoint resolution fails, fail activation cleanly before the tunnel is reported started.
- Preserve endpoint ports exactly.
- Do not create recursion between peer endpoint resolution and DoH endpoint routing.
- Add tests proving encrypted-DNS mode does not invoke the ordinary system resolver for hostname peer endpoints.

If implementation constraints make this unsafe or fundamentally incompatible with startup ordering, document that clearly and fail closed rather than silently leaking. Do not simply waive the issue in documentation without first attempting the controlled bootstrap design.

Likely areas:

- `tunnel/service.go`
- `conf/dnsresolver_windows.go`
- `bootstrap`

---

## Verification requirements

Run the deterministic supported test suite, including the newly-added tests. At minimum:

```text
go test -vet=off ./bootstrap ./dnsproxy ./doh ./dohruntime ./product
go test -vet=off ./tunnel ./tunnel/firewall ./manager ./ui
go test -vet=off ./conf -run 'Test(DNSParsing|DNSRoundTripPreservesDoH|DNSURLValidation|FromWgQuick)$'
cmd /c build.bat
```

Run the installer build if the local environment supports it. Do not claim ICE validation if the Windows Installer service/environment cannot perform it.

After pushing the branch, confirm GitHub Actions executes the expanded test coverage successfully.

---

## Completion report

Create `docs/task-015-report.md` containing:

- branch name and final commit SHA;
- exact fixes implemented for all five items;
- tests added;
- commands run and results;
- GitHub Actions result/run reference;
- any remaining limitations;
- anything that still requires Task 014 real-machine verification.

Do not mark Task 014 checks complete.

Open a PR to `main` when all Task 015 code work is complete. Do not merge it yourself unless explicitly instructed by the owner.

## Exit criteria

Task 015 is complete only when:

- temporary runtime peer `AllowedIPs` are correctly restored and failure-safe;
- DoH refresh cannot leave ownership/accounting partially transitioned;
- adapter/network reinitialization explicitly restores DoH runtime state or fails closed;
- CI executes the hardening packages/tests;
- encrypted-DNS mode no longer performs uncontrolled system-DNS resolution for WireGuard peer hostnames;
- focused tests and Windows build pass;
- Task 014 remains the only remaining Windows-v1 acceptance gate.
