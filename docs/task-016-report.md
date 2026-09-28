# Task 016 — Final Pre-Acceptance Cleanup Report

Branch: `codex/final-pre-acceptance-cleanup`

Implementation commit: `583c24a` (`Harden final pre-acceptance cleanup paths`)

Task 014 remains open. No owner/manual Windows acceptance checks are marked complete by this work.

## Fixes

1. **Windows host-route ownership.** Route bookkeeping now preserves TunnelMint ownership when recovery re-applies an existing route, keeps pre-existing routes unowned, and records recreated routes as owned. Focused tests cover reapply, pre-existing, and recreate cases.
2. **Adapter reinitialization fail-closed behavior.** Configuration and adapter bring-up are sequenced through a testable helper. A failure emits the matching service error and stops the recovery callback before encrypted-DNS recovery can run. Focused tests verify ordering and error mapping.
3. **Refresh rollback ownership.** Verification and proxy-swap rollback now report cleanup failures, retain prefixes whose deletion failed, and leave them available for `Close` retry. `Close` retains failed route cleanup state across calls. Focused tests inject both rollback failures and verify the retry.
4. **Mixed-DNS validation.** The parser rejects plain DNS addresses mixed with a DoH URL and rejects multiple DoH URLs, with clear errors. DoH plus search suffixes and ordinary plain DNS remain supported and round-trip tested.
5. **CI coverage.** The focused configuration test expression now includes `ResolveEndpointsWith` and the new DNS validation/round-trip tests. Privileged adapter, DPAPI, service, and protected-storage tests remain out of hosted CI.
6. **Endpoint refresh review.** No periodic refresh subsystem was added; the existing refresh path was limited to the required rollback correctness fixes.

## Verification

Passed locally with the bundled Go toolchain:

```text
go test -vet=off ./bootstrap ./dnsproxy ./doh ./dohruntime ./product
go test -vet=off ./tunnel ./tunnel/firewall ./manager ./ui
go test -vet=off ./conf -run 'Test(DNSParsing|DNSRoundTripPreservesDoH|DNSRoundTripPreservesDoHWithSearchSuffix|DNSRejectsMixedPlainAndDoH|DNSRejectsMultipleDoH|DNSURLValidation|FromWgQuick|ResolveEndpointsWith)'
cmd /c build.bat                         # x86, amd64, arm64 — passed
cmd /c installer\build.bat              # x86, amd64, arm64 MSI artifacts produced
```

The installer build completed, but WiX ICE validation remains environment-bound in this non-elevated VM because the Windows Installer service is unavailable. This is not treated as a source-code failure, and ICE validation is not claimed as passed.

## GitHub Actions

Windows validation passed on the pushed branch: [run 36386235313](https://github.com/neosmith20/TunnelMint/actions/runs/36386235313). The focused workflow includes the Task 016 parser and `ResolveEndpointsWith` tests. The separate repository-managed Code scanning AI findings run failed in its external processing step; it did not report a source finding and is outside the supported Windows validation workflow.

## Remaining limits

The real peer, elevated disposable VM, packet capture, installer isolation, service/manager lifecycle, network transitions, IPv6, and owner observation required by Task 014 remain unverified. This branch does not claim Windows v1 beta-readiness or production-readiness.
