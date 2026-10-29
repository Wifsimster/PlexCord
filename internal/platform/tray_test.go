package platform

import (
	"runtime"
	"testing"
)

func TestUpdateNoticeStoredBeforeMenuIsBuilt(t *testing.T) {
	tm := NewTrayManager(TrayCallbacks{}, TrayIcons{})

	notice := UpdateNotice{Label: "Restart to update to v1.5.0", Tooltip: "PlexCord v1.5.0 is installed", Actionable: true}
	tm.SetUpdateNotice(notice)

	if got := tm.UpdateNotice(); got != notice {
		t.Fatalf("UpdateNotice() = %+v, want %+v", got, notice)
	}
}

func TestUpdateNoticeOwnsTheTooltip(t *testing.T) {
	tm := NewTrayManager(TrayCallbacks{}, TrayIcons{})
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

	tm.SetTooltip("PlexCord — Idle")
	tm.mu.Lock()
	stillNotice := tm.effectiveTooltipLocked()
	tm.mu.Unlock()
	if stillNotice != "PlexCord v1.5.0 is available" {
		t.Errorf("tooltip after a base update = %q, want the notice to keep it", stillNotice)
	}

	tm.SetUpdateNotice(UpdateNotice{})
	tm.mu.Lock()
	restored := tm.effectiveTooltipLocked()
	tm.mu.Unlock()
	if restored != "PlexCord — Idle" {
		t.Errorf("tooltip after clearing = %q, want %q", restored, "PlexCord — Idle")
	}
}

func TestUpdateNoticeActive(t *testing.T) {
	if (UpdateNotice{}).active() {
		t.Error("zero UpdateNotice is active, want inactive so the menu item stays hidden")
	}
	if !(UpdateNotice{Label: "Update available"}).active() {
		t.Error("UpdateNotice with a label is inactive, want active")
	}
}

func TestApplyUpdateNoticeWithoutMenuItem(t *testing.T) {
	applyUpdateNotice(nil, UpdateNotice{Label: "Update available: v1.5.0"})
}

func TestTrayUpdateCallback(t *testing.T) {
	clicked := false
	tm := NewTrayManager(TrayCallbacks{OnUpdate: func() { clicked = true }}, TrayIcons{})

	tm.handleUpdate()

	if !clicked {
		t.Error("handleUpdate() did not invoke the OnUpdate callback")
	}
}

func TestTrayUpdateCallbackUnset(t *testing.T) {
	NewTrayManager(TrayCallbacks{}, TrayIcons{}).handleUpdate()
}

func TestTrayIconBadging(t *testing.T) {
	plainPNG := []byte("png")
	plainICO := []byte("ico")
	badgedPNG := []byte("png-badged")
	badgedICO := []byte("ico-badged")

	full := NewTrayManager(TrayCallbacks{}, TrayIcons{
		PNG: plainPNG, ICO: plainICO, UpdatePNG: badgedPNG, UpdateICO: badgedICO,
	})
	wantIdle, wantPending := plainPNG, badgedPNG
	if runtime.GOOS == "windows" {
		wantIdle, wantPending = plainICO, badgedICO
	}
	if got := full.icon(false); string(got) != string(wantIdle) {
		t.Errorf("icon(false) = %q, want %q", got, wantIdle)
	}
	if got := full.icon(true); string(got) != string(wantPending) {
		t.Errorf("icon(true) = %q, want %q", got, wantPending)
	}

	bare := NewTrayManager(TrayCallbacks{}, TrayIcons{PNG: plainPNG, ICO: plainICO})
	if got := bare.icon(true); string(got) != string(wantIdle) {
		t.Errorf("icon(true) without a badged variant = %q, want the plain icon %q", got, wantIdle)
	}

	if got := NewTrayManager(TrayCallbacks{}, TrayIcons{}).icon(true); len(got) != 0 {
		t.Errorf("icon(true) with no icons = %q, want empty", got)
	}
}
