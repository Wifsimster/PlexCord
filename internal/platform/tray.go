package platform

import (
	"log"
	"runtime"
	"sync"

	"github.com/energye/systray"
)

type TrayCallbacks struct {
	OnShow   func()
	OnQuit   func()
	OnUpdate func()
}

type UpdateNotice struct {
	Label      string
	Tooltip    string
	Actionable bool
}

func (n UpdateNotice) active() bool {
	return n.Label != ""
}

type TrayIcons struct {
	PNG       []byte
	ICO       []byte
	UpdatePNG []byte
	UpdateICO []byte
}

type TrayManager struct {
	callbacks  TrayCallbacks
	icons      TrayIcons
	tooltip    string
	notice     UpdateNotice
	noticeItem *systray.MenuItem
	mu         sync.Mutex
	running    bool
	menuReady  bool
}

func NewTrayManager(callbacks TrayCallbacks, icons TrayIcons) *TrayManager {
	return &TrayManager{
		callbacks: callbacks,
		icons:     icons,
		tooltip:   "PlexCord",
	}
}

func (tm *TrayManager) Start() {
	tm.mu.Lock()
	if tm.running {
		tm.mu.Unlock()
		return
	}
	tm.running = true
	tm.mu.Unlock()

	go systray.Run(tm.onReady, tm.onExit)
}

func (tm *TrayManager) onReady() {
	if icon := tm.icon(false); len(icon) > 0 {
		systray.SetIcon(icon)
	}
	systray.SetTitle("PlexCord")

	mShow := systray.AddMenuItem("Show PlexCord", "Bring the PlexCord window to the foreground")

	systray.AddSeparator()
	noticeItem := systray.AddMenuItem("", "")
	noticeItem.Hide()

	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "Quit PlexCord completely")

	mShow.Click(tm.handleShow)
	noticeItem.Click(tm.handleUpdate)
	mQuit.Click(tm.handleQuit)

	systray.SetOnClick(func(systray.IMenu) { tm.handleShow() })

	tm.mu.Lock()
	tm.noticeItem = noticeItem
	tm.menuReady = true
	notice, tooltip := tm.notice, tm.effectiveTooltipLocked()
	tm.mu.Unlock()

	systray.SetTooltip(tooltip)
	if icon := tm.icon(notice.active()); len(icon) > 0 {
		systray.SetIcon(icon)
	}
	applyUpdateNotice(noticeItem, notice)

	log.Printf("System tray: ready")
}

func applyUpdateNotice(item *systray.MenuItem, notice UpdateNotice) {
	if item == nil {
		return
	}
	if !notice.active() {
		item.Hide()
		return
	}
	item.SetTitle(notice.Label)
	item.SetTooltip(notice.Tooltip)
	if notice.Actionable {
		item.Enable()
	} else {
		item.Disable()
	}
	item.Show()
}

func (tm *TrayManager) onExit() {
	tm.mu.Lock()
	tm.menuReady = false
	tm.noticeItem = nil
	tm.mu.Unlock()

	log.Printf("System tray: exited")
}

func (tm *TrayManager) handleShow() {
	if tm.callbacks.OnShow != nil {
		tm.callbacks.OnShow()
	}
}

func (tm *TrayManager) handleQuit() {
	if tm.callbacks.OnQuit != nil {
		tm.callbacks.OnQuit()
	}
}

func (tm *TrayManager) handleUpdate() {
	if tm.callbacks.OnUpdate != nil {
		tm.callbacks.OnUpdate()
	}
}

func (tm *TrayManager) icon(badged bool) []byte {
	png, ico := tm.icons.PNG, tm.icons.ICO
	if badged {
		if len(tm.icons.UpdatePNG) > 0 {
			png = tm.icons.UpdatePNG
		}
		if len(tm.icons.UpdateICO) > 0 {
			ico = tm.icons.UpdateICO
		}
	}
	if runtime.GOOS == "windows" && len(ico) > 0 {
		return ico
	}
	return png
}

func (tm *TrayManager) Stop() {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if !tm.running {
		return
	}
	tm.running = false
	systray.Quit()
	log.Printf("System tray: stopped")
}

func (tm *TrayManager) IsRunning() bool {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	return tm.running
}

func (tm *TrayManager) SetTooltip(tooltip string) {
	tm.mu.Lock()
	tm.tooltip = tooltip
	ready, effective := tm.menuReady, tm.effectiveTooltipLocked()
	tm.mu.Unlock()

	if ready {
		systray.SetTooltip(effective)
	}
}

func (tm *TrayManager) SetUpdateNotice(notice UpdateNotice) {
	tm.mu.Lock()
	tm.notice = notice
	item, ready, tooltip := tm.noticeItem, tm.menuReady, tm.effectiveTooltipLocked()
	tm.mu.Unlock()

	if !ready {
		return
	}
	systray.SetTooltip(tooltip)
	if icon := tm.icon(notice.active()); len(icon) > 0 {
		systray.SetIcon(icon)
	}
	applyUpdateNotice(item, notice)
}

func (tm *TrayManager) UpdateNotice() UpdateNotice {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	return tm.notice
}

func (tm *TrayManager) effectiveTooltipLocked() string {
	if tm.notice.active() && tm.notice.Tooltip != "" {
		return tm.notice.Tooltip
	}
	return tm.tooltip
}
