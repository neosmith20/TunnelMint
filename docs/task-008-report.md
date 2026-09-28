# Task 008 report — DNS leak protection and failure recovery

## Leak model

Encrypted DNS configurations now enable the existing narrow Windows Filtering
Platform session before adapter configuration. DNS blocking applies in both
full-tunnel and split-tunnel encrypted modes, so ordinary UDP/TCP port 53
traffic cannot escape to arbitrary resolvers. Full-tunnel traffic remains
restricted by the existing WireGuard rules; split-tunnel traffic keeps its
existing routing semantics while DNS is blocked except for the explicitly
allowed bootstrap addresses and the local proxy.

The DoH endpoint is reached over HTTPS port 443 through host routes installed
for the resolved endpoint addresses. The local DNS proxy forwards only the
configured DoH wire-format queries and has no plaintext fallback.

## Bootstrap exception

The exception is installed by the TunnelMint tunnel service before the adapter
is configured and remains only until bootstrap resolution, tunnel-bound DoH
verification, proxy startup, and local DNS configuration complete. It permits
UDP and TCP port 53 to the six built-in bootstrap resolver addresses, plus the
loopback addresses used by the local proxy. The current WFP primitives do not
provide a process-path condition for this operation, so the exception is
IP-scoped for the short bootstrap window; another local process could use
those resolver addresses during that window. This is the narrowest robust
exception available in the existing firewall layer and is removed immediately
after verification by replacing TunnelMint's dynamic WFP session with one that
allows only loopback DNS.

If bootstrap or verification must be repeated, the lifecycle installs only
the newly required endpoint routes, verifies the replacement path, and then
reconfigures the WFP session again. A failed replacement keeps a restrictive
DNS-blocking session in place so a firewall setup error does not reopen plain
DNS.

## Recovery and cleanup

Activation is transactional. Bootstrap, endpoint routes, transport
verification, the proxy listener, local DNS, and final firewall tightening are
rolled back in reverse order on failure. Closing the encrypted-DNS session
restores the prior adapter DNS servers and search list, closes the local
listener, and deletes only routes owned by TunnelMint. Tunnel teardown then
destroys the adapter and closes the dynamic WFP session. The existing watcher
path remains responsible for disconnect cleanup and reconfiguration events.

Replacing the WFP session necessarily creates a very short transition between
the old and new dynamic sessions. The replacement path fails closed for DNS if
the new session cannot be installed.

## Failure-injection coverage

Focused tests cover bootstrap errors, route ownership and installation
failures, transport/DoH verification failures, proxy/listener failures,
finalization failures, reverse-order cleanup, endpoint refresh, proxy timeout
and malformed-query handling, and the absence of a plaintext fallback. The
proxy tests also cover concurrent UDP/TCP queries and unreachable DoH
behavior.

## Verification

- `go test -vet=off ./dohruntime ./dnsproxy ./doh ./bootstrap` — passed.
- `cmd /c build.bat` — passed for x86, amd64, and arm64.

Direct packet observation during sleep/wake, process restart, active-query
disconnect, and a real tunnel session was not available in this disposable VM.
Those scenarios remain the next Windows integration verification step; the
pure lifecycle and proxy failure paths are covered by the focused tests above.

