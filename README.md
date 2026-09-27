# TunnelMint

**A simple, user-friendly VPN client with smarter DNS.**

TunnelMint is built around one core idea: **keep it simple and make the software do the work.**

The first release is being developed for Windows, with Android planned after the Windows client is stable and polished.

## Why TunnelMint?

TunnelMint aims to provide a fast, simple tunnel client while adding useful features that should not require users to understand the plumbing underneath.

The first major addition is transparent encrypted DNS support.

A normal DNS entry continues to work normally:

```ini
DNS = 1.1.1.1
```

TunnelMint will also understand a DNS-over-HTTPS endpoint directly:

```ini
DNS = https://dns.example.com/dns-query
```

TunnelMint detects the HTTPS endpoint, bootstraps it automatically, and sends encrypted DNS traffic through the active tunnel.

No extra enable switch. No unnecessary configuration maze.

## Planned v1 Features

- Windows-first client
- Simple tunnel import and management
- Standard DNS support using IP addresses
- DNS-over-HTTPS support using `DNS = https://...`
- Automatic DNS bootstrap using sensible built-in defaults
- User-configurable bootstrap resolvers in Settings
- DNS leak protection
- No silent fallback from encrypted DNS to plain DNS
- Tunnel status, handshake information, and basic diagnostics
- Simple Windows installer

## Bootstrap DNS

Encrypted DNS endpoints use hostnames, so TunnelMint may need a traditional DNS resolver briefly to locate the encrypted DNS endpoint before it becomes available.

TunnelMint is planned to ship with multiple bootstrap resolvers for reliability while allowing users to change, disable, reorder, or replace them in **Settings**.

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

## Tunnel Engine

TunnelMint uses WireGuard as the underlying tunnel technology while providing its own client experience and additional functionality around it.
