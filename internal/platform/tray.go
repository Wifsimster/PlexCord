// Package platform provides OS-specific abstractions for system tray,
// auto-start, and other platform-specific functionality.
//
// The system tray is backed by github.com/energye/systray, a fork of
// getlantern/systray that coexists with Wails v2's main event loop (the
// upstream library requires ownership of the main thread on macOS). The tray
// icon is what makes "keep running in the background on close" usable: it is
// the visible affordance for restoring the window or quitting the app.
package platform

import (
	"log"
	"runtime"
	"sync"

	"github.com/energye/systray"
)

// TrayCallbacks holds callback functions for tray menu events.
type TrayCallbacks struct {
	OnShow   func() // Called when the user asks to show/restore the window
	OnQuit   func() // Called when the user asks to quit the application
	OnUpdate func() // Called when the user clicks the update notice (see SetUpdateNotice)
}

// UpdateNotice is the tray's rendering of a pending application update: a menu
// item that appears above Quit, plus the icon tooltip. It is the only update
// notification that reaches a user running PlexCord minimized to the tray —
// the in-window toast is invisible there.
//
// The zero value means "no update pending" and hides the menu item.
type UpdateNotice struct {
	// Label is the menu item text, e.g. "Restart to update to v1.5.0".
	Label string
	// Tooltip replaces the tray icon tooltip while the notice is active.
	Tooltip string
	// Actionable reports whether clicking the item does something (the
	// OnUpdate callback). A notice that is merely informational — a download
	// in progress — renders disabled.
	Actionable bool
}

// active reports whether the notice should be shown.
func (n UpdateNotice) active() bool {
	return n.Label != ""
}

// TrayManager manages the system tray icon and its menu.
type TrayManager struct {
	callbacks  TrayCallbacks
	iconPNG    []byte // PNG icon bytes (macOS/Linux)
	iconICO    []byte // ICO icon bytes (Windows)
	tooltip    string
	notice     UpdateNotice      // desired update notice, applied on/after onReady
	noticeItem *systray.MenuItem // nil until the menu is built
	mu         sync.Mutex
	running    bool
	// menuReady distinguishes "Start was called" from "systray has built the
	// menu": between the two, calls into systray are dropped on the floor, so
	// tooltip and notice changes are stored and replayed in onReady instead.
	menuReady bool
}

// NewTrayManager creates a new TrayManager with the provided callbacks and
// icon data. iconPNG is used on macOS/Linux and iconICO on Windows; either
// may be empty, in which case the icon is simply not set.
func NewTrayManager(callbacks TrayCallbacks, iconPNG, iconICO []byte) *TrayManager {
	return &TrayManager{
		callbacks: callbacks,
		iconPNG:   iconPNG,
		iconICO:   iconICO,
		tooltip:   "PlexCord",
	}
}

// Start creates the system tray icon and menu. It is safe to call once; a
// second call while already running is a no-op.
//
// systray.Run is launched on its own goroutine so it does not block the Wails
// main loop. The energye fork drives its own message loop on Windows and a
// DBus StatusNotifierItem on Linux; on macOS it attaches to the shared
// NSApplication.
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

// onReady builds the tray icon and menu once systray has initialized.
func (tm *TrayManager) onReady() {
	if icon := tm.icon(); len(icon) > 0 {
		systray.SetIcon(icon)
	}
	systray.SetTitle("PlexCord")

	mShow := systray.AddMenuItem("Show PlexCord", "Bring the PlexCord window to the foreground")

	// The update notice is built up front and hidden: systray has no API for
	// inserting an item into an existing menu, so the slot has to exist before
	// there is anything to put in it.
	systray.AddSeparator()
	noticeItem := systray.AddMenuItem("", "")
	noticeItem.Hide()

	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "Quit PlexCord completely")

	mShow.Click(tm.handleShow)
	noticeItem.Click(tm.handleUpdate)
	mQuit.Click(tm.handleQuit)

	// Left-clicking the tray icon also restores the window; right-click keeps
	// the default behavior of opening the menu.
	systray.SetOnClick(func(systray.IMenu) { tm.handleShow() })

	// Publish the item and replay whatever state was requested before the menu
	// existed — an update found during startup can land before this runs.
	tm.mu.Lock()
	tm.noticeItem = noticeItem
	tm.menuReady = true
	notice, tooltip := tm.notice, tm.effectiveTooltipLocked()
	tm.mu.Unlock()

	systray.SetTooltip(tooltip)
	applyUpdateNotice(noticeItem, notice)

	log.Printf("System tray: ready")
}

// applyUpdateNotice pushes a notice onto the menu item. Split out so the state
// transitions can be exercised without a live systray.
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

// onExit runs in the systray event loop when the tray is torn down.
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

// icon returns the icon bytes appropriate for the current OS.
func (tm *TrayManager) icon() []byte {
	if runtime.GOOS == "windows" && len(tm.iconICO) > 0 {
		return tm.iconICO
	}
	return tm.iconPNG
}

// Stop removes the system tray icon and stops its event loop.
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

// IsRunning reports whether the tray has been started.
func (tm *TrayManager) IsRunning() bool {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	return tm.running
}

// SetTooltip updates the tray tooltip text. Takes effect on the next Start if
// the tray is not yet running.
//
// An active update notice owns the tooltip, so this only reaches the tray once
// the notice is cleared — the pending update is the more important thing to say.
func (tm *TrayManager) SetTooltip(tooltip string) {
	tm.mu.Lock()
	tm.tooltip = tooltip
	ready, effective := tm.menuReady, tm.effectiveTooltipLocked()
	tm.mu.Unlock()

	if ready {
		systray.SetTooltip(effective)
	}
}

// SetUpdateNotice shows (or, with the zero UpdateNotice, hides) the update item
// in the tray menu and points the icon tooltip at it. Safe to call before the
// tray has finished starting: the notice is stored and applied in onReady.
func (tm *TrayManager) SetUpdateNotice(notice UpdateNotice) {
	tm.mu.Lock()
	tm.notice = notice
	item, ready, tooltip := tm.noticeItem, tm.menuReady, tm.effectiveTooltipLocked()
	tm.mu.Unlock()

	if !ready {
		return
	}
	systray.SetTooltip(tooltip)
	applyUpdateNotice(item, notice)
}

// UpdateNotice returns the notice currently displayed (the zero value when
// none is).
func (tm *TrayManager) UpdateNotice() UpdateNotice {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	return tm.notice
}

// effectiveTooltipLocked is the tooltip the tray should show: the update
// notice's when there is one, the base tooltip otherwise. Caller holds mu.
func (tm *TrayManager) effectiveTooltipLocked() string {
	if tm.notice.active() && tm.notice.Tooltip != "" {
		return tm.notice.Tooltip
	}
	return tm.tooltip
}
