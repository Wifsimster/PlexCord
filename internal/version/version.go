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

var (
	Version = "v0.0.0-dev"

	Commit = "unknown"

	BuildDate = "unknown"
)

type Info struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildDate string `json:"buildDate"`
}

func GetInfo() Info {
	return Info{
		Version:   Version,
		Commit:    Commit,
		BuildDate: BuildDate,
	}
}

type GitHubRelease struct {
	PublishedAt time.Time      `json:"published_at"`
	TagName     string         `json:"tag_name"`
	Name        string         `json:"name"`
	Body        string         `json:"body"`
	HTMLURL     string         `json:"html_url"`
	Assets      []ReleaseAsset `json:"assets"`
	Draft       bool           `json:"draft"`
	Prerelease  bool           `json:"prerelease"`
}

type ReleaseAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
}

type UpdateInfo struct {
	CurrentVersion string `json:"currentVersion"`
	LatestVersion  string `json:"latestVersion"`
	ReleaseURL     string `json:"releaseUrl"`
	ReleaseNotes   string `json:"releaseNotes"`
	PublishedAt    string `json:"publishedAt"`
	Available      bool   `json:"available"`
}

const GitHubRepo = "Wifsimster/PlexCord"

func newGitHubRequest(ctx context.Context, url string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "PlexCord/"+Version)
	return req, nil
}

func fetchLatestRelease(ctx context.Context) (release *GitHubRelease, notFound bool, err error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", GitHubRepo)

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := newGitHubRequest(ctx, url)
	if err != nil {
		return nil, false, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, false, fmt.Errorf("failed to check for updates: %w", err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			log.Printf("Warning: Failed to close response body: %v", cerr)
		}
	}()

	if resp.StatusCode == http.StatusNotFound {
		return nil, true, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var r GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, false, fmt.Errorf("failed to decode release info: %w", err)
	}
	return &r, false, nil
}

func upToDate() *UpdateInfo {
	return &UpdateInfo{
		Available:      false,
		CurrentVersion: Version,
		LatestVersion:  Version,
	}
}

func CheckForUpdate() (*UpdateInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	release, notFound, err := fetchLatestRelease(ctx)
	if err != nil {
		return nil, err
	}

	if notFound || release.Draft || release.Prerelease {
		return upToDate(), nil
	}

	available := isNewerVersion(release.TagName, Version)

	if available && !releaseIsInstallable(release) {
		log.Printf("Release %s is published but has no downloads yet; not offering it", release.TagName)
		return upToDate(), nil
	}

	return &UpdateInfo{
		Available:      available,
		CurrentVersion: Version,
		LatestVersion:  release.TagName,
		ReleaseURL:     release.HTMLURL,
		ReleaseNotes:   truncateReleaseNotes(release.Body, 500),
		PublishedAt:    release.PublishedAt.Format(time.RFC3339),
	}, nil
}

func IsDevBuild() bool {
	return Version == "dev" || strings.Contains(Version, "-dev")
}

func isNewerVersion(latest, current string) bool {
	latest = strings.TrimPrefix(latest, "v")
	current = strings.TrimPrefix(current, "v")

	if strings.Contains(current, "-dev") {
		return true
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

func parseVersion(v string) [3]int {
	var parts [3]int

	if idx := strings.IndexAny(v, "-+"); idx != -1 {
		v = v[:idx]
	}

	segments := strings.Split(v, ".")
	for i := 0; i < len(segments) && i < 3; i++ {
		if _, err := fmt.Sscanf(segments[i], "%d", &parts[i]); err != nil {
			parts[i] = 0
		}
	}

	return parts
}

func truncateReleaseNotes(notes string, maxLen int) string {
	if len(notes) <= maxLen {
		return notes
	}
	return notes[:maxLen] + "..."
}

func GetReleasesURL() string {
	return fmt.Sprintf("https://github.com/%s/releases", GitHubRepo)
}
