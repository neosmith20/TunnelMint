# Task 011 — Windows Installer and Release Packaging

## Goal

Produce a normal TunnelMint Windows installer and release layout that installs, upgrades, and uninstalls TunnelMint without colliding with or modifying an installed WireGuard client.

## Installer identity

Create TunnelMint-specific installer/product identity as appropriate, including:

- product/display name;
- package/output filenames;
- install directory;
- Windows product/package identifiers where required;
- shortcuts/start-menu entries;
- uninstall entry;
- service references;
- application executable name.

Do not reuse installer identity in a way that causes Windows Installer to treat TunnelMint as an upgrade/replacement for the upstream WireGuard application.

## Coexistence and safety

Verify install/uninstall behavior on the disposable VM with the upstream client installed.

TunnelMint install/upgrade/uninstall must not:

- remove or alter upstream application files;
- remove or alter upstream services;
- remove or alter upstream stored tunnel configurations;
- change the upstream application's installer registration;
- leave TunnelMint-owned services/routes/firewall/DNS state behind after uninstall.

## Update behavior

Do not ship a functional automatic updater pointing at infrastructure TunnelMint does not own.

If TunnelMint update infrastructure/signing is not available, keep automatic update checks disabled and document that clearly. Do not weaken signature verification or redirect the inherited updater to an unsigned source.

## Licensing/notices

Ensure distributed artifacts include or install the required TunnelMint license information and all required third-party/upstream notices, including `WIREGUARD-COPYING` / equivalent notices for incorporated upstream portions.

Do not claim upstream code is exclusively licensed under TunnelMint's PolyForm terms.

## Signing

If no TunnelMint code-signing certificate is available, produce an unsigned **development/release-candidate** build and document that limitation. Do not bypass Windows security mechanisms to hide the lack of a signature.

## Release version

Use an obviously pre-release/development version appropriate for owner testing. Do not label the build stable/production-ready before Task 012 and owner acceptance.

## Verification

Test as far as practical on the disposable VM:

- fresh install;
- launch after install;
- upgrade over an earlier TunnelMint development install if practical;
- uninstall;
- reinstall;
- coexistence with upstream WireGuard installation;
- service cleanup;
- TunnelMint data handling on uninstall as intentionally designed;
- no secrets/private configs included in artifacts.

Run `cmd /c build.bat` and the installer build.

## Report

Create `docs/task-011-report.md` with installer identity, output artifact paths, version, signing status, install/upgrade/uninstall test results, coexistence results, notices included, and remaining limitations.

Commit, push, and continue to Task 012.