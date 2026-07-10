package main

import (
	"context"
	"log"
	"net/url"
	"os"
	"os/exec"
	goruntime "runtime"
	"strings"
	"time"

	"plexcord/internal/errors"
	"plexcord/internal/events"
	"plexcord/internal/history"
	"plexcord/internal/version"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// ============================================================================
// Error Information (Story 6.2)
// ============================================================================

// GetErrorInfo returns user-friendly error information for an error code.
// This is used by the frontend to display actionable error messages.
func (a *App) GetErrorInfo(code string) errors.ErrorInfo {
	return errors.GetErrorInfo(code)
}

// IsRetryableError returns whether an error with the given code can be retried.
func (a *App) IsRetryableError(code string) bool {
	return errors.IsRetryable(code)
}

// IsAuthError returns whether the error indicates authentication is needed.
func (a *App) IsAuthError(code string) bool {
	return errors.IsAuthError(code)
}

// ============================================================================
// Connection History (Story 6.8)
// ============================================================================

// ConnectionHistory contains timestamps for last successful connections.
type ConnectionHistory struct {
	PlexLastConnected    *time.Time `json:"plexLastConnected"`
	DiscordLastConnected *time.Time `json:"discordLastConnected"`
}

// GetConnectionHistory returns the last successful connection timestamps.
func (a *App) GetConnectionHistory() ConnectionHistory {
	return ConnectionHistory{
		PlexLastConnected:    a.config.PlexLastConnected,
		DiscordLastConnected: a.config.DiscordLastConnected,
	}
}

// ============================================================================
// Version & Updates (Epic 7)
// ============================================================================

// GetVersion returns the current application version information.
// Version is set at build time via -ldflags.
// Example: go build -ldflags "-X plexcord/internal/version.Version=v1.0.0"
func (a *App) GetVersion() version.Info {
	return version.GetInfo()
}

// CheckForUpdate checks GitHub releases for a newer version.
// Returns update info including availability, latest version, and download URL.
// A loading indicator should be shown during the check (typically 1-5 seconds).
func (a *App) CheckForUpdate() (*version.UpdateInfo, error) {
	log.Printf("Checking for updates...")

	info, err := version.CheckForUpdate()
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

// OpenReleasesPage opens the GitHub releases page in the default browser.
// Used for viewing changelog and downloading updates.
func (a *App) OpenReleasesPage() error {
	releaseURL := version.GetReleasesURL()
	log.Printf("Opening releases page: %s", releaseURL)
	runtime.BrowserOpenURL(a.ctx, releaseURL)
	return nil
}

// OpenReleaseURL opens a specific release URL in the default browser.
// Used when an update notification provides a direct link to the new release.
func (a *App) OpenReleaseURL(releaseURL string) error {
	if releaseURL == "" {
		return a.OpenReleasesPage()
	}
	// Validate URL scheme and host to prevent opening arbitrary/malicious URLs
	parsed, err := url.Parse(releaseURL)
	if err != nil || (parsed.Scheme != "https" && parsed.Scheme != "http") {
		return errors.New(errors.CONFIG_READ_FAILED, "invalid release URL")
	}
	if !strings.HasSuffix(parsed.Host, "github.com") {
		return errors.New(errors.CONFIG_READ_FAILED, "release URL must be from github.com")
	}
	log.Printf("Opening release URL: %s", releaseURL)
	runtime.BrowserOpenURL(a.ctx, releaseURL)
	return nil
}

// CanSelfUpdate reports whether the running platform supports applying updates
// in place (Windows and Linux). macOS ships as a .dmg app bundle and must be
// updated manually, so the frontend should fall back to the download link.
func (a *App) CanSelfUpdate() bool {
	return version.CanSelfUpdate()
}

// DownloadAndInstallUpdate downloads the latest release for the current
// platform, verifies its checksum, and replaces the running executable.
// It shares the updater's single download path with the automatic background
// checker, so concurrent downloads are impossible. Download progress is
// streamed to the frontend via the UpdateDownloadProgress event; on
// completion an UpdateReady event carries the new version and on failure an
// UpdateError event carries a message.
//
// The application must be restarted for the update to take effect — the
// frontend is responsible for prompting the user to do so.
func (a *App) DownloadAndInstallUpdate() (*version.UpdateInfo, error) {
	if !version.CanSelfUpdate() {
		a.bus.Emit(events.UpdateError, version.ErrUpdateNotSupported.Error())
		return nil, errors.Wrap(version.ErrUpdateNotSupported, errors.UNKNOWN_ERROR, "update not supported")
	}

	log.Printf("Starting in-app update download...")

	info, err := a.updater.StartDownload(a.ctx, false)
	if err != nil {
		log.Printf("ERROR: Update failed: %v", err)
		return nil, errors.Wrap(err, errors.UNKNOWN_ERROR, "failed to install update")
	}
	return info, nil
}

// RestartApplication relaunches the application so a freshly installed update
// takes effect. It spawns the (now-updated) executable and quits the current
// process.
func (a *App) RestartApplication() error {
	exe, err := os.Executable()
	if err != nil {
		return errors.Wrap(err, errors.UNKNOWN_ERROR, "failed to locate executable")
	}
	// On Linux AppImage builds, relaunch the real .AppImage rather than the
	// temporary mount path.
	if ap := os.Getenv("APPIMAGE"); ap != "" {
		exe = ap
	}

	// Use context.Background (not a.ctx): a.ctx is cancelled by the
	// runtime.Quit below, which would otherwise terminate the relaunched
	// process. The child must outlive this one.
	// #nosec G204 G702 -- exe is our own executable ($APPIMAGE or os.Executable), not untrusted input
	cmd := exec.CommandContext(context.Background(), exe) //nolint:gosec
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return errors.Wrap(err, errors.UNKNOWN_ERROR, "failed to relaunch application")
	}

	log.Printf("Relaunching application to apply update")
	runtime.Quit(a.ctx)
	return nil
}

// ============================================================================
// Resource Monitoring (Story 6.9)
// ============================================================================

// ResourceStats contains runtime statistics for monitoring long-running stability.
type ResourceStats struct {
	Timestamp      string  `json:"timestamp"`
	MemoryAllocMB  float64 `json:"memoryAllocMB"`
	MemoryTotalMB  float64 `json:"memoryTotalMB"`
	GoroutineCount int     `json:"goroutineCount"`
}

// GetResourceStats returns current resource usage statistics.
// This is useful for debugging memory leaks and verifying long-running stability.
// Primarily used for development and troubleshooting - not displayed to users normally.
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

// ============================================================================
// Listening History
// ============================================================================

// GetListeningHistory returns the most recent listening history entries.
// Pass limit=0 or negative to get all entries (up to max stored).
func (a *App) GetListeningHistory(limit int) []history.Entry {
	if a.history == nil {
		return nil
	}
	return a.history.GetRecent(limit)
}

// GetListeningStats returns aggregate listening statistics.
func (a *App) GetListeningStats() history.Stats {
	if a.history == nil {
		return history.Stats{}
	}
	return a.history.GetStats()
}

// ClearListeningHistory removes all listening history entries.
func (a *App) ClearListeningHistory() {
	if a.history == nil {
		return
	}
	a.history.Clear()
}
