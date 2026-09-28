# Task 014 — Owner Windows v1 Acceptance Evidence

Test target: `codex/task-014-owner-acceptance` at `ca4b76f` (Task 016 implementation plus its verification report).

Task 014 is not complete. No owner/manual acceptance checkbox is marked passed by this report.

## Environment evidence

- The owner-provided file `C:\TunnelMint-Test\owner-acceptance.conf` is present (presence and file size were checked only). Its contents and credentials were not read, copied, logged, staged, or published.
- The current shell is not elevated (`admin=False`).
- No `TunnelMintManager` service is installed in this environment.
- Official WireGuard is installed at `C:\Program Files\WireGuard\wireguard.exe`.
- The Task 016 client and installer builds produced unsigned development artifacts. The three MSI files are under `installer/dist/`.

## Automated evidence available before owner testing

The Task 016 implementation and its deterministic tests passed locally, and the pushed implementation SHA passed the hosted Windows validation workflow at [run 36386235313](https://github.com/neosmith20/TunnelMint/actions/runs/36386235313). Those results establish build/test readiness for manual testing; they do not satisfy the real-machine acceptance checks below.

## Owner/manual checks not run

The following groups remain pending on an elevated disposable Windows 11 VM with the owner’s real peer and observation:

- install, launch, manager/service isolation, coexistence with official WireGuard, uninstall/reinstall, and upgrade;
- plain-DNS handshake, traffic, DNS behavior, and disconnect restoration;
- full-tunnel and split-tunnel DoH handshake, HTTPS routing, endpoint path preservation, packet capture, and runtime prefix cleanup;
- normal-user Bootstrap Settings persistence, ordering, custom resolver behavior, and validation errors;
- failure-closed behavior for bootstrap, TLS, endpoint status, unreachable DoH, activation transition, disconnect, and force-stop cases;
- reboot, sleep/resume, Ethernet/Wi-Fi changes, and physical network loss/reconnect;
- IPv6 endpoint routing and DNS leak checks; and
- installed-package notices, TunnelMint naming, uninstall scope, upstream WireGuard preservation, and expected unsigned-build warnings.

These checks require native adapter/service access, elevation, real tunnel traffic, packet capture, and owner observation that are unavailable in this shell. No production-readiness or beta-readiness claim is made.

## Required owner next step

Run the checklist in `TASKS/014-owner-windows-v1-acceptance.md` on the elevated disposable Windows VM using the local-only configuration at `C:\TunnelMint-Test\owner-acceptance.conf`. Record only redacted evidence and turn every failure into a focused follow-up task before release consideration.
