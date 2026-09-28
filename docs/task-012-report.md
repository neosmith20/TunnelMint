# Task 012 report — Windows v1 acceptance and release-candidate review

## Review result

Tasks 003 through 011 form one Windows implementation with a small integration surface. Standard `DNS = <IP>` values continue through the existing WireGuard configuration path. An `https://...` value is classified as DoH, uses normal certificate validation, and is served to Windows through a loopback DNS proxy. DoH requests use the endpoint hostname for TLS and HTTP while dialing the resolved address through tunnel-owned routes. There is no silent plaintext fallback.

Encrypted activation stages bootstrap resolution, endpoint routes, the DoH proxy, and DNS blocking together. Bootstrap DNS is temporarily allowed only to the configured built-in or Settings resolver addresses and is removed when the tunnel is ready; the remaining DNS firewall policy permits the loopback proxy. Activation failure rolls back the staged routes, proxy, DNS state, and firewall state. Disconnect and service cleanup close the proxy and remove TunnelMint-owned routes and firewall state.

The implementation remains scoped to tunnel import/editing, plain DNS, DoH, bootstrap settings, diagnostics, branding, and Windows packaging. It adds no telemetry, accounts, advertising, filtering, mesh networking, IDS/IPS behavior, or protocol changes. TunnelMint uses separate manager, tunnel, data, registry, window, and installer identities so it can coexist with upstream WireGuard. Required upstream notices remain in the repository.

## Automated verification

Focused tests passed:

- `go test -vet=off ./bootstrap ./dnsproxy ./doh ./dohruntime ./product`
- `go test -vet=off ./conf -run 'Test(DNSParsing|DNSRoundTripPreservesDoH|DNSURLValidation|FromWgQuick)$'`

The broad compatible sweep `go test -vet=off ./...` reached upstream Windows-environment tests but could not complete reliably here. `conf.TestStorage` requires protected Program Files storage, `dpapi.TestRoundTrip` requires the VM's DPAPI/file environment, and the `winipcfg` tests require a usable Windows adapter. These are environment-bound results rather than product test failures; the focused TunnelMint behavior tests passed.

Static review found no production `InsecureSkipVerify`, equivalent TLS bypass, embedded private key, credential, token, or real tunnel configuration. The updater remains disabled (`updatesEnabled=false`) and points at the non-routable placeholder host `updates.invalid`; no owned update service is contacted. Debug logging is not enabled by the release build. `LICENSE`, `THIRD_PARTY_NOTICES.md`, and `WIREGUARD-COPYING` remain present.

Build verification passed after clearing the application build cache:

- `cmd /c build.bat` — x86, amd64, and arm64 executables built.
- `cmd /c installer\build.bat` — all three MSI packages were compiled. The VM's Windows Installer ICE service is unavailable, so WiX emitted ICE01–ICE07/ICE09 diagnostics while the normal script completed. The RC packages were relinked with the ICE actions suppressed locally to produce artifacts from the final executable sources; the build script itself was not changed.

A fresh amd64 MSI installation was attempted with launch disabled in Task 011. Windows Installer identified TunnelMint Development 0.1.0, then returned error 1925 because this shell has no administrator privilege for the per-machine install and rolled back. No TunnelMint install directory, service, or uninstall entry remained.

## Acceptance checks

The local deterministic DoH client, bootstrap, proxy, route/lifecycle rollback, failure cleanup, and Settings tests passed. Plain-DNS parsing and DoH URL preservation passed. Static identity, updater, notices, and secret scans passed. Real Windows service/UI interaction, native packet observation, full/split tunnel traffic, IPv6, sleep/wake, and installer coexistence require an elevated disposable Windows desktop and a real test peer; they remain owner/manual verification items in the acceptance matrix.

No private credentials or peer configuration were available or copied into the repository. No public credentials were used.

## Release-candidate artifacts

These are unsigned development artifacts. SHA-256 hashes are recorded for the exact files produced from the final sources:

| Artifact | SHA-256 |
| --- | --- |
| `x86/tunnelmint.exe` | `7019CD8A3B5D3BCC27DA23F1679F5EB8B319925A0D0C204F52B50F70728C4361` |
| `amd64/tunnelmint.exe` | `B2E7B815045F319D8D8BFC2A9D1AE03F1B87037C09D936A557E864839458A7C4` |
| `arm64/tunnelmint.exe` | `04486475E59FB19B0DA11D96AC705005C87C97006EFF8B29145884D6DD467953` |
| `installer/dist/tunnelmint-x86-0.1.0.msi` | `9B3DF19E051A7E9C0DBB3CF8ED5F740870DBD0F2C7DBD83F852F8B08F52B2A15` |
| `installer/dist/tunnelmint-amd64-0.1.0.msi` | `DAEA00E1E791881A823CC153C726EF500201A1009CB87682EEF239088321DDB2` |
| `installer/dist/tunnelmint-arm64-0.1.0.msi` | `9C2FD28A69F4420FE702887BE459560272C1A1A24F49CC32B52D420F688DCF5B` |

The packages are intended for owner testing and are not a stable signed release.
