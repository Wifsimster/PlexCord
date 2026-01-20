// Package platform provides OS-specific abstractions for system tray,
// auto-start, and other platform-specific functionality.
package platform

import (
	"log"
	"os"
	"path/filepath"
)

// AutoStartManager handles auto-start registration for the application.
// Uses native OS mechanisms:
// - Windows: Registry (HKCU\Software\Microsoft\Windows\CurrentVersion\Run)
// - macOS: LaunchAgents (~/Library/LaunchAgents/)
// - Linux: XDG Autostart (~/.config/autostart/)
type AutoStartManager struct {
	appName    string
	executable string
}

// NewAutoStartManager creates a new AutoStartManager for PlexCord.
func NewAutoStartManager() *AutoStartManager {
	executable, err := os.Executable()
	if err != nil {
		log.Printf("Warning: Could not determine executable path: %v", err)
		executable = "plexcord"
	}

	// Resolve any symlinks to get the real path
	executable, err = filepath.EvalSymlinks(executable)
	if err != nil {
		log.Printf("Warning: Could not resolve executable symlinks: %v", err)
	}

	return &AutoStartManager{
		appName:    "PlexCord",
		executable: executable,
	}
}

// SetEnabled enables or disables auto-start based on the provided value.
func (m *AutoStartManager) SetEnabled(enabled bool) error {
	if enabled {
		return m.Enable()
	}
	return m.Disable()
}
