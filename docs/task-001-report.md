# Task 001 completion report

- Branch: `codex/windows-baseline`
- Upstream source: `https://github.com/WireGuard/wireguard-windows.git`
- Upstream commit: `6ece77bc487c8aa697e3c092197621c4f3e5ccb8`
- Build command: `cmd /c build.bat`
- Build result: passed; the upstream script completed x86, amd64, and arm64 builds.
- amd64 executable: `amd64\wireguard.exe`
- amd64 SHA-256: `29CD946F751F5628EB0D34D4F32D7D2DFF76F284AFA519E22A7A6A1C9DD6DD55`

## Tests

The official build completed successfully. The available Go tests were also
run with Go 1.27.1 from `.deps`:

```text
go test ./...
```

That run did not complete cleanly in this VM. Existing upstream tests reported
that `conf.TestStorage` could not access the Windows config store, and Go vet
reported existing `%w` directives passed to `testing.T.Errorf` in
`tunnel/winipcfg`. The run was stopped after it continued into environment-
dependent tests.

An additional bounded run disabled vet and excluded the config-store package:

```text
go test -vet=off -timeout 30s ./conf/dpapi ./driver/... ./elevate ./l18n ./manager ./ringlogger ./services ./tunnel/firewall ./tunnel/winipcfg ./ui/... ./updater/... ./version
```

`conf/dpapi`, `driver`, `elevate`, `l18n`, `manager`, `services`,
`tunnel/firewall`, `ui`, and packages without tests passed. The following
environment-dependent upstream tests failed or timed out:

- `ringlogger.TestWriteText` and `ringlogger.TestFollow` (the latter timed out
  after 30 seconds while emitting the test stream);
- `tunnel/winipcfg` interface and DNS tests (the VM has no matching test
  interface; Windows error 1168);
- `updater.TestUpdate` (updates are disabled for an unofficial build);
- `updater/winhttp.TestResponse` (the test host could not be resolved);
- `version.TestExtractCertificateExtension` (no expected certificate object in
  this development environment).

## Launch and remaining verification

Running `amd64\wireguard.exe` on the Windows development VM returned exit code
0. The VM already had WireGuard processes and a manager session, so this
verifies the executable can hand off to the existing upstream workflow. A
fresh visible UI launch, signed-driver installation, tunnel import, connect,
disconnect, handshake, and traffic checks remain unverified because performing
them would interfere with the existing development session and requires a
usable test tunnel configuration.
