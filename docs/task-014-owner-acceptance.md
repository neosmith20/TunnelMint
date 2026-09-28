# Task 014 — Owner Windows v1 Acceptance Evidence

Test target: Task 016 head `ca4b76f8044e6fa98b48e9cbcccf2ef6cc25f010`, tested from branch `codex/task-014-owner-acceptance`.

Task 014 is blocked by a reproducible startup failure. No tunnel, DNS, DoH, leak, lifecycle, or beta-readiness check is marked passed.

## Environment and secret handling

- The current Windows VM was used, as required by the latest Task 014 instructions.
- Normal Windows elevation works on this VM; an elevated probe reported `admin=True`.
- The owner-provided `C:\TunnelMint-Test\owner-acceptance.conf` is present. Its contents and credentials were not read, copied, logged, staged, or published.
- Before this report commit, staged-file verification found no `.conf` file and no `TunnelMint-Test` artifact in the repository.

## Checks completed

- Elevated amd64 MSI installation succeeded (`msiexec` exit 0).
- Installed files are under `C:\Program Files\TunnelMint`, including `tunnelmint.exe`, `wg.exe`, and the required `Notices` files.
- Official WireGuard remains installed at `C:\Program Files\WireGuard\wireguard.exe`.
- Elevated uninstall succeeded; the TunnelMint directory was removed while the official WireGuard executable remained.
- Elevated reinstall succeeded; TunnelMint binaries/notices and official WireGuard were present afterward.

## Blocking failure

The installed client cannot start. Running the built amd64 executable with `/update` reproduces a startup panic before any UI or service command runs:

```text
panic:
crypto/internal/fips140.CAST(...)
crypto/internal/fips140/sha3.init.0()
```

The stack maps to the repository overlay at `.overlay/crypto/internal/fips140/fips140.go:26`, where the disabled-FIPS `CAST` stub panics. The installed executable exits before TunnelMintManager installation, so no product-created `TunnelMintManager` service exists and no UI window appears. A temporary direct Go diagnostic confirmed the manager installer can create a service when invoked from test code; that diagnostic service was deleted and all temporary files were removed. It is not counted as a product acceptance pass.

## Checks not run

Because the installed client cannot start, these checks remain pending: manager/service lifecycle through the product UI, tunnel import, real handshake and traffic, plain DNS, full/split-tunnel DoH, bootstrap Settings, packet capture and leak checks, failure-closed cases, disconnect cleanup, reboot/sleep/network transitions, IPv6, upstream coexistence through both clients, and package upgrade behavior.

The failed startup must become a focused source fix and be retested before Task 014 can continue. No beta-readiness or production-readiness claim is made.
