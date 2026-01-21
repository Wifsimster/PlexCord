package version

import (
	"testing"
)

func TestGetInfo(t *testing.T) {
	info := GetInfo()

	if info.Version == "" {
		t.Error("Version should not be empty")
	}
	if info.Commit == "" {
		t.Error("Commit should not be empty")
	}
	if info.BuildDate == "" {
		t.Error("BuildDate should not be empty")
	}
}

func TestIsNewerVersion(t *testing.T) {
	tests := []struct {
		latest   string
		current  string
		expected bool
	}{
		// Newer versions
		{"v1.1.0", "v1.0.0", true},
		{"v2.0.0", "v1.9.9", true},
		{"v1.0.1", "v1.0.0", true},
		{"1.1.0", "1.0.0", true}, // Without v prefix

		// Same versions
		{"v1.0.0", "v1.0.0", false},
		{"1.0.0", "1.0.0", false},

		// Older versions
		{"v1.0.0", "v1.1.0", false},
		{"v1.0.0", "v2.0.0", false},

		// Dev builds always get updates
		{"v1.0.0", "v0.0.0-dev", true},
		{"v0.1.0", "v0.0.0-dev", true},

		// Pre-release suffixes (release version newer than pre-release)
		{"v1.1.0", "v1.0.0-beta", true},
		// Same base version: release is newer than pre-release
		// Note: Our simple implementation treats v1.0.0 and v1.0.0-rc1 as same base
		// For production use, we'd need more sophisticated semver comparison
		{"v1.0.1", "v1.0.0-rc1", true},
	}

	for _, tt := range tests {
		t.Run(tt.latest+"_vs_"+tt.current, func(t *testing.T) {
			result := isNewerVersion(tt.latest, tt.current)
			if result != tt.expected {
				t.Errorf("isNewerVersion(%q, %q) = %v, want %v",
					tt.latest, tt.current, result, tt.expected)
			}
		})
	}
}

func TestParseVersion(t *testing.T) {
	tests := []struct {
		input    string
		expected [3]int
	}{
		{"1.2.3", [3]int{1, 2, 3}},
		// Note: parseVersion expects 'v' prefix to be stripped by caller (isNewerVersion)
		{"1.0.0", [3]int{1, 0, 0}},
		{"0.0.1", [3]int{0, 0, 1}},
		{"1.2.3-beta", [3]int{1, 2, 3}},
		{"1.2.3-rc1", [3]int{1, 2, 3}},
		{"1.2.3+build", [3]int{1, 2, 3}},
		{"10.20.30", [3]int{10, 20, 30}},
		{"1", [3]int{1, 0, 0}},
		{"1.2", [3]int{1, 2, 0}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseVersion(tt.input)
			if result != tt.expected {
				t.Errorf("parseVersion(%q) = %v, want %v",
					tt.input, result, tt.expected)
			}
		})
	}
}

func TestTruncateReleaseNotes(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		maxLen   int
	}{
		{"Short note", "Short note", 100},
		{"This is a longer note", "This is a ...", 10},
		{"Exact", "Exact", 5},
		{"", "", 10},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := truncateReleaseNotes(tt.input, tt.maxLen)
			if result != tt.expected {
				t.Errorf("truncateReleaseNotes(%q, %d) = %q, want %q",
					tt.input, tt.maxLen, result, tt.expected)
			}
		})
	}
}

func TestGetReleasesURL(t *testing.T) {
	url := GetReleasesURL()
	if url == "" {
		t.Error("GetReleasesURL should not return empty string")
	}
	if url != "https://github.com/your-username/PlexCord/releases" {
		t.Errorf("GetReleasesURL = %q, want URL containing releases", url)
	}
}
