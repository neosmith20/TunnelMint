# Task 014 — Owner Windows v1 Acceptance Evidence

## Test target

The initial Task 014 attempt tested Task 016 head
`ca4b76f8044e6fa98b48e9cbcccf2ef6cc25f010` from
`codex/task-014-owner-acceptance`. This focused startup-fix branch starts at
`4f468faef945dc9eaa1361ae051c1321eb1e57d0` (the current `main` at branch
creation) and contains the uncommitted fix described below.

Task 016 has not been merged into `main`, so its acceptance work remains
separate from this startup-fix branch.

## Environment and secret handling

- The current Windows VM is the sole acceptance machine.
- The owner-provided `C:\TunnelMint-Test\owner-acceptance.conf` was present.
  Its contents and credentials were not read, copied, logged, staged, or
  published.
- No `.conf` file, key material, or `TunnelMint-Test` artifact is included in
  this branch.

## Initial acceptance evidence and blocker

The initial elevated amd64 MSI installation, uninstall, and reinstall passed.
TunnelMint installed under `C:\Program Files\TunnelMint`, required Notices
were present, and official WireGuard remained installed at
`C:\Program Files\WireGuard\wireguard.exe`.

The installed client then failed before any UI or service command ran:

```text
panic:
crypto/internal/fips140.CAST(...)
crypto/internal/fips140/sha3.init.0()
```

The stack mapped to `.overlay/crypto/internal/fips140/fips140.go`, where the
disabled-FIPS `CAST` stub panicked. The failure prevented product-created
`TunnelMintManager` installation and deferred all tunnel, DNS, DoH, leak,
lifecycle, IPv6, and upgrade checks. A temporary direct Go diagnostic that
created a service was removed and is not counted as product acceptance.

## Focused source fix

This branch restores the upstream WireGuard overlay entry for
`crypto/rand/util.go`. Go 1.27 requires `crypto/rand.Prime` and
`crypto/rand.Int` through `crypto/rsa`, so the overlay supplies the small
compatible Go implementation using the caller-provided system-backed reader.
It does not import `crypto/internal/rand`, the FIPS DRBG tree, or disable TLS
verification.

Go 1.27 also initializes FIPS self-test packages from standard TLS
dependencies even when FIPS is disabled. The existing overlay sentinel made
every self-test panic. `CAST` and `PCT` now execute their supplied checks and
panic if a check fails; they are not no-ops. This preserves the self-test
failure behavior while allowing disabled-FIPS standard-library initialization.

The GitHub Actions Windows build now runs the actual built executable with
`amd64\tunnelmint.exe /update` after compilation.

## Verification completed on this branch

- `go test -vet=off ./bootstrap ./dnsproxy ./doh ./dohruntime ./product`
- `go test -vet=off ./tunnel ./tunnel/firewall ./manager ./ui`
- Focused `./conf` DNS/DoH parsing tests
- `build.bat` for x86, amd64, and arm64
- `installer\build.bat` for x86, amd64, and arm64
- Built `amd64\tunnelmint.exe /update`, which completed without the former
  startup panic and reported that updates are disabled until signing
  infrastructure is available.

## Installed-product retest status

The pre-fix installed executable remains present and still reproduces the
original FIPS startup panic. The new MSI was built successfully, but the VM's
Windows Installer custom action cannot obtain a primary token after the VM
restart: uninstall returns 1603 with Windows Installer error 1719 at
`EvaluateTunnelMintServices`. Normal `RunAs` elevation from the current agent
session also fails with `0xc0000142`.

The new MSI therefore could not be installed on this current agent session.
The remaining Task 014 owner-acceptance checks, including product manager/UI
startup, must resume from an Administrator terminal on this same VM once its
Windows Installer/UAC session can elevate again. No beta-readiness or
production-readiness claim is made.
