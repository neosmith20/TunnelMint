# Contributing to TunnelMint

Thanks for helping improve TunnelMint.

TunnelMint is intentionally focused on a small, reliable user experience. Contributions should follow the project's KISS principle: solve the problem with the smallest safe change that preserves compatibility and security.

## Licensing and contribution ownership

TunnelMint is source-available under the [PolyForm Noncommercial License 1.0.0](LICENSE).

- **Bug reports, feature suggestions, logs, testing results, and general feedback** may be submitted freely without transferring ownership.
- **Code, documentation, artwork, tests, or any other copyrightable contribution** will not be merged until the contributor has explicitly accepted the [Contributor Copyright Assignment](CONTRIBUTOR_COPYRIGHT_ASSIGNMENT.md).
- Opening a pull request, issue, or other submission by itself is **not** treated as a copyright assignment.
- The assignment must be accepted through a designated signed or electronic acceptance process before the contribution is merged.
- Once accepted, copyright ownership in the covered Contribution is transferred to the TunnelMint Project Owner as described in the assignment agreement.
- If a Contribution may be owned in whole or in part by an employer or another entity, the contributor must disclose that before merge. Authorization from that entity may be required.
- TunnelMint may decline any contribution whose ownership, provenance, licensing, or third-party rights are unclear.

Until an acceptance workflow is published, external contributors are welcome to open issues, submit testing results, suggest changes, and discuss implementation, but copyrightable contributions intended for merge should not be merged.

## Before Contributing

Please:

1. Read `README.md`, `AGENTS.md`, `ROADMAP.md`, `SECURITY.md`, `LICENSE`, and `CONTRIBUTOR_COPYRIGHT_ASSIGNMENT.md`.
2. Search existing issues and pull requests before starting duplicate work.
3. For anything beyond a small fix, open an issue first to discuss the approach.
4. Keep changes focused. Large unrelated refactors should be discussed before implementation.
5. Do not submit private keys, credentials, real tunnel configurations, personal data, or other secrets.
6. Security vulnerabilities must be reported through the process in `SECURITY.md`, not disclosed publicly before remediation.

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

## Third-Party Material

Do not submit code or assets copied from another project unless its license permits the proposed use and all required notices are included. Identify third-party material clearly in the pull request.

See `THIRD_PARTY_NOTICES.md` for the project's third-party notice policy.
