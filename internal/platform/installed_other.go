//go:build !windows

package platform

// SyncInstalledVersion is a no-op outside Windows: macOS and Linux have no
// per-user "Apps & features" registration to keep in step with a self-update.
func SyncInstalledVersion(string) error { return nil }
