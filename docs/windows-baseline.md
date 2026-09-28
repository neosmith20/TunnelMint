# Windows baseline

TunnelMint's Windows tunnel foundation is the official WireGuard for Windows
source at commit `6ece77bc487c8aa697e3c092197621c4f3e5ccb8` from
`https://github.com/WireGuard/wireguard-windows.git` (the `master` branch
HEAD queried on 2026-09-27). The source is incorporated without behavioral
changes for this baseline task.

The upstream notices are preserved in [WIREGUARD-COPYING](../WIREGUARD-COPYING)
and [the upstream README](wireguard-windows-upstream-README.md). The source
files retain their upstream copyright and SPDX headers. The required upstream
source commit is recorded here so a clean checkout can be reproduced exactly.

## Build

Requirements are Windows 10 64-bit or Windows Server 2019, Git for Windows,
and network access for the pinned dependencies fetched by `build.bat`.

From a clean checkout at the repository root, run:

```text
build.bat
```

The script downloads and verifies its pinned Go, LLVM-MinGW, ImageMagick,
Make, wireguard-tools, and WireGuardNT inputs into `.deps`, then builds the
supported architectures. The amd64 application is written to:

```text
amd64\wireguard.exe
```

To run the development build, launch `amd64\wireguard.exe`. The upstream
workflow requires a Microsoft-signed WireGuard driver; install an official
WireGuard for Windows release first when the development machine does not
already have the driver. The development executable can then install its
manager service and show the UI.

## Verification record

The build and test results for this checkout are recorded in the Task 001
completion report. Driver installation, tunnel import, connect/disconnect,
handshake, and traffic tests require a Windows development VM with a usable
WireGuard configuration and are reported separately when unavailable.
