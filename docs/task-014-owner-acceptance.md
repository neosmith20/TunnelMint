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

The corrected current-main tree was revalidated at source commit
`ac715cc3aa2c964d3a2a188c529bed5a47401d31`; the following evidence-only
commit does not change the built source:

- `go test -vet=off ./bootstrap ./dnsproxy ./doh ./dohruntime ./product`
- `go test -vet=off ./tunnel ./tunnel/firewall ./manager ./ui`
- Focused `./conf` DNS/DoH parser tests, including `ResolveEndpointsWith`
- `cmd /c build.bat` for x86, amd64, and arm64
- `amd64\tunnelmint.exe /update`, which exited normally without a
  package-initialization panic and reported that updates are disabled until
  signing infrastructure is available
- `cmd /c installer\build.bat` for x86, amd64, and arm64

GitHub Actions validation remains pending until the corrected branch is
pushed and its pull request runs.

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

## Current-main retest attempt

Current `main` was pulled at merge commit
`767a3ffb3719abe20f65bb2bb0013650fa90a3d7` (PR #8). It contains the merged
startup fix. `cmd /c build.bat` and `cmd /c installer\build.bat` both
succeeded on this VM, and the rebuilt `amd64\tunnelmint.exe /update` exited
normally without the former FIPS/package-initialization panic.

The existing pre-fix installation could not be removed from the agent's
current process: the process has a non-elevated token, `RunAs` fails with
`0xc0000142`, and MSI uninstall returns 1603 because the installer custom
action `EvaluateTunnelMintServices` reports Windows Installer error 1719.
The installed fixed-MSI `/update` and normal manager/UI checks therefore did
not run. This is a Windows desktop-session access failure, not a TunnelMint
product result. The owner configuration was not read, copied, or modified.

`scripts/task014-elevated-setup.ps1` is prepared for one manual run from an
Administrator PowerShell window on this same VM. It performs only the blocked
uninstall/install/startup/service checks and writes sanitized local results;
it does not read or print the owner tunnel configuration.

## Elevated current-main installation and startup result

The owner ran the prepared elevated script on this VM on 2026-09-28. Its
sanitized result file records that it ran elevated, removed one old
`TunnelMint Development` product successfully (MSI exit 0), and installed the
current-main amd64 MSI successfully (MSI exit 0). The script then ran the
installed `C:\Program Files\TunnelMint\tunnelmint.exe /update` successfully
(exit 0), with no FIPS or package-initialization panic.

The script also observed `TunnelMintManager` installed with Automatic startup
and Running state, and a separately running TunnelMint UI window titled
`TunnelMint (unsigned build, no updates)`. This completes the fixed-MSI
installation, manager-service, update-smoke, and normal-startup checks on the
real acceptance VM. The installed product files and the existing official
WireGuard installation remain present as previously recorded.

The installed-program registry identifies the product as `TunnelMint
Development` version `0.1.0`, published by `TunnelMint contributors`; the
side-by-side package remains `WireGuard` version `1.1.1`, published by
`WireGuard LLC`. The common Start Menu contains separate `TunnelMint.lnk` and
`WireGuard.lnk` entries. The installed TunnelMint notices directory contains
`LICENSE`, `GPL-2.0.txt`, `THIRD_PARTY_NOTICES.md`, and
`WIREGUARD-COPYING`.

The owner configuration remains unread, unmodified, and absent from this
repository. Live import, connection, traffic, DNS, DoH, leak, lifecycle,
IPv6, and upgrade acceptance checks remain unexecuted and are not marked
passed.

## Plain-DNS pass and DoH TLS regression

On 2026-09-28, the owner imported the local acceptance configuration and
successfully connected it using a normal IP-address DNS setting. This is real
plain-DNS tunnel evidence, but it does not complete the remaining traffic,
route-restoration, or lifecycle checks.

Changing that test to DoH exposed a release-blocking TLS startup panic before
the encrypted resolver could make a request. The sanitized stack identified
`crypto/internal/fips140.Version` called by `crypto/tls` while constructing a
ClientHello. The disabled-FIPS compatibility overlay incorrectly panicked for
that ordinary TLS metadata query.

The pending focused fix returns Go's normal non-frozen FIPS metadata version
(`latest`) and makes disabled-mode metadata/service-indicator accessors
non-panicking, while retaining disabled FIPS mode and TLS certificate
validation. A deterministic ClientHello test passes both normally and with
the production overlay enabled; it is included in Windows CI with that
overlay. All three Windows client architectures rebuilt successfully, and the
rebuilt amd64 `/update` smoke test exits successfully.

The current restricted agent sandbox cannot relink a new MSI: WiX's ICE
validation cannot access Windows Installer services from that account. The
existing MSI therefore has not been retested with this fix. DoH live-machine
acceptance remains blocked until the fixed amd64 MSI is built and installed by
the owner from an Administrator desktop session, then the same DoH test is
repeated.

## DoH endpoint candidate failover regression

After the TLS-panic fix was installed, the owner repeated the DoH activation.
TLS no longer panicked, but the bounded DoH verification request timed out and
the tunnel correctly shut down rather than falling back to plaintext DNS. The
result is an activation failure, not a DoH acceptance pass.

The transport had a candidate failover defect: it passed the whole request
deadline to the first bootstrapped endpoint address. If that address is not
reachable over the configured tunnel family, a later reachable candidate never
gets a dial attempt. The focused pending fix divides the existing request
deadline among remaining candidates, while retaining the same overall timeout,
endpoint hostname, HTTPS certificate validation, and fail-closed behavior. A
deterministic test makes the first candidate time out and verifies that the
second is tried; both normal and production-overlay `doh` tests pass. The
three Windows client architectures rebuilt successfully.

This remains unverified on the live VM until a newly packaged and installed
amd64 MSI repeats the owner DoH activation successfully.

## DoH activation ordering regression

The owner installed the candidate-failover fix and repeated the same DoH
activation. It produced the same bounded transport timeout, confirming that
candidate failover was not the sole cause. No plaintext fallback occurred.

The remaining startup defect was an ordering race: the tunnel service began
DoH verification immediately after bringing the adapter up, while the adapter
watcher configured addresses and routes asynchronously. The verification could
therefore start before the selected tunnel family had an address and route.
The pending focused fix waits for every address family used by the tunnel to
finish its initial interface configuration before beginning encrypted-DNS
activation. It does not relax certificate validation, firewall policy, routing
ownership, bootstrap rules, or the request timeout.

Focused tunnel and DoH tests, including the production-overlay TLS smoke test,
pass. All three Windows client architectures rebuilt, and the rebuilt amd64
`/update` smoke test exits successfully. This fix still requires packaging,
installation, and a third live DoH retry before any DoH acceptance item can be
marked passed.

No beta-readiness or production-readiness claim is made.
