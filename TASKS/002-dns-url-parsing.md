# Task 002 — DNS URL Parsing and Model Support

## Goal

Extend TunnelMint's configuration parser and model so the existing `DNS =` setting can represent either normal DNS server IP addresses or an HTTPS DNS-over-HTTPS endpoint, without changing tunnel or DNS runtime behavior yet.

This task is parser/model work only. Do not implement DoH transport in this task.

## Starting point

Create a new focused feature branch from the completed Windows baseline branch:

- baseline branch: `codex/windows-baseline`
- suggested new branch: `codex/dns-url-parsing`

Read `AGENTS.md`, `README.md`, `ROADMAP.md`, and `docs/task-001-report.md` before changing code.

## Required behavior

Existing configurations must continue to behave exactly as before.

Normal DNS examples must remain valid:

```ini
DNS = 1.1.1.1
```

```ini
DNS = 1.1.1.1, 1.0.0.1
```

IPv6 DNS addresses already supported by the baseline must remain supported.

TunnelMint must additionally accept an HTTPS DoH endpoint directly in the same setting:

```ini
DNS = https://dns.example.com/dns-query
```

Paths and provider-specific identifiers must be preserved, for example:

```ini
DNS = https://dns.example.com/dns-query/client-id
```

The parser/model must preserve enough information to distinguish:

- normal IP-based DNS;
- HTTPS DoH endpoint.

Do not convert a DoH URL into a DNS search suffix or silently discard it.

Configuration write/serialize behavior must preserve the DoH endpoint so a parse → write → parse round trip does not lose or alter the URL.

## Validation

For this milestone:

- `https://` is the only encrypted-DNS URL scheme to accept.
- Plain IP addresses continue using existing behavior.
- Reject malformed HTTPS URLs with a clear parse error.
- Reject unsupported URL schemes such as `http://`, `tls://`, or `quic://` for now rather than guessing.
- Do not disable or weaken existing validation to make HTTPS values pass.

Keep validation minimal and standards-based. Do not impose arbitrary hostname/path restrictions that would reject legitimate provider-specific DoH URLs.

## Tests

Add focused tests covering at least:

1. IPv4 DNS continues to parse unchanged.
2. Multiple normal IP DNS servers continue to parse unchanged.
3. Existing IPv6 DNS behavior remains unchanged.
4. A valid `https://.../dns-query` endpoint parses as DoH.
5. A DoH URL containing an additional path/client identifier is preserved.
6. Parse → serialize → parse preserves the DoH endpoint.
7. Malformed HTTPS URLs fail cleanly.
8. Unsupported schemes fail cleanly.
9. Existing parser tests continue to pass except for already-documented environment-dependent baseline issues.

## Do not do yet

Do **not**:

- send DNS-over-HTTPS requests;
- start a local DNS listener;
- change Windows DNS settings;
- add bootstrap resolver logic;
- alter routing or firewall rules;
- add a Settings UI;
- rebrand or redesign the UI;
- change tunnel behavior;
- add DoT, DoQ, DNSCrypt, filtering, or unrelated DNS features.

## Completion criteria

Task 002 is complete when:

- ordinary `DNS = IP` configurations retain existing behavior;
- `DNS = https://...` is represented explicitly and safely in the configuration model;
- DoH URLs survive config round trips unchanged;
- focused parser/model tests pass;
- the normal Windows build still succeeds;
- no runtime DNS/tunnel behavior has been added or changed;
- the feature branch is pushed to GitHub.

## Completion report

Add `docs/task-002-report.md` containing:

- branch name;
- commit SHA;
- files changed;
- model/parser design used;
- build command and result;
- tests run and results;
- known limitations;
- explicit confirmation that no DoH network transport was implemented in this task.
