package main

import (
	"context"
	"testing"
)

// The window manager is the piece that used to be untestable: every path went
// straight into the Wails runtime. These tests drive all of it against a fake.

func TestWindowManagerShowRaisesWindow(t *testing.T) {
	desktop := &fakeDesktop{}
	w := newWindowManager(desktop, desktop)
	w.MarkReady(context.Background())

	w.Show()

	shown, _, _ := desktop.counts()
	if shown != 1 {
		t.Fatalf("Show() drove the window %d time(s), want 1", shown)
	}
	// The always-on-top flick is what actually raises the window; it must be
	// turned back off, or PlexCord would sit on top of everything.
	want := []bool{true, false}
	if len(desktop.alwaysOnTop) != len(want) {
		t.Fatalf("always-on-top calls = %v, want %v", desktop.alwaysOnTop, want)
	}
	for i, v := range want {
		if desktop.alwaysOnTop[i] != v {
			t.Fatalf("always-on-top calls = %v, want %v", desktop.alwaysOnTop, want)
		}
	}
}

func TestWindowManagerUnminimisesOnlyWhenMinimised(t *testing.T) {
	// A window hidden while maximised must not be dropped back to its normal
	// size by an unconditional restore.
	desktop := &fakeDesktop{isMinimised: false}
	w := newWindowManager(desktop, desktop)
	w.MarkReady(context.Background())
	w.Show()

	if desktop.unminimised != 0 {
		t.Errorf("un-minimised a window that was not minimised (%d calls)", desktop.unminimised)
	}

	minimised := &fakeDesktop{isMinimised: true}
	w2 := newWindowManager(minimised, minimised)
	w2.MarkReady(context.Background())
	w2.Show()

	if minimised.unminimised != 1 {
		t.Errorf("un-minimise calls = %d, want 1 for a minimised window", minimised.unminimised)
	}
}

func TestWindowManagerIgnoresHideAndMinimiseBeforeReady(t *testing.T) {
	// Hiding or minimising a window that does not exist yet is meaningless —
	// and against a nil context it used to panic inside the Wails runtime.
	desktop := &fakeDesktop{}
	w := newWindowManager(desktop, desktop)

	w.Hide()
	w.Minimise()
	w.Resize(800, 600)

	if _, hidden, _ := desktop.counts(); hidden != 0 {
		t.Errorf("Hide() before ready reached the runtime (%d calls)", hidden)
	}
	if desktop.minimised != 0 {
		t.Errorf("Minimise() before ready reached the runtime (%d calls)", desktop.minimised)
	}
	if len(desktop.sizes) != 0 {
		t.Errorf("Resize() before ready reached the runtime (%v)", desktop.sizes)
	}
}

func TestWindowManagerQuitMarksExplicitExit(t *testing.T) {
	desktop := &fakeDesktop{}
	w := newWindowManager(desktop, desktop)
	w.MarkReady(context.Background())

	if w.IsQuitting() {
		t.Fatal("a fresh window manager reports an exit in progress")
	}

	w.Quit()

	if !w.IsQuitting() {
		t.Error("Quit() did not flag the exit as explicit; beforeClose would hide the window instead")
	}
	if _, _, quits := desktop.counts(); quits != 1 {
		t.Errorf("Quit() calls reaching the runtime = %d, want 1", quits)
	}
}

func TestWindowManagerMarkQuittingDoesNotExit(t *testing.T) {
	// The update restart flags the exit itself, then quits through its own path.
	desktop := &fakeDesktop{}
	w := newWindowManager(desktop, desktop)
	w.MarkReady(context.Background())

	w.MarkQuitting()

	if !w.IsQuitting() {
		t.Error("MarkQuitting() did not flag the exit")
	}
	if _, _, quits := desktop.counts(); quits != 0 {
		t.Errorf("MarkQuitting() quit the app (%d calls)", quits)
	}
}

func TestWindowManagerResizeCentersWindow(t *testing.T) {
	desktop := &fakeDesktop{}
	w := newWindowManager(desktop, desktop)
	w.MarkReady(context.Background())

	w.Resize(1024, 768)

	if len(desktop.sizes) != 1 || desktop.sizes[0] != [2]int{1024, 768} {
		t.Fatalf("sizes = %v, want one 1024x768 call", desktop.sizes)
	}
	// Wails centered the window against the previous size, so it has to be
	// re-centered after a resize.
	if desktop.centered != 1 {
		t.Errorf("center calls = %d, want 1 after a resize", desktop.centered)
	}
}

// TestBeforeCloseHidesWhenMinimizingToTray covers the close-button behaviour
// that depends on both the quit flag and the setting.
func TestBeforeCloseHidesWhenMinimizingToTray(t *testing.T) {
	app := newTestApp(defaultTestConfig())
	app.config.MinimizeToTray = true
	desktop := app.desktop.(*fakeDesktop)

	if prevent := app.beforeClose(context.Background()); !prevent {
		t.Error("beforeClose() = false with minimize-to-tray on, want the shutdown prevented")
	}
	if _, hidden, _ := desktop.counts(); hidden != 1 {
		t.Errorf("window hide calls = %d, want 1", hidden)
	}

	// An explicit quit must get through.
	app.windows.MarkQuitting()
	if prevent := app.beforeClose(context.Background()); prevent {
		t.Error("beforeClose() prevented an explicit quit")
	}
}
