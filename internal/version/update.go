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
	"time"

	"github.com/minio/selfupdate"
)

// ErrUpdateNotSupported is returned when in-place self-update is not possible
// for the current platform. macOS is distributed as a .dmg containing an .app
// bundle, which cannot be swapped in place while running, so those users must
// download and install updates manually from the releases page.
var ErrUpdateNotSupported = errors.New("in-app update is not supported on this platform; please download the update manually")

// ProgressFunc receives download progress. total is the expected size in bytes
// (0 when the server does not report a Content-Length).
type ProgressFunc func(downloaded, total int64)

// sha256Pattern matches a bare 64-character hex SHA-256 digest. It is used to
// pull the digest out of a checksum file regardless of the tool that produced
// it (sha256sum, shasum, or Windows certutil all wrap the digest differently).
var sha256Pattern = regexp.MustCompile(`[0-9a-fA-F]{64}`)

// updatableAssetName returns the release-asset filename for the current
// platform and whether an in-place self-update is supported. The names must
// match the assets uploaded by .github/workflows/release.yml.
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
	// macOS (.dmg app bundle) and unsupported architectures fall through.
	return "", false
}

// updateTargetPath returns the path of the file that should be replaced by the
// update. For AppImage builds the running executable resolves to a read-only
// mount inside /tmp, so the real .AppImage path is taken from $APPIMAGE. An
// empty return lets selfupdate default to os.Executable().
func updateTargetPath() string {
	if runtime.GOOS == "linux" {
		if p := os.Getenv("APPIMAGE"); p != "" {
			return p
		}
	}
	return ""
}

// CanSelfUpdate reports whether the current platform supports applying updates
// in place. The frontend uses this to decide between a "Download & Install"
// button and a plain "Download" link.
func CanSelfUpdate() bool {
	_, supported := updatableAssetName()
	return supported
}

// parseChecksum extracts the SHA-256 digest bytes from the contents of a
// checksum file.
func parseChecksum(content string) ([]byte, error) {
	match := sha256Pattern.FindString(content)
	if match == "" {
		return nil, errors.New("no SHA-256 digest found in checksum file")
	}
	return hex.DecodeString(match)
}

// progressReader wraps a reader and reports cumulative bytes read through a
// callback, throttled so a large download does not flood the event bus.
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
			// Emit at most every 256 KiB, plus a final update at EOF.
			if pr.downloaded-pr.reported >= 256*1024 || err == io.EOF {
				pr.progress(pr.downloaded, pr.total)
				pr.reported = pr.downloaded
			}
		}
	}
	return n, err
}

// findAsset returns the asset with the exact given name, if present.
func findAsset(assets []ReleaseAsset, name string) (ReleaseAsset, bool) {
	for _, a := range assets {
		if a.Name == name {
			return a, true
		}
	}
	return ReleaseAsset{}, false
}

// downloadBody performs a GET and returns the response for streaming. The
// caller is responsible for closing resp.Body.
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

// DownloadAndApplyUpdate downloads the latest release binary for the current
// platform, verifies its SHA-256 checksum, and atomically replaces the running
// executable. On success the caller should prompt the user to restart the
// application for the new version to take effect.
//
// The provided progress callback is invoked periodically during the download.
// It returns the release that was applied so the caller can surface the new
// version.
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
		return nil, fmt.Errorf("release %s has no asset named %s", release.TagName, assetName)
	}

	client := &http.Client{Timeout: 5 * time.Minute}

	// Fetch and parse the checksum first so we fail fast on a bad release
	// before downloading the (much larger) binary.
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
		// Attempt rollback if selfupdate left the binary in a bad state.
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

// fetchChecksum downloads and parses the ".sha256" companion asset for the
// given binary asset.
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

	// Checksum files are tiny; cap the read defensively.
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
