package main

import (
	"log"
	"os"
	"strconv"
	"time"
)

const relaunchPIDEnv = "PLEXCORD_RELAUNCH_PID"

const relaunchWaitTimeout = 15 * time.Second

const relaunchLockSettle = 250 * time.Millisecond

func waitForPreviousInstanceExit() bool {
	raw := os.Getenv(relaunchPIDEnv)
	if raw == "" {
		return false
	}
	if err := os.Unsetenv(relaunchPIDEnv); err != nil {
		log.Printf("Warning: failed to clear %s: %v", relaunchPIDEnv, err)
	}

	pid, err := strconv.Atoi(raw)
	if err != nil || pid <= 0 {
		return true
	}

	log.Printf("Update relaunch: waiting for previous instance (pid %d) to exit", pid)
	deadline := time.Now().Add(relaunchWaitTimeout)
	for time.Now().Before(deadline) {
		if !processExists(pid) {
			time.Sleep(relaunchLockSettle)
			log.Printf("Previous instance exited; continuing startup")
			return true
		}
		time.Sleep(100 * time.Millisecond)
	}
	log.Printf("Timed out waiting for previous instance (pid %d) to exit; starting anyway", pid)
	return true
}
