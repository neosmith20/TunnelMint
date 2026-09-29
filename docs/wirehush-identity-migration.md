# WireHush Identity Migration

WireHush is the public product name beginning with this release. The executable, application metadata, installer display name, Start Menu shortcut, UI, tray, documentation, and release artifacts use `WireHush` and `wirehush.exe`.

## Deliberately Retained Compatibility Identifiers

The following identifiers remain unchanged because they are persisted by existing Windows installations. They are internal implementation details and are not shown as the product name.

| Identifier | Retained value | Reason |
| --- | --- | --- |
| MSI UpgradeCode | Existing per-architecture UpgradeCode values | Windows Installer uses these to recognize a prior TunnelMint installation as an upgrade. ProductCode remains generated per MSI, as before. |
| MSI component GUIDs | Existing component GUIDs | Preserves Windows Installer component ownership through the executable rename. |
| Manager service key | `TunnelMintManager` | Existing manager service and upgrade custom actions can find and stop it. Its display name is `WireHush Manager`. |
| Tunnel service prefix | `TunnelMintTunnel$` | Existing active tunnel services remain discoverable and do not collide with official WireGuard services. Their display names use `WireHush Tunnel:`. |
| Installed application and protected data directory | `C:\Program Files\TunnelMint` and `Data` beneath it | The legacy install location is retained so an in-place upgrade continues to find the executable, encrypted configuration, and bootstrap settings. |
| Administrative registry key | `HKLM\Software\TunnelMint` | Existing installation state remains removable by the installer. |
| Adapter GUID namespace | `Deterministic TunnelMint Windows GUID v1` and `Fixed TunnelMint Windows GUID v1` | Changing either label would create different adapter GUIDs for existing tunnels. |

The transient manager window class and HTTP user agent now use WireHush. WireHush retains separate service, registry, data, and adapter identities from official WireGuard, so both applications can coexist.

## Upgrade and Uninstall Behavior

A major upgrade is recognized through the unchanged UpgradeCode. The new MSI installs `wirehush.exe` in the retained `Program Files\TunnelMint` location so configurations remain discoverable. Its Add/Remove Programs entry and Start Menu shortcut are WireHush. Uninstall removes WireHush-owned installed files and uses the retained internal identifiers only for its own service, registry, data, and adapter cleanup. It does not target official WireGuard services, paths, registry keys, or adapters.