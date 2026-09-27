# Contributing to TunnelMint

Thanks for helping improve TunnelMint.

TunnelMint is intentionally focused on a small, reliable user experience. Contributions should follow the project's KISS principle: solve the problem with the smallest safe change that preserves compatibility and security.

## Before Contributing

Please:

1. Read `README.md`, `AGENTS.md`, `ROADMAP.md`, `SECURITY.md`, and `LICENSE`.
2. Search existing issues and pull requests before starting duplicate work.
3. Keep changes focused. Large unrelated refactors should be discussed before implementation.
4. Do not submit private keys, credentials, real tunnel configurations, personal data, or other secrets.
5. Security vulnerabilities must be reported through the process in `SECURITY.md`, not disclosed in a public issue or pull request before remediation.

## Development Expectations

Contributions should:

- preserve ordinary WireGuard tunnel behavior unless a change is explicitly required;
- avoid unnecessary dependencies and background services;
- preserve required third-party copyright and license notices;
- include focused tests when practical;
- build successfully for the affected target;
- avoid disabling TLS validation, leak protections, or other security controls to make a test pass;
- avoid unrelated formatting or cleanup changes in feature commits.

## Pull Requests

A pull request should clearly describe:

- what changed;
- why the change is needed;
- how it was tested;
- any known limitations or untested behavior;
- any new dependency, license, or third-party code introduced.

Keep commits reasonably focused and descriptive.

## Contribution License

By intentionally submitting code, documentation, artwork, tests, or other material to TunnelMint for inclusion in the project (a "Contribution"), you represent that you have the right to submit that Contribution.

You grant the TunnelMint project and its maintainers a perpetual, worldwide, non-exclusive, irrevocable, royalty-free license to use, reproduce, modify, prepare derivative works of, publicly display, publicly perform, distribute, sublicense, relicense, and commercially use or license your Contribution, in source or binary form, as part of TunnelMint or related TunnelMint distributions.

You understand that TunnelMint may be distributed under the license in `LICENSE`, under another license in the future, or as part of a commercial offering, and that you are not entitled to royalties or other compensation solely because your Contribution is used.

You retain ownership of any copyright you hold in your original Contribution.

Submitting a pull request or other Contribution after these terms are published indicates acceptance of these contribution terms.

If you cannot grant these rights, do not submit the Contribution for inclusion.

## Third-Party Material

Do not submit code or assets copied from another project unless its license permits the proposed use and all required notices are included. Identify third-party material clearly in the pull request.

See `THIRD_PARTY_NOTICES.md` for the project's third-party notice policy.
