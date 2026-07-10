package main

import (
	"log"
	"os"
	"strconv"
	"time"
)

// relaunchPIDEnv carries the PID of the process being replaced during an
// update relaunch. RestartApplication sets it on the child it spawns so the
// child can wait for the old instance to fully exit — and thereby release the
// Wails single-instance lock — before its own Wails runtime tries to acquire
// that lock. Without this wait the freshly-downloaded binary would start while
// the old one still holds the lock, be treated as a second instance (it would
// just restore the old window and exit), and leave the OLD version running.
const relaunchPIDEnv = "PLEXCORD_RELAUNCH_PID"

// relaunchWaitTimeout bounds how long a relaunched instance waits for its
// predecessor to exit before starting anyway, so a stuck old process can never
// hang the update relaunch indefinitely.
const relaunchWaitTimeout = 15 * time.Second

// relaunchLockSettle is a short grace period after the old process disappears,
// giving the OS time to release the single-instance lock it held (a named
// mutex on Windows, a lock file/socket elsewhere) before this instance's Wails
// runtime tries to acquire it.
const relaunchLockSettle = 250 * time.Millisecond

// waitForPreviousInstanceExit blocks until the process identified by
// relaunchPIDEnv has exited (or relaunchWaitTimeout elapses). It must run
// before wails.Run acquires the single-instance lock. For a normal launch the
// env var is absent and this returns immediately.
func waitForPreviousInstanceExit() {
	raw := os.Getenv(relaunchPIDEnv)
	if raw == "" {
		return
	}
	// Clear it so this instance's own environment (and any future relaunch it
	// spawns) starts clean rather than inheriting a stale predecessor PID.
	if err := os.Unsetenv(relaunchPIDEnv); err != nil {
		log.Printf("Warning: failed to clear %s: %v", relaunchPIDEnv, err)
	}

	pid, err := strconv.Atoi(raw)
	if err != nil || pid <= 0 {
		return
	}

	log.Printf("Update relaunch: waiting for previous instance (pid %d) to exit", pid)
	deadline := time.Now().Add(relaunchWaitTimeout)
	for time.Now().Before(deadline) {
		if !processExists(pid) {
			time.Sleep(relaunchLockSettle)
			log.Printf("Previous instance exited; continuing startup")
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	log.Printf("Timed out waiting for previous instance (pid %d) to exit; starting anyway", pid)
}
