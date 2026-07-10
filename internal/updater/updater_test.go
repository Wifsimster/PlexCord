package updater

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"plexcord/internal/events"
	"plexcord/internal/version"
)

// newTestUpdater builds an Updater with fast timings and controllable fakes.
// Override the injected funcs after construction as needed.
func newTestUpdater(bus events.Bus) *Updater {
	u := New(bus, 20*time.Millisecond)
	u.initialDelay = time.Millisecond
	u.isDev = func() bool { return false }
	u.canSelf = func() bool { return true }
	u.check = func() (*version.UpdateInfo, error) {
		return &version.UpdateInfo{Available: false}, nil
	}
	u.download = func(context.Context, version.ProgressFunc) (*version.UpdateInfo, error) {
		return &version.UpdateInfo{LatestVersion: "v9.9.9"}, nil
	}
	return u
}

func updateInfo(v string) *version.UpdateInfo {
	return &version.UpdateInfo{Available: true, CurrentVersion: "v1.0.0", LatestVersion: v}
}

// waitFor polls until cond returns true or the timeout elapses.
func waitFor(t *testing.T, cond func() bool, msg string) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(2 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for: %s", msg)
}

func TestAutoCheckDownloadsOnceAndDedupsAcrossTicks(t *testing.T) {
	bus := events.NewRecordingBus()
	u := newTestUpdater(bus)

	var checks, downloads atomic.Int32
	u.check = func() (*version.UpdateInfo, error) {
		checks.Add(1)
		return updateInfo("v9.9.9"), nil
	}
	u.download = func(context.Context, version.ProgressFunc) (*version.UpdateInfo, error) {
		downloads.Add(1)
		return &version.UpdateInfo{LatestVersion: "v9.9.9"}, nil
	}

	u.StartChecker(context.Background())
	defer u.StopChecker()

	// Wait for several ticks to pass.
	waitFor(t, func() bool { return checks.Load() >= 3 }, "at least 3 checks")

	if got := downloads.Load(); got != 1 {
		t.Errorf("download called %d times, want 1 (dedup by version)", got)
	}
	if got := bus.Count(events.UpdateAvailable); got != 1 {
		t.Errorf("UpdateAvailable emitted %d times, want 1", got)
	}
	if got := bus.Count(events.UpdateReady); got != 1 {
		t.Errorf("UpdateReady emitted %d times, want 1", got)
	}
	if st := u.GetStatus(); st.State != StateReady || !st.Auto {
		t.Errorf("status = %+v, want ready/auto", st)
	}
}

func TestAutoCheckNewVersionSupersedesDedup(t *testing.T) {
	bus := events.NewRecordingBus()
	u := newTestUpdater(bus)
	u.canSelf = func() bool { return false } // notify-only keeps state simple

	var checks atomic.Int32
	u.check = func() (*version.UpdateInfo, error) {
		n := checks.Add(1)
		if n <= 2 {
			return updateInfo("v9.9.9"), nil
		}
		return updateInfo("v10.0.0"), nil
	}

	u.StartChecker(context.Background())
	defer u.StopChecker()

	waitFor(t, func() bool { return bus.Count(events.UpdateAvailable) >= 2 }, "second UpdateAvailable for new version")

	if got := bus.Count(events.UpdateAvailable); got != 2 {
		t.Errorf("UpdateAvailable emitted %d times, want 2 (once per distinct version)", got)
	}
}

func TestAutoCheckNotifyOnlyWhenSelfUpdateUnsupported(t *testing.T) {
	bus := events.NewRecordingBus()
	u := newTestUpdater(bus)
	u.canSelf = func() bool { return false }

	var checks, downloads atomic.Int32
	u.check = func() (*version.UpdateInfo, error) {
		checks.Add(1)
		return updateInfo("v9.9.9"), nil
	}
	u.download = func(context.Context, version.ProgressFunc) (*version.UpdateInfo, error) {
		downloads.Add(1)
		return nil, errors.New("should not be called")
	}

	u.StartChecker(context.Background())
	defer u.StopChecker()

	waitFor(t, func() bool { return checks.Load() >= 2 }, "at least 2 checks")

	if downloads.Load() != 0 {
		t.Error("download must not run on platforms without self-update support")
	}
	if got := bus.Count(events.UpdateAvailable); got != 1 {
		t.Errorf("UpdateAvailable emitted %d times, want 1", got)
	}
	if st := u.GetStatus(); st.State != StateAvailable {
		t.Errorf("status state = %s, want available", st.State)
	}
}

func TestDevBuildNeverChecks(t *testing.T) {
	bus := events.NewRecordingBus()
	u := newTestUpdater(bus)
	u.isDev = func() bool { return true }

	var checks atomic.Int32
	u.check = func() (*version.UpdateInfo, error) {
		checks.Add(1)
		return updateInfo("v9.9.9"), nil
	}

	u.StartChecker(context.Background())
	defer u.StopChecker()

	time.Sleep(60 * time.Millisecond)
	if checks.Load() != 0 {
		t.Error("dev builds must not run automatic update checks")
	}
}

func TestStopCheckerStopsAndStartIsIdempotent(t *testing.T) {
	bus := events.NewRecordingBus()
	u := newTestUpdater(bus)

	var checks atomic.Int32
	u.check = func() (*version.UpdateInfo, error) {
		checks.Add(1)
		return &version.UpdateInfo{Available: false}, nil
	}

	ctx := context.Background()
	u.StartChecker(ctx)
	u.StartChecker(ctx) // idempotent: must not spawn a second loop

	waitFor(t, func() bool { return checks.Load() >= 1 }, "first check")
	// With a 1ms initial delay, a duplicate goroutine would have doubled the
	// early check count; give both a moment then compare against ticks.
	u.StopChecker()
	u.StopChecker() // idempotent

	stopped := checks.Load()
	time.Sleep(80 * time.Millisecond)
	if checks.Load() != stopped {
		t.Errorf("checks continued after StopChecker: %d -> %d", stopped, checks.Load())
	}

	// Restart works after a stop.
	u.StartChecker(ctx)
	waitFor(t, func() bool { return checks.Load() > stopped }, "check after restart")
	u.StopChecker()
}

func TestStartDownloadRejectsConcurrentAndCachesReady(t *testing.T) {
	bus := events.NewRecordingBus()
	u := newTestUpdater(bus)

	release := make(chan struct{})
	var downloads atomic.Int32
	u.download = func(context.Context, version.ProgressFunc) (*version.UpdateInfo, error) {
		downloads.Add(1)
		<-release
		return &version.UpdateInfo{LatestVersion: "v9.9.9"}, nil
	}

	done := make(chan error, 1)
	go func() {
		_, err := u.StartDownload(context.Background(), false)
		done <- err
	}()
	waitFor(t, func() bool { return downloads.Load() == 1 }, "first download started")

	if _, err := u.StartDownload(context.Background(), false); !errors.Is(err, ErrDownloadInProgress) {
		t.Errorf("concurrent StartDownload error = %v, want ErrDownloadInProgress", err)
	}

	close(release)
	if err := <-done; err != nil {
		t.Fatalf("first download failed: %v", err)
	}

	// A download after the update is ready must not re-download.
	info, err := u.StartDownload(context.Background(), false)
	if err != nil {
		t.Fatalf("StartDownload after ready: %v", err)
	}
	if info == nil || info.LatestVersion != "v9.9.9" {
		t.Errorf("cached ready info = %+v, want v9.9.9", info)
	}
	if downloads.Load() != 1 {
		t.Errorf("download ran %d times, want 1 (ready state short-circuits)", downloads.Load())
	}
}

func TestAutoCheckDoesNotClobberInFlightDownload(t *testing.T) {
	bus := events.NewRecordingBus()
	u := newTestUpdater(bus)
	// No prior auto-announce: a purely manual download, notifiedVersion == "".
	u.check = func() (*version.UpdateInfo, error) {
		return updateInfo("v9.9.9"), nil
	}

	release := make(chan struct{})
	u.download = func(context.Context, version.ProgressFunc) (*version.UpdateInfo, error) {
		<-release
		return &version.UpdateInfo{LatestVersion: "v9.9.9"}, nil
	}

	// Start a manual download and let it reach StateDownloading.
	done := make(chan error, 1)
	go func() {
		_, err := u.StartDownload(context.Background(), false)
		done <- err
	}()
	waitFor(t, func() bool { return u.GetStatus().State == StateDownloading }, "manual download reaches downloading")

	// An automatic check firing now must not overwrite the downloading status
	// nor emit a spurious UpdateAvailable.
	u.runAutoCheck(context.Background())

	if st := u.GetStatus().State; st != StateDownloading {
		t.Errorf("status = %s after auto-check during download, want downloading (no clobber)", st)
	}
	if got := bus.Count(events.UpdateAvailable); got != 0 {
		t.Errorf("UpdateAvailable emitted %d times during an in-flight download, want 0", got)
	}

	close(release)
	if err := <-done; err != nil {
		t.Fatalf("manual download failed: %v", err)
	}
}

func TestAutoDownloadFailureRetriesWithoutReannouncing(t *testing.T) {
	bus := events.NewRecordingBus()
	u := newTestUpdater(bus)

	var downloads atomic.Int32
	u.check = func() (*version.UpdateInfo, error) { return updateInfo("v9.9.9"), nil }
	u.download = func(context.Context, version.ProgressFunc) (*version.UpdateInfo, error) {
		downloads.Add(1)
		return nil, errors.New("network exploded")
	}

	u.StartChecker(context.Background())
	defer u.StopChecker()

	waitFor(t, func() bool { return downloads.Load() >= 3 }, "download retried across ticks")

	// The update is announced exactly once even though the download keeps
	// failing and retrying every tick.
	if got := bus.Count(events.UpdateAvailable); got != 1 {
		t.Errorf("UpdateAvailable emitted %d times, want 1 (no re-announce on retry)", got)
	}
	if got := bus.Count(events.UpdateError); got != 0 {
		t.Errorf("UpdateError emitted %d times for automatic downloads, want 0", got)
	}
}

func TestAutoDownloadFailureIsSilentAndRetries(t *testing.T) {
	bus := events.NewRecordingBus()
	u := newTestUpdater(bus)

	var downloads atomic.Int32
	u.check = func() (*version.UpdateInfo, error) {
		return updateInfo("v9.9.9"), nil
	}
	u.download = func(context.Context, version.ProgressFunc) (*version.UpdateInfo, error) {
		downloads.Add(1)
		return nil, errors.New("network exploded")
	}

	u.StartChecker(context.Background())
	defer u.StopChecker()

	waitFor(t, func() bool { return downloads.Load() >= 2 }, "download retried on a later tick")

	if got := bus.Count(events.UpdateError); got != 0 {
		t.Errorf("UpdateError emitted %d times for automatic downloads, want 0", got)
	}
	if st := u.GetStatus(); st.State != StateAvailable {
		t.Errorf("status state after failed auto download = %s, want available", st.State)
	}
}

func TestManualDownloadFailureEmitsUpdateError(t *testing.T) {
	bus := events.NewRecordingBus()
	u := newTestUpdater(bus)
	u.download = func(context.Context, version.ProgressFunc) (*version.UpdateInfo, error) {
		return nil, errors.New("checksum mismatch")
	}

	if _, err := u.StartDownload(context.Background(), false); err == nil {
		t.Fatal("expected error from failing download")
	}
	if got := bus.Count(events.UpdateError); got != 1 {
		t.Errorf("UpdateError emitted %d times for manual download, want 1", got)
	}
}

func TestDownloadEmitsProgressAndReady(t *testing.T) {
	bus := events.NewRecordingBus()
	u := newTestUpdater(bus)
	u.download = func(_ context.Context, progress version.ProgressFunc) (*version.UpdateInfo, error) {
		progress(50, 100)
		progress(100, 100)
		return &version.UpdateInfo{LatestVersion: "v9.9.9"}, nil
	}

	if _, err := u.StartDownload(context.Background(), false); err != nil {
		t.Fatalf("StartDownload: %v", err)
	}

	if got := bus.Count(events.UpdateDownloadProgress); got != 2 {
		t.Errorf("UpdateDownloadProgress emitted %d times, want 2", got)
	}
	if got := bus.Count(events.UpdateReady); got != 1 {
		t.Errorf("UpdateReady emitted %d times, want 1", got)
	}

	// Payload shape: percent must be present and computed.
	for _, e := range bus.Snapshot() {
		if e.Name != events.UpdateDownloadProgress {
			continue
		}
		p, ok := e.Payload[0].(DownloadProgress)
		if !ok {
			t.Fatalf("progress payload type = %T, want DownloadProgress", e.Payload[0])
		}
		if p.Percent <= 0 || p.Percent > 100 {
			t.Errorf("progress percent = %v, want within (0,100]", p.Percent)
		}
	}
}
