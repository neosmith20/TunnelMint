# Third-Party Notices

TunnelMint may incorporate or depend on third-party software. Each third-party component remains subject to its own copyright, license terms, and required notices.

## Notice Policy

When third-party source code, binaries, libraries, drivers, or other components are added to TunnelMint:

- preserve all copyright notices required by that component's license;
- preserve the component's license text when required;
- do not represent third-party code as original TunnelMint code;
- record the component, source, license, and any required attribution in this file or an accompanying notice file;
- ensure release packages include all notices required for distributed third-party components.

## WireGuard

TunnelMint uses WireGuard as its underlying VPN tunnel technology. Any WireGuard-derived code or distributed WireGuard components included in TunnelMint remain subject to their applicable upstream copyright and license notices.

The Windows tunnel foundation is incorporated from the official
`wireguard-windows` repository at commit
`6ece77bc487c8aa697e3c092197621c4f3e5ccb8`:

- Source: https://github.com/WireGuard/wireguard-windows
- License: MIT
- Required notice: [WIREGUARD-COPYING](WIREGUARD-COPYING)
- Upstream documentation: [wireguard-windows-upstream-README.md](docs/wireguard-windows-upstream-README.md)
- Build/provenance record: [docs/windows-baseline.md](docs/windows-baseline.md)

Upstream source files retain their original copyright and SPDX headers.
