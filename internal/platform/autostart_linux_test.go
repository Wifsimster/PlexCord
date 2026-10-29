//go:build linux

package platform

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func newTestManager(t *testing.T) *AutoStartManager {
	t.Helper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	return &AutoStartManager{appName: "PlexCord", executable: "/opt/plexcord/plexcord"}
}

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

func TestEnsureRegisteredLeavesAutoStartOff(t *testing.T) {
	m := newTestManager(t)

	if err := m.EnsureRegistered(); err != nil {
		t.Fatalf("EnsureRegistered() error = %v", err)
	}
	if m.IsEnabled() {
		t.Error("EnsureRegistered() registered auto-start while it was disabled")
	}
}

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
