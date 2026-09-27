# TunnelMint — Claude Code Notes

Read and follow [`AGENTS.md`](AGENTS.md) before making changes.

TunnelMint is a Windows-first WireGuard-based VPN client focused on simplicity and transparent encrypted DNS support.

Key product behavior:

- `DNS = 1.1.1.1` → normal DNS behavior.
- `DNS = https://dns.example.com/dns-query` → DNS-over-HTTPS through the active WireGuard tunnel.
- Bootstrap DNS should be automatic by default and configurable in Settings.
- Encrypted DNS must not silently fall back to plain DNS.
- Preserve standard WireGuard behavior and compatibility wherever possible.
- Keep the UI simple and avoid unnecessary dependencies or scope expansion.

Before coding, inspect the existing implementation and use the smallest safe change. Build and test before claiming completion.
