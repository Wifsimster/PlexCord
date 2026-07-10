package main

import (
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"testing"
	"time"
)

// deadPID starts a helper process that exits immediately, reaps it, and returns
// its now-defunct PID. Skips the test if a helper cannot be started.
func deadPID(t *testing.T) int {
	t.Helper()
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "exit")
	} else {
		cmd = exec.Command("sh", "-c", "exit 0")
	}
	if err := cmd.Start(); err != nil {
		t.Skipf("could not start helper process: %v", err)
	}
	pid := cmd.Process.Pid
	_ = cmd.Wait() // reap it so the PID is fully released
	return pid
}

// TestProcessExistsSelf verifies the current process is reported as alive.
func TestProcessExistsSelf(t *testing.T) {
	if !processExists(os.Getpid()) {
		t.Fatalf("processExists(self) = false, want true")
	}
}

// TestProcessExistsDead verifies a process that has already exited is reported
// as not running. This underpins the relaunch wait: the new instance must be
// able to observe the old one going away.
func TestProcessExistsDead(t *testing.T) {
	pid := deadPID(t)
	if processExists(pid) {
		t.Fatalf("processExists(dead pid %d) = true, want false", pid)
	}
}

// TestWaitForPreviousInstanceExitNoEnv returns immediately when the relaunch
// marker is absent (the normal-launch path).
func TestWaitForPreviousInstanceExitNoEnv(t *testing.T) {
	_ = os.Unsetenv(relaunchPIDEnv)

	done := make(chan struct{})
	go func() {
		waitForPreviousInstanceExit()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("waitForPreviousInstanceExit blocked with no relaunch env set")
	}
}

// TestWaitForPreviousInstanceExitAlreadyGone returns promptly (well under the
// timeout) when the referenced PID is already dead, and clears the env var so
// this instance does not inherit a stale predecessor PID.
func TestWaitForPreviousInstanceExitAlreadyGone(t *testing.T) {
	pid := deadPID(t)
	t.Setenv(relaunchPIDEnv, strconv.Itoa(pid))

	start := time.Now()
	waitForPreviousInstanceExit()
	elapsed := time.Since(start)

	if elapsed >= relaunchWaitTimeout {
		t.Fatalf("waitForPreviousInstanceExit took %v, expected far less than timeout %v", elapsed, relaunchWaitTimeout)
	}
	if v := os.Getenv(relaunchPIDEnv); v != "" {
		t.Fatalf("relaunch env not cleared, still %q", v)
	}
}
