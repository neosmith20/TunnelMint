# TunnelMint

**A simple, user-friendly WireGuard-based VPN client with smarter DNS.**

TunnelMint is an independent VPN client built around WireGuard with one core goal: **keep it simple and make the software do the work.**

The first release is being developed for Windows, with Android planned after the Windows client is stable and polished.

## Why TunnelMint?

TunnelMint aims to preserve a simple tunnel workflow while adding quality-of-life features that should not require users to understand the plumbing underneath.

The first major addition is transparent encrypted DNS support.

A normal DNS entry should continue to work normally:

```ini
DNS = 1.1.1.1
```

TunnelMint will also understand a DoH endpoint directly:

```ini
DNS = https://dns.example.com/dns-query
```

TunnelMint will detect the HTTPS endpoint, bootstrap it automatically, and send the encrypted DNS traffic through the active tunnel.

No separate `EncryptedDNS=true` switch. No unnecessary configuration maze.

## Planned v1 Features

- Windows-first client
- Familiar, simple tunnel workflow
- Import existing WireGuard tunnel configurations
- Standard DNS support using IP addresses
- DNS-over-HTTPS support using `DNS = https://...`
- Automatic DNS bootstrap using sensible built-in defaults
- User-configurable bootstrap resolvers in Settings
- DNS leak protection
- No silent fallback from encrypted DNS to plain DNS
- Tunnel status, handshake information, and basic diagnostics
- Simple Windows installer

## Bootstrap DNS

Encrypted DNS endpoints use hostnames, so TunnelMint may need a traditional DNS resolver briefly to locate the DoH endpoint before encrypted DNS is available.

TunnelMint is planned to ship with multiple bootstrap resolvers for reliability, while allowing users to change, disable, reorder, or replace them in **Settings**.

Bootstrap DNS is only intended to locate the encrypted DNS endpoint. Normal DNS queries should then use the configured encrypted resolver.

## Project Philosophy

**KISS — Keep It Simple.**

The user should be able to:

1. Download TunnelMint.
2. Install it.
3. Import a tunnel.
4. Connect.

Advanced networking details belong inside the software whenever they can be handled safely and automatically.

## Project Status

> **Very early development.**

TunnelMint is not currently ready for production use. The initial Windows client and encrypted DNS functionality have not yet been implemented.

See [ROADMAP.md](ROADMAP.md) for the initial development plan.

## Security

Please report security issues according to [SECURITY.md](SECURITY.md). Do not publish private keys, credentials, tunnel configurations, or exploit details in public issues.

## Licensing

TunnelMint's original project code is licensed under the terms in [LICENSE](LICENSE). Third-party components remain subject to their own licenses and required notices; those notices will be maintained in [THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md) as dependencies are incorporated.

## WireGuard

TunnelMint uses WireGuard as its VPN tunnel technology while providing its own client experience and additional functionality around it.

TunnelMint is an independent project and is not affiliated with or endorsed by the WireGuard project.
