//go:build linux

package platform

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// newTestManager points an AutoStartManager at a throwaway XDG config dir so
// the tests never touch the developer's own autostart entries.
func newTestManager(t *testing.T) *AutoStartManager {
	t.Helper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	return &AutoStartManager{appName: "PlexCord", executable: "/opt/plexcord/plexcord"}
}

// TestEnableRegistersTheAutoStartFlag verifies the registration launches
// PlexCord with the flag that marks a login launch: without it the app cannot
// tell the OS starting it from the user opening it, and would show its window
// on every boot.
func TestEnableRegistersTheAutoStartFlag(t *testing.T) {
	m := newTestManager(t)

	if err := m.Enable(); err != nil {
		t.Fatalf("Enable() error = %v", err)
	}
	if !m.IsEnabled() {
		t.Fatal("IsEnabled() = false after Enable()")
	}

	entry, err := os.ReadFile(m.getDesktopFilePath())
	if err != nil {
		t.Fatalf("reading the .desktop file: %v", err)
	}
	want := "Exec=" + m.executable + " " + AutoStartFlag
	if !strings.Contains(string(entry), want) {
		t.Errorf("desktop entry does not contain %q:\n%s", want, entry)
	}
}

// TestEnsureRegisteredRefreshesAStaleEntry covers the upgrade path: an entry
// written by an older version launches the bare executable, and is rewritten
// in place rather than left as it was.
func TestEnsureRegisteredRefreshesAStaleEntry(t *testing.T) {
	m := newTestManager(t)

	stale := "[Desktop Entry]\nType=Application\nName=PlexCord\nExec=/old/path/plexcord\n"
	path := m.getDesktopFilePath()
	if err := os.MkdirAll(filepath.Dir(path), 0750); err != nil {
		t.Fatalf("creating the autostart dir: %v", err)
	}
	if err := os.WriteFile(path, []byte(stale), 0600); err != nil {
		t.Fatalf("writing the stale entry: %v", err)
	}

	if err := m.EnsureRegistered(); err != nil {
		t.Fatalf("EnsureRegistered() error = %v", err)
	}

	entry, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading the .desktop file: %v", err)
	}
	if string(entry) != m.desktopEntry() {
		t.Errorf("stale entry was not refreshed:\n%s", entry)
	}
}

// TestEnsureRegisteredLeavesAutoStartOff verifies refreshing never turns the
// setting on behind the user's back.
func TestEnsureRegisteredLeavesAutoStartOff(t *testing.T) {
	m := newTestManager(t)

	if err := m.EnsureRegistered(); err != nil {
		t.Fatalf("EnsureRegistered() error = %v", err)
	}
	if m.IsEnabled() {
		t.Error("EnsureRegistered() registered auto-start while it was disabled")
	}
}

// TestDisableRemovesTheEntry rounds out the lifecycle.
func TestDisableRemovesTheEntry(t *testing.T) {
	m := newTestManager(t)

	if err := m.Enable(); err != nil {
		t.Fatalf("Enable() error = %v", err)
	}
	if err := m.Disable(); err != nil {
		t.Fatalf("Disable() error = %v", err)
	}
	if m.IsEnabled() {
		t.Error("IsEnabled() = true after Disable()")
	}
}
