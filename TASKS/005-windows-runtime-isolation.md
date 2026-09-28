# Task 005 — Windows Runtime Identity and Coexistence

## Goal

Make TunnelMint a distinct Windows application/runtime so it can coexist with an installed WireGuard Windows client without sharing manager state, tunnel services, application data, singleton/IPC identity, or updater behavior.

The Task 001 baseline currently still contains upstream identities such as `WireGuardManager`, `WireGuardTunnel$...`, `Program Files\WireGuard\Data`, and the upstream update service constants. Those must not remain TunnelMint's runtime identity.

## Required behavior

Audit the Windows client for runtime/product identity and isolate TunnelMint where appropriate. At minimum address:

1. Manager service name and display name.
2. Tunnel service name prefix and display name.
3. Application data/configuration/log root directory.
4. IPC/named-pipe/singleton/mutex/window identity used to locate or raise an existing manager/UI process.
5. Registry paths or product-specific settings.
6. Scheduled tasks or other persistent Windows identifiers, if any.
7. Product/user-agent strings where they identify the application rather than the tunnel protocol.
8. Upstream self-update behavior.

Use TunnelMint-specific names for TunnelMint-owned runtime resources. Preserve required upstream notices in source files.

## Coexistence

An installed upstream WireGuard client must be able to remain installed while TunnelMint is built and run.

TunnelMint must not:

- report the upstream manager as its own running instance;
- read or overwrite the upstream client's stored tunnel configurations;
- install/delete upstream manager or tunnel services;
- write TunnelMint logs/configuration into the upstream application data directory;
- uninstall, replace, or modify the upstream application;
- contact the upstream product's update service as though TunnelMint were that product.

## Updater

Until TunnelMint has its own signed release/update infrastructure, disable TunnelMint's automatic upstream-update path cleanly. Do not point TunnelMint at the upstream vendor's update service and do not invent an insecure replacement updater.

Keep the updater code separable so a TunnelMint update service can be added later.

## Scope note

This task is runtime isolation and minimal development identity, not the full visual redesign. It is acceptable for UI layout to remain baseline-like for now, but visible development builds should identify themselves as TunnelMint sufficiently to avoid operator confusion.

## Verification

On the disposable Windows VM, with the normal WireGuard client installed:

- verify the upstream app can remain installed;
- verify TunnelMint no longer collides with `WireGuardManager`;
- verify TunnelMint data goes to its own root;
- verify TunnelMint tunnel service names are distinct;
- verify launching TunnelMint does not merely hand off to the upstream manager;
- verify the normal Windows build still passes.

Do not import or expose real private tunnel credentials just for this test.

## Tests and report

Add focused tests for deterministic naming/path behavior where practical.

Create `docs/task-005-report.md` documenting every identity changed, every intentionally retained upstream identity and why, coexistence tests, updater behavior, build result, and anything still unverified.

Commit, push, and continue to Task 006.