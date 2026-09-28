# Known limitations

- The available VM shell is not an elevated desktop and has no installed TunnelMint manager service. Interactive UI, service lifecycle, per-machine install, uninstall/reinstall, and coexistence checks need owner testing on a disposable elevated Windows VM.
- No real VPN peer or tunnel configuration was available. Real handshake, traffic, plain DNS, DoH full-tunnel, DoH split-tunnel, packet capture, and DNS leak observation remain manual verification items.
- IPv6, sleep/wake, and Ethernet/Wi-Fi transition behavior were not exercised in this environment.
- The broad upstream `conf`, DPAPI, and adapter test suites depend on protected storage, DPAPI, or a usable Windows adapter and did not complete reliably here.
- The release-candidate executables and MSIs are unsigned development artifacts. WiX ICE validation could not execute because the VM Windows Installer service was unavailable; the installer script still produced the three MSI artifacts, but their ICE results require owner or CI validation on a machine with Windows Installer available.
- A TunnelMint tunnel currently supports one configured DoH endpoint per activation. The DNS proxy accepts IPv4 loopback by default; its implementation also supports IPv6 loopback where the host configuration uses it.
