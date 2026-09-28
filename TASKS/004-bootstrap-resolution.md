# Task 004 — Bootstrap Resolution and DoH Endpoint Dialing

## Goal

Resolve a configured DoH endpoint hostname without depending on the system DNS path that TunnelMint will later replace, then connect to the resolved endpoint while preserving correct HTTPS/TLS identity.

## Required behavior

Implement a small bootstrap subsystem that:

1. Accepts a validated DoH endpoint URL.
2. Skips bootstrap when the endpoint host is already an IP literal.
3. Resolves endpoint hostnames using an ordered list of configured **IP-address bootstrap resolvers**.
4. Ships with these sensible defaults in the model/config layer:
   - `1.1.1.1`, `1.0.0.1`
   - `9.9.9.9`, `149.112.112.112`
   - `8.8.8.8`, `8.8.4.4`
5. Uses the bootstrap resolvers only for resolving the DoH endpoint hostname.
6. Never silently falls back to the operating system resolver when encrypted DNS mode is selected.
7. Returns usable IPv4 and IPv6 endpoint candidates where available.
8. Applies bounded timeouts and ordered failover.
9. Uses a small bounded endpoint-address cache where useful; document the cache lifetime and invalidation rules. Do not invent permanent DNS pinning.
10. Allows cache invalidation on tunnel/network lifecycle changes.

Use the simplest reliable implementation available in the existing dependency set. Do not add a heavyweight DNS dependency just for convenience.

## HTTPS dialing

Integrate bootstrap results with the Task 003 DoH transport so the HTTP connection can dial a resolved IP address **without losing the original endpoint hostname**.

TLS ServerName/SNI and HTTP host semantics must continue to use the configured DoH hostname. Certificate validation must therefore validate the configured hostname, not the bootstrap IP.

The DoH transport must not perform a second system-DNS lookup behind the bootstrap subsystem.

## Tests

Tests must be local and deterministic. Cover at least:

- ordered bootstrap resolver failover;
- A/IPv4 resolution;
- AAAA/IPv6 resolution where the test environment supports it;
- endpoint IP-literal bypass;
- no system-resolver fallback;
- timeout behavior;
- endpoint cache hit/invalidation behavior;
- custom dial to a bootstrapped IP while TLS validates the original hostname;
- correct DoH URL path preservation after bootstrap;
- total bootstrap failure causes encrypted DNS setup to fail closed.

A local fake DNS server and local TLS server are preferred over public Internet dependencies.

## Do not do yet

Do not configure Windows adapter DNS, start a local DNS listener, change tunnel routes, add firewall/leak rules, or build Settings UI in this task.

## Verification and report

Run focused tests and `cmd /c build.bat`.

Create `docs/task-004-report.md` with design, cache behavior, resolver failover behavior, tests, build result, and remaining unverified items.

Commit, push, and continue to Task 005 under `TASKS/000-windows-v1-execution-queue.md`.