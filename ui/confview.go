//go:build windows

package ui

import (
	"fmt"
	"log"
	"net/netip"
	"runtime/debug"
	"strings"
	"time"

	"github.com/lxn/walk"
	"github.com/lxn/win"
	"golang.zx2c4.com/wireguard/windows/conf"
	"golang.zx2c4.com/wireguard/windows/manager"
)

type dashboardSection int

const (
	dashboardOverview dashboardSection = iota
	dashboardNetwork
	dashboardDNS
	dashboardPeer
	dashboardAllowedIPs
)

type ConfView struct {
	*walk.ScrollView
	empty, dashboard      *walk.Composite
	emptyImport, emptyAdd *darkButton
	title, state          *walk.Label
	connect               *darkButton
	nav                   map[dashboardSection]*darkButton
	pages                 map[dashboardSection]*walk.Composite
	tunnel                *manager.Tunnel
	tunnelChangedCB       *manager.TunnelChangeCallback
	updateTicker          *time.Ticker
	quit                  chan struct{}
	traffic               trafficHistory
	trafficTunnel         string
	trafficGraph          *trafficGraph
	trafficSummary        *walk.Label
	summaryValues         []*walk.Label
	rows                  map[string]*walk.Label
	dnsHeading            *walk.Label
	observedStart         time.Time
	lastState             manager.TunnelState
	lastTunnel            string
	detailKey             string
	trafficRx             uint64
	trafficTx             uint64
}

func newDashboardCard(parent walk.Container, title string) (*walk.Composite, error) {
	card, err := walk.NewComposite(parent)
	if err != nil {
		return nil, err
	}
	l := walk.NewVBoxLayout()
	l.SetMargins(walk.Margins{16, 14, 16, 14})
	l.SetSpacing(8)
	card.SetLayout(l)
	applyDarkSurface(card, uiCardBrush)
	h, _ := walk.NewLabel(card)
	h.SetText(title)
	h.SetTextColor(uiTextColor)
	f, _ := walk.NewFont("Segoe UI Semibold", 12, 0)
	h.SetFont(f)
	return card, nil
}
func addDashboardRow(parent walk.Container, label, value string) *walk.Label {
	row, _ := walk.NewComposite(parent)
	l := walk.NewHBoxLayout()
	l.SetMargins(walk.Margins{})
	row.SetLayout(l)
	k, _ := walk.NewLabel(row)
	k.SetText(label)
	applyMutedText(k)
	walk.NewHSpacer(row)
	v, _ := walk.NewLabel(row)
	v.SetText(value)
	v.SetTextColor(uiTextColor)
	return v
}

func (v *ConfView) row(parent walk.Container, key, label, value string) {
	v.rows[key] = addDashboardRow(parent, label, value)
}

func NewConfView(parent walk.Container) (*ConfView, error) {
	v := &ConfView{nav: map[dashboardSection]*darkButton{}, pages: map[dashboardSection]*walk.Composite{}, rows: map[string]*walk.Label{}}
	var err error
	v.ScrollView, err = walk.NewScrollView(parent)
	if err != nil {
		return nil, err
	}
	l := walk.NewVBoxLayout()
	l.SetMargins(walk.Margins{18, 18, 18, 18})
	l.SetSpacing(14)
	v.SetLayout(l)
	applyDarkSurface(v, uiCanvasBrush)
	v.empty, _ = walk.NewComposite(v)
	el := walk.NewVBoxLayout()
	el.SetMargins(walk.Margins{80, 100, 80, 80})
	el.SetAlignment(walk.AlignHCenterVNear)
	el.SetSpacing(12)
	v.empty.SetLayout(el)
	applyDarkSurface(v.empty, uiCanvasBrush)
	logo, _ := loadLogoIcon(64)
	if logo != nil {
		im, _ := walk.NewImageView(v.empty)
		im.SetImage(logo)
		im.SetMinMaxSize(walk.Size{64, 64}, walk.Size{64, 64})
	}
	et, _ := walk.NewLabel(v.empty)
	et.SetText("No Connection Selected")
	et.SetTextColor(uiTextColor)
	ef, _ := walk.NewFont("Segoe UI Semibold", 18, 0)
	et.SetFont(ef)
	ed, _ := walk.NewLabel(v.empty)
	ed.SetText("Select a connection from the left or import a tunnel to get started.")
	applyMutedText(ed)
	v.emptyImport, err = newDarkButton(v.empty, "Import Tunnel(s)", true)
	if err != nil {
		return nil, fmt.Errorf("empty import button: %w", err)
	}
	v.emptyAdd, err = newDarkButton(v.empty, "Add Tunnel", false)
	if err != nil {
		return nil, fmt.Errorf("empty add button: %w", err)
	}
	v.dashboard, _ = walk.NewComposite(v)
	dl := walk.NewVBoxLayout()
	dl.SetMargins(walk.Margins{})
	dl.SetSpacing(12)
	v.dashboard.SetLayout(dl)
	applyDarkSurface(v.dashboard, uiCanvasBrush)
	v.dashboard.SetVisible(false)
	header, _ := newDashboardCard(v.dashboard, "")
	hl := walk.NewHBoxLayout()
	hl.SetMargins(walk.Margins{18, 16, 18, 16})
	header.SetLayout(hl)
	left, _ := walk.NewComposite(header)
	left.SetLayout(walk.NewVBoxLayout())
	applyDarkSurface(left, uiCardBrush)
	v.title, _ = walk.NewLabel(left)
	v.title.SetTextColor(uiTextColor)
	tf, _ := walk.NewFont("Segoe UI Semibold", 22, 0)
	v.title.SetFont(tf)
	v.state, _ = walk.NewLabel(left)
	applyMutedText(v.state)
	walk.NewHSpacer(header)
	v.connect, err = newDarkButton(header, "Connect", true)
	if err != nil {
		return nil, err
	}
	v.connect.SetMinMaxSize(walk.Size{170, 52}, walk.Size{170, 52})
	v.connect.Clicked().Attach(v.onToggle)
	summary, _ := walk.NewComposite(v.dashboard)
	sl := walk.NewHBoxLayout()
	sl.SetMargins(walk.Margins{})
	sl.SetSpacing(12)
	summary.SetLayout(sl)
	applyDarkSurface(summary, uiCanvasBrush)
	for _, label := range []string{"VPN IP (IPv4)", "VPN IP (IPv6)", "Endpoint", "Latest Handshake"} {
		c, _ := newDashboardCard(summary, label)
		v.summaryValues = append(v.summaryValues, addDashboardRow(c, "", "Not Assigned"))
		c.SetMinMaxSize(walk.Size{180, 88}, walk.Size{0, 88})
	}
	navigation, _ := walk.NewComposite(v.dashboard)
	nl := walk.NewHBoxLayout()
	nl.SetMargins(walk.Margins{})
	nl.SetSpacing(4)
	navigation.SetLayout(nl)
	for i, label := range []string{"Overview", "Network", "DNS", "Peer", "Allowed IPs"} {
		s := dashboardSection(i)
		b, buttonErr := newDarkButton(navigation, label, false)
		if buttonErr != nil {
			return nil, fmt.Errorf("navigation button %q: %w", label, buttonErr)
		}
		v.nav[s] = b
		b.Clicked().Attach(func() { v.showSection(s) })
	}
	for i := dashboardOverview; i <= dashboardAllowedIPs; i++ {
		p, _ := walk.NewComposite(v.dashboard)
		p.SetLayout(walk.NewVBoxLayout())
		applyDarkSurface(p, uiCanvasBrush)
		p.SetVisible(false)
		v.pages[i] = p
	}
	ov := v.pages[dashboardOverview]
	ol := walk.NewHBoxLayout()
	ol.SetMargins(walk.Margins{})
	ol.SetSpacing(12)
	ov.SetLayout(ol)
	connection, _ := newDashboardCard(ov, "Connection")
	v.row(connection, "connection.status", "Status", "Disconnected")
	v.row(connection, "connection.name", "Tunnel Name", "—")
	v.row(connection, "connection.addresses", "Addresses", "Not Assigned")
	v.row(connection, "connection.uptime", "Uptime", "—")
	v.row(connection, "connection.listen", "Listen Port", "Not configured")
	traffic, _ := newDashboardCard(ov, "Traffic")
	v.trafficGraph, err = newTrafficGraph(traffic, &v.traffic)
	if err != nil {
		return nil, fmt.Errorf("traffic graph: %w", err)
	}
	if v.trafficGraph == nil {
		return nil, fmt.Errorf("traffic graph constructor returned nil widget")
	}
	v.trafficSummary, err = walk.NewLabel(traffic)
	if err != nil {
		return nil, fmt.Errorf("traffic summary label: %w", err)
	}
	applyMutedText(v.trafficSummary)
	dns, _ := newDashboardCard(ov, "DNS")
	// The first child is the heading; retain it so DoH can truthfully be
	// called out as encrypted when configured.
	v.dnsHeading, _ = dns.Children().At(0).(*walk.Label)
	v.row(dns, "dns.mode", "Mode", "Not Configured")
	v.row(dns, "dns.resolver", "Resolver", "Not configured")
	v.row(dns, "dns.family", "Address Family", "—")
	v.row(dns, "dns.fallback", "Fallback", "Disabled")
	peer, _ := newDashboardCard(ov, "Peer")
	v.row(peer, "peer.endpoint", "Endpoint", "Not Configured")
	v.row(peer, "peer.allowed", "Allowed IPs", "Not configured")
	v.row(peer, "peer.keepalive", "Persistent Keepalive", "Disabled")
	v.row(peer, "peer.handshake", "Latest Handshake", "No handshake yet")
	v.row(peer, "peer.psk", "Preshared Key", "Not configured")
	v.showSection(dashboardOverview)
	v.tunnelChangedCB = manager.IPCClientRegisterTunnelChange(v.onChanged)
	v.updateTicker = time.NewTicker(time.Second)
	v.quit = make(chan struct{})
	go v.loop()
	return v, nil
}

// SetEmptyActions keeps the deliberate empty state connected to the existing
// import and editor paths without giving the dashboard its own tunnel logic.
func (v *ConfView) SetEmptyActions(importTunnel, addTunnel func()) {
	v.emptyImport.Clicked().Attach(importTunnel)
	v.emptyAdd.Clicked().Attach(addTunnel)
}
func (v *ConfView) showSection(s dashboardSection) {
	for k, p := range v.pages {
		p.SetVisible(k == s)
		v.nav[k].primary = k == s
		v.nav[k].Invalidate()
	}
}
func (v *ConfView) loop() {
	log.Printf("ConfView.loop ENTER")
	defer func() {
		if recovered := recover(); recovered != nil {
			log.Printf("PANIC IN ConfView.loop: %v\n%s", recovered, debug.Stack())
			if syncer, ok := log.Writer().(interface{ Sync() error }); ok {
				_ = syncer.Sync()
			}
			panic(recovered)
		}
	}()
	for {
		select {
		case <-v.updateTicker.C:
			log.Printf("ConfView.loop TICK begin")
			if v.tunnel != nil && v.Visible() {
				t := v.tunnel
				log.Printf("ConfView.loop before State")
				state, _ := t.State()
				log.Printf("ConfView.loop after State")
				c := conf.Config{}
				if state == manager.TunnelStarted {
					log.Printf("ConfView.loop before RuntimeConfig")
					c, _ = t.RuntimeConfig()
				}
				if c.Name == "" {
					log.Printf("ConfView.loop before StoredConfig")
					c, _ = t.StoredConfig()
				}
				log.Printf("ConfView.loop before Synchronize")
				v.Synchronize(func() { v.setTunnel(t, &c, state) })
				log.Printf("ConfView.loop TICK queued")
			}
		case <-v.quit:
			return
		}
	}
}
func (v *ConfView) Dispose() {
	if v.tunnelChangedCB != nil {
		v.tunnelChangedCB.Unregister()
	}
	v.updateTicker.Stop()
	close(v.quit)
	v.ScrollView.Dispose()
}
func (v *ConfView) SetTunnel(t *manager.Tunnel) {
	v.tunnel = t
	if t == nil {
		v.setTunnel(nil, &conf.Config{}, manager.TunnelUnknown)
		return
	}
	go func() {
		s, _ := t.State()
		c, _ := t.StoredConfig()
		if s == manager.TunnelStarted {
			if r, e := t.RuntimeConfig(); e == nil {
				c = r
			}
		}
		v.Synchronize(func() { v.setTunnel(t, &c, s) })
	}()
}
func (v *ConfView) onChanged(t *manager.Tunnel, s, global manager.TunnelState, err error) {
	if v.tunnel != nil && t != nil && v.tunnel.Name == t.Name {
		c, _ := t.StoredConfig()
		if s == manager.TunnelStarted {
			if r, e := t.RuntimeConfig(); e == nil {
				c = r
			}
		}
		v.Synchronize(func() { v.setTunnel(t, &c, s) })
	}
}
func (v *ConfView) onToggle() {
	if v.tunnel == nil {
		return
	}
	v.connect.SetEnabled(false)
	go v.tunnel.Toggle()
}
func (v *ConfView) setTunnel(t *manager.Tunnel, c *conf.Config, s manager.TunnelState) {
	v.empty.SetVisible(t == nil)
	v.dashboard.SetVisible(t != nil)
	if t == nil {
		v.lastTunnel, v.lastState, v.observedStart = "", manager.TunnelUnknown, time.Time{}
		return
	}
	if v.lastTunnel != t.Name {
		v.traffic.reset()
		v.trafficTunnel = t.Name
		v.observedStart = time.Time{}
		v.lastState = manager.TunnelUnknown
		v.lastTunnel = t.Name
	}
	if s == manager.TunnelStarted && v.lastState != manager.TunnelStarted && v.lastState != manager.TunnelUnknown {
		v.observedStart = time.Now()
	}
	if s != manager.TunnelStarted {
		v.observedStart = time.Time{}
	}
	v.lastState = s
	v.title.SetText(c.Name)
	status := textForState(s, false)
	if s == manager.TunnelStarted {
		if !v.observedStart.IsZero() {
			status = "Connected for " + formatDuration(time.Since(v.observedStart))
		} else {
			status = "Connected"
		}
		v.state.SetTextColor(uiHealthyColor)
		v.connect.SetText("Disconnect")
	} else {
		applyMutedText(v.state)
		v.connect.SetText("Connect")
	}
	v.state.SetText(status)
	v.connect.SetEnabled(s == manager.TunnelStarted || s == manager.TunnelStopped)
	ipv4, ipv6 := splitAddresses(c.Interface.Addresses)
	endpoint, extraEndpoints := firstEndpoint(c.Peers)
	if extraEndpoints > 0 {
		endpoint += fmt.Sprintf("  +%d more", extraEndpoints)
	}
	handshake := newestHandshake(c.Peers)
	for i, value := range []string{ipv4, ipv6, endpoint, handshake} {
		v.summaryValues[i].SetText(value)
	}
	setRow := func(key, value string) {
		if label := v.rows[key]; label != nil {
			label.SetText(value)
		}
	}
	setRow("connection.status", textForState(s, false))
	setRow("connection.name", c.Name)
	setRow("connection.addresses", strings.Join(joinAddresses(c.Interface.Addresses), ", "))
	if c.Interface.ListenPort > 0 {
		setRow("connection.listen", fmt.Sprintf("%d", c.Interface.ListenPort))
	}
	if !v.observedStart.IsZero() && s == manager.TunnelStarted {
		setRow("connection.uptime", formatDuration(time.Since(v.observedStart)))
	} else {
		setRow("connection.uptime", "—")
	}
	mode, resolver, family := dnsDisplay(c)
	if len(c.Interface.DNSOverHTTPS) > 0 {
		v.dnsHeading.SetText("DNS (Encrypted)")
	} else {
		v.dnsHeading.SetText("DNS")
	}
	setRow("dns.mode", mode)
	setRow("dns.resolver", resolver)
	setRow("dns.family", family)
	setRow("dns.fallback", "Disabled")
	peer := firstPeer(c.Peers)
	setRow("peer.endpoint", endpoint)
	setRow("peer.allowed", allowedSummary(c.Peers))
	setRow("peer.handshake", handshake)
	if peer != nil && peer.PersistentKeepalive > 0 {
		setRow("peer.keepalive", fmt.Sprintf("%d seconds", peer.PersistentKeepalive))
	} else {
		setRow("peer.keepalive", "Disabled")
	}
	if peer != nil && !peer.PresharedKey.IsZero() {
		setRow("peer.psk", "Enabled")
	} else {
		setRow("peer.psk", "Not configured")
	}
	if s == manager.TunnelStarted {
		rx, tx := peerCounters(c.Peers)
		v.trafficRx, v.trafficTx = rx, tx
		v.traffic.sample(time.Now(), rx, tx)
	}
	x := v.traffic.latest()
	v.trafficSummary.SetText("Download " + formatRate(x.rxBps) + "   Upload " + formatRate(x.txBps) + "   Received " + formatBytes(v.trafficRx) + "   Sent " + formatBytes(v.trafficTx))
	v.trafficGraph.Invalidate()
	if key := detailFingerprint(c); key != v.detailKey {
		v.detailKey = key
		v.buildDetailPages(c)
	}
}

func (v *ConfView) buildDetailPages(c *conf.Config) {
	v.replaceDetails(v.pages[dashboardNetwork], func(parent *walk.Composite) {
		card, _ := newDashboardCard(parent, "Network")
		ipv4, ipv6 := splitAddresses(c.Interface.Addresses)
		addDashboardRow(card, "IPv4 Address(es)", ipv4)
		addDashboardRow(card, "IPv6 Address(es)", ipv6)
		if c.Interface.ListenPort > 0 {
			addDashboardRow(card, "Listen Port", fmt.Sprintf("%d", c.Interface.ListenPort))
		}
		if c.Interface.MTU > 0 {
			addDashboardRow(card, "MTU", fmt.Sprintf("%d", c.Interface.MTU))
		}
		endpoint, extra := firstEndpoint(c.Peers)
		if extra > 0 {
			endpoint += fmt.Sprintf("  +%d more", extra)
		}
		addDashboardRow(card, "Endpoint(s)", endpoint)
		if c.Interface.TableOff {
			addDashboardRow(card, "Routing Table", "Disabled by configuration")
		}
	})
	v.replaceDetails(v.pages[dashboardDNS], func(parent *walk.Composite) {
		card, _ := newDashboardCard(parent, "DNS")
		mode, resolver, family := dnsDisplay(c)
		addDashboardRow(card, "DNS Mode", mode)
		addDashboardRow(card, "Resolver", resolver)
		if len(c.Interface.DNSSearch) > 0 {
			addDashboardRow(card, "Search Suffixes", strings.Join(c.Interface.DNSSearch, ", "))
		}
		if len(c.Interface.DNSOverHTTPS) > 0 {
			addDashboardRow(card, "Bootstrap Policy", "Family-Aware Bootstrap")
			addDashboardRow(card, "Fallback", "Disabled")
		}
		addDashboardRow(card, "Address Family", family)
	})
	v.replaceDetails(v.pages[dashboardPeer], func(parent *walk.Composite) {
		if len(c.Peers) == 0 {
			card, _ := newDashboardCard(parent, "Peer")
			addDashboardRow(card, "Status", "No peers configured")
			return
		}
		for index, peer := range c.Peers {
			card, _ := newDashboardCard(parent, fmt.Sprintf("Peer %d", index+1))
			addDashboardRow(card, "Public Key", shortKey(peer.PublicKey.String()))
			if peer.Endpoint.IsEmpty() {
				addDashboardRow(card, "Endpoint", "Not configured")
			} else {
				addDashboardRow(card, "Endpoint", peer.Endpoint.String())
			}
			addDashboardRow(card, "Allowed IPs", joinPrefixes(peer.AllowedIPs))
			addDashboardRow(card, "Latest Handshake", handshakeDisplay(peer.LastHandshakeTime))
			addDashboardRow(card, "Received", formatBytes(uint64(peer.RxBytes)))
			addDashboardRow(card, "Sent", formatBytes(uint64(peer.TxBytes)))
			if peer.PersistentKeepalive > 0 {
				addDashboardRow(card, "Persistent Keepalive", fmt.Sprintf("%d seconds", peer.PersistentKeepalive))
			}
			if !peer.PresharedKey.IsZero() {
				addDashboardRow(card, "Preshared Key", "Enabled")
			}
		}
	})
	v.replaceDetails(v.pages[dashboardAllowedIPs], func(parent *walk.Composite) {
		if len(c.Peers) == 0 {
			card, _ := newDashboardCard(parent, "Allowed IPs")
			addDashboardRow(card, "Status", "No peers configured")
			return
		}
		for index, peer := range c.Peers {
			card, _ := newDashboardCard(parent, "Peer "+shortKey(peer.PublicKey.String()))
			addDashboardRow(card, fmt.Sprintf("Peer %d Prefixes", index+1), joinPrefixes(peer.AllowedIPs))
		}
	})
}

func (v *ConfView) replaceDetails(page *walk.Composite, build func(*walk.Composite)) {
	for page.Children().Len() > 0 {
		page.Children().At(0).Dispose()
	}
	build(page)
}

func splitAddresses(addresses []netip.Prefix) (string, string) {
	var ipv4, ipv6 []string
	for _, address := range addresses {
		if address.Addr().Is4() {
			ipv4 = append(ipv4, address.String())
		} else if address.Addr().Is6() {
			ipv6 = append(ipv6, address.String())
		}
	}
	if len(ipv4) == 0 {
		ipv4 = []string{"Not Assigned"}
	}
	if len(ipv6) == 0 {
		ipv6 = []string{"Not Assigned"}
	}
	return strings.Join(ipv4, ", "), strings.Join(ipv6, ", ")
}
func joinAddresses(addresses []netip.Prefix) []string {
	result := make([]string, 0, len(addresses))
	for _, address := range addresses {
		result = append(result, address.String())
	}
	if len(result) == 0 {
		return []string{"Not Assigned"}
	}
	return result
}
func firstEndpoint(peers []conf.Peer) (string, int) {
	for index, peer := range peers {
		if !peer.Endpoint.IsEmpty() {
			return peer.Endpoint.String(), len(peers) - index - 1
		}
	}
	return "Not Configured", 0
}
func newestHandshake(peers []conf.Peer) string {
	var newest conf.HandshakeTime
	for _, peer := range peers {
		if !peer.LastHandshakeTime.IsEmpty() && (newest.IsEmpty() || peer.LastHandshakeTime > newest) {
			newest = peer.LastHandshakeTime
		}
	}
	return handshakeDisplay(newest)
}
func handshakeDisplay(handshake conf.HandshakeTime) string {
	if handshake.IsEmpty() {
		return "No handshake yet"
	}
	return handshake.String()
}
func firstPeer(peers []conf.Peer) *conf.Peer {
	if len(peers) == 0 {
		return nil
	}
	return &peers[0]
}
func allowedSummary(peers []conf.Peer) string {
	if len(peers) == 0 {
		return "Not configured"
	}
	value := joinPrefixes(peers[0].AllowedIPs)
	if len(peers) > 1 {
		value += fmt.Sprintf("  +%d peer(s)", len(peers)-1)
	}
	return value
}
func joinPrefixes(prefixes []netip.Prefix) string {
	if len(prefixes) == 0 {
		return "Not configured"
	}
	values := make([]string, 0, len(prefixes))
	for _, prefix := range prefixes {
		values = append(values, prefix.String())
	}
	return strings.Join(values, ", ")
}
func dnsDisplay(c *conf.Config) (string, string, string) {
	has4, has6 := false, false
	for _, address := range c.Interface.Addresses {
		has4 = has4 || address.Addr().Is4()
		has6 = has6 || address.Addr().Is6()
	}
	family := "—"
	if has4 && has6 {
		family = "Dual-Stack"
	} else if has4 {
		family = "IPv4"
	} else if has6 {
		family = "IPv6"
	}
	if len(c.Interface.DNSOverHTTPS) > 0 {
		return "DNS over HTTPS (DoH)", strings.Join(c.Interface.DNSOverHTTPS, ", "), family
	}
	if len(c.Interface.DNS) > 0 {
		values := make([]string, 0, len(c.Interface.DNS))
		for _, dns := range c.Interface.DNS {
			values = append(values, dns.String())
		}
		return "Plain DNS", strings.Join(values, ", "), family
	}
	return "Not Configured", "Not configured", family
}
func shortKey(key string) string {
	if len(key) <= 16 {
		return key
	}
	return key[:8] + "..." + key[len(key)-6:]
}
func formatDuration(duration time.Duration) string {
	if duration < time.Minute {
		return "less than a minute"
	}
	hours, minutes := int(duration.Hours()), int(duration.Minutes())%60
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}
func detailFingerprint(c *conf.Config) string {
	values := []string{c.Name, strings.Join(joinAddresses(c.Interface.Addresses), ","), strings.Join(c.Interface.DNSOverHTTPS, ","), fmt.Sprint(c.Interface.DNS), fmt.Sprint(c.Interface.DNSSearch), fmt.Sprint(c.Interface.ListenPort), fmt.Sprint(c.Interface.MTU), fmt.Sprint(c.Interface.TableOff)}
	for _, peer := range c.Peers {
		values = append(values, peer.PublicKey.String(), peer.Endpoint.String(), joinPrefixes(peer.AllowedIPs), fmt.Sprint(peer.PersistentKeepalive), fmt.Sprint(!peer.PresharedKey.IsZero()))
	}
	return strings.Join(values, "|")
}

var _ = win.IsIconic
