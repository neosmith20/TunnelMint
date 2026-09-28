# Task 006 — Local DNS Proxy

## Design

The new `dnsproxy` package exposes a lifecycle-managed local DNS service. It binds UDP and TCP listeners only to the explicitly supplied loopback address; the default is `127.0.0.1:53`. A wildcard address, hostname, malformed port, or other non-loopback address is rejected before binding. Port conflicts are returned from `Start` with the listener protocol and address.

Each UDP datagram and TCP DNS frame is validated as a bounded raw DNS message, forwarded through the existing `doh.Client`, and returned with its wire data and DNS ID unchanged. TCP responses use the two-byte DNS length prefix. The proxy has bounded query and response sizes (65,535 bytes by default), a bounded concurrent query semaphore, per-query timeout, TCP read/write deadlines, cancellation, and clean listener shutdown/restart. DoH failures close the request path and record the error; there is no plaintext fallback, DNS cache, filtering, query logging, or payload logging.

`Status()` reports running state, bound UDP/TCP addresses, active query count, and the last internal error for later diagnostics. The constructor requires an HTTPS endpoint when it creates the DoH client, and `ValidateEndpoint` is available for configuration validation.

## Verification

- Focused proxy tests: `go test -vet=off -timeout 30s ./dnsproxy` — passed.
- Existing DoH/bootstrap tests remain passing from the previous checkpoints.
- Windows build: `cmd /c build.bat` — pending for this task checkpoint.

Tests use only local TLS and loopback servers and cover UDP, TCP framing, concurrent queries, malformed and oversized input, upstream failure, timeout, loopback-only binding, no fallback, and restart behavior.

Windows adapter DNS settings, tunnel routes, firewall rules, and full tunnel lifecycle integration remain intentionally unmodified for Tasks 007–008.

