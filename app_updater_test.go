package main

import (
	"testing"

	"plexcord/internal/platform"
	"plexcord/internal/updater"
	"plexcord/internal/version"
)

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
			name:           "downloading is informational only",
			status:         updater.Status{State: updater.StateDownloading, Info: info, Progress: 42},
			wantLabel:      "Downloading update v1.5.0…",
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

func TestPublishUpdateNoticeWithoutTray(t *testing.T) {
	app := &App{}

	app.publishUpdateNotice(updater.Status{State: updater.StateReady, Info: &version.UpdateInfo{LatestVersion: "v1.5.0"}})
}

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

	app.publishUpdateNotice(updater.Status{State: updater.StateIdle})
	notice, _ = tray.lastNotice()
	if notice != (platform.UpdateNotice{}) {
		t.Errorf("tray notice = %+v, want it withdrawn", notice)
	}
}

func TestUpdaterStatusReachesTrayThroughListener(t *testing.T) {
	tray := &fakeTray{}
	fake := &fakeUpdater{}
	app := &App{tray: tray, updater: fake}

	fake.OnStatusChange(app.publishUpdateNotice)

	notice, ok := tray.lastNotice()
	if !ok {
		t.Fatal("registering the listener delivered nothing to the tray")
	}
	if notice != (platform.UpdateNotice{}) {
		t.Errorf("initial tray notice = %+v, want none for an idle updater", notice)
	}

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

func TestGetUpdateStatusWithoutUpdater(t *testing.T) {
	app := &App{}

	if got := app.GetUpdateStatus(); got.State != updater.StateIdle {
		t.Errorf("GetUpdateStatus() state = %q, want idle before the updater exists", got.State)
	}
}
