//go:build windows

package platform

import (
	"log"
	"strings"

	"golang.org/x/sys/windows/registry"
)

const uninstallKeyPath = `Software\Microsoft\Windows\CurrentVersion\Uninstall\PlexCord`

func SyncInstalledVersion(version string) error {
	version = strings.TrimPrefix(version, "v")
	if version == "" {
		return nil
	}

	key, err := registry.OpenKey(registry.CURRENT_USER, uninstallKeyPath, registry.QUERY_VALUE|registry.SET_VALUE)
	if err != nil {
		return nil
	}
	defer func() {
		if err := key.Close(); err != nil {
			log.Printf("Warning: Failed to close registry key: %v", err)
		}
	}()

	if current, _, err := key.GetStringValue("DisplayVersion"); err == nil && current == version {
		return nil
	}

	if err := key.SetStringValue("DisplayVersion", version); err != nil {
		return err
	}

	log.Printf("Installed version registration updated to %s", version)
	return nil
}
