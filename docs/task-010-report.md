# Task 010 report — TunnelMint branding and user-facing UI

## Product identity and assets

The main window, tray, About dialog, diagnostics, and user-facing activation
messages use TunnelMint naming. The resource metadata keeps the required
WireGuard copyright and manufacturer attribution while identifying the product
as TunnelMint. A small original teal TunnelMint icon was added and is now used
for the application resource; the upstream WireGuard artwork remains available
only where attribution or underlying technology references require it.

## Main workflow and DNS status

The existing tabs keep the workflow to tunnels, logs, and Settings. The tunnel
view continues to expose import, selection, activation, editing, peer
handshake, and transfer information without adding a dashboard. Its DNS line
now states `Plain: ...` for ordinary IP DNS and `Encrypted (DoH): ...` for an
HTTPS endpoint, with initializing, ready-on-activation, and error text for the
encrypted path. Detailed bootstrap, TLS/HTTP, proxy, route, and firewall
errors remain in the existing activation error and log paths rather than
displaying query data or secrets in the main view.

## Verification

- `go test -vet=off ./conf -run 'Test(DNSParsing|DNSRoundTripPreservesDoH|DNSURLValidation|FromWgQuick)$'` — passed.
- `cmd /c build.bat` — passed for x86, amd64, and arm64.
- The amd64 development executable was launched from the VM. It exited before
  exposing a UI because this disposable environment has no installed
  TunnelMint manager service and the normal entry point delegates to the
  elevated service installer. No upstream WireGuard window was opened.

Interactive UI state, manager-service installation, and real tunnel activation
remain unverified in this VM because they require the Windows service and an
elevated desktop session. The compiled resource and code paths were verified by
the three-architecture build.

