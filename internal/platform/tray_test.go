package platform

import "testing"

// TestUpdateNoticeStoredBeforeMenuIsBuilt verifies a notice that arrives before
// systray has built the menu is remembered rather than dropped. This is the
// realistic ordering: the tray starts on its own goroutine, while an update
// found at startup can land at any point after that.
func TestUpdateNoticeStoredBeforeMenuIsBuilt(t *testing.T) {
	tm := NewTrayManager(TrayCallbacks{}, nil, nil)

	notice := UpdateNotice{Label: "Restart to update to v1.5.0", Tooltip: "PlexCord v1.5.0 is installed", Actionable: true}
	tm.SetUpdateNotice(notice)

	if got := tm.UpdateNotice(); got != notice {
		t.Fatalf("UpdateNotice() = %+v, want %+v", got, notice)
	}
}

// TestUpdateNoticeOwnsTheTooltip verifies a pending update takes over the tray
// tooltip, and hands it back once cleared.
func TestUpdateNoticeOwnsTheTooltip(t *testing.T) {
	tm := NewTrayManager(TrayCallbacks{}, nil, nil)
	tm.SetTooltip("PlexCord — Live")

	tm.mu.Lock()
	base := tm.effectiveTooltipLocked()
	tm.mu.Unlock()
	if base != "PlexCord — Live" {
		t.Fatalf("tooltip with no notice = %q, want the base tooltip", base)
	}

	tm.SetUpdateNotice(UpdateNotice{Label: "Update available: v1.5.0", Tooltip: "PlexCord v1.5.0 is available"})

	tm.mu.Lock()
	withNotice := tm.effectiveTooltipLocked()
	tm.mu.Unlock()
	if withNotice != "PlexCord v1.5.0 is available" {
		t.Errorf("tooltip with a notice = %q, want the notice tooltip", withNotice)
	}

	// A tooltip update while a notice is active must not steal it back.
	tm.SetTooltip("PlexCord — Idle")
	tm.mu.Lock()
	stillNotice := tm.effectiveTooltipLocked()
	tm.mu.Unlock()
	if stillNotice != "PlexCord v1.5.0 is available" {
		t.Errorf("tooltip after a base update = %q, want the notice to keep it", stillNotice)
	}

	// Clearing the notice restores the most recent base tooltip.
	tm.SetUpdateNotice(UpdateNotice{})
	tm.mu.Lock()
	restored := tm.effectiveTooltipLocked()
	tm.mu.Unlock()
	if restored != "PlexCord — Idle" {
		t.Errorf("tooltip after clearing = %q, want %q", restored, "PlexCord — Idle")
	}
}

// TestUpdateNoticeActive covers the zero value meaning "nothing pending".
func TestUpdateNoticeActive(t *testing.T) {
	if (UpdateNotice{}).active() {
		t.Error("zero UpdateNotice is active, want inactive so the menu item stays hidden")
	}
	if !(UpdateNotice{Label: "Update available"}).active() {
		t.Error("UpdateNotice with a label is inactive, want active")
	}
}

// TestApplyUpdateNoticeWithoutMenuItem verifies applying a notice before the
// menu exists is a no-op rather than a nil dereference.
func TestApplyUpdateNoticeWithoutMenuItem(t *testing.T) {
	applyUpdateNotice(nil, UpdateNotice{Label: "Update available: v1.5.0"})
}

// TestTrayUpdateCallback verifies a click on the update item reaches the app.
func TestTrayUpdateCallback(t *testing.T) {
	clicked := false
	tm := NewTrayManager(TrayCallbacks{OnUpdate: func() { clicked = true }}, nil, nil)

	tm.handleUpdate()

	if !clicked {
		t.Error("handleUpdate() did not invoke the OnUpdate callback")
	}
}

// TestTrayUpdateCallbackUnset verifies a click with no callback wired is safe.
func TestTrayUpdateCallbackUnset(t *testing.T) {
	NewTrayManager(TrayCallbacks{}, nil, nil).handleUpdate()
}
