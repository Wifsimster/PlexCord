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

// Enable creates a LaunchAgent plist to start PlexCord on login.
func (m *AutoStartManager) Enable() error {
	if m.IsEnabled() {
		log.Printf("Auto-start already enabled")
		return nil
	}

	plistPath := m.getLaunchAgentPath()
	if plistPath == "" {
		return fmt.Errorf("could not determine home directory")
	}

	// Ensure LaunchAgents directory exists
	if err := os.MkdirAll(filepath.Dir(plistPath), 0755); err != nil {
		log.Printf("ERROR: Failed to create LaunchAgents directory: %v", err)
		return err
	}

	// Create plist content
	plistContent := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.plexcord.app</string>
    <key>ProgramArguments</key>
    <array>
        <string>%s</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <false/>
</dict>
</plist>
`, m.executable)

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
