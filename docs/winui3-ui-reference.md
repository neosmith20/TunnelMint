# WireHush WinUI 3 UI Reference

Status: owner-approved visual source of truth. This document is an implementation contract, not a suggestion.

Visual target: `UI Mockup.png` supplied by the owner, reference canvas 1536x1024 at 100% scale.

## 1. Non-negotiable platform choice

- Frontend: C# on .NET 10 LTS.
- UI: WinUI 3 on Windows App SDK stable channel only.
- Current implementation target: Windows App SDK 2.5.1 stable.
- Existing Go networking/service code remains authoritative for tunnel, DNS, DoH, bootstrap, WFP, routing, adapter, service, and installer behavior.
- The WinUI frontend is non-elevated.
- The Go manager/service owns privileged work.
- `github.com/lxn/walk` is not used by the new frontend.
- The existing Walk UI remains only as temporary migration/reference code until the WinUI frontend passes acceptance; do not extend it.

## 2. Visual contract

At 1536x1024 / 100% scaling the reference image defines these primary regions:

- Window: 1536x1024 reference canvas.
- Integrated header/title area: 108 DIP high.
- Left Connections rail: 342 DIP wide below header.
- Main content begins at X=343 DIP.
- Main content outer padding: approximately 16 DIP.
- Main selected-connection summary card begins around X=359, Y=123 and spans nearly full main width.
- Main section navigation sits immediately below the summary card.
- Overview uses a two-column card layout with a narrow consistent gap.
- Bottom card row aligns to the same two-column grid.

The layout must scale with DPI and window size; the pixel values above are the 100% reference, not hard-coded physical pixels.

### Color system

Use a restrained dark Fluent palette matching the image:

- Header base: approximately #131B22.
- Rail base: approximately #0C1218.
- Main canvas/card base: approximately #0F171E / #10181E.
- Dividers/borders: approximately #1A232A to #243642.
- Selected tunnel surface: approximately #09334B with cyan border/accent.
- Primary cyan: approximately #04C1F3.
- Healthy green: approximately #17DB55.
- Primary text: near-white, approximately #F3F7FA.
- Muted text: blue-gray, approximately #8FA8B9.

Do not introduce neon glows, giant gradients, gamer RGB, or arbitrary accent colors.

## 3. Window chrome and header

Use WinUI custom title-bar support, not a separate old-style Win32 title bar.

- `ExtendsContentIntoTitleBar = true`.
- Preserve native minimize, maximize/restore, and close caption buttons.
- Caption button background stays transparent/dark and follows theme.
- Define a real drag region that does not overlap clickable controls.
- Header height: 108 DIP reference.

Header left:

- WireHush icon from the approved brand kit.
- Icon visual size approximately 96x64 DIP reference.
- Product name `WireHush`, large semibold/bold.
- Subtitle `Private network control`, smaller regular text.

Header right:

- `Settings`, `Log`, `Help` as compact icon+text commands.
- They are not giant equal-width navigation slabs.
- Hit targets minimum 40x40 DIP.
- Use Fluent icons from the system glyph set where practical; do not rasterize ordinary UI icons.

## 4. Connections rail

Use a Grid with fixed width 342 DIP at the reference size.

Structure:

- Heading `Connections`.
- Search box approximately 300 DIP wide and 48 DIP high.
- Scrollable tunnel list fills remaining height.
- Bottom action area remains anchored to the bottom and contains:
  - Add Tunnel
  - Import Tunnel(s)
  - Delete

Tunnel row contract:

- Approx. 76 DIP high.
- Left status dot.
- Tunnel name on first line, semibold.
- State on second line.
- Ellipsis menu at far right.
- Connected state uses green.
- Disconnected uses muted gray-blue.
- Selected row uses the dark teal selected surface plus a 3-4 DIP cyan left accent.
- Selection must not depend only on color; maintain clear contrast/border/focus indication.

Delete is disabled when no applicable tunnel is selected.

Do not duplicate Add/Import actions in the center of the normal dashboard. The rail is the normal action location.

## 5. Selected connection summary

Top summary card contains:

Top line:

- Tunnel name, e.g. `Home`.
- Small edit glyph next to name when editing is permitted.
- State row: status dot + `Connected` / `Disconnected` / transitional state.
- Connected duration when reliable.
- Primary Connect/Disconnect button aligned top-right.

Second line: four equal summary tiles:

1. VPN IP (IPv4)
2. VPN IP (IPv6)
3. Endpoint
4. Latest Handshake

Each tile:

- small icon tile on left,
- muted label,
- larger value,
- no fake value; use `Not Assigned`, `Not Configured`, or `No handshake yet` when unknown.

## 6. Section navigation

Immediately below summary card:

- Overview
- Network
- DNS
- Peer
- Allowed IPs

Use a tab-like command strip matching the mockup, not a stock legacy TabControl appearance.

Selected section:

- cyan icon/text,
- cyan underline,
- no giant filled rectangle.

## 7. Overview page

Two-column responsive card grid.

Left upper card: Connection

Rows:

- Status
- Tunnel Name
- Configuration File (only if product policy permits path display; show actual path, never a mock path)
- Interface / adapter display name
- Uptime

Right upper card: Traffic

- `Last 5 minutes` selector shown only if additional windows are actually supported. If v1 only implements five minutes, render a non-interactive caption instead of a fake ComboBox.
- Real traffic graph only.
- Download line: cyan/blue.
- Upload line: healthy green.
- Dynamic Y-axis.
- 300 one-second samples for five-minute history.
- Runtime data comes from manager-provided peer Rx/Tx counters.
- Rate calculation: `(currentBytes - previousBytes) * 8 / elapsedSeconds`.
- Counter rollback/reset produces zero for that sample, never negative/overflow.
- Current download/upload rate displayed below graph.
- Cumulative received/sent bytes displayed below rate.
- Do not create network polling in XAML code-behind; ViewModel/service owns sampling and cancellation.

Left lower card: DNS

For DoH:

- Heading `DNS (Encrypted)`.
- green healthy dot only when encrypted DNS is actually established/ready.
- `Using DNS over HTTPS (DoH)`.
- Server: configured DoH URL, with sensitive client identifiers redacted if policy requires.
- Address Family: actual tunnel family / bootstrap family decision.
- IMPORTANT semantic correction to the mock image: bootstrap resolvers are not plaintext fallback after encrypted DNS readiness. Label them `Bootstrap` or `Bootstrap Resolvers`; render `Fallback: Disabled` separately if space permits. Never show the bootstrap resolver list as active fallback.

For plain DNS:

- Heading `DNS`.
- Mode `Plain DNS`.
- Resolver list from actual config.
- Do not claim encryption.

Right lower card: Peer

Rows:

- Public Key: truncated representation only; copy command may copy public key.
- Endpoint
- Allowed IPs
- Persistent Keepalive

Never expose tunnel private key or preshared key material. Preshared key may only be represented as `Enabled` / `Not configured`.

## 8. Empty, disconnected, transitional, and error states

The provided mockup is the connected-state visual source of truth. Other states must preserve the same shell and dimensions.

Empty state:

- Connections rail remains fully rendered.
- Main content uses a centered WireHush icon, `No Connection Selected`, and a short instructional sentence.
- At most one primary `Import Tunnel(s)` action and one secondary `Add Tunnel` action in the empty state.
- Do not stretch these buttons across most of the content width.
- Do not duplicate them if the rail action area is already fully visible unless owner approves the duplication.

Disconnected selected tunnel:

- same summary/card geometry,
- muted disconnected state,
- primary `Connect` action,
- runtime-only fields use honest unavailable states,
- stored configuration fields remain visible.

Starting/stopping:

- disable repeat action,
- show `Connecting…` / `Disconnecting…`,
- no fake connected duration.

Error:

- keep window usable,
- show compact non-destructive InfoBar/toast for recoverable manager/tunnel errors,
- preserve underlying error in diagnostics,
- never silently retry privileged mutations forever.

## 9. Theme behavior

Implement `System`, `Dark`, and `Light` preferences unless a blocker is proven.

- Default: System.
- The owner-approved dark theme must match the mockup.
- Theme selection persists per user.
- Window chrome and caption buttons follow the selected theme.
- All owned dialogs, menus, flyouts, InfoBars, Settings, Log, and tunnel editor must respect theme.

## 10. Accessibility and interaction

- Full keyboard navigation.
- Logical tab order.
- Visible keyboard focus.
- AutomationProperties.Name on icon-only controls.
- Minimum target size 40x40 DIP for primary interactive icons.
- High DPI support through WinUI DIPs; do not manually multiply pixel values.
- Text must not be rendered into bitmaps.
- Respect Windows text scaling where possible.

## 11. Frontend architecture

Create a separate project, conceptually:

`windows-ui/WireHush.UI/`

Suggested folders:

- `Views/`
- `ViewModels/`
- `Models/`
- `Services/`
- `Controls/`
- `Converters/`
- `Assets/`

Required services:

- `IManagerClient`
- `ManagerClient`
- `INavigationService`
- `ThemeService`
- `DialogService`
- `TrafficSampler`
- `AppLogger`

No networking/tunnel privilege logic belongs in the UI project.

Avoid heavy MVVM frameworks initially. Use `INotifyPropertyChanged`, `ObservableCollection<T>`, and small command helpers unless a specific framework is approved.

## 12. Manager/UI IPC redesign

Do not expose the existing Go `encoding/gob` protocol directly to C#.

Use a versioned local Windows named-pipe protocol.

Transport requirements:

- Local machine only.
- Reject remote pipe clients.
- Per-session pipe instance or equivalent client authorization.
- Pipe ACL includes LocalSystem plus the intended interactive user SID; do not grant Everyone/Authenticated Users broad write access.
- Service validates connecting client identity/capabilities in addition to ACL.
- UI runs non-elevated.
- Manager remains privileged.
- Message size limit: 1 MiB maximum unless a documented need requires less/more.
- Length-prefixed UTF-8 messages.
- Protocol version in initial handshake.
- Request IDs for request/response correlation.
- Server push events supported on the same authenticated session or a second authenticated event pipe.
- Unknown method/version: explicit error, no crash.
- Parse errors: close offending client session safely.
- No private keys, preshared-key bytes, credentials, or DoH secrets in routine UI snapshots/events.

Suggested envelope:

```json
{
  "protocol": 1,
  "kind": "request",
  "id": "uuid",
  "method": "tunnel.list",
  "payload": {}
}
```

Responses/events use the same envelope with `kind`, `id`, `result`, or `error`.

Minimum API surface:

- `session.hello`
- `tunnel.list`
- `tunnel.get`
- `tunnel.runtime`
- `tunnel.start`
- `tunnel.stop`
- `tunnel.import`
- `tunnel.create`
- `tunnel.update`
- `tunnel.delete`
- `settings.bootstrap.get`
- `settings.bootstrap.save`
- `log.snapshot`
- event `tunnel.changed`
- event `tunnels.changed`
- event `manager.stopping`

Capabilities returned by `session.hello` drive UI enablement. Do not infer authorization only from local Windows group membership in the frontend.

## 13. Security rules

- No UI elevation requirement.
- No secrets in logs.
- No private key in ordinary UI models.
- DoH TLS verification remains mandatory.
- No silent plaintext DNS fallback for DoH.
- No remote named-pipe access.
- No unbounded message allocation.
- All UI strings derived from external/config data are treated as untrusted text, not markup.
- File import uses Windows picker and passes selected file content/path to a manager API designed for import; manager validates format and destination.
- Manager is authoritative for configuration persistence and privileged mutation.
- UI crash must not restart the manager service.
- Manager may restart UI with bounded backoff, but never a tight infinite visual relaunch loop.

## 14. Deployment direction

Keep the existing MSI installation model during migration.

- WinUI project is unpackaged (`WindowsPackageType=None`) unless later migration evidence supports a different choice.
- Use Windows App SDK stable runtime only.
- Installer must ensure required Windows App SDK runtime is present, or move to an explicitly approved self-contained model.
- Do not require Microsoft Store installation.
- Existing manager/tunnel service compatibility identifiers remain unchanged unless a migration task explicitly changes them.

## 15. Acceptance gates before Walk removal

The WinUI frontend is not accepted until all are true:

1. Running window materially matches the owner mockup at 100% reference scale.
2. Dark theme matches the owner mockup.
3. System and Light themes do not break layout.
4. Window opens reliably; no startup loop.
5. Manager connection/reconnection is bounded and user-visible when unavailable.
6. Tunnel list uses real manager data.
7. Selection works.
8. Connect/disconnect works.
9. Add/import/edit/delete works with correct authorization.
10. Settings/bootstrap controls work.
11. Log view works.
12. DoH and plain DNS behavior are not regressed.
13. Traffic graph uses real counters and survives counter reset/tunnel switch.
14. No private key or sensitive token is exposed.
15. DPI tests at 100%, 125%, 150%, 200% pass.
16. Keyboard navigation and screen-reader labels are sane.
17. Existing WireGuard installation remains unaffected.
18. Upgrade/uninstall/reinstall acceptance passes.
19. Dependency inventory and license notices are updated.
20. Walk is removed from the production UI dependency graph only after the WinUI path is owner-approved and release-tested.

## 16. Implementation discipline

This reference is deliberately explicit because the previous Task019 implementation drifted far from the approved mockup.

Rules for the worker:

- Do not reinterpret the layout.
- Do not add giant navigation slabs.
- Do not move the Connections rail halfway down the page.
- Do not create duplicate full-width empty-state buttons.
- Do not use placeholder/fake runtime data.
- Do not mark the task complete based on unit tests alone.
- Owner visual approval of the actual running app is mandatory.

This document defines the target. Any deliberate visual or architectural deviation requires explicit owner approval.