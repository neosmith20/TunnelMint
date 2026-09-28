# Task 004 completion report

- Branch: `codex/windows-v1`
- Commit: `efdbefe048437190c494826213c9d7941cf91a77`
- Files changed: `bootstrap/resolver.go`, `bootstrap/resolver_test.go`,
  `doh/client.go`, and `ROADMAP.md`

## Design

The `bootstrap` package validates HTTPS DoH endpoints, bypasses DNS when the
endpoint host is already an IP literal, and otherwise resolves the hostname
through an ordered list of explicit IP bootstrap resolvers. The built-in list
is `1.1.1.1`, `1.0.0.1`, `9.9.9.9`, `149.112.112.112`, `8.8.8.8`, and
`8.8.4.4`. Each lookup uses Go's resolver with `PreferGo` and a custom dial
function that connects only to the selected resolver IP, so there is no system
resolver fallback.

The resolver returns both IPv4 and IPv6 candidates when available, applies a
five-second bounded operation timeout, and fails closed when all configured
resolvers fail. Its cache is limited to 64 host entries with a five-minute
expiry. `InvalidateEndpoint` and `InvalidateAll` are available for tunnel and
network lifecycle changes; cache keys are endpoint hostnames and do not pin
addresses permanently.

`doh.NewHTTPClientForAddresses` supplies the resolved candidates to a cloned
HTTP transport. It dials only those IPs while the original endpoint URL stays
in the request, preserving the HTTP Host header, TLS ServerName/SNI, and
certificate validation for the configured hostname. Proxy use is disabled for
this endpoint path so the connection cannot be diverted through an unrelated
system proxy.

## Verification

Focused tests passed:

```text
go test -vet=off ./doh ./bootstrap
```

Coverage includes ordered resolver failover, IPv4 and IPv6 candidates,
IP-literal bypass, no system-resolver fallback, timeout behavior, cache hit and
invalidation, custom IP dialing with hostname TLS validation, and DoH path
preservation. Tests use injected deterministic lookup functions and local TLS
servers only.

The official Windows build passed for x86, amd64, and arm64:

```text
cmd /c build.bat
```

Windows adapter DNS changes, local DNS proxy startup, tunnel route ownership,
and full runtime activation remain intentionally deferred to later tasks.
