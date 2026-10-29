package main

import (
	stderrors "errors"
	"log"
	"net/url"
	goruntime "runtime"
	"strings"
	"time"

	"plexcord/internal/errors"
	"plexcord/internal/events"
	"plexcord/internal/history"
	"plexcord/internal/updater"
	"plexcord/internal/version"
)

func (a *App) GetErrorInfo(code string) errors.ErrorInfo {
	return errors.GetErrorInfo(code)
}

func (a *App) IsRetryableError(code string) bool {
	return errors.IsRetryable(code)
}

func (a *App) IsAuthError(code string) bool {
	return errors.IsAuthError(code)
}

type ConnectionHistory struct {
	PlexLastConnected    *time.Time `json:"plexLastConnected"`
	DiscordLastConnected *time.Time `json:"discordLastConnected"`
}

func (a *App) GetConnectionHistory() ConnectionHistory {
	return ConnectionHistory{
		PlexLastConnected:    a.config.PlexLastConnected,
		DiscordLastConnected: a.config.DiscordLastConnected,
	}
}

func (a *App) GetVersion() version.Info {
	return version.GetInfo()
}

func (a *App) CheckForUpdate() (*version.UpdateInfo, error) {
	log.Printf("Checking for updates...")

	check := version.CheckForUpdate
	if a.updater != nil {
		check = a.updater.Check
	}

	info, err := check()
	if err != nil {
		log.Printf("ERROR: Update check failed: %v", err)
		return nil, errors.Wrap(err, errors.TIMEOUT, "failed to check for updates")
	}

	if info.Available {
		log.Printf("Update available: %s -> %s", info.CurrentVersion, info.LatestVersion)
	} else {
		log.Printf("Application is up to date: %s", info.CurrentVersion)
	}

	return info, nil
}

func (a *App) OpenReleasesPage() error {
	releaseURL := version.GetReleasesURL()
	log.Printf("Opening releases page: %s", releaseURL)
	a.desktop.OpenURL(a.ctx, releaseURL)
	return nil
}

func (a *App) OpenReleaseURL(releaseURL string) error {
	if releaseURL == "" {
		return a.OpenReleasesPage()
	}
	parsed, err := url.Parse(releaseURL)
	if err != nil || parsed.Scheme != "https" {
		return errors.New(errors.CONFIG_READ_FAILED, "invalid release URL")
	}
	if host := parsed.Hostname(); host != "github.com" && !strings.HasSuffix(host, ".github.com") {
		return errors.New(errors.CONFIG_READ_FAILED, "release URL must be from github.com")
	}
	log.Printf("Opening release URL: %s", releaseURL)
	a.desktop.OpenURL(a.ctx, releaseURL)
	return nil
}

func (a *App) CanSelfUpdate() bool {
	return version.CanSelfUpdate()
}

func (a *App) DownloadAndInstallUpdate() (*version.UpdateInfo, error) {
	if !version.CanSelfUpdate() {
		a.bus.Emit(events.UpdateError, version.ErrUpdateNotSupported.Error())
		return nil, errors.Wrap(version.ErrUpdateNotSupported, errors.UNKNOWN_ERROR, "update not supported")
	}
	if a.updater == nil {
		return nil, errors.New(errors.UNKNOWN_ERROR, "updater is not ready yet")
	}

	log.Printf("Starting in-app update download...")

	info, err := a.updater.StartDownload(a.ctx, false)
	if err != nil {
		if stderrors.Is(err, updater.ErrDownloadInProgress) {
			log.Printf("Update download already in progress; joining it")
			return a.updater.GetStatus().Info, nil
		}
		log.Printf("ERROR: Update failed: %v", err)
		return nil, errors.Wrap(err, errors.UNKNOWN_ERROR, "failed to install update")
	}
	return info, nil
}

func (a *App) RestartApplication() error {
	if err := a.relauncher.Relaunch(); err != nil {
		return err
	}

	log.Printf("Relaunching application to apply update")
	a.windows.MarkQuitting()
	a.desktop.Quit(a.ctx)
	return nil
}

type ResourceStats struct {
	Timestamp      string  `json:"timestamp"`
	MemoryAllocMB  float64 `json:"memoryAllocMB"`
	MemoryTotalMB  float64 `json:"memoryTotalMB"`
	GoroutineCount int     `json:"goroutineCount"`
}

func (a *App) GetResourceStats() ResourceStats {
	var m goruntime.MemStats
	goruntime.ReadMemStats(&m)

	return ResourceStats{
		MemoryAllocMB:  float64(m.Alloc) / 1024 / 1024,
		MemoryTotalMB:  float64(m.TotalAlloc) / 1024 / 1024,
		GoroutineCount: goruntime.NumGoroutine(),
		Timestamp:      time.Now().Format(time.RFC3339),
	}
}

func (a *App) GetListeningHistory(limit int) []history.Entry {
	if a.history == nil {
		return nil
	}
	return a.history.GetRecent(limit)
}

func (a *App) GetListeningStats() history.Stats {
	if a.history == nil {
		return history.Stats{}
	}
	return a.history.GetStats()
}

func (a *App) ClearListeningHistory() {
	if a.history == nil {
		return
	}
	a.history.Clear()
}
