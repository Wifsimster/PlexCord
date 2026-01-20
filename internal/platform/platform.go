package platform

import "runtime"

// OS platform types
const (
	Windows = "windows"
	MacOS   = "darwin"
	Linux   = "linux"
)

// GetPlatform returns the current operating system platform
func GetPlatform() string {
	return runtime.GOOS
}
