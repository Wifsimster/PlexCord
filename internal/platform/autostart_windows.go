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

// runCommand is the command line registered under the Run key: the quoted
// executable path (quoted so spaces in it do not split the command) followed by
// AutoStartFlag, which tells the launched process it was started by Windows at
// login rather than by the user.
func (m *AutoStartManager) runCommand() string {
	return `"` + m.executable + `" ` + AutoStartFlag
}

// Enable adds PlexCord to the Windows Run registry key.
//
// An entry that is already there is rewritten unless it matches exactly, so a
// registration left by an older version (or by the executable at a previous
// path) is brought up to date instead of being left as it was.
func (m *AutoStartManager) Enable() error {
	key, _, err := registry.CreateKey(registry.CURRENT_USER, runKeyPath, registry.SET_VALUE|registry.QUERY_VALUE)
	if err != nil {
		log.Printf("ERROR: Failed to open registry key: %v", err)
		return err
	}
	defer func() {
		if err := key.Close(); err != nil {
			log.Printf("Warning: Failed to close registry key: %v", err)
		}
	}()

	command := m.runCommand()
	if current, _, err := key.GetStringValue(m.appName); err == nil && current == command {
		log.Printf("Auto-start already enabled")
		return nil
	}

	if err := key.SetStringValue(m.appName, command); err != nil {
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
