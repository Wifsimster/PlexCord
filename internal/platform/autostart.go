package platform

import (
	"log"
	"os"
	"path/filepath"
)

type AutoStartManager struct {
	appName    string
	executable string
}

func NewAutoStartManager() *AutoStartManager {
	executable, err := os.Executable()
	if err != nil {
		log.Printf("Warning: Could not determine executable path: %v", err)
		executable = "plexcord"
	}

	executable, err = filepath.EvalSymlinks(executable)
	if err != nil {
		log.Printf("Warning: Could not resolve executable symlinks: %v", err)
	}

	return &AutoStartManager{
		appName:    "PlexCord",
		executable: executable,
	}
}

func (m *AutoStartManager) SetEnabled(enabled bool) error {
	if enabled {
		return m.Enable()
	}
	return m.Disable()
}

const AutoStartFlag = "--autostart"

func (m *AutoStartManager) EnsureRegistered() error {
	if !m.IsEnabled() {
		return nil
	}
	return m.Enable()
}
