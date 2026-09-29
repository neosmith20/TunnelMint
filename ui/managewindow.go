/* SPDX-License-Identifier: MIT
 *
 * Copyright (C) 2019-2026 WireGuard LLC. All Rights Reserved.
 */

package ui

import (
	"sync"
	"unsafe"

	"github.com/lxn/walk"
	"github.com/lxn/win"
	"golang.org/x/sys/windows"

	"golang.zx2c4.com/wireguard/windows/l18n"
	"golang.zx2c4.com/wireguard/windows/manager"
	"golang.zx2c4.com/wireguard/windows/product"
)

type ManageTunnelsWindow struct {
	walk.FormBase

	tabs         *walk.TabWidget
	tunnelsPage  *TunnelsPage
	logPage      *LogPage
	settingsPage *SettingsPage
	updatePage   *UpdatePage

	tunnelChangedCB *manager.TunnelChangeCallback
}

const (
	manageWindowWindowClass = product.ManagerWindowClass
	raiseMsg                = win.WM_USER + 0x3510
	aboutWireHushCmd        = 0x37
)

var taskbarButtonCreatedMsg uint32

var initedManageTunnels sync.Once

func NewManageTunnelsWindow() (*ManageTunnelsWindow, error) {
	initedManageTunnels.Do(func() {
		walk.AppendToWalkInit(func() {
			walk.MustRegisterWindowClass(manageWindowWindowClass)
			taskbarButtonCreatedMsg = win.RegisterWindowMessage(windows.StringToUTF16Ptr("TaskbarButtonCreated"))
		})
	})

	var err error
	var disposables walk.Disposables
	defer disposables.Treat()

	font, err := walk.NewFont("Segoe UI", 9, 0)
	if err != nil {
		return nil, err
	}

	mtw := new(ManageTunnelsWindow)
	mtw.SetName(product.Name)

	err = walk.InitWindow(mtw, nil, manageWindowWindowClass, win.WS_OVERLAPPEDWINDOW, win.WS_EX_CONTROLPARENT)
	if err != nil {
		return nil, err
	}
	disposables.Add(mtw)
	win.ChangeWindowMessageFilterEx(mtw.Handle(), raiseMsg, win.MSGFLT_ALLOW, nil)
	mtw.SetPersistent(true)

	if icon, err := loadLogoIcon(32); err == nil {
		mtw.SetIcon(icon)
	}
	mtw.SetTitle(product.ManagerWindowTitle)
	mtw.SetFont(font)
	mtw.SetSize(walk.Size{1120, 720})
	mtw.SetMinMaxSize(walk.Size{900, 600}, walk.Size{0, 0})
	applyDarkWindow(mtw.Handle())
	mtw.SetBackground(uiCanvasBrush)
	vlayout := walk.NewVBoxLayout()
	vlayout.SetMargins(walk.Margins{5, 5, 5, 5})
	vlayout.SetSpacing(0)
	mtw.SetLayout(vlayout)
	if err = addProductHeader(mtw, &disposables); err != nil {
		return nil, err
	}
	mtw.Closing().Attach(func(canceled *bool, reason walk.CloseReason) {
		// "Close to tray" instead of exiting application
		*canceled = true
		if !noTrayAvailable {
			mtw.Hide()
		} else {
			win.ShowWindow(mtw.Handle(), win.SW_MINIMIZE)
		}
	})

	if mtw.tabs, err = walk.NewTabWidget(mtw); err != nil {
		return nil, err
	}

	if mtw.tunnelsPage, err = NewTunnelsPage(); err != nil {
		return nil, err
	}
	mtw.tabs.Pages().Add(mtw.tunnelsPage.TabPage)
	mtw.tunnelsPage.CreateToolbar()

	if mtw.logPage, err = NewLogPage(); err != nil {
		return nil, err
	}
	mtw.tabs.Pages().Add(mtw.logPage.TabPage)

	if mtw.settingsPage, err = NewSettingsPage(); err != nil {
		return nil, err
	}
	mtw.tabs.Pages().Add(mtw.settingsPage.TabPage)

	mtw.VisibleChanged().Attach(func() {
		if mtw.Visible() {
			mtw.tunnelsPage.updateConfView()
			win.SetForegroundWindow(mtw.Handle())
			win.BringWindowToTop(mtw.Handle())
			mtw.logPage.scrollToBottom()
		}
	})

	mtw.tunnelChangedCB = manager.IPCClientRegisterTunnelChange(mtw.onTunnelChange)
	globalState, _ := manager.IPCClientGlobalState()
	mtw.onTunnelChange(nil, manager.TunnelUnknown, globalState, nil)

	systemMenu := win.GetSystemMenu(mtw.Handle(), false)
	if systemMenu != 0 {
		win.InsertMenuItem(systemMenu, 0, true, &win.MENUITEMINFO{
			CbSize:     uint32(unsafe.Sizeof(win.MENUITEMINFO{})),
			FMask:      win.MIIM_ID | win.MIIM_STRING | win.MIIM_FTYPE,
			FType:      win.MIIM_STRING,
			DwTypeData: windows.StringToUTF16Ptr(l18n.Sprintf("&About WireHush…")),
			WID:        uint32(aboutWireHushCmd),
		})
		win.InsertMenuItem(systemMenu, 1, true, &win.MENUITEMINFO{
			CbSize: uint32(unsafe.Sizeof(win.MENUITEMINFO{})),
			FMask:  win.MIIM_TYPE,
			FType:  win.MFT_SEPARATOR,
		})
	}

	disposables.Spare()

	return mtw, nil
}

func addProductHeader(parent walk.Container, disposables *walk.Disposables) error {
	header, err := walk.NewComposite(parent)
	if err != nil {
		return err
	}
	headerLayout := walk.NewHBoxLayout()
	headerLayout.SetMargins(walk.Margins{14, 12, 14, 10})
	header.SetLayout(headerLayout)
	header.SetMinMaxSize(walk.Size{0, 62}, walk.Size{0, 62})
	applyDarkSurface(header, uiHeaderBrush)

	icon, err := loadLogoIcon(40)
	if err == nil {
		imageView, imageErr := walk.NewImageView(header)
		if imageErr != nil {
			return imageErr
		}
		imageView.SetMode(walk.ImageViewModeCenter)
		imageView.SetMinMaxSize(walk.Size{40, 40}, walk.Size{40, 40})
		if err := imageView.SetImage(icon); err != nil {
			return err
		}
	}

	labels, err := walk.NewComposite(header)
	if err != nil {
		return err
	}
	labelsLayout := walk.NewVBoxLayout()
	labelsLayout.SetMargins(walk.Margins{10, 0, 0, 0})
	labelsLayout.SetSpacing(0)
	labels.SetLayout(labelsLayout)
	title, err := walk.NewLabel(labels)
	if err != nil {
		return err
	}
	title.SetText(l18n.Sprintf("WireHush"))
	title.SetTextColor(uiTextColor)
	titleFont, err := walk.NewFont("Segoe UI Semibold", 13, 0)
	if err == nil {
		title.SetFont(titleFont)
		disposables.Add(titleFont)
	}
	subtitle, err := walk.NewLabel(labels)
	if err != nil {
		return err
	}
	subtitle.SetText(l18n.Sprintf("Private network control"))
	applyMutedText(subtitle)

	accent, err := walk.NewComposite(parent)
	if err != nil {
		return err
	}
	accent.SetMinMaxSize(walk.Size{0, 3}, walk.Size{0, 3})
	brush, err := walk.NewSolidColorBrush(walk.RGB(23, 195, 210))
	if err != nil {
		return err
	}
	accent.SetBackground(brush)
	disposables.Add(brush)
	return nil
}

func (mtw *ManageTunnelsWindow) Dispose() {
	if mtw.tunnelChangedCB != nil {
		mtw.tunnelChangedCB.Unregister()
		mtw.tunnelChangedCB = nil
	}
	mtw.FormBase.Dispose()
}

func (mtw *ManageTunnelsWindow) updateProgressIndicator(globalState manager.TunnelState) {
	pi := mtw.ProgressIndicator()
	if pi == nil {
		return
	}
	switch globalState {
	case manager.TunnelStopping, manager.TunnelStarting:
		pi.SetState(walk.PIIndeterminate)
	default:
		pi.SetState(walk.PINoProgress)
	}
	if icon, err := iconForState(globalState, 16); err == nil {
		if globalState == manager.TunnelStopped {
			icon = nil
		}
		pi.SetOverlayIcon(icon, textForState(globalState, false))
	}
}

func (mtw *ManageTunnelsWindow) onTunnelChange(tunnel *manager.Tunnel, state, globalState manager.TunnelState, err error) {
	mtw.Synchronize(func() {
		mtw.updateProgressIndicator(globalState)

		if err != nil && mtw.Visible() {
			errMsg := err.Error()
			if len(errMsg) > 0 && errMsg[len(errMsg)-1] != '.' {
				errMsg += "."
			}
			showWarningCustom(mtw, l18n.Sprintf("Tunnel Error"), l18n.Sprintf("%s\n\nPlease consult the log for more information.", errMsg))
		}
	})
}

func (mtw *ManageTunnelsWindow) UpdateFound() {
	if mtw.updatePage != nil {
		return
	}
	if IsAdmin {
		mtw.SetTitle(l18n.Sprintf("%s (out of date)", mtw.Title()))
	}
	updatePage, err := NewUpdatePage()
	if err == nil {
		mtw.updatePage = updatePage
		mtw.tabs.Pages().Add(updatePage.TabPage)
	}
}

func (mtw *ManageTunnelsWindow) WndProc(hwnd win.HWND, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case win.WM_QUERYENDSESSION:
		if lParam == win.ENDSESSION_CLOSEAPP {
			return win.TRUE
		}
	case win.WM_ENDSESSION:
		if lParam == win.ENDSESSION_CLOSEAPP && wParam == 1 {
			walk.App().Exit(198)
		}
	case win.WM_SYSCOMMAND:
		if wParam == aboutWireHushCmd {
			onAbout(mtw)
			return 0
		}
	case raiseMsg:
		if mtw.tunnelsPage == nil || mtw.tabs == nil {
			mtw.Synchronize(func() {
				mtw.SendMessage(msg, wParam, lParam)
			})
			return 0
		}
		if !mtw.Visible() {
			mtw.tunnelsPage.listView.SelectFirstActiveTunnel()
			mtw.tabs.SetCurrentIndex(0)
		}
		if mtw.updatePage != nil {
			mtw.tabs.SetCurrentIndex(mtw.tabs.Pages().Index(mtw.updatePage.TabPage))
		}
		raise(mtw.Handle())
		return 0
	case taskbarButtonCreatedMsg:
		ret := mtw.FormBase.WndProc(hwnd, msg, wParam, lParam)
		go func() {
			globalState, err := manager.IPCClientGlobalState()
			if err == nil {
				mtw.Synchronize(func() {
					mtw.updateProgressIndicator(globalState)
				})
			}
		}()
		return ret
	}

	return mtw.FormBase.WndProc(hwnd, msg, wParam, lParam)
}
