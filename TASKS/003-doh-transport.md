# Task 003 — DNS-over-HTTPS Transport Core

## Goal

Implement the small, isolated DNS-over-HTTPS transport primitive that TunnelMint will later connect to bootstrap logic and the active tunnel.

This task is deliberately **not** Windows DNS integration and **not** tunnel routing. Build and test the DoH engine as a standalone component first.

## Starting point

Start from the current `main` branch after Task 002 has been merged.

Create a fresh focused branch, preferably:

`codex/doh-transport`

Before changing anything, read `AGENTS.md`, `ROADMAP.md`, `SECURITY.md`, `THIRD_PARTY_NOTICES.md`, and `docs/task-002-report.md`.

## Required behavior

Implement a minimal reusable DoH transport component that:

1. Accepts a validated HTTPS DoH endpoint URL.
2. Accepts a DNS query as raw DNS wire-format bytes.
3. Sends the query with HTTP `POST`.
4. Uses `Content-Type: application/dns-message`.
5. Sends `Accept: application/dns-message`.
6. Returns the successful response body as raw DNS wire-format bytes.
7. Uses normal system TLS certificate validation in production code.
8. Supports request cancellation and a bounded timeout.
9. Treats non-success HTTP responses as errors.
10. Rejects responses that are clearly not DNS messages when the server supplies an incompatible `Content-Type`.
11. Uses a reasonable bounded response size so a malicious or broken resolver cannot return an unbounded body.
12. Does not silently retry through plaintext DNS or downgrade HTTPS.

Keep this component small and separable so later tasks can control how the endpoint hostname is resolved and which network path the HTTP connection uses.

## TLS and testability

Production behavior must **never** disable certificate verification.

Tests may use a locally trusted test TLS server or dependency injection of an HTTP client/transport. Do not add an insecure production switch such as `InsecureSkipVerify` just to make tests easy.

Do not pin public CA certificates or hard-code a specific DoH provider.

## Redirect behavior

Do not allow redirects to downgrade from HTTPS to HTTP.

Prefer the simplest secure behavior. If redirects are supported at all, the final request must remain HTTPS and the behavior must be covered by tests. It is acceptable to reject redirects entirely in this first implementation if that keeps the security model clearer.

## Tests

Add focused automated tests that cover at least:

- successful POST request/response using DNS wire-format bytes;
- required request headers;
- endpoint/path preservation, including provider-specific path components;
- non-2xx HTTP status;
- incompatible response content type;
- request timeout or cancellation;
- TLS certificate validation failure using an untrusted test certificate;
- successful TLS test using a test client/root that explicitly trusts the test server;
- HTTPS-to-HTTP redirect rejection if redirect handling exists;
- oversized response rejection.

The tests must not depend on a public DoH provider or the public Internet.

## Build verification

Run the focused tests for the new DoH component and any directly affected packages.

Then run the normal Windows build:

`cmd /c build.bat`

Do not claim success unless the build completes.

## Do not do yet

Do **not** implement any of the following in Task 003:

- automatic bootstrap DNS;
- built-in bootstrap resolver lists;
- endpoint IP caching;
- Windows local DNS listener/proxy;
- changes to Windows adapter DNS configuration;
- routing the DoH socket through WireGuard;
- DNS leak protection/firewall changes;
- settings UI;
- TunnelMint visual branding;
- DoT or DoQ;
- ad blocking or DNS filtering.

Do not change ordinary `DNS = <IP>` runtime behavior.

## Documentation

Create `docs/task-003-report.md` containing:

- branch name;
- commit SHA(s);
- files/packages changed;
- transport design summary;
- timeout and response-size limits chosen and why;
- redirect policy;
- TLS validation behavior;
- tests run and results;
- Windows build result;
- anything still unverified.

Update `ROADMAP.md` only for Phase 2 items that were genuinely completed and tested by this task.

## Completion criteria

Task 003 is complete when TunnelMint has a tested, provider-neutral DoH POST transport that securely exchanges raw DNS messages over HTTPS, validates TLS normally, fails closed on transport/protocol errors, builds successfully on Windows, and has **no** bootstrap, tunnel-routing, or Windows DNS integration mixed into it.

Commit and push the completed work to the feature branch. Do not merge it into `main` yourself unless explicitly instructed.
