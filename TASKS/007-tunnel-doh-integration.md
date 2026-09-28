# Task 007 — Tunnel and DoH Runtime Integration

## Goal

Connect the encrypted-DNS stack to the active TunnelMint tunnel so `DNS = https://...` works through the tunnel rather than over an arbitrary host route.

## Required lifecycle

For an encrypted-DNS tunnel, implement and document a deterministic activation sequence broadly equivalent to:

1. Bring up the tunnel interface using the normal tunnel configuration.
2. Resolve/bootstrap the configured DoH endpoint as needed.
3. Ensure the selected DoH endpoint address(es) have a route through the active tunnel, including split-tunnel configurations.
4. Establish/verify the DoH transport through that route.
5. Start the local DNS proxy.
6. Configure the tunnel/Windows DNS path to use the local proxy.
7. Mark encrypted DNS ready only after the above succeeds.

On teardown, reverse owned state cleanly.

## Routing requirements

- Full-tunnel configurations must continue to work normally.
- Split-tunnel configurations must still route the configured DoH endpoint through the active tunnel even when its public IP is not otherwise in `AllowedIPs`.
- Use Windows/tunnel interface routing primitives already present in the codebase where practical.
- Prefer host routes (`/32` for IPv4 and `/128` for IPv6) for DoH endpoint candidates rather than broad route expansion.
- Track routes TunnelMint adds so only TunnelMint-owned routes are removed later.
- Re-bootstrap/re-route safely when endpoint addresses change.
- Avoid route recursion. If the selected DoH address is unsafe/impossible to route through the tunnel (for example because doing so would break the tunnel's own transport path), fail with a clear error rather than creating a loop.

## Windows DNS integration

When encrypted DNS is selected, Windows should send DNS requests to the local TunnelMint DNS proxy rather than directly to the DoH provider or a plaintext resolver.

Do not change ordinary `DNS = <IP>` runtime behavior.

Persist enough pre-change state to restore Windows networking correctly on disconnect/failure.

## Tests

Add deterministic tests for route-selection/lifecycle logic where possible, plus disposable-VM integration checks for:

- encrypted-DNS activation ordering;
- host-route add/remove;
- full-tunnel case;
- split-tunnel case;
- IPv4;
- IPv6 when available;
- endpoint-address refresh;
- activation failure rolls back partial state;
- plain-DNS tunnels remain on the existing path.

Do not use real user private keys in repository fixtures.

## Verification and report

Run focused tests and `cmd /c build.bat`.

Create `docs/task-007-report.md` documenting lifecycle order, route ownership, Windows DNS changes, rollback behavior, tests, build result, and remaining unverified items.

Commit, push, and continue to Task 008.