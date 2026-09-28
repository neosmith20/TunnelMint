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
- License: mixed file-level licensing; the core WireGuard Windows source is
  generally MIT-marked, while the imported installer and fetcher sources carry
  GPL-2.0 SPDX markers. The applicable license is the marker in each file.
- Required notices: [WIREGUARD-COPYING](WIREGUARD-COPYING) for MIT-marked
  WireGuard components and [GPL-2.0.txt](GPL-2.0.txt) for the GPL-2.0-marked
  installer sources. The distributed package also includes this inventory.
- Upstream documentation: [wireguard-windows-upstream-README.md](docs/wireguard-windows-upstream-README.md)
- Build/provenance record: [docs/windows-baseline.md](docs/windows-baseline.md)

Upstream source files retain their original copyright and SPDX headers.
In particular, `installer/wireguard.wxs` and the sources under
`installer/fetcher/` are GPL-2.0-marked; they are not relicensed as
TunnelMint-owned code. TunnelMint's original code remains governed by the
project [LICENSE](LICENSE).
