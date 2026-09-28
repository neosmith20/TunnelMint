# Task 015 — Pre-acceptance recovery hardening report

Branch: `codex/pre-acceptance-recovery-hardening`

Implementation commit: `1e312f7` (`Harden pre-acceptance recovery paths`)

Task 014 remains the only Windows v1 acceptance gate. No owner/manual checks are marked complete by this work.

## Fixes

1. **Runtime peer cleanup.** `tunnel/doh_peer_routes_windows.go` now treats the active temporary prefix set as a transaction. It restores the selected peer to the exact configured `AllowedIPs` when the final temporary prefix is removed, retains bookkeeping when `SetConfiguration` fails, and reapplies the live set idempotently after adapter recovery. Tests cover add/delete, multiple prefixes, final restoration, replacement, and injected update failures.
2. **Refresh consistency.** `dohruntime.Session` serializes refresh, recovery, and close operations. Refresh commits `s.routes` and the selected endpoint addresses after the proxy swap; stale-route failures are retained in `s.routes` so ownership matches live state and a later refresh can retry cleanup. Route-add, verification, proxy-swap, stale-cleanup, successful replacement, repeated refresh, and recovery tests were added.
3. **Adapter/network recovery.** `interfaceWatcher` now accepts an explicit recovery callback. After interface setup or adapter reinitialization, the encrypted-DNS session restores its owned routes and runtime peer prefixes, verifies the DoH transport, swaps the existing proxy transport, reapplies loopback DNS, and reconfigures the firewall. Any failure is sent to the tunnel service so it stops instead of reporting stale ready state. Plain-DNS watcher behavior is unchanged.
4. **CI coverage.** `.github/workflows/windows.yml` now runs the deterministic `tunnel`, `tunnel/firewall`, `manager`, and `ui` tests in addition to the existing bootstrap, proxy, DoH, runtime, product, and focused configuration tests. Privileged adapter, DPAPI, service, and storage tests remain intentionally excluded from hosted CI.
5. **Pre-tunnel endpoint resolution.** `conf.ResolveEndpointsWith` and `bootstrap.Resolver.ResolveHost` allow encrypted-DNS startup to resolve peer hostnames through the same ordered enabled bootstrap resolver list used for DoH bootstrap. IP literals bypass resolution, ports remain unchanged, and failure stops activation before the tunnel reports started. Plain-DNS tunnels keep the inherited system resolver path.

## Verification

Passed locally with the bundled Go toolchain and a task-local cache:

```text
go test -vet=off ./bootstrap ./dnsproxy ./doh ./dohruntime ./product
go test -vet=off ./tunnel ./tunnel/firewall ./manager ./ui
go test -vet=off ./conf -run 'Test(DNSParsing|DNSRoundTripPreservesDoH|DNSURLValidation|FromWgQuick|ResolveEndpointsWith)$'
cmd /c build.bat                         # x86, amd64, arm64
cmd /c installer\build.bat              # x86, amd64, arm64 artifacts produced
```

The full `./conf` package remains environment-bound: inherited storage tests fail with access denied in this non-elevated VM. WiX compiler/linker output was produced for all three MSI architectures, but ICE validation cannot run because the VM has no Windows Installer service.

## Remaining limitations

Real peer traffic, packet capture, install/service/manager behavior, reboot, sleep/wake, network transitions, IPv6, and upstream WireGuard coexistence still require the elevated disposable Windows owner environment described by Task 014. The branch does not claim production readiness.

GitHub Actions Windows validation passed on the pushed implementation at [run 36382538291](https://github.com/neosmith20/TunnelMint/actions/runs/36382538291). The separate repository-managed Code scanning AI findings job failed in its external processing step and did not report a source finding; it is outside the supported Task 015 workflow.
