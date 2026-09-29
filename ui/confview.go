/* SPDX-License-Identifier: MIT
 *
 * Copyright (C) 2019-2026 WireGuard LLC. All Rights Reserved.
 */

package ui

import (
	"strconv"
	"strings"
	"time"

	"github.com/lxn/walk"
	"github.com/lxn/win"

	"golang.zx2c4.com/wireguard/windows/conf"
	"golang.zx2c4.com/wireguard/windows/l18n"
	"golang.zx2c4.com/wireguard/windows/manager"
)

type widgetsLine interface {
	widgets() (walk.Widget, walk.Widget)
}

type widgetsLinesView interface {
	widgetsLines() []widgetsLine
}

type labelStatusLine struct {
	label           *walk.TextLabel
	statusComposite *walk.Composite
	statusImage     *walk.ImageView
	statusLabel     *walk.LineEdit
}

type labelTextLine struct {
	label *walk.TextLabel
	text  *walk.TextEdit
}

type toggleActiveLine struct {
	composite *walk.Composite
	button    *walk.PushButton
}

type interfaceView struct {
	status       *labelStatusLine
	publicKey    *labelTextLine
	listenPort   *labelTextLine
	mtu          *labelTextLine
	addresses    *labelTextLine
	dns          *labelTextLine
	scripts      *labelTextLine
	table        *labelTextLine
	toggleActive *toggleActiveLine
	lines        []widgetsLine
}

type peerView struct {
	publicKey           *labelTextLine
	presharedKey        *labelTextLine
	allowedIPs          *labelTextLine
	endpoint            *labelTextLine
	persistentKeepalive *labelTextLine
	latestHandshake     *labelTextLine
	transfer            *labelTextLine
	lines               []widgetsLine
}

type ConfView struct {
	*walk.ScrollView
	title           *walk.Label
	state           *walk.Label
	summaries       []*summaryCard
	name            *walk.GroupBox
	interfaze       *interfaceView
	peers           map[conf.Key]*peerView
	tunnelChangedCB *manager.TunnelChangeCallback
	tunnel          *manager.Tunnel
	updateTicker    *time.Ticker
	quit            chan struct{}
}

type summaryCard struct {
	value *walk.Label
}

func newSummaryCard(parent walk.Container, heading string) (*summaryCard, error) {
	card, err := walk.NewComposite(parent)
	if err != nil {
		return nil, err
	}
	card.SetMinMaxSize(walk.Size{150, 64}, walk.Size{0, 64})
	layout := walk.NewVBoxLayout()
	layout.SetMargins(walk.Margins{12, 8, 12, 8})
	layout.SetSpacing(2)
	card.SetLayout(layout)
	applyDarkSurface(card, uiCardBrush)
	label, err := walk.NewLabel(card)
	if err != nil {
		return nil, err
	}
	label.SetText(heading)
	applyMutedText(label)
	value, err := walk.NewLabel(card)
	if err != nil {
		return nil, err
	}
	value.SetText("—")
	value.SetTextColor(uiTextColor)
	return &summaryCard{value: value}, nil
}

func (lsl *labelStatusLine) widgets() (walk.Widget, walk.Widget) {
	return lsl.label, lsl.statusComposite
}

func (lsl *labelStatusLine) update(state manager.TunnelState) {
	icon, err := iconForState(state, 14)
	if err == nil {
		lsl.statusImage.SetImage(icon)
	} else {
		lsl.statusImage.SetImage(nil)
	}

	s, e := lsl.statusLabel.TextSelection()
	lsl.statusLabel.SetText(textForState(state, false))
	lsl.statusLabel.SetTextSelection(s, e)
}

func (lsl *labelStatusLine) Dispose() {
	lsl.label.Dispose()
	lsl.statusComposite.Dispose()
}

func newLabelStatusLine(parent walk.Container) (*labelStatusLine, error) {
	var err error
	var disposables walk.Disposables
	defer disposables.Treat()

	lsl := new(labelStatusLine)

	if lsl.label, err = walk.NewTextLabel(parent); err != nil {
		return nil, err
	}
	disposables.Add(lsl.label)
	lsl.label.SetText(l18n.Sprintf("Status:"))
	lsl.label.SetTextAlignment(walk.AlignHFarVNear)
	lsl.label.SetTextColor(uiMutedColor)

	if lsl.statusComposite, err = walk.NewComposite(parent); err != nil {
		return nil, err
	}
	disposables.Add(lsl.statusComposite)
	layout := walk.NewHBoxLayout()
	layout.SetMargins(walk.Margins{})
	layout.SetAlignment(walk.AlignHNearVNear)
	layout.SetSpacing(0)
	lsl.statusComposite.SetLayout(layout)
	applyDarkSurface(lsl.statusComposite, uiCardBrush)

	if lsl.statusImage, err = walk.NewImageView(lsl.statusComposite); err != nil {
		return nil, err
	}
	disposables.Add(lsl.statusImage)
	lsl.statusImage.SetMargin(2)
	lsl.statusImage.SetMode(walk.ImageViewModeIdeal)

	if lsl.statusLabel, err = walk.NewLineEdit(lsl.statusComposite); err != nil {
		return nil, err
	}
	disposables.Add(lsl.statusLabel)
	win.SetWindowLong(lsl.statusLabel.Handle(), win.GWL_EXSTYLE, win.GetWindowLong(lsl.statusLabel.Handle(), win.GWL_EXSTYLE)&^win.WS_EX_CLIENTEDGE)
	lsl.statusLabel.SetReadOnly(true)
	lsl.statusLabel.SetBackground(uiCardBrush)
	lsl.statusLabel.SetTextColor(uiTextColor)
	lsl.statusLabel.FocusedChanged().Attach(func() {
		lsl.statusLabel.SetTextSelection(0, 0)
	})
	lsl.update(manager.TunnelUnknown)
	lsl.statusLabel.Accessibility().SetRole(walk.AccRoleStatictext)

	disposables.Spare()

	return lsl, nil
}

func (lt *labelTextLine) widgets() (walk.Widget, walk.Widget) {
	return lt.label, lt.text
}

func (lt *labelTextLine) show(text string) {
	s, e := lt.text.TextSelection()
	lt.text.SetText(text)
	lt.label.SetVisible(true)
	lt.text.SetVisible(true)
	lt.text.SetTextSelection(s, e)
}

func (lt *labelTextLine) hide() {
	lt.text.SetText("")
	lt.label.SetVisible(false)
	lt.text.SetVisible(false)
}

func (lt *labelTextLine) hideWithText(text string) {
	lt.text.SetText(text)
	lt.label.SetVisible(false)
	lt.text.SetVisible(false)
}

func (lt *labelTextLine) Dispose() {
	lt.label.Dispose()
	lt.text.Dispose()
}

func newLabelTextLine(fieldName string, parent walk.Container) (*labelTextLine, error) {
	var err error
	var disposables walk.Disposables
	defer disposables.Treat()

	lt := new(labelTextLine)

	if lt.label, err = walk.NewTextLabel(parent); err != nil {
		return nil, err
	}
	disposables.Add(lt.label)
	lt.label.SetText(fieldName)
	lt.label.SetTextAlignment(walk.AlignHFarVNear)
	lt.label.SetTextColor(uiMutedColor)
	lt.label.SetVisible(false)

	if lt.text, err = walk.NewTextEdit(parent); err != nil {
		return nil, err
	}
	disposables.Add(lt.text)
	win.SetWindowLong(lt.text.Handle(), win.GWL_EXSTYLE, win.GetWindowLong(lt.text.Handle(), win.GWL_EXSTYLE)&^win.WS_EX_CLIENTEDGE)
	lt.text.SetCompactHeight(true)
	lt.text.SetReadOnly(true)
	lt.text.SetBackground(uiCardBrush)
	lt.text.SetTextColor(uiTextColor)
	lt.text.SetVisible(false)
	lt.text.FocusedChanged().Attach(func() {
		lt.text.SetTextSelection(0, 0)
	})
	lt.text.Accessibility().SetRole(walk.AccRoleStatictext)

	disposables.Spare()

	return lt, nil
}

func (tal *toggleActiveLine) widgets() (walk.Widget, walk.Widget) {
	return nil, tal.composite
}

func (tal *toggleActiveLine) updateGlobal(globalState manager.TunnelState) {
	tal.button.SetEnabled(globalState == manager.TunnelStarted || globalState == manager.TunnelStopped)
}

func (tal *toggleActiveLine) update(state manager.TunnelState) {
	var text string

	switch state {
	case manager.TunnelStarted:
		text = l18n.Sprintf("&Deactivate")
	case manager.TunnelStopped:
		text = l18n.Sprintf("&Activate")
	case manager.TunnelStarting, manager.TunnelStopping:
		text = textForState(state, true)
	default:
		text = ""
	}

	tal.button.SetText(text)
	tal.button.SetVisible(state != manager.TunnelUnknown)
}

func (tal *toggleActiveLine) Dispose() {
	tal.composite.Dispose()
}

func newToggleActiveLine(parent walk.Container) (*toggleActiveLine, error) {
	var err error
	var disposables walk.Disposables
	defer disposables.Treat()

	tal := new(toggleActiveLine)

	if tal.composite, err = walk.NewComposite(parent); err != nil {
		return nil, err
	}
	disposables.Add(tal.composite)
	layout := walk.NewHBoxLayout()
	layout.SetMargins(walk.Margins{0, 0, 0, 6})
	tal.composite.SetLayout(layout)

	if tal.button, err = walk.NewPushButton(tal.composite); err != nil {
		return nil, err
	}
	disposables.Add(tal.button)
	tal.button.SetBackground(uiAccentBrush)
	walk.NewHSpacer(tal.composite)
	tal.update(manager.TunnelStopped)

	disposables.Spare()

	return tal, nil
}

type labelTextLineItem struct {
	label string
	ptr   **labelTextLine
}

func createLabelTextLines(items []labelTextLineItem, parent walk.Container, disposables *walk.Disposables) ([]widgetsLine, error) {
	var err error
	var disps walk.Disposables
	defer disps.Treat()

	wls := make([]widgetsLine, len(items))
	for i, item := range items {
		if *item.ptr, err = newLabelTextLine(item.label, parent); err != nil {
			return nil, err
		}
		disps.Add(*item.ptr)
		if disposables != nil {
			disposables.Add(*item.ptr)
		}
		wls[i] = *item.ptr
	}

	disps.Spare()

	return wls, nil
}

func newInterfaceView(parent walk.Container) (*interfaceView, error) {
	var err error
	var disposables walk.Disposables
	defer disposables.Treat()

	iv := new(interfaceView)

	if iv.status, err = newLabelStatusLine(parent); err != nil {
		return nil, err
	}
	disposables.Add(iv.status)

	items := []labelTextLineItem{
		{l18n.Sprintf("Public Key:"), &iv.publicKey},
		{l18n.Sprintf("Listen Port:"), &iv.listenPort},
		{l18n.Sprintf("MTU:"), &iv.mtu},
		{l18n.Sprintf("Addresses:"), &iv.addresses},
		{l18n.Sprintf("DNS Servers:"), &iv.dns},
		{l18n.Sprintf("Scripts:"), &iv.scripts},
		{l18n.Sprintf("Table:"), &iv.table},
	}
	if iv.lines, err = createLabelTextLines(items, parent, &disposables); err != nil {
		return nil, err
	}

	if iv.toggleActive, err = newToggleActiveLine(parent); err != nil {
		return nil, err
	}
	disposables.Add(iv.toggleActive)

	iv.lines = append([]widgetsLine{iv.status}, append(iv.lines, iv.toggleActive)...)

	layoutInGrid(iv, parent.Layout().(*walk.GridLayout))

	disposables.Spare()

	return iv, nil
}

func newPeerView(parent walk.Container) (*peerView, error) {
	pv := new(peerView)

	items := []labelTextLineItem{
		{l18n.Sprintf("Public Key:"), &pv.publicKey},
		{l18n.Sprintf("Preshared Key:"), &pv.presharedKey},
		{l18n.Sprintf("Allowed IPs:"), &pv.allowedIPs},
		{l18n.Sprintf("Endpoint:"), &pv.endpoint},
		{l18n.Sprintf("Persistent Keepalive:"), &pv.persistentKeepalive},
		{l18n.Sprintf("Latest Handshake:"), &pv.latestHandshake},
		{l18n.Sprintf("Transfer:"), &pv.transfer},
	}
	var err error
	if pv.lines, err = createLabelTextLines(items, parent, nil); err != nil {
		return nil, err
	}

	layoutInGrid(pv, parent.Layout().(*walk.GridLayout))

	return pv, nil
}

func layoutInGrid(view widgetsLinesView, layout *walk.GridLayout) {
	for i, l := range view.widgetsLines() {
		w1, w2 := l.widgets()

		if w1 != nil {
			layout.SetRange(w1, walk.Rectangle{0, i, 1, 1})
		}
		if w2 != nil {
			layout.SetRange(w2, walk.Rectangle{2, i, 1, 1})
		}
	}
}

func (iv *interfaceView) widgetsLines() []widgetsLine {
	return iv.lines
}

func (iv *interfaceView) apply(c *conf.Interface, state manager.TunnelState) {
	if IsAdmin {
		iv.publicKey.show(c.PrivateKey.Public().String())
	} else {
		iv.publicKey.hideWithText(c.PrivateKey.Public().String())
	}

	if c.ListenPort > 0 {
		iv.listenPort.show(strconv.Itoa(int(c.ListenPort)))
	} else {
		iv.listenPort.hide()
	}

	if c.MTU > 0 {
		iv.mtu.show(strconv.Itoa(int(c.MTU)))
	} else {
		iv.mtu.hide()
	}

	if len(c.Addresses) > 0 {
		addrStrings := make([]string, len(c.Addresses))
		for i, address := range c.Addresses {
			addrStrings[i] = address.String()
		}
		iv.addresses.show(strings.Join(addrStrings[:], l18n.EnumerationSeparator()))
	} else {
		iv.addresses.hide()
	}

	if len(c.DNSOverHTTPS) > 0 {
		if state == manager.TunnelStarting {
			iv.dns.show(l18n.Sprintf("Encrypted (DoH): initializing"))
		} else {
			status := l18n.Sprintf("Encrypted (DoH)")
			if state == manager.TunnelStopped {
				status += l18n.Sprintf(" (ready on activation)")
			}
			iv.dns.show(status + ": " + strings.Join(c.DNSOverHTTPS, l18n.EnumerationSeparator()))
		}
	} else if len(c.DNS)+len(c.DNSSearch) > 0 {
		addrStrings := make([]string, 0, len(c.DNS)+len(c.DNSSearch))
		for _, address := range c.DNS {
			addrStrings = append(addrStrings, address.String())
		}
		addrStrings = append(addrStrings, c.DNSSearch...)
		iv.dns.show(l18n.Sprintf("Plain: %s", strings.Join(addrStrings[:], l18n.EnumerationSeparator())))
	} else {
		iv.dns.hide()
	}

	var scriptsInUse []string
	if len(c.PreUp) > 0 {
		scriptsInUse = append(scriptsInUse, l18n.Sprintf("pre-up"))
	}
	if len(c.PostUp) > 0 {
		scriptsInUse = append(scriptsInUse, l18n.Sprintf("post-up"))
	}
	if len(c.PreDown) > 0 {
		scriptsInUse = append(scriptsInUse, l18n.Sprintf("pre-down"))
	}
	if len(c.PostDown) > 0 {
		scriptsInUse = append(scriptsInUse, l18n.Sprintf("post-down"))
	}
	if len(scriptsInUse) > 0 {
		if conf.AdminBool("DangerousScriptExecution") {
			iv.scripts.show(strings.Join(scriptsInUse, l18n.EnumerationSeparator()))
		} else {
			iv.scripts.show(l18n.Sprintf("disabled, per policy"))
		}
	} else {
		iv.scripts.hide()
	}

	if c.TableOff {
		iv.table.show(l18n.Sprintf("off"))
	} else {
		iv.table.hide()
	}
}

func (iv *interfaceView) showDNSError(err error) {
	if err != nil {
		iv.dns.show(l18n.Sprintf("Encrypted (DoH): error: %v", err))
	}
}

func (pv *peerView) widgetsLines() []widgetsLine {
	return pv.lines
}

func (pv *peerView) apply(c *conf.Peer) {
	if IsAdmin {
		pv.publicKey.show(c.PublicKey.String())
	} else {
		pv.publicKey.hideWithText(c.PublicKey.String())
	}

	if !c.PresharedKey.IsZero() && IsAdmin {
		pv.presharedKey.show(l18n.Sprintf("enabled"))
	} else {
		pv.presharedKey.hide()
	}

	if len(c.AllowedIPs) > 0 {
		addrStrings := make([]string, len(c.AllowedIPs))
		for i, address := range c.AllowedIPs {
			addrStrings[i] = address.String()
		}
		pv.allowedIPs.show(strings.Join(addrStrings[:], l18n.EnumerationSeparator()))
	} else {
		pv.allowedIPs.hide()
	}

	if !c.Endpoint.IsEmpty() {
		pv.endpoint.show(c.Endpoint.String())
	} else {
		pv.endpoint.hide()
	}

	if c.PersistentKeepalive > 0 {
		pv.persistentKeepalive.show(strconv.Itoa(int(c.PersistentKeepalive)))
	} else {
		pv.persistentKeepalive.hide()
	}

	if !c.LastHandshakeTime.IsEmpty() {
		pv.latestHandshake.show(c.LastHandshakeTime.String())
	} else {
		pv.latestHandshake.hide()
	}

	if c.RxBytes > 0 || c.TxBytes > 0 {
		pv.transfer.show(l18n.Sprintf("%s received, %s sent", c.RxBytes.String(), c.TxBytes.String()))
	} else {
		pv.transfer.hide()
	}
}

func newPaddedGroupGrid(parent walk.Container) (group *walk.GroupBox, err error) {
	group, err = walk.NewGroupBox(parent)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			group.Dispose()
		}
	}()
	layout := walk.NewGridLayout()
	layout.SetMargins(walk.Margins{10, 5, 10, 5})
	layout.SetSpacing(0)
	err = group.SetLayout(layout)
	if err != nil {
		return nil, err
	}
	applyDarkSurface(group, uiCardBrush)
	spacer, err := walk.NewSpacerWithCfg(group, &walk.SpacerCfg{walk.GrowableHorz | walk.GreedyHorz, walk.Size{10, 0}, false})
	if err != nil {
		return nil, err
	}
	layout.SetRange(spacer, walk.Rectangle{1, 0, 1, 1})
	return group, nil
}

func NewConfView(parent walk.Container) (*ConfView, error) {
	var err error
	var disposables walk.Disposables
	defer disposables.Treat()

	cv := new(ConfView)
	if cv.ScrollView, err = walk.NewScrollView(parent); err != nil {
		return nil, err
	}
	disposables.Add(cv)
	vlayout := walk.NewVBoxLayout()
	vlayout.SetMargins(walk.Margins{16, 14, 16, 16})
	vlayout.SetSpacing(12)
	cv.SetLayout(vlayout)
	applyDarkSurface(cv, uiCanvasBrush)

	header, err := walk.NewComposite(cv)
	if err != nil {
		return nil, err
	}
	header.SetLayout(walk.NewVBoxLayout())
	applyDarkSurface(header, uiCanvasBrush)
	if cv.title, err = walk.NewLabel(header); err != nil {
		return nil, err
	}
	cv.title.SetText(l18n.Sprintf("Select a connection"))
	cv.title.SetTextColor(uiTextColor)
	if titleFont, fontErr := walk.NewFont("Segoe UI Semibold", 20, 0); fontErr == nil {
		cv.title.SetFont(titleFont)
		disposables.Add(titleFont)
	}
	if cv.state, err = walk.NewLabel(header); err != nil {
		return nil, err
	}
	cv.state.SetText(l18n.Sprintf("Choose a tunnel from Connections to view its details."))
	applyMutedText(cv.state)

	summaryRow, err := walk.NewComposite(cv)
	if err != nil {
		return nil, err
	}
	summaryLayout := walk.NewHBoxLayout()
	summaryLayout.SetMargins(walk.Margins{})
	summaryLayout.SetSpacing(10)
	summaryRow.SetLayout(summaryLayout)
	applyDarkSurface(summaryRow, uiCanvasBrush)
	for _, heading := range []string{"Network", "DNS", "Endpoint", "Latest Handshake"} {
		card, cardErr := newSummaryCard(summaryRow, l18n.Sprintf(heading))
		if cardErr != nil {
			return nil, cardErr
		}
		cv.summaries = append(cv.summaries, card)
	}
	if cv.name, err = newPaddedGroupGrid(cv); err != nil {
		return nil, err
	}
	if cv.interfaze, err = newInterfaceView(cv.name); err != nil {
		return nil, err
	}
	cv.interfaze.toggleActive.button.Clicked().Attach(cv.onToggleActiveClicked)
	cv.peers = make(map[conf.Key]*peerView)
	cv.tunnelChangedCB = manager.IPCClientRegisterTunnelChange(cv.onTunnelChanged)
	cv.SetTunnel(nil)
	globalState, err := manager.IPCClientGlobalState()
	if err != nil {
		return nil, err
	}
	cv.interfaze.toggleActive.updateGlobal(globalState)

	if err := walk.InitWrapperWindow(cv); err != nil {
		return nil, err
	}
	cv.SetDoubleBuffering(true)
	cv.updateTicker = time.NewTicker(time.Second)
	cv.quit = make(chan struct{})
	go func() {
		for {
			select {
			case <-cv.updateTicker.C:
				if !cv.Visible() || !cv.Form().Visible() || win.IsIconic(cv.Form().Handle()) {
					continue
				}
				if cv.tunnel != nil {
					tunnel := cv.tunnel
					var state manager.TunnelState
					var config conf.Config
					if state, _ = tunnel.State(); state == manager.TunnelStarted {
						config, _ = tunnel.RuntimeConfig()
					}
					if config.Name == "" {
						config, _ = tunnel.StoredConfig()
					}
					cv.Synchronize(func() {
						cv.setTunnel(tunnel, &config, state)
					})
				}
			case <-cv.quit:
				return
			}
		}
	}()

	disposables.Spare()

	return cv, nil
}

func (cv *ConfView) Dispose() {
	if cv.tunnelChangedCB != nil {
		cv.tunnelChangedCB.Unregister()
		cv.tunnelChangedCB = nil
	}
	if cv.updateTicker != nil {
		cv.updateTicker.Stop()
		close(cv.quit)
		cv.updateTicker = nil
	}
	cv.ScrollView.Dispose()
}

func (cv *ConfView) onToggleActiveClicked() {
	cv.interfaze.toggleActive.button.SetEnabled(false)
	go func() {
		oldState, err := cv.tunnel.Toggle()
		if err != nil {
			cv.Synchronize(func() {
				if oldState == manager.TunnelUnknown {
					showErrorCustom(cv.Form(), l18n.Sprintf("Failed to determine tunnel state"), err.Error())
				} else if oldState == manager.TunnelStopped {
					showErrorCustom(cv.Form(), l18n.Sprintf("Failed to activate tunnel"), err.Error())
				} else if oldState == manager.TunnelStarted {
					showErrorCustom(cv.Form(), l18n.Sprintf("Failed to deactivate tunnel"), err.Error())
				}
			})
		}
	}()
}

func (cv *ConfView) onTunnelChanged(tunnel *manager.Tunnel, state, globalState manager.TunnelState, err error) {
	cv.Synchronize(func() {
		cv.interfaze.toggleActive.updateGlobal(globalState)
		if cv.tunnel != nil && cv.tunnel.Name == tunnel.Name {
			cv.interfaze.status.update(state)
			cv.interfaze.toggleActive.update(state)
		}
	})
	if cv.tunnel != nil && cv.tunnel.Name == tunnel.Name {
		var config conf.Config
		if state == manager.TunnelStarted {
			config, _ = tunnel.RuntimeConfig()
		}
		if config.Name == "" {
			config, _ = tunnel.StoredConfig()
		}
		cv.Synchronize(func() {
			cv.setTunnel(tunnel, &config, state)
			if err != nil && len(config.Interface.DNSOverHTTPS) > 0 {
				cv.interfaze.showDNSError(err)
			}
		})
	}
}

func (cv *ConfView) SetTunnel(tunnel *manager.Tunnel) {
	cv.tunnel = tunnel // XXX: This races with the read in the updateTicker, but it's pointer-sized!

	var config conf.Config
	var state manager.TunnelState
	if tunnel != nil {
		go func() {
			if state, _ = tunnel.State(); state == manager.TunnelStarted {
				config, _ = tunnel.RuntimeConfig()
			}
			if config.Name == "" {
				config, _ = tunnel.StoredConfig()
			}
			cv.Synchronize(func() {
				cv.setTunnel(tunnel, &config, state)
			})
		}()
	} else {
		cv.setTunnel(tunnel, &config, state)
	}
}

func (cv *ConfView) setTunnel(tunnel *manager.Tunnel, config *conf.Config, state manager.TunnelState) {
	if !(cv.tunnel == nil || tunnel == nil || tunnel.Name == cv.tunnel.Name) {
		return
	}

	if tunnel == nil {
		cv.title.SetText(l18n.Sprintf("Select a connection"))
		cv.state.SetText(l18n.Sprintf("Choose a tunnel from Connections to view its details."))
		applyMutedText(cv.state)
		for _, summary := range cv.summaries {
			summary.value.SetText("—")
		}
	} else {
		cv.title.SetText(config.Name)
		if state == manager.TunnelStarted {
			cv.state.SetText(l18n.Sprintf("Connected"))
			cv.state.SetTextColor(uiHealthyColor)
		} else {
			cv.state.SetText(textForState(state, false))
			applyMutedText(cv.state)
		}
		addresses := l18n.Sprintf("No address")
		if len(config.Interface.Addresses) > 0 {
			values := make([]string, len(config.Interface.Addresses))
			for i, address := range config.Interface.Addresses {
				values[i] = address.String()
			}
			addresses = strings.Join(values, l18n.EnumerationSeparator())
		}
		dns := l18n.Sprintf("Not configured")
		if len(config.Interface.DNSOverHTTPS) > 0 {
			dns = l18n.Sprintf("Encrypted DNS (DoH)")
		} else if len(config.Interface.DNS) > 0 {
			dns = l18n.Sprintf("Plain DNS")
		}
		endpoint := l18n.Sprintf("Not configured")
		handshake := l18n.Sprintf("No handshake yet")
		if len(config.Peers) > 0 {
			if !config.Peers[0].Endpoint.IsEmpty() {
				endpoint = config.Peers[0].Endpoint.String()
			}
			if !config.Peers[0].LastHandshakeTime.IsEmpty() {
				handshake = config.Peers[0].LastHandshakeTime.String()
			}
		}
		for i, value := range []string{addresses, dns, endpoint, handshake} {
			cv.summaries[i].value.SetText(value)
		}
	}

	title := l18n.Sprintf("Connection")
	if cv.name.Title() != title {
		cv.SetSuspended(true)
		defer cv.SetSuspended(false)
		cv.name.SetTitle(title)
	}
	cv.name.SetVisible(tunnel != nil)

	cv.interfaze.apply(&config.Interface, state)
	cv.interfaze.status.update(state)
	cv.interfaze.toggleActive.update(state)
	inverse := make(map[*peerView]bool, len(cv.peers))
	all := make([]*peerView, 0, len(cv.peers))
	for _, pv := range cv.peers {
		inverse[pv] = true
		all = append(all, pv)
	}
	someMatch := false
	for _, peer := range config.Peers {
		_, ok := cv.peers[peer.PublicKey]
		if ok {
			someMatch = true
			break
		}
	}
	for _, peer := range config.Peers {
		if pv := cv.peers[peer.PublicKey]; (!someMatch && len(all) > 0) || pv != nil {
			if pv == nil {
				pv = all[0]
				all = all[1:]
				k, e := conf.NewPrivateKeyFromString(pv.publicKey.text.Text())
				if e != nil {
					continue
				}
				delete(cv.peers, *k)
				cv.peers[peer.PublicKey] = pv
			}
			pv.apply(&peer)
			inverse[pv] = false
		} else {
			group, err := newPaddedGroupGrid(cv)
			if err != nil {
				continue
			}
			group.SetTitle(l18n.Sprintf("Peer"))
			pv, err := newPeerView(group)
			if err != nil {
				group.Dispose()
				continue
			}
			pv.apply(&peer)
			cv.peers[peer.PublicKey] = pv
		}
	}
	for pv, remove := range inverse {
		if !remove {
			continue
		}
		k, e := conf.NewPrivateKeyFromString(pv.publicKey.text.Text())
		if e != nil {
			continue
		}
		delete(cv.peers, *k)
		groupBox := pv.publicKey.label.Parent().AsContainerBase().Parent().(*walk.GroupBox)
		groupBox.SetVisible(false)
		groupBox.Parent().Children().Remove(groupBox)
		groupBox.Dispose()
	}
}
