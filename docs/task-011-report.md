# Task 011 report — Windows installer and release packaging

## Installer identity

The development release is version `0.1.0`. The generated packages are:

- `installer/dist/tunnelmint-x86-0.1.0.msi`
- `installer/dist/tunnelmint-amd64-0.1.0.msi`
- `installer/dist/tunnelmint-arm64-0.1.0.msi`

The MSI product is named **TunnelMint Development**, installs to
`Program Files\TunnelMint`, creates a TunnelMint Start Menu shortcut, and
ships `tunnelmint.exe` plus the `wg.exe` command-line utility. Each architecture
uses a new TunnelMint-specific UpgradeCode, so Windows Installer cannot treat
it as an upgrade or replacement for upstream WireGuard. Service enumeration,
process cleanup, driver removal, and data cleanup are scoped to the
TunnelMintManager/TunnelMintTunnel$ identities and TunnelMint-owned paths.

The installer keeps automatic update checks disabled because TunnelMint does
not yet have owned signed release infrastructure.

## Notices and signing

The MSI contains the built executable and required runtime resources. The
repository distribution set retains `LICENSE`, `THIRD_PARTY_NOTICES.md`, and
`WIREGUARD-COPYING`; the installer build does not remove or relabel those
upstream notices. No TunnelMint code-signing certificate is available, so
these are explicitly unsigned development artifacts. Signature verification
was not bypassed.

## Verification

- `cmd /c build.bat` — passed for x86, amd64, and arm64, producing
  `x86/tunnelmint.exe`, `amd64/tunnelmint.exe`, and `arm64/tunnelmint.exe`.
- `cmd /c installer\build.bat` — passed for all three MSI architectures.
- A fresh amd64 MSI install was attempted with launch disabled. Windows
  Installer identified the package as TunnelMint Development 0.1.0 and the
  TunnelMint install directory, but returned error 1925 because this VM shell
  has no administrator privileges for a per-machine install. The transaction
  rolled back; no TunnelMint directory, service, or uninstall entry remained.

Upgrade, uninstall/reinstall, launch-after-install, coexistence with an
installed upstream client, and post-uninstall adapter cleanup could not be
executed without an elevated disposable Windows desktop and an installed
upstream client. The package database and custom-action build verify the
separate identities and scoped cleanup paths; those VM behaviors remain for
Task 012 acceptance.

