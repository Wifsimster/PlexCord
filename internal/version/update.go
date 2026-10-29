package version

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/minio/selfupdate"
)

var ErrUpdateNotSupported = errors.New("in-app update is not supported on this platform; please download the update manually")

type ProgressFunc func(downloaded, total int64)

var sha256Pattern = regexp.MustCompile(`[0-9a-fA-F]{64}`)

func updatableAssetName() (name string, supported bool) {
	switch runtime.GOOS {
	case "windows":
		if runtime.GOARCH == "amd64" {
			return "PlexCord-windows-amd64.exe", true
		}
	case "linux":
		if runtime.GOARCH == "amd64" {
			return "PlexCord-linux-amd64.AppImage", true
		}
	}
	return "", false
}

var (
	launchPathOnce sync.Once
	launchPathVal  string
)

func resolveLaunchPath() string {
	if runtime.GOOS == "linux" {
		if p := os.Getenv("APPIMAGE"); p != "" {
			return p
		}
	}
	if exe, err := os.Executable(); err == nil {
		return exe
	}
	return ""
}

func CaptureLaunchPath() {
	launchPathOnce.Do(func() {
		launchPathVal = resolveLaunchPath()
	})
}

func LaunchPath() string {
	launchPathOnce.Do(func() {
		launchPathVal = resolveLaunchPath()
	})
	return launchPathVal
}

func updateTargetPath() string {
	return LaunchPath()
}

func CanSelfUpdate() bool {
	_, supported := updatableAssetName()
	return supported
}

func parseChecksum(content string) ([]byte, error) {
	match := sha256Pattern.FindString(content)
	if match == "" {
		return nil, errors.New("no SHA-256 digest found in checksum file")
	}
	return hex.DecodeString(match)
}

type progressReader struct {
	reader     io.Reader
	total      int64
	downloaded int64
	reported   int64
	progress   ProgressFunc
}

func (pr *progressReader) Read(p []byte) (int, error) {
	n, err := pr.reader.Read(p)
	if n > 0 {
		pr.downloaded += int64(n)
		if pr.progress != nil {
			if pr.downloaded-pr.reported >= 256*1024 || err == io.EOF {
				pr.progress(pr.downloaded, pr.total)
				pr.reported = pr.downloaded
			}
		}
	}
	return n, err
}

func findAsset(assets []ReleaseAsset, name string) (ReleaseAsset, bool) {
	for _, a := range assets {
		if a.Name == name {
			return a, true
		}
	}
	return ReleaseAsset{}, false
}

func releaseIsInstallable(release *GitHubRelease) bool {
	assetName, supported := updatableAssetName()
	if !supported {
		return len(release.Assets) > 0
	}
	if _, ok := findAsset(release.Assets, assetName); !ok {
		return false
	}
	_, ok := findAsset(release.Assets, assetName+".sha256")
	return ok
}

func downloadBody(ctx context.Context, client *http.Client, url string) (*http.Response, error) {
	req, err := newGitHubRequest(ctx, url)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		if cerr := resp.Body.Close(); cerr != nil {
			log.Printf("Warning: Failed to close response body: %v", cerr)
		}
		return nil, fmt.Errorf("download failed with status %d", resp.StatusCode)
	}
	return resp, nil
}

func DownloadAndApplyUpdate(ctx context.Context, progress ProgressFunc) (*UpdateInfo, error) {
	assetName, supported := updatableAssetName()
	if !supported {
		return nil, ErrUpdateNotSupported
	}

	release, notFound, err := fetchLatestRelease(ctx)
	if err != nil {
		return nil, err
	}
	if notFound || release.Draft || release.Prerelease {
		return nil, errors.New("no installable release is available")
	}
	if !isNewerVersion(release.TagName, Version) {
		return nil, errors.New("already running the latest version")
	}

	asset, ok := findAsset(release.Assets, assetName)
	if !ok {
		return nil, fmt.Errorf("release %s is published but its downloads are not ready yet; try again in a few minutes", release.TagName)
	}

	client := &http.Client{Timeout: 5 * time.Minute}

	checksum, err := fetchChecksum(ctx, client, release.Assets, assetName)
	if err != nil {
		return nil, err
	}

	log.Printf("Downloading update %s (%s, %d bytes)...", release.TagName, asset.Name, asset.Size)

	resp, err := downloadBody(ctx, client, asset.BrowserDownloadURL)
	if err != nil {
		return nil, fmt.Errorf("failed to download update: %w", err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			log.Printf("Warning: Failed to close download body: %v", cerr)
		}
	}()

	total := asset.Size
	if total == 0 {
		total = resp.ContentLength
	}
	reader := &progressReader{reader: resp.Body, total: total, progress: progress}

	opts := selfupdate.Options{
		TargetPath: updateTargetPath(),
		Checksum:   checksum,
	}
	if err := selfupdate.Apply(reader, opts); err != nil {
		if rerr := selfupdate.RollbackError(err); rerr != nil {
			return nil, fmt.Errorf("failed to apply update and rollback also failed: %w (rollback: %v)", err, rerr)
		}
		return nil, fmt.Errorf("failed to apply update: %w", err)
	}

	log.Printf("Update to %s applied successfully; restart required", release.TagName)

	return &UpdateInfo{
		Available:      false,
		CurrentVersion: Version,
		LatestVersion:  release.TagName,
		ReleaseURL:     release.HTMLURL,
		ReleaseNotes:   truncateReleaseNotes(release.Body, 500),
		PublishedAt:    release.PublishedAt.Format(time.RFC3339),
	}, nil
}

func fetchChecksum(ctx context.Context, client *http.Client, assets []ReleaseAsset, assetName string) ([]byte, error) {
	checksumAsset, ok := findAsset(assets, assetName+".sha256")
	if !ok {
		return nil, fmt.Errorf("release has no checksum asset for %s", assetName)
	}

	resp, err := downloadBody(ctx, client, checksumAsset.BrowserDownloadURL)
	if err != nil {
		return nil, fmt.Errorf("failed to download checksum: %w", err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			log.Printf("Warning: Failed to close checksum body: %v", cerr)
		}
	}()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if err != nil {
		return nil, fmt.Errorf("failed to read checksum: %w", err)
	}

	checksum, err := parseChecksum(strings.TrimSpace(string(body)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse checksum: %w", err)
	}
	return checksum, nil
}
