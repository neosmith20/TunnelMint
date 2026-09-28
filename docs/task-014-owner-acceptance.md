# Task 014 — Owner Windows v1 Acceptance Evidence

## Initial Task 014 acceptance evidence

Task 016 final pre-acceptance cleanup was merged to `main` in pull request #6
(`177628537da325c58a22bd4905236780e3e152c0`). Task 014 real-machine
acceptance then began from Task 016 head
`ca4b76f8044e6fa98b48e9cbcccf2ef6cc25f010` on branch
`codex/task-014-owner-acceptance`.

Task 014 remains incomplete. Plain-DNS and Cloudflare DoH activation now have
live evidence, while the owner endpoint, leak, lifecycle, and beta-readiness
checks remain unresolved or unexecuted.

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
repository. Live traffic, leak, lifecycle, IPv6, and upgrade acceptance checks
remain incomplete and are not marked passed. Plain-DNS and Cloudflare DoH
evidence is recorded below; it does not establish acceptance for every
endpoint or every remaining check.

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

## DoH transport diagnostic evidence

The third DoH retry produced the same bounded transport timeout. No additional
code change is claimed from that result. The next evidence step is the
sanitized `scripts/task014-doh-connectivity.ps1` probe while the previously
working plain-DNS tunnel is active. It records only service/adapter counts,
route counts, TCP/HTTPS success categories, and tunnel adapter byte deltas;
it does not read or report the owner configuration, tunnel name, endpoint IPs,
or keys.

The owner completed the initial probe with the known-good plain-DNS tunnel
active. It observed one running TunnelMint tunnel service, one active
TunnelMint adapter, and four adapter routes. TCP port 443 to the configured
DoH host connected, and the direct HTTPS DoH POST returned HTTP 200. The
adapter counters increased during both probes, which is evidence that the
plain-tunnel test exchanged traffic while the checks ran.

That initial helper wrote the counter increases under unintended `sent` and
`received` result properties instead of its documented
`tunnelAdapter*BytesDelta` properties. It did not log configuration contents,
keys, tunnel aliases, or endpoint addresses. The helper is corrected and now
also records only aggregate route ownership for resolved endpoint candidates;
the owner repeated it with the same plain-DNS tunnel active. All four resolved
endpoint candidates selected a TunnelMint-adapter route (zero selected a
non-tunnel route or lacked a route). TCP 443 connected and the direct HTTPS
DoH POST again returned HTTP 200, with positive sent and received tunnel-byte
deltas for both checks. This establishes the plain-tunnel route and endpoint
baseline but does not establish that the TunnelMint service's DoH startup path
is working.

The next focused DoH retry will add only sanitized service-side transport
events to the tunnel log: candidate order, IPv4/IPv6 family, connection
success, and a coarse failure category. It will not log endpoint addresses,
configuration data, credentials, or keys. No change to routing, timeout,
TLS validation, DNS fallback, or firewall policy is claimed from that
diagnostic. After that retry,
`scripts/task014-doh-log-extract.ps1` exports only those fixed-format events
to a local sanitized result file for this acceptance work; it never copies or
prints the raw product log.

## Live DoH endpoint results

On 2026-09-28, the owner connected the current installed build using the
Cloudflare DoH endpoint. This confirms the former TLS panic and bounded
transport timeout are no longer present for that endpoint. It is live DoH
activation evidence, but it does not complete the remaining DNS leak, traffic,
disconnect, route-restoration, or upgrade checks.

The owner then tried an external DoH endpoint and the tunnel failed closed with
`DoH endpoint returned HTTP status 302 Found`. The server answered with an HTTP
redirect instead of a successful DoH response. TunnelMint intentionally rejects
redirects so it does not silently change the configured endpoint identity or
TLS policy. The external service must expose the exact HTTPS DoH URL that
returns a 2xx `application/dns-message` response, or the configuration must
use the redirect's final HTTPS URL after verifying it is the intended resolver.
This result does not by itself identify a TunnelMint transport defect.

The corrected redirect probe then captured the comparison chain without
publishing the tokenized URL. The original POST returned `302` to a Cloudflare
Access login endpoint. Following that `Location` while preserving the POST
returned `404`. The automatic-follow comparison returned `200`, but its
content type was `text/html`, confirming it was an access/login response rather
than `application/dns-message`. The earlier direct diagnostic therefore had a
false-positive HTTP-status check: it followed redirects and did not require a
DoH content type. The connectivity helper now requires
`application/dns-message` before marking a direct probe successful.

After the owner corrected the endpoint, the redirect probe returned HTTP 200
directly with `application/dns-message`, no `Location` header, and a one-hop
manual chain. This clears the external redirect diagnosis for the corrected
endpoint. TunnelMint activation with that endpoint is still the next live
acceptance step; this probe alone does not establish service startup,
DNS-leak protection, traffic, or lifecycle behavior.

Before accepting the custom endpoint, the owner must run the elevated
`scripts/task014-doh-leak-acceptance.ps1` probe while that tunnel is active.
It checks loopback-only DNS on the TunnelMint adapter, captures port-53
traffic during a fresh DNS query, observes service-owned HTTPS connections to
the endpoint's resolved addresses, and requires a fresh query to fail after
the endpoint is broken. Use `-ManualServerBreak` after arranging a real
server-side outage or access revocation; the local firewall mode cannot end an
already-established HTTPS session. The temporary outbound block (when used)
and packet capture are removed in script cleanup. The result stores only an
endpoint SHA-256 fingerprint and aggregate counts; exact HTTPS path/client-ID
use still requires the DoH server's request log because TLS hides the HTTP
path from Windows packet capture.

The first leak-acceptance attempt did not pass. The TunnelMint adapter had one
loopback DNS address and no non-loopback address, and the service had an
endpoint HTTPS connection. However, the capture recorded two non-loopback
port-53 events during the baseline query. The temporary local endpoint block
was created and removed successfully, but the subsequent fresh query still
returned a response over an existing HTTPS connection. This is insufficient
evidence for either zero DNS leakage or fail-closed behavior; no custom DoH
acceptance item is marked passed.

The owner then reran the probe with `-ManualServerBreak` and disabled or
revoked the custom DoH endpoint during the scripted pause. The TunnelMint
service and adapter were running, the TunnelMint adapter still had only its
loopback DNS address, and the baseline query received a response while an
endpoint HTTPS connection was observed. After the server-side break, the
fresh query received no DNS response, which is the expected fail-closed
signal. The same result recorded eight non-loopback port-53 events, however,
and the system-wide DNS summary still contained two non-loopback configured
addresses. The system-wide count is retained as context because physical
adapters may keep their ordinary DNS settings; the acceptance decision relies
on the TunnelMint adapter's loopback-only setting and observed wire traffic.
Because the zero-port-53 condition is not met, this rerun does not accept
custom DoH. Exact endpoint path and client-ID use still require the sanitized
DoH server request log.

The port-53 counts above came from the initial conservative parser, which
treated every non-loopback pktmon record as possible leakage. The acceptance
helper now parses pktmon's per-packet IP, direction, wire-type, and drop fields
and reports outbound wire candidates separately from unclassified records.
The result must be rerun with that helper before deciding whether the observed
records were blocked WFP attempts or packets that reached a physical
interface; either an outbound wire candidate or incomplete classification
keeps the zero-leak check failed.

The next rerun used that parser and found zero outbound wire candidates, but
six baseline and eight broken records remained unclassified. The fail-closed
DNS result therefore remains useful, but the capture is still insufficient to
accept the zero-leak condition. The helper now narrows matching to pktmon
records that explicitly identify TCP or UDP port 53 and treats an empty,
readable capture as complete evidence; one more run is required.

That follow-up run recorded zero port-53 packet records, zero outbound wire
candidates, and no unclassified records during both queries. The TunnelMint
adapter remained loopback-only and the baseline query received a DNS response.
It did not pass the failure-closed check: the query after the purported manual
endpoint break also received a DNS response. The server-side change therefore
did not interrupt requests from the running client, or did not take effect for
the configured endpoint. No custom DoH acceptance item is marked passed.

The owner then corrected the test so the exact same external DoH URL was used
in the TunnelMint `DNS =` setting and as the probe endpoint. With that matched
configuration, the elevated run found a running service and adapter,
loopback-only DNS on the TunnelMint adapter, and a successful baseline DNS
response while an endpoint HTTPS connection was observed. Both the baseline
and broken phases recorded zero port-53 packet records, zero outbound wire
candidates, and zero unclassified records. After the owner disabled the same
external endpoint during the scripted pause, the fresh DNS query received no
response. This is real-machine evidence that the tested encrypted-DNS path
fails closed without observed plaintext DNS traffic. Exact endpoint
path/client-ID preservation still requires the sanitized resolver request log.

No beta-readiness or production-readiness claim is made.
