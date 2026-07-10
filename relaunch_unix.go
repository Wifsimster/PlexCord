//go:build !windows

package main

import (
	"errors"
	"syscall"
)

// processExists reports whether the process with the given PID is still
// running. Signal 0 performs the kernel's permission/existence checks without
// delivering an actual signal: a nil error means the process is alive, EPERM
// means it exists but is owned by another user (still alive), and ESRCH means
// it is gone.
func processExists(pid int) bool {
	err := syscall.Kill(pid, 0)
	if err == nil {
		return true
	}
	return errors.Is(err, syscall.EPERM)
}
