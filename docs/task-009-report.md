# Task 009 report — bootstrap DNS settings

## UI behavior

The TunnelMint Settings tab shows the ordered built-in bootstrap resolver list
and marks each entry as enabled or disabled. The selected entry can be toggled,
moved up or down, and custom entries can be added or removed. Restore defaults
returns the six built-in resolvers to their original enabled order. The page
explains that changes apply on the next tunnel connection, so it does not
attempt fragile live reload of an active tunnel.

## Persistence and validation

Settings are stored as `bootstrap-dns.json` below the TunnelMint-owned data
root established by Task 005. Writes use a private temporary file followed by
an atomic rename. Every entry must be an IPv4 or IPv6 literal; hostnames,
unspecified addresses, multicast addresses, duplicates, malformed JSON, and an
empty enabled set are rejected. Built-in entries may be disabled but cannot be
removed. Custom entries are enabled when added and can be removed later.

Tunnel activation loads this file through the bootstrap package and passes the
enabled addresses to the ordered failover resolver. A missing file uses the
built-in defaults, so normal users do not need to open Settings.

## Security behavior

The Settings page exposes no plaintext-DNS fallback switch. DoH certificate
validation and the encrypted-DNS failure-closed behavior remain unchanged.

## Verification

- Bootstrap settings serialization, persistence, ordering, validation,
  restore-defaults, and activation handoff tests passed.
- `go test -vet=off ./bootstrap ./dohruntime ./dnsproxy ./doh` — passed.
- `cmd /c build.bat` — passed for x86, amd64, and arm64.

The Settings page was compiled but not interactively exercised on a desktop
session in this VM. File permissions and the next-activation handoff are
covered by code paths and focused tests; a non-administrator Windows UI run
against the installed data-root ACL remains unverified.

