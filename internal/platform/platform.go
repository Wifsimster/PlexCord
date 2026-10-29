package platform

import "runtime"

const (
	Windows = "windows"
	MacOS   = "darwin"
	Linux   = "linux"
)

func GetPlatform() string {
	return runtime.GOOS
}
