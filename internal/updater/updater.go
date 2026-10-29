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

var ErrDownloadInProgress = errors.New("an update download is already in progress")

const initialCheckDelay = 30 * time.Second

type State string

const (
	StateIdle        State = "idle"
	StateAvailable   State = "available"
	StateDownloading State = "downloading"
	StateReady       State = "ready"
)

type Status struct {
	State    State               `json:"state"`
	Info     *version.UpdateInfo `json:"info,omitempty"`
	Progress float64             `json:"progress"`
	Auto     bool                `json:"auto"`
}

type Updater struct {
	bus      events.Bus
	interval time.Duration

	check    func() (*version.UpdateInfo, error)
	download func(ctx context.Context, progress version.ProgressFunc) (*version.UpdateInfo, error)
	canSelf  func() bool
	isDev    func() bool

	initialDelay time.Duration

	mu              sync.Mutex
	status          Status
	notifiedVersion string
	stop            context.CancelFunc
	onStatusChange  func(Status)

	downloadMu sync.Mutex
}

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

func (u *Updater) StopChecker() {
	u.mu.Lock()
	defer u.mu.Unlock()
	if u.stop != nil {
		u.stop()
		u.stop = nil
	}
}

func (u *Updater) GetStatus() Status {
	u.mu.Lock()
	defer u.mu.Unlock()
	return u.status
}

func (u *Updater) OnStatusChange(fn func(Status)) {
	u.mu.Lock()
	u.onStatusChange = fn
	status := u.status
	u.mu.Unlock()

	if fn != nil {
		fn(status)
	}
}

func (u *Updater) notifyStatusChange() {
	u.mu.Lock()
	fn, status := u.onStatusChange, u.status
	u.mu.Unlock()

	if fn != nil {
		fn(status)
	}
}

func (u *Updater) Check() (*version.UpdateInfo, error) {
	info, err := u.check()
	if err != nil {
		return nil, err
	}
	if !info.Available {
		return info, nil
	}

	u.mu.Lock()
	if u.status.State == StateDownloading || u.status.State == StateReady {
		u.mu.Unlock()
		return info, nil
	}
	u.notifiedVersion = info.LatestVersion
	u.status = Status{State: StateAvailable, Info: info, Auto: false}
	u.mu.Unlock()

	u.notifyStatusChange()
	return info, nil
}

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
	if u.status.State == StateDownloading || u.status.State == StateReady {
		u.mu.Unlock()
		return
	}
	newVersion := info.LatestVersion != u.notifiedVersion
	if newVersion {
		u.notifiedVersion = info.LatestVersion
		u.status = Status{State: StateAvailable, Info: info, Auto: true}
	}
	u.mu.Unlock()

	if newVersion {
		log.Printf("Automatic update check: %s -> %s available", info.CurrentVersion, info.LatestVersion)
		u.bus.Emit(events.UpdateAvailable, info)
		u.notifyStatusChange()
	}

	if !u.canSelf() {
		return
	}

	if _, err := u.StartDownload(ctx, true); err != nil && !errors.Is(err, ErrDownloadInProgress) {
		log.Printf("Automatic update download failed (will retry next check): %v", err)
	}
}

func (u *Updater) StartDownload(ctx context.Context, auto bool) (*version.UpdateInfo, error) {
	u.mu.Lock()
	if u.status.State == StateReady {
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
	u.notifyStatusChange()

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
		u.notifyStatusChange()
		return nil, err
	}

	u.mu.Lock()
	u.status = Status{State: StateReady, Info: info, Progress: 100, Auto: auto}
	u.notifiedVersion = info.LatestVersion
	u.mu.Unlock()

	log.Printf("Update installed: %s (restart required)", info.LatestVersion)
	u.bus.Emit(events.UpdateReady, info)
	u.notifyStatusChange()
	return info, nil
}

func (u *Updater) setState(mutate func(*Status)) {
	u.mu.Lock()
	defer u.mu.Unlock()
	mutate(&u.status)
}

type DownloadProgress struct {
	Downloaded int64   `json:"downloaded"`
	Total      int64   `json:"total"`
	Percent    float64 `json:"percent"`
}
