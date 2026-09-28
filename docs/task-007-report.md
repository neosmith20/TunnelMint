# Task 007 — Tunnel and DoH Runtime Integration

## Lifecycle

Encrypted-DNS tunnel startup now follows a deterministic coordinator after the WireGuard adapter is configured and brought up:

1. Bootstrap the configured DoH hostname through the ordered bootstrap resolver set.
2. Reject endpoint candidates that are invalid, unspecified, multicast, or equal to a tunnel peer transport endpoint.
3. Add only host routes (`/32` or `/128`) for safe candidates through the active tunnel interface, recording routes that TunnelMint actually created.
4. Build the resolved-IP DoH transport while preserving the original HTTPS hostname for TLS SNI and HTTP host handling, then verify it with a DNS probe.
5. Start the loopback DNS proxy.
6. Save the adapter's existing IPv4 DNS addresses and point the tunnel adapter at `127.0.0.1`.
7. Mark the session ready only after all stages succeed.

The coordinator owns reverse-order cleanup: restore DNS, stop the proxy, and remove only TunnelMint-owned host routes. Activation failures use the same rollback path. `Session.Refresh` supports endpoint-address changes by verifying a replacement transport and proxy before removing stale owned routes. Plain `DNS = <IP>` configurations do not enter this path and retain the existing tunnel behavior.

The Windows service closes the encrypted-DNS session before the existing tunnel watcher teardown. Existing adapter routes, firewall rules, and ordinary tunnel lifecycle code remain in place; this task does not add broad routes or modify the firewall policy.

## Verification

- Lifecycle tests: `go test -vet=off ./dohruntime` — passed, including ordering, recursion rejection, rollback, route ownership, and refresh.
- Local DNS proxy tests remain passing.
- DoH/bootstrap tests remain passing. The existing Windows configuration-store test was attempted but could not write its protected test storage in this VM (`Access is denied`).
- Windows build: `cmd /c build.bat` — passed for x86, amd64, and arm64.
- The build overlay was updated only for Go 1.27 API compatibility (no-BoringCrypto stubs and CPU feature symbols); upstream crypto notices and the existing overlay behavior remain represented.

Disposable-VM checks for a real tunnel, split-tunnel host-route behavior, adapter DNS restoration across sleep/wake, and IPv6 Windows DNS selection remain unverified in this environment. The runtime currently uses the IPv4 loopback proxy address; the proxy package itself supports explicit IPv6 loopback listeners.

