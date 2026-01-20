# Story 7.2: Manual Update Check

Status: done

## Story

As a user,
I want to check if a newer version of PlexCord is available,
So that I can stay up to date with improvements and fixes.

## Acceptance Criteria

1. **AC1: Check Button**
   - **Given** the user is in settings
   - **When** clicking "Check for Updates"
   - **Then** PlexCord checks for newer versions via GitHub releases

2. **AC2: Loading State**
   - **Given** update check is initiated
   - **When** the check is in progress
   - **Then** a loading indicator is shown during the check

3. **AC3: Update Available**
   - **Given** a new version exists
   - **When** the check completes
   - **Then** the new version number is displayed

4. **AC4: Up to Date**
   - **Given** no new version exists
   - **When** the check completes
   - **Then** "You're up to date" is shown

5. **AC5: Error Handling**
   - **Given** network issues
   - **When** the check fails
   - **Then** a clear error message is shown

## Tasks / Subtasks

- [x] **Task 1: GitHub API Integration** (AC: 1, 3, 4)
  - [x] CheckForUpdate() function
  - [x] Parse GitHub releases API response
  - [x] Compare versions

- [x] **Task 2: UpdateInfo Struct** (AC: 3)
  - [x] Available, CurrentVersion, LatestVersion
  - [x] ReleaseURL, ReleaseNotes, PublishedAt

- [x] **Task 3: Error Handling** (AC: 5)
  - [x] Network timeout handling
  - [x] API error handling

- [x] **Task 4: Wails Binding** (AC: 1, 2)
  - [x] CheckForUpdate() method on App

## Dev Notes

### Implementation

Update check in `internal/version/version.go`:

```go
type UpdateInfo struct {
    Available      bool   `json:"available"`
    CurrentVersion string `json:"currentVersion"`
    LatestVersion  string `json:"latestVersion"`
    ReleaseURL     string `json:"releaseUrl"`
    ReleaseNotes   string `json:"releaseNotes"`
    PublishedAt    string `json:"publishedAt"`
}

func CheckForUpdate() (*UpdateInfo, error) {
    // Fetches GitHub releases API
    // Compares latest vs current version
    // Returns availability and info
}
```

Version comparison:

```go
func isNewerVersion(latest, current string) bool {
    // Semantic version comparison
    // Handles v prefix, suffixes
    // Dev builds always show updates
}
```

### Frontend Usage

```typescript
// Check for updates
const info = await CheckForUpdate();
if (info.available) {
    console.log(`Update available: ${info.latestVersion}`);
} else {
    console.log("You're up to date!");
}
```

### References

- [Source: internal/version/version.go:52-65] - UpdateInfo struct
- [Source: internal/version/version.go:70-116] - CheckForUpdate function
- [Source: internal/version/version.go:119-145] - Version comparison
- [Source: app.go:1034-1053] - CheckForUpdate Wails binding

## Dev Agent Record

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Completion Notes List

1. **GitHub API**: Uses releases/latest endpoint for simplicity.

2. **Version Compare**: Semantic version comparison with suffix handling.

3. **Timeout**: 10 second HTTP client timeout.

4. **Release Notes**: Truncated to 500 chars if too long.

5. **Frontend Ready**: CheckForUpdate() exposed as Wails binding.

### File List

Files implementing this story:
- `internal/version/version.go` - CheckForUpdate, version comparison
- `internal/version/version_test.go` - Version comparison tests
- `app.go` - CheckForUpdate binding
