//go:build windows

package main

import "golang.org/x/sys/windows"

// processExists reports whether the process with the given PID is still
// running. It opens the process with SYNCHRONIZE rights and probes its wait
// state: WAIT_TIMEOUT means the process is still alive, WAIT_OBJECT_0 means it
// has terminated. A failed open (the process is gone, so its handle can no
// longer be obtained) is treated as "not running".
func processExists(pid int) bool {
	handle, err := windows.OpenProcess(windows.SYNCHRONIZE, false, uint32(pid))
	if err != nil {
		return false
	}
	defer func() { _ = windows.CloseHandle(handle) }()

	event, err := windows.WaitForSingleObject(handle, 0)
	if err != nil {
		return false
	}
	return event == uint32(windows.WAIT_TIMEOUT)
}
