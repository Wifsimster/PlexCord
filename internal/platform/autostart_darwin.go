//go:build darwin

package platform

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

// getLaunchAgentPath returns the path to the LaunchAgent plist file.
func (m *AutoStartManager) getLaunchAgentPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(homeDir, "Library", "LaunchAgents", "com.plexcord.app.plist")
}

// IsEnabled checks if PlexCord LaunchAgent plist exists.
func (m *AutoStartManager) IsEnabled() bool {
	plistPath := m.getLaunchAgentPath()
	if plistPath == "" {
		return false
	}
	_, err := os.Stat(plistPath)
	return err == nil
}

// launchAgentPlist is the LaunchAgent loaded at login. Its arguments carry
// AutoStartFlag so the launched process knows launchd started it rather than
// the user opening PlexCord.
func (m *AutoStartManager) launchAgentPlist() string {
	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.plexcord.app</string>
    <key>ProgramArguments</key>
    <array>
        <string>%s</string>
        <string>%s</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <false/>
</dict>
</plist>
`, m.executable, AutoStartFlag)
}

// Enable creates a LaunchAgent plist to start PlexCord on login.
//
// An existing plist is rewritten unless it already matches, so an agent left by
// an older version (or by the executable at a previous path) is brought up to
// date instead of being left as it was.
func (m *AutoStartManager) Enable() error {
	plistPath := m.getLaunchAgentPath()
	if plistPath == "" {
		return fmt.Errorf("could not determine home directory")
	}

	plistContent := m.launchAgentPlist()
	//nolint:gosec // plistPath is derived from the user home dir, not user input
	if current, err := os.ReadFile(plistPath); err == nil && string(current) == plistContent {
		log.Printf("Auto-start already enabled")
		return nil
	}

	// Ensure LaunchAgents directory exists
	if err := os.MkdirAll(filepath.Dir(plistPath), 0755); err != nil {
		log.Printf("ERROR: Failed to create LaunchAgents directory: %v", err)
		return err
	}

	if err := os.WriteFile(plistPath, []byte(plistContent), 0644); err != nil {
		log.Printf("ERROR: Failed to write LaunchAgent plist: %v", err)
		return err
	}

	log.Printf("Auto-start enabled successfully")
	return nil
}

// Disable removes the LaunchAgent plist.
func (m *AutoStartManager) Disable() error {
	if !m.IsEnabled() {
		log.Printf("Auto-start already disabled")
		return nil
	}

	plistPath := m.getLaunchAgentPath()
	if plistPath == "" {
		return fmt.Errorf("could not determine home directory")
	}

	if err := os.Remove(plistPath); err != nil {
		log.Printf("ERROR: Failed to remove LaunchAgent plist: %v", err)
		return err
	}

	log.Printf("Auto-start disabled successfully")
	return nil
}
