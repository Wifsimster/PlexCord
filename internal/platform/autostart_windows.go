//go:build windows

package platform

import (
	"log"

	"golang.org/x/sys/windows/registry"
)

const (
	runKeyPath = `Software\Microsoft\Windows\CurrentVersion\Run`
)

// IsEnabled checks if PlexCord is in the Windows Run registry key.
func (m *AutoStartManager) IsEnabled() bool {
	key, err := registry.OpenKey(registry.CURRENT_USER, runKeyPath, registry.QUERY_VALUE)
	if err != nil {
		return false
	}
	defer func() {
		if err := key.Close(); err != nil {
			log.Printf("Warning: Failed to close registry key: %v", err)
		}
	}()

	_, _, err = key.GetStringValue(m.appName)
	return err == nil
}

// Enable adds PlexCord to the Windows Run registry key.
func (m *AutoStartManager) Enable() error {
	if m.IsEnabled() {
		log.Printf("Auto-start already enabled")
		return nil
	}

	key, _, err := registry.CreateKey(registry.CURRENT_USER, runKeyPath, registry.SET_VALUE)
	if err != nil {
		log.Printf("ERROR: Failed to open registry key: %v", err)
		return err
	}
	defer func() {
		if err := key.Close(); err != nil {
			log.Printf("Warning: Failed to close registry key: %v", err)
		}
	}()

	// Use quoted path to handle spaces in the executable path
	if err := key.SetStringValue(m.appName, `"`+m.executable+`"`); err != nil {
		log.Printf("ERROR: Failed to set registry value: %v", err)
		return err
	}

	log.Printf("Auto-start enabled successfully")
	return nil
}

// Disable removes PlexCord from the Windows Run registry key.
func (m *AutoStartManager) Disable() error {
	if !m.IsEnabled() {
		log.Printf("Auto-start already disabled")
		return nil
	}

	key, err := registry.OpenKey(registry.CURRENT_USER, runKeyPath, registry.SET_VALUE)
	if err != nil {
		log.Printf("ERROR: Failed to open registry key: %v", err)
		return err
	}
	defer func() {
		if err := key.Close(); err != nil {
			log.Printf("Warning: Failed to close registry key: %v", err)
		}
	}()

	if err := key.DeleteValue(m.appName); err != nil {
		log.Printf("ERROR: Failed to delete registry value: %v", err)
		return err
	}

	log.Printf("Auto-start disabled successfully")
	return nil
}
