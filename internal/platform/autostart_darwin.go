//go:build darwin

package platform

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func (m *AutoStartManager) getLaunchAgentPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(homeDir, "Library", "LaunchAgents", "com.plexcord.app.plist")
}

func (m *AutoStartManager) IsEnabled() bool {
	plistPath := m.getLaunchAgentPath()
	if plistPath == "" {
		return false
	}
	_, err := os.Stat(plistPath)
	return err == nil
}

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

func (m *AutoStartManager) Enable() error {
	plistPath := m.getLaunchAgentPath()
	if plistPath == "" {
		return fmt.Errorf("could not determine home directory")
	}

	plistContent := m.launchAgentPlist()
	if current, err := os.ReadFile(plistPath); err == nil && string(current) == plistContent { //nolint:gosec
		log.Printf("Auto-start already enabled")
		return nil
	}

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
