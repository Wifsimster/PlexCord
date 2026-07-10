//go:build windows

package main

import (
	"log"

	"golang.org/x/sys/windows"
)

// processExists reports whether the process with the given PID is still
// running. It opens the process with SYNCHRONIZE rights and probes its wait
// state: WAIT_TIMEOUT means the process is still alive, WAIT_OBJECT_0 means it
// has terminated. A failed open (the process is gone, so its handle can no
// longer be obtained) is treated as "not running".
func processExists(pid int) bool {
	if pid <= 0 {
		return false
	}
	// Windows process IDs are DWORDs (uint32) and pid is non-negative here.
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
