//go:build linux

package platform

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

// getDesktopFilePath returns the path to the XDG autostart .desktop file.
func (m *AutoStartManager) getDesktopFilePath() string {
	// Check XDG_CONFIG_HOME first, fall back to ~/.config
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		configDir = filepath.Join(homeDir, ".config")
	}
	return filepath.Join(configDir, "autostart", "plexcord.desktop")
}

// IsEnabled checks if PlexCord .desktop file exists in autostart.
func (m *AutoStartManager) IsEnabled() bool {
	desktopPath := m.getDesktopFilePath()
	if desktopPath == "" {
		return false
	}
	_, err := os.Stat(desktopPath)
	return err == nil
}

// desktopEntry is the .desktop file the session autostarts. Exec carries
// AutoStartFlag so the launched process knows the session started it at login
// rather than the user opening PlexCord.
func (m *AutoStartManager) desktopEntry() string {
	return fmt.Sprintf(`[Desktop Entry]
Type=Application
Name=PlexCord
Comment=Plex to Discord Rich Presence
Exec=%s %s
Icon=plexcord
Terminal=false
Categories=AudioVideo;Audio;
X-GNOME-Autostart-enabled=true
`, m.executable, AutoStartFlag)
}

// Enable creates an XDG autostart .desktop file.
//
// An existing file is rewritten unless it already matches, so an entry left by
// an older version (or by the executable at a previous path) is brought up to
// date instead of being left as it was.
func (m *AutoStartManager) Enable() error {
	desktopPath := m.getDesktopFilePath()
	if desktopPath == "" {
		return fmt.Errorf("could not determine config directory")
	}

	desktopContent := m.desktopEntry()
	//nolint:gosec // desktopPath is derived from the user config dir, not user input
	if current, err := os.ReadFile(desktopPath); err == nil && string(current) == desktopContent {
		log.Printf("Auto-start already enabled")
		return nil
	}

	// Ensure autostart directory exists
	if err := os.MkdirAll(filepath.Dir(desktopPath), 0750); err != nil {
		log.Printf("ERROR: Failed to create autostart directory: %v", err)
		return err
	}

	if err := os.WriteFile(desktopPath, []byte(desktopContent), 0600); err != nil {
		log.Printf("ERROR: Failed to write .desktop file: %v", err)
		return err
	}

	log.Printf("Auto-start enabled successfully")
	return nil
}

// Disable removes the XDG autostart .desktop file.
func (m *AutoStartManager) Disable() error {
	if !m.IsEnabled() {
		log.Printf("Auto-start already disabled")
		return nil
	}

	desktopPath := m.getDesktopFilePath()
	if desktopPath == "" {
		return fmt.Errorf("could not determine config directory")
	}

	if err := os.Remove(desktopPath); err != nil {
		log.Printf("ERROR: Failed to remove .desktop file: %v", err)
		return err
	}

	log.Printf("Auto-start disabled successfully")
	return nil
}
