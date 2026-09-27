# Task 001 — Windows Baseline

## Goal

Establish a clean, reproducible Windows development baseline for TunnelMint before any encrypted-DNS feature work begins.

## Required reading

Before changing anything, read:

- `README.md`
- `AGENTS.md`
- `ROADMAP.md`
- `LICENSE`
- `CONTRIBUTING.md`
- `SECURITY.md`
- `THIRD_PARTY_NOTICES.md`

## Scope

Use the current official WireGuard Windows codebase as the tunnel/client foundation for TunnelMint.

For this task only:

1. Integrate the Windows source cleanly into the TunnelMint repository.
2. Preserve all required upstream copyright and license notices.
3. Record the exact upstream commit used.
4. Build the unmodified Windows client successfully for `amd64`.
5. Verify that the resulting application launches on the Windows development VM.
6. Verify that the installed signed tunnel driver can be used by the development build if required by the existing upstream development workflow.
7. Run available relevant tests.
8. Document the exact repeatable build and launch procedure for future agents.
9. Commit the completed baseline on a focused feature branch.

## Do not do yet

Do **not**:

- add DoH or encrypted DNS support;
- modify DNS parsing behavior;
- change tunnel behavior;
- redesign or rebrand the UI;
- add unrelated dependencies;
- remove or rewrite required third-party notices;
- claim functionality works unless it was actually built and tested.

## Completion criteria

This task is complete only when:

- the Windows source foundation is present in the TunnelMint repo;
- the exact upstream baseline is documented;
- an `amd64` build completes successfully;
- the resulting executable launches successfully;
- relevant baseline tests have been run or any untestable items are clearly documented;
- the build procedure is repeatable from a clean development checkout.

When finished, report:

- branch name;
- commit SHA;
- upstream source commit used;
- build command(s);
- tests run and results;
- path to the built executable;
- anything that remains unverified.
