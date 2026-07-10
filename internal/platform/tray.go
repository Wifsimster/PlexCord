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
	OnShow func() // Called when the user asks to show/restore the window
	OnQuit func() // Called when the user asks to quit the application
}

// TrayManager manages the system tray icon and its menu.
type TrayManager struct {
	callbacks TrayCallbacks
	iconPNG   []byte // PNG icon bytes (macOS/Linux)
	iconICO   []byte // ICO icon bytes (Windows)
	tooltip   string
	mu        sync.Mutex
	running   bool
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
	systray.SetTooltip(tm.tooltip)

	mShow := systray.AddMenuItem("Show PlexCord", "Bring the PlexCord window to the foreground")
	mQuit := systray.AddMenuItem("Quit", "Quit PlexCord completely")

	mShow.Click(tm.handleShow)
	mQuit.Click(tm.handleQuit)

	// Left-clicking the tray icon also restores the window; right-click keeps
	// the default behavior of opening the menu.
	systray.SetOnClick(func(systray.IMenu) { tm.handleShow() })

	log.Printf("System tray: ready")
}

// onExit runs in the systray event loop when the tray is torn down.
func (tm *TrayManager) onExit() {
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
func (tm *TrayManager) SetTooltip(tooltip string) {
	tm.mu.Lock()
	running := tm.running
	tm.tooltip = tooltip
	tm.mu.Unlock()

	if running {
		systray.SetTooltip(tooltip)
	}
}
