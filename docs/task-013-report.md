# Task 013 — Windows v1 hardening report

Branch: `codex/windows-v1-hardening`

Base: `2dba28f` (`origin/main`, after the owner acceptance checklist)

The branch hardens the merged Windows v1 implementation before owner acceptance. It does not claim production readiness and does not mark real-machine acceptance complete.

## Findings and fixes

1. **Split-tunnel DoH routing.** `tunnel/doh_peer_routes_windows.go` tracks only the endpoint `/32` or `/128` prefixes needed at runtime. It selects a single safe peer, preserves the configured `AllowedIPs`, applies the runtime peer update through WireGuardNT, and removes the temporary peer prefixes on refresh or close. Full-tunnel prefixes are detected as already covered. Ambiguous multi-peer ownership fails closed. `tunnel/doh_integration_windows.go` couples this manager with the existing Windows host-route ownership so pre-existing routes are not deleted.
2. **WFP replacement.** `tunnel/firewall/blocker.go` installs the replacement dynamic session completely before closing the old session. An injected install failure leaves the old restrictive session active; the ordering and failure behavior are covered by `tunnel/firewall/blocker_test.go`.
3. **Normal-user Settings.** `ui/settingspage.go` now reads and writes through manager IPC. `manager/ipc_server.go` validates the complete payload again, resolves the protected machine-wide data path on the privileged side, and uses the existing atomic settings persistence. `manager/bootstrap_settings_test.go` covers serialization and validation. The protected data-root ACL is unchanged.
4. **Adapter identity.** `tunnel/deterministicguid.go` uses TunnelMint-owned deterministic and fixed labels. `tunnel/deterministicguid_test.go` proves an identical configuration does not collide with the upstream WireGuard label while remaining deterministic within TunnelMint.
5. **DoH refresh.** `dnsproxy.Proxy.SetClient` swaps the verified upstream under a lock while keeping the existing TCP/UDP listeners. `dohruntime.Session.Refresh` verifies the new transport, swaps the existing proxy, then removes stale endpoint routes. `dnsproxy/proxy_test.go` uses real local TLS servers and confirms the listener address remains unchanged.
6. **Bootstrap failover.** `bootstrap/resolver.go` gives each configured resolver a bounded attempt inside one overall deadline, preserves resolver order, and propagates caller cancellation. `bootstrap/resolver_test.go` covers a timed-out first resolver followed by a successful second resolver.
7. **Bootstrap firewall settings.** `tunnel/service.go` loads the enabled ordered resolver list before encrypted-DNS firewall setup. `tunnel/addressconfig.go` uses exactly that list plus loopback for the temporary exception. Tests cover custom-only, reordered, and disabled-default inputs.
8. **Project status.** `README.md` now describes the alpha/development state, the implemented Windows v1 scope, and the remaining real-machine verification. The contributor agreement link points to `CONTRIBUTOR_LICENSE_AGREEMENT.md`.
9. **License inventory.** `THIRD_PARTY_NOTICES.md` records mixed file-level upstream licensing and identifies the GPL-2.0-marked installer/fetcher sources. `GPL-2.0.txt` is included without changing upstream headers or relicensing third-party code.
10. **Installer notices.** `installer/wireguard.wxs` installs TunnelMint `LICENSE`, `WIREGUARD-COPYING`, `THIRD_PARTY_NOTICES.md`, and `GPL-2.0.txt` under a separate `Notices` directory.
11. **CI.** `.github/workflows/windows.yml` runs the focused bootstrap, DNS proxy, DoH, runtime, product, and configuration tests on pull requests and pushes to `main`, then runs the supported Windows build. Installer linking remains a separate release step because the available environment does not provide deterministic Windows Installer ICE validation.
12. **Limitations.** `KNOWN_LIMITATIONS.md` retains the real-machine, owner-only acceptance items and records the unavailable Windows Installer service. `ROADMAP.md` leaves end-to-end tunnel, leak, sleep/wake, network-transition, and release acceptance unchecked.

## Verification

Passed with the bundled Go toolchain and a task-local build cache:

```text
go test -vet=off ./bootstrap ./dnsproxy ./doh ./dohruntime ./product
go test -vet=off ./tunnel ./tunnel/firewall ./manager ./ui
go test -vet=off ./conf -run 'Test(DNSParsing|DNSRoundTripPreservesDoH|DNSURLValidation|FromWgQuick)$'
cmd /c build.bat                         # x86, amd64, arm64
cmd /c installer\build.bat              # x86, amd64, arm64 artifacts produced
```

The installer compiler and linker completed for all three architectures. WiX ICE actions that require the Windows Installer service failed in this VM with `LGHT0217`/`LGHT0216`; that service-bound validation remains an owner or CI environment check.

Static review found no production TLS verification bypass, plaintext DoH fallback, embedded credential/private-key material, or broad bootstrap firewall exception. Required notice files are present in the source tree and in the MSI component definitions.

## Remaining owner/manual checks

An elevated disposable Windows system with a real WireGuard peer is still required to verify import, connect/disconnect, handshake and traffic, full- and split-tunnel DoH packet routing, DNS leak behavior, IPv4/IPv6, adapter coexistence with upstream WireGuard, manager IPC through the installed service, protected-directory install/upgrade/uninstall, reboot, sleep/wake, Ethernet/Wi-Fi changes, and failure recovery. The unsigned development artifacts are not a production release.

## Commit

Hardening implementation: `515f6db` (`Harden Windows v1 before owner acceptance`). The report update is the subsequent documentation commit at the branch head.
