//go:build windows

package platform

import (
	"log"
	"strings"

	"golang.org/x/sys/windows/registry"
)

// uninstallKeyPath is where the NSIS installer registers PlexCord for
// "Apps & features" (build/windows/installer/project.nsi). It lives under
// HKCU because PlexCord installs per user.
const uninstallKeyPath = `Software\Microsoft\Windows\CurrentVersion\Uninstall\PlexCord`

// SyncInstalledVersion brings the DisplayVersion shown in "Apps & features"
// in line with the version actually installed on disk.
//
// The installer writes that value once, but PlexCord then updates itself by
// replacing its executable in place — no installer runs, so without this the
// entry would keep advertising whatever version was last installed by hand.
//
// A missing key means this is not an installed copy (the portable executable,
// or a build run from a checkout): that is the normal case, not an error, and
// nothing is written.
func SyncInstalledVersion(version string) error {
	version = strings.TrimPrefix(version, "v")
	if version == "" {
		return nil
	}

	key, err := registry.OpenKey(registry.CURRENT_USER, uninstallKeyPath, registry.QUERY_VALUE|registry.SET_VALUE)
	if err != nil {
		// Not an installed copy, or the key is not ours to write.
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
