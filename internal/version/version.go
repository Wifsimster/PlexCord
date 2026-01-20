// Package version provides build-time version information for PlexCord.
// Version is injected at build time using -ldflags.
package version

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

// Build-time variables - set via -ldflags
// Example: go build -ldflags "-X plexcord/internal/version.Version=v1.0.0 -X plexcord/internal/version.Commit=abc123 -X plexcord/internal/version.BuildDate=2024-01-15"
var (
	// Version is the semantic version (e.g., "v1.0.0")
	Version = "v0.0.0-dev"

	// Commit is the git commit hash
	Commit = "unknown"

	// BuildDate is the build timestamp
	BuildDate = "unknown"
)

// Info contains version information for the application.
type Info struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildDate string `json:"buildDate"`
}

// GetInfo returns the current version information.
func GetInfo() Info {
	return Info{
		Version:   Version,
		Commit:    Commit,
		BuildDate: BuildDate,
	}
}

// GitHubRelease represents a release from GitHub's API.
type GitHubRelease struct {
	TagName     string    `json:"tag_name"`
	Name        string    `json:"name"`
	Body        string    `json:"body"`
	Draft       bool      `json:"draft"`
	Prerelease  bool      `json:"prerelease"`
	PublishedAt time.Time `json:"published_at"`
	HTMLURL     string    `json:"html_url"`
}

// UpdateInfo contains information about an available update.
type UpdateInfo struct {
	Available      bool   `json:"available"`
	CurrentVersion string `json:"currentVersion"`
	LatestVersion  string `json:"latestVersion"`
	ReleaseURL     string `json:"releaseUrl"`
	ReleaseNotes   string `json:"releaseNotes"`
	PublishedAt    string `json:"publishedAt"`
}

// GitHubRepo is the repository to check for updates.
const GitHubRepo = "your-username/PlexCord"

// CheckForUpdate checks GitHub releases for a newer version.
func CheckForUpdate() (*UpdateInfo, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", GitHubRepo)

	client := &http.Client{Timeout: 10 * time.Second}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "PlexCord/"+Version)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to check for updates: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Warning: Failed to close response body: %v", err)
		}
	}()

	if resp.StatusCode == http.StatusNotFound {
		// No releases yet
		return &UpdateInfo{
			Available:      false,
			CurrentVersion: Version,
			LatestVersion:  Version,
		}, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to decode release info: %w", err)
	}

	// Skip draft and prerelease versions
	if release.Draft || release.Prerelease {
		return &UpdateInfo{
			Available:      false,
			CurrentVersion: Version,
			LatestVersion:  Version,
		}, nil
	}

	// Compare versions
	available := isNewerVersion(release.TagName, Version)

	return &UpdateInfo{
		Available:      available,
		CurrentVersion: Version,
		LatestVersion:  release.TagName,
		ReleaseURL:     release.HTMLURL,
		ReleaseNotes:   truncateReleaseNotes(release.Body, 500),
		PublishedAt:    release.PublishedAt.Format(time.RFC3339),
	}, nil
}

// isNewerVersion compares two semantic version strings.
// Returns true if latest is newer than current.
func isNewerVersion(latest, current string) bool {
	// Strip 'v' prefix if present
	latest = strings.TrimPrefix(latest, "v")
	current = strings.TrimPrefix(current, "v")

	// Handle dev versions
	if strings.Contains(current, "-dev") {
		return true // Always offer updates for dev builds
	}

	latestParts := parseVersion(latest)
	currentParts := parseVersion(current)

	for i := 0; i < 3; i++ {
		if latestParts[i] > currentParts[i] {
			return true
		}
		if latestParts[i] < currentParts[i] {
			return false
		}
	}

	return false
}

// parseVersion extracts major, minor, patch from a version string.
func parseVersion(v string) [3]int {
	var parts [3]int

	// Remove any suffix (e.g., -beta, -rc1)
	if idx := strings.IndexAny(v, "-+"); idx != -1 {
		v = v[:idx]
	}

	segments := strings.Split(v, ".")
	for i := 0; i < len(segments) && i < 3; i++ {
		_, _ = fmt.Sscanf(segments[i], "%d", &parts[i])
	}

	return parts
}

// truncateReleaseNotes limits release notes to a maximum length.
func truncateReleaseNotes(notes string, maxLen int) string {
	if len(notes) <= maxLen {
		return notes
	}
	return notes[:maxLen] + "..."
}

// GetReleasesURL returns the URL to the GitHub releases page.
func GetReleasesURL() string {
	return fmt.Sprintf("https://github.com/%s/releases", GitHubRepo)
}
