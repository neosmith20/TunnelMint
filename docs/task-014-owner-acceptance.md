# Task 014 — Owner Windows v1 Acceptance Evidence

## Initial Task 014 acceptance evidence

Task 016 final pre-acceptance cleanup was merged to `main` in pull request #6
(`177628537da325c58a22bd4905236780e3e152c0`). Task 014 real-machine
acceptance then began from Task 016 head
`ca4b76f8044e6fa98b48e9cbcccf2ef6cc25f010` on branch
`codex/task-014-owner-acceptance`.

Task 014 is blocked by a reproducible startup failure. No tunnel, DNS, DoH,
leak, lifecycle, or beta-readiness check is marked passed.

## Environment and secret handling

- The current Windows VM was used, as required by the Task 014 instructions.
- Normal Windows elevation worked in the original acceptance session; an
  elevated probe reported `admin=True`.
- The owner-provided `C:\TunnelMint-Test\owner-acceptance.conf` is present.
  Its contents and credentials were not read, copied, logged, staged, or
  published.
- Before every commit for this work, staged-file verification excludes `.conf`
  files, key material, and `TunnelMint-Test` artifacts.

## Initial checks completed

- Elevated amd64 MSI installation succeeded (`msiexec` exit 0).
- Installed files are under `C:\Program Files\TunnelMint`, including
  `tunnelmint.exe`, `wg.exe`, and the required Notices files.
- Official WireGuard remains installed at
  `C:\Program Files\WireGuard\wireguard.exe`.
- Elevated uninstall succeeded; the TunnelMint directory was removed while the
  official WireGuard executable remained.
- Elevated reinstall succeeded; TunnelMint binaries/notices and official
  WireGuard were present afterward.

## Initial blocking failure

The installed client could not start. Running the built amd64 executable with
`/update` reproduced a startup panic before any UI or service command ran:

```text
panic:
crypto/internal/fips140.CAST(...)
crypto/internal/fips140/sha3.init.0()
```

The stack mapped to the repository overlay at
`.overlay/crypto/internal/fips140/fips140.go`, where the disabled-FIPS `CAST`
stub panicked. The installed executable exited before TunnelMintManager
installation, so no product-created `TunnelMintManager` service existed and
no UI window appeared. A temporary direct Go diagnostic confirmed the manager
installer can create a service when invoked from test code; that diagnostic
service was deleted and all temporary files were removed. It is not counted as
a product acceptance pass.

## Focused startup-fix retest

PR #7 merged this initial evidence to `main`
(`6a526bfbebf51cf795974dcf8a27b00b618c104c`). The focused startup-fix branch
is rebased on that current `main`, preserving the merged Task 016 behavior:
persistent DoH host-route ownership across recovery, fail-closed adapter
reinitialization, retryable failed-route cleanup, the DNS/DoH parser
restrictions, and the expanded focused test coverage.

The fix restores the upstream WireGuard overlay entry for `crypto/rand/util.go`.
Go 1.27 requires `crypto/rand.Prime` and `crypto/rand.Int` through
`crypto/rsa`, so the overlay supplies the small compatible Go implementation
using the caller-provided system-backed reader. It does not import
`crypto/internal/rand`, the FIPS DRBG tree, or weaken TLS verification.

The disabled-FIPS `CAST` and `PCT` compatibility functions now match Go
1.27.1 semantics: because this overlay defines `Enabled = false`, they return
without executing supplied FIPS self-tests. The FIPS module and DRBG tree stay
disabled.

GitHub Actions runs `amd64\tunnelmint.exe /update` immediately after the
Windows client build to catch future package-initialization failures.

The corrected current-main tree must be fully revalidated before recording
built-executable, MSI, or CI results here.

## Installed-product retest status

The old pre-fix installed executable remains present and reproduces the
original FIPS startup panic. After a VM reboot, the agent was moved to the
wrong non-interactive session (`tunnelmint\codexsandboxoffline`): normal
`RunAs` elevation fails with `0xc0000142`, and MSI uninstall returns 1603 with
Windows Installer error 1719 at `EvaluateTunnelMintServices`.

These session failures are not a TunnelMint product acceptance failure. Task
014 must resume on this same VM from the original `neosmith20` desktop session
launched from an Administrator terminal. At that point, install the corrected
amd64 MSI, verify the installed `/update` smoke test and normal manager/UI
startup, and continue the remaining real-machine checks without marking them
passed until they are actually performed.

No beta-readiness or production-readiness claim is made.
