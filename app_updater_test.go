package main

import (
	"testing"

	"plexcord/internal/platform"
	"plexcord/internal/updater"
	"plexcord/internal/version"
)

// TestUpdateNoticeWording covers how each updater state is presented in the
// system tray — the only update notification a user running PlexCord in the
// background ever sees.
func TestUpdateNoticeWording(t *testing.T) {
	info := &version.UpdateInfo{Available: true, CurrentVersion: "v1.4.0", LatestVersion: "v1.5.0"}

	tests := []struct {
		name           string
		status         updater.Status
		wantLabel      string
		wantActionable bool
	}{
		{
			name:           "ready invites a restart",
			status:         updater.Status{State: updater.StateReady, Info: info},
			wantLabel:      "Restart to update to v1.5.0",
			wantActionable: true,
		},
		{
			name:      "downloading is informational only",
			status:    updater.Status{State: updater.StateDownloading, Info: info, Progress: 42},
			wantLabel: "Downloading update v1.5.0…",
			// Clicking would only interrupt work already under way.
			wantActionable: false,
		},
		{
			name:           "available opens the window to decide",
			status:         updater.Status{State: updater.StateAvailable, Info: info},
			wantLabel:      "Update available: v1.5.0",
			wantActionable: true,
		},
		{
			name:   "idle shows nothing",
			status: updater.Status{State: updater.StateIdle},
		},
		{
			// Defensive: a state without info has no version to name, so there
			// is nothing worth putting in the menu.
			name:   "no info shows nothing",
			status: updater.Status{State: updater.StateReady},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := updateNotice(tt.status)
			if got.Label != tt.wantLabel {
				t.Errorf("Label = %q, want %q", got.Label, tt.wantLabel)
			}
			if got.Actionable != tt.wantActionable {
				t.Errorf("Actionable = %v, want %v", got.Actionable, tt.wantActionable)
			}
			if tt.wantLabel == "" && got != (platform.UpdateNotice{}) {
				t.Errorf("notice = %+v, want the zero value so the menu item hides", got)
			}
			if tt.wantLabel != "" && got.Tooltip == "" {
				t.Error("Tooltip is empty, want text for the tray icon hover")
			}
		})
	}
}

// TestPublishUpdateNoticeWithoutTray verifies mirroring update state is safe
// before the tray exists — the updater's listener fires as soon as it is
// registered, and QuitApp can tear the tray down while a check is in flight.
func TestPublishUpdateNoticeWithoutTray(t *testing.T) {
	app := &App{}

	app.publishUpdateNotice(updater.Status{State: updater.StateReady, Info: &version.UpdateInfo{LatestVersion: "v1.5.0"}})
}

// TestPublishUpdateNoticeReachesTray verifies the mapping is what the tray ends
// up holding.
func TestPublishUpdateNoticeReachesTray(t *testing.T) {
	tray := &fakeTray{}
	app := &App{tray: tray}

	app.publishUpdateNotice(updater.Status{
		State: updater.StateReady,
		Info:  &version.UpdateInfo{LatestVersion: "v1.5.0"},
	})
	notice, ok := tray.lastNotice()
	if !ok {
		t.Fatal("no notice reached the tray")
	}
	if notice.Label != "Restart to update to v1.5.0" {
		t.Errorf("tray notice label = %q, want the restart invitation", notice.Label)
	}

	// Back to idle (for example after a reset): the notice must be withdrawn.
	app.publishUpdateNotice(updater.Status{State: updater.StateIdle})
	notice, _ = tray.lastNotice()
	if notice != (platform.UpdateNotice{}) {
		t.Errorf("tray notice = %+v, want it withdrawn", notice)
	}
}

// TestUpdaterStatusReachesTrayThroughListener covers the wiring the tray notice
// depends on end to end: the updater announces a lifecycle transition and the
// tray menu item follows it. Before the UpdateService and TrayController
// interfaces existed, this needed a real GitHub poll and a real desktop tray.
func TestUpdaterStatusReachesTrayThroughListener(t *testing.T) {
	tray := &fakeTray{}
	fake := &fakeUpdater{}
	app := &App{tray: tray, updater: fake}

	// Registering the listener is what startup does; it must deliver the
	// current state immediately so a listener registered after an early
	// transition is not left behind.
	fake.OnStatusChange(app.publishUpdateNotice)

	notice, ok := tray.lastNotice()
	if !ok {
		t.Fatal("registering the listener delivered nothing to the tray")
	}
	if notice != (platform.UpdateNotice{}) {
		t.Errorf("initial tray notice = %+v, want none for an idle updater", notice)
	}

	// A download starting is informational: the tray says so but the item
	// is not clickable, since a click could only interrupt work in progress.
	fake.publish(updater.Status{
		State: updater.StateDownloading,
		Info:  &version.UpdateInfo{LatestVersion: "v2.0.0"},
	})
	notice, _ = tray.lastNotice()
	if notice.Label != "Downloading update v2.0.0…" {
		t.Errorf("downloading notice = %q, want the download label", notice.Label)
	}
	if notice.Actionable {
		t.Error("a download in progress was offered as an actionable tray item")
	}

	// Ready is the one worth acting on: the update is on disk, a restart applies it.
	fake.publish(updater.Status{
		State: updater.StateReady,
		Info:  &version.UpdateInfo{LatestVersion: "v2.0.0"},
	})
	notice, _ = tray.lastNotice()
	if notice.Label != "Restart to update to v2.0.0" {
		t.Errorf("ready notice = %q, want the restart invitation", notice.Label)
	}
	if !notice.Actionable {
		t.Error("a ready update was not offered as an actionable tray item")
	}
}

// TestGetUpdateStatusWithoutUpdater verifies the binding hydrates the frontend
// with a sane state before the checker has been constructed.
func TestGetUpdateStatusWithoutUpdater(t *testing.T) {
	app := &App{}

	if got := app.GetUpdateStatus(); got.State != updater.StateIdle {
		t.Errorf("GetUpdateStatus() state = %q, want idle before the updater exists", got.State)
	}
}
