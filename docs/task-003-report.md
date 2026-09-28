# Task 003 completion report

- Branch: `codex/windows-v1`
- Commit: `1f0673417dff9a84cda994bdabeac292ef73c041`
- Files changed: `doh/client.go`, `doh/client_test.go`, and `ROADMAP.md`

## Transport design

The new `doh` package exposes a small `Client` that accepts one validated
HTTPS endpoint and sends raw DNS wire-format queries with HTTP POST. It sets
both `Content-Type` and `Accept` to `application/dns-message`, returns the raw
response bytes, and rejects non-2xx responses or incompatible response media
types. The endpoint URL, including provider-specific paths, is used unchanged.

The default request timeout is 10 seconds. A request is also bounded by the
caller context, so cancellation remains available to later lifecycle code.
The default response limit is 1 MiB and the accepted DNS query limit is 65,535
bytes. Responses are read through a limit reader and oversized bodies fail
closed.

Redirects are rejected entirely. This keeps the configured HTTPS endpoint and
TLS identity explicit and prevents an HTTPS-to-HTTP downgrade. The production
client uses Go's normal system certificate verification; no TLS bypass or
insecure production switch was added. Tests inject the `httptest` trusted
client only to trust their local test certificate.

The client accepts an injected HTTP client so the bootstrap and tunnel tasks
can later supply a custom dial path without moving endpoint identity or TLS
validation into this package.

## Verification

Focused tests passed:

```text
go test -vet=off -v -timeout 15s ./doh
```

Coverage includes successful POST headers and raw bytes, endpoint path
preservation, non-2xx status, incompatible content type, timeout,
cancellation, untrusted TLS rejection, trusted local TLS success, redirect
rejection, endpoint validation, and oversized response rejection.

The official Windows build passed for x86, amd64, and arm64:

```text
cmd /c build.bat
```

Automatic bootstrap DNS, endpoint IP caching, Windows DNS integration, local
DNS listeners, tunnel routing, firewall/leak protection, settings UI, and
runtime behavior remain intentionally unimplemented for later tasks.
