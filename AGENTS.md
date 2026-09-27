# TunnelMint Agent Instructions

## Project Goal

TunnelMint is a simple, user-friendly VPN client built around WireGuard. The first target is Windows. Android is planned only after the Windows implementation is stable and polished.

The project follows one rule above all others: **KISS — Keep It Simple.**

Do not add complexity unless it is required for correctness, security, compatibility, or a clearly approved user-facing feature.

## Core Behavior

TunnelMint must preserve normal WireGuard behavior for ordinary configurations.

Examples:

```ini
DNS = 1.1.1.1
```

Must behave as normal DNS.

```ini
DNS = https://dns.example.com/dns-query
```

Must be recognized as DNS-over-HTTPS (DoH).

For DoH:

1. Bring up the WireGuard tunnel.
2. Automatically bootstrap the DoH hostname when needed.
3. Establish the DoH connection through the WireGuard tunnel.
4. Use the DoH endpoint for DNS queries.
5. Do not silently fall back to plain DNS if encrypted DNS fails.
6. Restore networking state cleanly when the tunnel stops.

Bootstrap resolvers must have sensible built-in defaults and be configurable from a Settings area. The user should not be required to configure bootstrap DNS for normal use.

## Product Principles

- Preserve compatibility with normal WireGuard tunnel configurations whenever possible.
- Do not change the WireGuard protocol.
- Do not reimplement WireGuard cryptography.
- Reuse the proven WireGuard Windows implementation and platform integration where appropriate.
- Keep encrypted DNS and related additions outside the WireGuard protocol itself.
- Prefer automatic behavior with safe defaults over exposing unnecessary configuration.
- Keep the UI simple and familiar, but make TunnelMint visually and functionally its own application.
- Avoid unnecessary dependencies.
- Avoid unnecessary background services, frameworks, runtimes, and telemetry.
- Never add tracking, analytics, advertising, or account requirements unless explicitly approved by the project owner.

## Initial Windows Scope

The initial Windows milestone is limited to:

- Existing WireGuard tunnel functionality
- Importing and editing standard tunnel configurations
- Plain DNS using IP addresses
- DoH using `DNS = https://...`
- Automatic bootstrap DNS
- Configurable bootstrap resolver settings
- DNS leak protection
- Clean tunnel connect/disconnect behavior
- Basic tunnel status and diagnostics
- Simple installer

Do not expand scope into ad blocking, DNS filtering, mesh networking, IDS/IPS, commercial VPN accounts, or unrelated networking features unless explicitly requested.

## Development Workflow

Before modifying code:

1. Read this file and `ROADMAP.md`.
2. Inspect the existing implementation before proposing replacements.
3. Prefer the smallest safe change that accomplishes the task.
4. Preserve upstream notices and licensing when importing or modifying upstream WireGuard code.

For each meaningful change:

1. Build the relevant target.
2. Run existing tests.
3. Add focused tests for new parsing, DNS, bootstrap, routing, or failure behavior where practical.
4. Verify normal WireGuard behavior has not regressed.
5. Document any behavior that could not be tested.

Do not claim a feature works unless it was actually built and tested or the limitation is clearly stated.

## Git Discipline

- Keep commits focused and descriptive.
- Do not mix unrelated cleanup with feature work.
- Do not rewrite shared history unless explicitly instructed.
- Do not force-push without explicit approval.
- Use feature branches for significant work.
- One coding agent should own a branch at a time to avoid concurrent-edit conflicts.

## Security and Networking Rules

- DoH certificate validation must not be bypassed in production code.
- Do not disable TLS verification as a convenience fix.
- Do not silently leak DNS outside the requested encrypted resolver path.
- Do not silently route around the VPN when tunnel-only behavior is expected.
- Fail closed for encrypted DNS unless an explicit fallback setting exists and the user enabled it.
- Treat configuration values, keys, endpoints, and logs as potentially sensitive.
- Never commit private keys, credentials, tokens, or real user tunnel configurations.

## Current Priority

The first engineering objective is to establish a clean, reproducible Windows baseline from the WireGuard Windows codebase before adding features.

After that, implement DoH support incrementally:

1. Config parsing and model support
2. DoH transport
3. Bootstrap resolution
4. Tunnel-bound routing
5. Windows DNS integration
6. Failure/leak handling
7. Bootstrap Settings UI
8. Branding/UI polish
9. Installer and release packaging
