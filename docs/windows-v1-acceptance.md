# TunnelMint Windows v1 acceptance matrix

| Check | Status | Evidence / remaining work |
| --- | --- | --- |
| Plain `DNS = <IP>` parsing | PASS | Existing configuration tests pass; plain values remain on the standard tunnel path. |
| DoH URL parsing and path preservation | PASS | `conf` focused tests pass. |
| DoH TLS and HTTP behavior | PASS | `doh` tests pass; normal certificate validation and bounded requests are enforced. |
| Bootstrap resolver defaults and settings | PASS | `bootstrap` tests pass; defaults, validation, ordering, persistence, and custom entries are covered. |
| Loopback DNS proxy | PASS | `dnsproxy` tests pass for UDP/TCP forwarding and bounded failure behavior. |
| Activation, rollback, and cleanup state machine | PASS | `dohruntime` tests pass for staged activation, rollback, refresh, and route ownership. |
| No silent plaintext fallback | PASS | No fallback path in DoH runtime; DNS blocking remains active when replacement fails. |
| No TLS verification bypass | PASS | Static scan found no production `InsecureSkipVerify` or equivalent. |
| Updater isolation | PASS | Updater is disabled and uses the `updates.invalid` placeholder. |
| TunnelMint/upstream identity separation | PASS | Separate service, route, registry, data, window, and MSI identities are implemented. |
| Notices and secret scan | PASS | Required notices remain; no private keys, credentials, tokens, or real configs found. |
| Three-architecture application build | PASS | `build.bat` passed for x86, amd64, and arm64. |
| Three-architecture MSI artifacts | PASS | RC packages produced for x86, amd64, and arm64; hashes are in `docs/task-012-report.md`. |
| Full WiX ICE validation | NOT TESTED | The VM Windows Installer service cannot execute ICE actions; local relinking suppressed those environment checks. |
| Fresh elevated install/uninstall/reinstall | NOT TESTED | Per-machine install returned Windows Installer error 1925 in this non-elevated shell. |
| Upstream WireGuard coexistence | NOT TESTED | No elevated desktop with an installed upstream client was available. |
| Interactive import/edit/status UI | NOT TESTED | Manager service/elevated desktop is unavailable in this VM. |
| Plain DNS through a real tunnel | NOT TESTED | No usable test peer/configuration was available. |
| DoH through real full and split tunnels | NOT TESTED | No usable test peer/configuration was available. |
| Native packet/DNS leak observation | NOT TESTED | Requires a real tunnel and Windows packet observation. |
| IPv4 real tunnel traffic | NOT TESTED | Synthetic/local code paths passed; no real peer was available. |
| IPv6 tunnel behavior | NOT TESTED | Requires a VM/network with IPv6 tunnel coverage. |
| Sleep/wake and network switching | NOT TESTED | Requires a reliable desktop VM lifecycle and network changes. |
