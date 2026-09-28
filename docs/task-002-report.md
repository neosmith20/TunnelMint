# Task 002 completion report

- Branch: `codex/dns-url-parsing`
- Implementation commits: `9da41e50512e0580f6b4ef6a993332c476092a0a`,
  `a9deed4b4da66c0de4a52fcb61e4606dc7d54b3e`
- Files changed: `conf/config.go`, `conf/parser.go`, `conf/parser_test.go`,
  `conf/writer.go`, `ui/confview.go`, and `ROADMAP.md`

## Design

The configuration model keeps existing `Interface.DNS` and `Interface.DNSSearch`
fields unchanged and adds `Interface.DNSOverHTTPS []string` for explicit HTTPS
endpoints. The parser first accepts normal IP addresses with the existing
`netip` behavior, recognizes URL-shaped values with `net/url`, and accepts only
the `https` scheme with a non-empty host. It rejects malformed URLs and
unsupported schemes with parse errors. The original endpoint string is stored
unchanged, so paths and provider-specific identifiers survive serialization.

The writer emits stored DoH endpoints in the existing `DNS =` setting, and the
configuration view displays them. Tunnel and Windows DNS runtime consumers were
left unchanged; no DoH transport or runtime DNS behavior was added.

## Verification

Focused parser tests passed:

```text
go test -vet=off ./conf -run 'Test(DNS|FromWgQuick|Comments|ParseEndpoint)'
```

The tests cover IPv4, multiple IPv4 values, IPv6, DoH paths and client
identifiers, parse/serialize/parse round trips, malformed HTTPS URLs,
unsupported schemes, and existing DNS search suffix behavior.

The official Windows build also passed for all upstream targets:

```text
cmd /c build.bat
```

No DoH network transport, bootstrap resolution, DNS listener, Windows DNS
changes, routing, firewall changes, or tunnel behavior changes were
implemented in this task.
