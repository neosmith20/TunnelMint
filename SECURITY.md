# Security Policy

WireHush is networking and security-sensitive software. Vulnerability reports are welcome and should be handled privately whenever possible.

## Supported Versions

WireHush is currently in early development and has no production-supported release yet.

Once releases begin, security fixes will target the latest supported release unless otherwise stated in release notes.

## Reporting a Vulnerability

Please do **not** publish exploit details, private keys, credentials, tunnel configurations, or other sensitive information in a public issue.

Use GitHub's private vulnerability reporting feature for this repository when available.

If private vulnerability reporting is unavailable, open a minimal public issue stating that you need a private channel for a security report. Do not include technical exploit details in that issue.

Useful information in a private report includes:

- affected WireHush version or commit
- operating system and version
- clear reproduction steps
- expected and observed behavior
- security impact
- relevant logs with secrets removed
- any suggested remediation, if known

Please allow reasonable time for investigation and remediation before public disclosure.

## Security Expectations

WireHush should fail safely when security-sensitive behavior is uncertain. Production code must not disable certificate validation, silently leak DNS, expose tunnel secrets, or silently route around an expected tunnel path.
