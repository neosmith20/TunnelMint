# Task 008 — DNS Leak Protection and Failure Recovery

## Goal

Make encrypted-DNS mode fail closed and recover cleanly from tunnel, DNS, network, sleep/wake, and process failures without leaking ordinary DNS or leaving Windows networking broken.

## Required behavior

When `DNS = https://...` is active:

1. Prevent Windows/app DNS from escaping to arbitrary plaintext DNS servers.
2. Allow plaintext bootstrap DNS only when required to locate the configured DoH endpoint, only to enabled bootstrap resolver IPs, and only for the bootstrap operation.
3. Do not silently fall back to plaintext DNS if DoH fails.
4. Keep DoH endpoint traffic routed through the active tunnel.
5. If the encrypted DNS path cannot be established or is lost, DNS should fail closed with a clear state/error rather than leaking.
6. Restore all TunnelMint-owned DNS settings, routes, listeners, and firewall state on normal disconnect.
7. Roll back partial changes when activation fails midway.
8. Re-establish state correctly after sleep/wake and meaningful network/interface changes.
9. Avoid deleting or rewriting unrelated user/system firewall rules, routes, or DNS settings.

Reuse the existing narrow Windows firewall/routing infrastructure where practical instead of shelling out to broad system commands.

## Bootstrap exception model

The bootstrap exception should be narrowly scoped and temporary. Document exactly:

- what process/path can use it;
- destination IP/port/protocol allowed;
- when it is installed;
- when it is removed;
- how re-bootstrap is handled safely.

If process-specific enforcement is not practical with the existing architecture, use the narrowest robust alternative and document the tradeoff. Do not pretend a broad exception is process-specific if it is not.

## Failure-injection tests

Cover at least:

- bootstrap failure;
- TLS validation failure;
- DoH HTTP failure;
- local DNS listener failure/port conflict;
- tunnel route installation failure;
- tunnel disconnect while queries are active;
- DoH endpoint becoming unreachable;
- process/service restart recovery;
- cleanup after partial activation;
- no plaintext fallback after any of the above.

On the disposable VM, use Windows-native observation tools where useful to verify that unexpected UDP/TCP port 53 traffic is not leaving the machine during steady encrypted-DNS operation. Do not require Wireshark if built-in tooling is sufficient.

## Plain DNS compatibility

Existing `DNS = <IP>` tunnels must retain their expected behavior. Leak protection for encrypted mode must not accidentally rewrite the semantics of plain-DNS configs.

## Verification and report

Run focused tests and `cmd /c build.bat`.

Create `docs/task-008-report.md` describing the leak model, firewall/routing rules, bootstrap exception, recovery paths, failure-injection results, build result, and anything still unverified.

Commit, push, and continue to Task 009.