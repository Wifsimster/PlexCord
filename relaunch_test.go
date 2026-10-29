package main

import (
	"context"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"testing"
	"time"
)

func deadPID(t *testing.T) int {
	t.Helper()
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(context.Background(), "cmd", "/c", "exit")
	} else {
		cmd = exec.CommandContext(context.Background(), "sh", "-c", "exit 0")
	}
	if err := cmd.Start(); err != nil {
		t.Skipf("could not start helper process: %v", err)
	}
	pid := cmd.Process.Pid
	_ = cmd.Wait()
	return pid
}

func TestProcessExistsSelf(t *testing.T) {
	if !processExists(os.Getpid()) {
		t.Fatalf("processExists(self) = false, want true")
	}
}

func TestProcessExistsDead(t *testing.T) {
	pid := deadPID(t)
	if processExists(pid) {
		t.Fatalf("processExists(dead pid %d) = true, want false", pid)
	}
}

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
