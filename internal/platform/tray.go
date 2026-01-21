// Package platform provides OS-specific abstractions for system tray,
// auto-start, and other platform-specific functionality.
//
// NOTE: Full system tray support requires Wails v3 or a separate process
// approach due to threading conflicts between systray library and Wails v2.
// This implementation provides a basic abstraction that can be enhanced
// when Wails v3 is adopted.
package platform

import (
	"log"
	"sync"
)

// TrayStatus represents the current connection status shown in tray
type TrayStatus string

const (
	// TrayStatusConnected indicates both Plex and Discord are connected
	TrayStatusConnected TrayStatus = "connected"
	// TrayStatusDisconnected indicates one or both services are disconnected
	TrayStatusDisconnected TrayStatus = "disconnected"
	// TrayStatusError indicates an error state
	TrayStatusError TrayStatus = "error"
)

// TrayCallbacks holds callback functions for tray events
type TrayCallbacks struct {
	OnShow func() // Called when user clicks to show window
	OnQuit func() // Called when user clicks quit
}

// TrayManager manages the system tray icon and menu.
// NOTE: This is a placeholder implementation for Wails v2.
// Full tray functionality requires Wails v3 or platform-specific code.
type TrayManager struct {
	callbacks TrayCallbacks
	status    TrayStatus
	tooltip   string
	mu        sync.Mutex
	running   bool
}

// NewTrayManager creates a new TrayManager with the provided callbacks.
func NewTrayManager(callbacks TrayCallbacks) *TrayManager {
	return &TrayManager{
		callbacks: callbacks,
		status:    TrayStatusDisconnected,
		tooltip:   "PlexCord",
		running:   false,
	}
}

// Start initializes the system tray (placeholder for Wails v2).
// In Wails v2, we rely on HideWindowOnClose for minimize-to-tray behavior.
// Full tray icon support will be added when migrating to Wails v3.
func (tm *TrayManager) Start() {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if tm.running {
		return
	}

	tm.running = true
	log.Printf("System tray: Started (window management mode)")
	// NOTE: Full systray implementation deferred to Wails v3
	// The HideWindowOnClose option in main.go provides basic minimize-to-tray
}

// Stop stops the system tray
func (tm *TrayManager) Stop() {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if !tm.running {
		return
	}

	tm.running = false
	log.Printf("System tray: Stopped")
}

// IsRunning returns whether the tray is running
func (tm *TrayManager) IsRunning() bool {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	return tm.running
}

// SetStatus updates the tray status.
// In full implementation, this would change the tray icon.
func (tm *TrayManager) SetStatus(status TrayStatus) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tm.status = status
	log.Printf("System tray: Status changed to %s", status)
	// NOTE: Icon changes deferred to Wails v3
}

// SetTooltip updates the tray tooltip.
// In full implementation, this would update the tray tooltip text.
func (tm *TrayManager) SetTooltip(tooltip string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tm.tooltip = tooltip
	// NOTE: Tooltip changes deferred to Wails v3
}

// GetStatus returns the current tray status
func (tm *TrayManager) GetStatus() TrayStatus {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	return tm.status
}
