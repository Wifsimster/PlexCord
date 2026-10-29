//go:build !windows

package platform

func SyncInstalledVersion(string) error { return nil }
