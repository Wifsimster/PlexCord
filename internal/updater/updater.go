// Package updater runs the automatic update checker: it polls GitHub releases
// at startup and on a fixed interval, downloads and applies updates in the
// background on platforms that support self-update, and notifies the frontend
// through the event bus. The actual check/download/apply logic lives in
// internal/version; this package only adds scheduling, deduplication, and
// concurrency control on top, and is also the single download path for the
// user-initiated "Download & Install" action.
package updater

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"

	"plexcord/internal/events"
	"plexcord/internal/version"
)

// ErrDownloadInProgress is returned by StartDownload when another download is
// already running (for example the background checker started one just before
// the user clicked "Download & Install").
var ErrDownloadInProgress = errors.New("an update download is already in progress")

// initialCheckDelay is how long the checker waits after startup before its
// first check, so it does not compete with the Discord/Plex auto-connect.
const initialCheckDelay = 30 * time.Second

// State describes where the updater currently is in the update lifecycle.
type State string

// Updater lifecycle states, serialized to the frontend via Status.
const (
	StateIdle        State = "idle"        // no update known
	StateAvailable   State = "available"   // update found, not (yet) downloaded
	StateDownloading State = "downloading" // download/apply in progress
	StateReady       State = "ready"       // update applied on disk, restart pending
)

// Status is a snapshot of the updater state for the frontend, exposed through
// the GetUpdateStatus binding so the UI can hydrate after a page load.
type Status struct {
	State    State               `json:"state"`
	Info     *version.UpdateInfo `json:"info,omitempty"`
	Progress float64             `json:"progress"`
	Auto     bool                `json:"auto"` // last transition came from the background checker
}

// Updater schedules automatic update checks and serializes downloads.
// Construct with New; the zero value is not usable.
type Updater struct {
	bus      events.Bus
	interval time.Duration

	// Injected version-package functions, replaceable in tests.
	check    func() (*version.UpdateInfo, error)
	download func(ctx context.Context, progress version.ProgressFunc) (*version.UpdateInfo, error)
	canSelf  func() bool
	isDev    func() bool

	// initialDelay defaults to initialCheckDelay; tests shorten it.
	initialDelay time.Duration

	mu              sync.Mutex
	status          Status
	notifiedVersion string             // dedup: one announcement/download per version
	stop            context.CancelFunc // non-nil while the checker goroutine runs

	// downloadMu is held for the whole duration of a download so the manual
	// button and the background checker can never download concurrently.
	downloadMu sync.Mutex
}

// New creates an Updater wired to the production version-package functions.
func New(bus events.Bus, interval time.Duration) *Updater {
	return &Updater{
		bus:          bus,
		interval:     interval,
		check:        version.CheckForUpdate,
		download:     version.DownloadAndApplyUpdate,
		canSelf:      version.CanSelfUpdate,
		isDev:        version.IsDevBuild,
		initialDelay: initialCheckDelay,
		status:       Status{State: StateIdle},
	}
}

// StartChecker starts the background check loop: one check after a short
// startup delay, then one per interval. It is a no-op if the checker is
// already running or this is a dev build (dev builds always report an update
// available, which would otherwise cause a download on every launch).
func (u *Updater) StartChecker(parent context.Context) {
	if u.isDev() {
		log.Printf("Automatic update checks disabled for dev build")
		return
	}

	u.mu.Lock()
	if u.stop != nil {
		u.mu.Unlock()
		return
	}
	ctx, cancel := context.WithCancel(parent)
	u.stop = cancel
	u.mu.Unlock()

	go u.run(ctx)
}

// StopChecker stops the background check loop. Idempotent. An in-flight
// download is deliberately not interrupted: at that point the update is
// already being applied and aborting midway would risk a corrupt binary
// (selfupdate rolls back, but there is no reason to take the risk).
func (u *Updater) StopChecker() {
	u.mu.Lock()
	defer u.mu.Unlock()
	if u.stop != nil {
		u.stop()
		u.stop = nil
	}
}

// GetStatus returns a snapshot of the current updater state.
func (u *Updater) GetStatus() Status {
	u.mu.Lock()
	defer u.mu.Unlock()
	return u.status
}

// run is the checker goroutine body.
func (u *Updater) run(ctx context.Context) {
	select {
	case <-time.After(u.initialDelay):
	case <-ctx.Done():
		return
	}
	u.runAutoCheck(ctx)

	ticker := time.NewTicker(u.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			u.runAutoCheck(ctx)
		case <-ctx.Done():
			return
		}
	}
}

// runAutoCheck performs one automatic check cycle. Errors are logged and
// swallowed — the next tick retries; background work must never surface
// UpdateError toasts meant for user-initiated installs.
func (u *Updater) runAutoCheck(ctx context.Context) {
	info, err := u.check()
	if err != nil {
		log.Printf("Automatic update check failed (will retry): %v", err)
		return
	}
	if !info.Available {
		return
	}

	u.mu.Lock()
	// A download or applied-and-pending-restart (manual or automatic) already
	// owns the status; leave it alone and revisit on the next tick so we never
	// clobber an in-flight StateDownloading or emit a spurious event mid-work.
	if u.status.State == StateDownloading || u.status.State == StateReady {
		u.mu.Unlock()
		return
	}
	// Announce each version once. Retrying a previously failed download for the
	// same version must not re-emit, so gate the event on the version, not on
	// whether we then (re)start a download.
	newVersion := info.LatestVersion != u.notifiedVersion
	if newVersion {
		u.notifiedVersion = info.LatestVersion
		u.status = Status{State: StateAvailable, Info: info, Auto: true}
	}
	u.mu.Unlock()

	if newVersion {
		log.Printf("Automatic update check: %s -> %s available", info.CurrentVersion, info.LatestVersion)
		u.bus.Emit(events.UpdateAvailable, info)
	}

	if !u.canSelf() {
		// Platform without in-place self-update (macOS): notify-only, the
		// frontend links to the releases page.
		return
	}

	// Download the update, or retry a download that failed on an earlier tick
	// (state is still StateAvailable in that case).
	if _, err := u.StartDownload(ctx, true); err != nil && !errors.Is(err, ErrDownloadInProgress) {
		log.Printf("Automatic update download failed (will retry next check): %v", err)
	}
}

// StartDownload downloads, verifies, and applies the latest release. It is
// the single download path for both the background checker (auto=true) and
// the user-facing DownloadAndInstallUpdate binding (auto=false). Progress and
// completion are emitted as UpdateDownloadProgress / UpdateReady events; a
// failure emits UpdateError only for manual downloads.
func (u *Updater) StartDownload(ctx context.Context, auto bool) (*version.UpdateInfo, error) {
	u.mu.Lock()
	if u.status.State == StateReady {
		// Update already applied on disk; re-applying would download again
		// for nothing. Hand back the ready info so the caller can prompt
		// for a restart.
		info := u.status.Info
		u.mu.Unlock()
		return info, nil
	}
	u.mu.Unlock()

	if !u.downloadMu.TryLock() {
		return nil, ErrDownloadInProgress
	}
	defer u.downloadMu.Unlock()

	u.setState(func(s *Status) {
		s.State = StateDownloading
		s.Progress = 0
		s.Auto = auto
	})

	progress := func(downloaded, total int64) {
		var percent float64
		if total > 0 {
			percent = float64(downloaded) / float64(total) * 100
		}
		u.setState(func(s *Status) { s.Progress = percent })
		u.bus.Emit(events.UpdateDownloadProgress, DownloadProgress{
			Downloaded: downloaded,
			Total:      total,
			Percent:    percent,
		})
	}

	info, err := u.download(ctx, progress)
	if err != nil {
		u.mu.Lock()
		// Fall back to StateAvailable so the next automatic tick retries the
		// download (runAutoCheck retries whenever the state is available);
		// notifiedVersion is left intact so the retry does not re-announce.
		if u.status.Info != nil {
			u.status.State = StateAvailable
		} else {
			u.status.State = StateIdle
		}
		u.status.Progress = 0
		u.mu.Unlock()

		if !auto {
			u.bus.Emit(events.UpdateError, err.Error())
		}
		return nil, err
	}

	u.mu.Lock()
	u.status = Status{State: StateReady, Info: info, Progress: 100, Auto: auto}
	u.notifiedVersion = info.LatestVersion
	u.mu.Unlock()

	log.Printf("Update installed: %s (restart required)", info.LatestVersion)
	u.bus.Emit(events.UpdateReady, info)
	return info, nil
}

// setState mutates the status under the lock.
func (u *Updater) setState(mutate func(*Status)) {
	u.mu.Lock()
	defer u.mu.Unlock()
	mutate(&u.status)
}

// DownloadProgress is the payload of the UpdateDownloadProgress event. The
// field shape must stay identical to what the frontend progress bar consumes.
type DownloadProgress struct {
	Downloaded int64   `json:"downloaded"`
	Total      int64   `json:"total"`
	Percent    float64 `json:"percent"`
}
