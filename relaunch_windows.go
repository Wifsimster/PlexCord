//go:build windows

package main

import (
	"log"

	"golang.org/x/sys/windows"
)

func processExists(pid int) bool {
	if pid <= 0 {
		return false
	}
	handle, err := windows.OpenProcess(windows.SYNCHRONIZE, false, uint32(pid)) //nolint:gosec // G115: pid guarded non-negative above; PIDs fit in uint32
	if err != nil {
		return false
	}
	defer func() {
		if cerr := windows.CloseHandle(handle); cerr != nil {
			log.Printf("Warning: failed to close process handle: %v", cerr)
		}
	}()

	event, err := windows.WaitForSingleObject(handle, 0)
	if err != nil {
		return false
	}
	return event == uint32(windows.WAIT_TIMEOUT)
}
