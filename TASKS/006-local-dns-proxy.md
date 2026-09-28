# Task 006 — Local DNS Proxy

## Goal

Provide the local DNS service Windows can later point at when a tunnel uses `DNS = https://...`.

This component accepts ordinary DNS queries locally and forwards them through the tested DoH transport/bootstrap stack.

## Required behavior

Implement a small local-only DNS proxy that:

1. Listens only on loopback addresses, never wildcard/public interfaces.
2. Supports normal DNS over UDP and TCP locally.
3. Accepts raw DNS wire-format queries and forwards each query through the configured DoH endpoint.
4. Returns the resulting DNS wire-format response to the requesting client.
5. Handles TCP DNS framing correctly.
6. Preserves request/response DNS IDs and wire data correctly.
7. Uses bounded request sizes, response sizes, timeouts, concurrency, and shutdown behavior.
8. Supports cancellation and clean listener shutdown.
9. Starts only for encrypted-DNS configurations.
10. Fails closed if the DoH path is unavailable; it must not fall back to plaintext DNS.
11. Exposes enough internal status/error information for later UI diagnostics without logging DNS payload contents by default.

Do not add filtering, ad blocking, query logging, analytics, or caching beyond what is strictly needed for correct operation.

## Port/listener handling

Prefer the simplest robust loopback listener design that works with Windows DNS client behavior. Detect and report port conflicts clearly rather than silently choosing an arbitrary externally invisible configuration that Windows cannot use.

If IPv6 loopback support is implemented, test it separately and do not let lack of IPv6 on a test VM break IPv4 correctness.

## Tests

Use local deterministic test servers. Cover at least:

- UDP query success;
- TCP query success;
- concurrent queries;
- malformed/oversized local input rejection;
- upstream DoH failure;
- timeout/cancellation;
- clean start/stop/restart;
- loopback-only binding;
- no plaintext fallback.

Do not require a public resolver or real VPN peer for unit/integration tests.

## Do not do yet

Do not change Windows adapter DNS settings, tunnel routes, or firewall/leak rules in this task.

## Verification and report

Run focused tests and `cmd /c build.bat`.

Create `docs/task-006-report.md` describing listener addresses, limits, lifecycle behavior, tests, build result, and remaining unverified items.

Commit, push, and continue to Task 007.