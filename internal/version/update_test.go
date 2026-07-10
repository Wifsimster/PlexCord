package version

import (
	"encoding/hex"
	"runtime"
	"strings"
	"testing"
)

func TestParseChecksum(t *testing.T) {
	digest := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"

	tests := []struct {
		name    string
		content string
		want    string
		wantErr bool
	}{
		{
			name:    "sha256sum format",
			content: digest + "  PlexCord-linux-amd64.AppImage",
			want:    digest,
		},
		{
			name:    "shasum format",
			content: digest + "  PlexCord-darwin-universal.dmg",
			want:    digest,
		},
		{
			name: "certutil format",
			content: "SHA256 hash of PlexCord-windows-amd64.exe:\r\n" +
				digest + "\r\nCertUtil: -hashfile command completed successfully.",
			want: digest,
		},
		{
			name:    "uppercase digest",
			content: strings.ToUpper(digest),
			want:    digest,
		},
		{
			name:    "no digest present",
			content: "not a checksum file",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseChecksum(tt.content)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("parseChecksum(%q) expected error, got nil", tt.content)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseChecksum(%q) unexpected error: %v", tt.content, err)
			}
			if hex.EncodeToString(got) != strings.ToLower(tt.want) {
				t.Errorf("parseChecksum(%q) = %x, want %s", tt.content, got, tt.want)
			}
		})
	}
}

func TestUpdatableAssetName(t *testing.T) {
	name, supported := updatableAssetName()

	switch runtime.GOOS {
	case "windows":
		if runtime.GOARCH == "amd64" {
			if !supported || name != "PlexCord-windows-amd64.exe" {
				t.Errorf("windows/amd64: got (%q, %v)", name, supported)
			}
		}
	case "linux":
		if runtime.GOARCH == "amd64" {
			if !supported || name != "PlexCord-linux-amd64.AppImage" {
				t.Errorf("linux/amd64: got (%q, %v)", name, supported)
			}
		}
	case "darwin":
		if supported {
			t.Errorf("darwin should not support in-place self-update, got supported=true")
		}
	}
}

func TestCanSelfUpdateMatchesAssetSupport(t *testing.T) {
	_, supported := updatableAssetName()
	if CanSelfUpdate() != supported {
		t.Errorf("CanSelfUpdate() = %v, want %v", CanSelfUpdate(), supported)
	}
}

func TestFindAsset(t *testing.T) {
	assets := []ReleaseAsset{
		{Name: "PlexCord-linux-amd64.AppImage", BrowserDownloadURL: "https://example/app"},
		{Name: "PlexCord-linux-amd64.AppImage.sha256", BrowserDownloadURL: "https://example/sum"},
	}

	if a, ok := findAsset(assets, "PlexCord-linux-amd64.AppImage"); !ok || a.BrowserDownloadURL != "https://example/app" {
		t.Errorf("findAsset binary: got (%+v, %v)", a, ok)
	}
	if a, ok := findAsset(assets, "PlexCord-linux-amd64.AppImage.sha256"); !ok || a.BrowserDownloadURL != "https://example/sum" {
		t.Errorf("findAsset checksum: got (%+v, %v)", a, ok)
	}
	if _, ok := findAsset(assets, "does-not-exist"); ok {
		t.Errorf("findAsset missing: expected ok=false")
	}
}
