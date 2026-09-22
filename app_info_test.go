package main

import (
	"testing"

	"plexcord/internal/config"
)

// TestOpenReleaseURLAcceptsOnlyGitHub verifies the host check is exact: a
// look-alike domain that merely ends in "github.com" must not be opened.
func TestOpenReleaseURLAcceptsOnlyGitHub(t *testing.T) {
	cases := []struct {
		url  string
		want bool
	}{
		{"https://github.com/wifsimster/PlexCord/releases/tag/v1.2.3", true},
		{"https://objects.github.com/x", true},
		{"https://evilgithub.com/releases", false},
		{"https://github.com.evil.example/releases", false},
		{"http://github.com/wifsimster/PlexCord/releases", false},
		{"javascript:alert(1)", false},
	}
	for _, tc := range cases {
		app := newTestApp(config.DefaultConfig())
		desktop := app.desktop.(*fakeDesktop)

		err := app.OpenReleaseURL(tc.url)
		opened := len(desktop.openedURLs) == 1
		if opened != tc.want || (err == nil) != tc.want {
			t.Errorf("OpenReleaseURL(%q): opened=%v err=%v, want opened=%v", tc.url, opened, err, tc.want)
		}
	}
}
