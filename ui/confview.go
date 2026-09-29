//go:build windows

package ui

import (
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
	empty, dashboard *walk.Composite
	title, state     *walk.Label
	connect          *darkButton
	nav              map[dashboardSection]*darkButton
	pages            map[dashboardSection]*walk.Composite
	tunnel           *manager.Tunnel
	tunnelChangedCB  *manager.TunnelChangeCallback
	updateTicker     *time.Ticker
	quit             chan struct{}
	traffic          trafficHistory
	trafficTunnel    string
	trafficGraph     *trafficGraph
	trafficSummary   *walk.Label
	summaryValues    []*walk.Label
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

func NewConfView(parent walk.Container) (*ConfView, error) {
	v := &ConfView{nav: map[dashboardSection]*darkButton{}, pages: map[dashboardSection]*walk.Composite{}}
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
		b, _ := newDarkButton(navigation, label, false)
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
	addDashboardRow(connection, "Status", "Disconnected")
	addDashboardRow(connection, "Tunnel Name", "—")
	addDashboardRow(connection, "Addresses", "Not Assigned")
	traffic, _ := newDashboardCard(ov, "Traffic")
	v.trafficGraph, _ = newTrafficGraph(traffic, &v.traffic)
	v.trafficSummary, _ = walk.NewLabel(traffic)
	applyMutedText(v.trafficSummary)
	dns, _ := newDashboardCard(ov, "DNS")
	addDashboardRow(dns, "Mode", "Not Configured")
	addDashboardRow(dns, "Fallback", "Disabled")
	peer, _ := newDashboardCard(ov, "Peer")
	addDashboardRow(peer, "Endpoint", "Not Configured")
	addDashboardRow(peer, "Preshared Key", "Not exposed")
	for s, title := range map[dashboardSection]string{dashboardNetwork: "Network", dashboardDNS: "DNS", dashboardPeer: "Peer", dashboardAllowedIPs: "Allowed IPs"} {
		c, _ := newDashboardCard(v.pages[s], title)
		addDashboardRow(c, "Details", "Select a tunnel to view configured values.")
	}
	v.showSection(dashboardOverview)
	v.tunnelChangedCB = manager.IPCClientRegisterTunnelChange(v.onChanged)
	v.updateTicker = time.NewTicker(time.Second)
	v.quit = make(chan struct{})
	go v.loop()
	return v, nil
}
func (v *ConfView) showSection(s dashboardSection) {
	for k, p := range v.pages {
		p.SetVisible(k == s)
		v.nav[k].primary = k == s
		v.nav[k].Invalidate()
	}
}
func (v *ConfView) loop() {
	for {
		select {
		case <-v.updateTicker.C:
			if v.tunnel != nil && v.Visible() {
				t := v.tunnel
				state, _ := t.State()
				c := conf.Config{}
				if state == manager.TunnelStarted {
					c, _ = t.RuntimeConfig()
				}
				if c.Name == "" {
					c, _ = t.StoredConfig()
				}
				v.Synchronize(func() { v.setTunnel(t, &c, state) })
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
		return
	}
	v.title.SetText(c.Name)
	v.state.SetText(textForState(s, false))
	if s == manager.TunnelStarted {
		v.state.SetTextColor(uiHealthyColor)
		v.connect.SetText("Disconnect")
	} else {
		applyMutedText(v.state)
		v.connect.SetText("Connect")
	}
	v.connect.SetEnabled(s == manager.TunnelStarted || s == manager.TunnelStopped)
	ipv4, ipv6 := "Not Assigned", "Not Assigned"
	for _, address := range c.Interface.Addresses {
		if address.Addr().Is4() {
			ipv4 = address.String()
		} else if address.Addr().Is6() {
			ipv6 = address.String()
		}
	}
	endpoint, handshake := "Not Configured", "No handshake yet"
	for _, peer := range c.Peers {
		if endpoint == "Not Configured" && !peer.Endpoint.IsEmpty() {
			endpoint = peer.Endpoint.String()
		}
		if !peer.LastHandshakeTime.IsEmpty() {
			handshake = peer.LastHandshakeTime.String()
		}
	}
	for i, value := range []string{ipv4, ipv6, endpoint, handshake} {
		v.summaryValues[i].SetText(value)
	}
	if v.trafficTunnel != t.Name {
		v.traffic.reset()
		v.trafficTunnel = t.Name
	}
	if s == manager.TunnelStarted {
		rx, tx := peerCounters(c.Peers)
		v.traffic.sample(time.Now(), rx, tx)
		x := v.traffic.latest()
		v.trafficSummary.SetText("Download " + formatRate(x.rxBps) + "   Upload " + formatRate(x.txBps))
		v.trafficGraph.Invalidate()
	}
}

var _ = win.IsIconic
