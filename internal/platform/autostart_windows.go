//go:build windows

package platform

import (
	"log"

	"golang.org/x/sys/windows/registry"
)

const (
	runKeyPath = `Software\Microsoft\Windows\CurrentVersion\Run`
)

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

func (m *AutoStartManager) runCommand() string {
	return `"` + m.executable + `" ` + AutoStartFlag
}

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
