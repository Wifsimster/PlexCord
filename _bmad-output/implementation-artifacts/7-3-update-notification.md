# Story 7.3: Update Notification

Status: done

## Story

As a user,
I want to be notified when a new version is available,
So that I can update when convenient.

## Acceptance Criteria

1. **AC1: Update Info**
   - **Given** a new version is detected
   - **When** the update check completes
   - **Then** the update information includes the new version number

2. **AC2: Download Button**
   - **Given** update info is displayed
   - **When** viewing the notification
   - **Then** a "Download" or "View Release" button is available

3. **AC3: Browser Open**
   - **Given** download button is clicked
   - **When** the user clicks it
   - **Then** it opens the download page in the default browser

4. **AC4: Non-Intrusive**
   - **Given** the notification is displayed
   - **When** interacting with the app
   - **Then** the notification is non-intrusive (doesn't block usage)

## Tasks / Subtasks

- [x] **Task 1: Release URL** (AC: 1, 2)
  - [x] Include release URL in UpdateInfo
  - [x] Include release notes snippet

- [x] **Task 2: Browser Open** (AC: 3)
  - [x] OpenReleaseURL() method
  - [x] Uses Wails runtime.BrowserOpenURL

- [x] **Task 3: Frontend Support** (AC: 4)
  - [x] Non-blocking API design
  - [x] Data for frontend notification

## Dev Notes

### Implementation

UpdateInfo includes release URL in `internal/version/version.go`:

```go
type UpdateInfo struct {
    Available      bool   `json:"available"`
    CurrentVersion string `json:"currentVersion"`
    LatestVersion  string `json:"latestVersion"`
    ReleaseURL     string `json:"releaseUrl"`     // GitHub release page
    ReleaseNotes   string `json:"releaseNotes"`   // Truncated notes
    PublishedAt    string `json:"publishedAt"`
}
```

Browser open in `app.go`:

```go
func (a *App) OpenReleaseURL(url string) error {
    if url == "" {
        return a.OpenReleasesPage()
    }
    runtime.BrowserOpenURL(a.ctx, url)
    return nil
}
```

### Frontend Usage

```typescript
// After checking for update
if (info.available) {
    // Show notification with button
    // Button calls: await OpenReleaseURL(info.releaseUrl)
}
```

### Notification Design

The frontend should:
1. Show non-modal notification (toast or banner)
2. Display version number and brief changelog
3. Include "Download" button that calls OpenReleaseURL
4. Allow dismissal without blocking app usage

### References

- [Source: internal/version/version.go:52-65] - UpdateInfo with ReleaseURL
- [Source: app.go:1064-1073] - OpenReleaseURL binding

## Dev Agent Record

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Completion Notes List

1. **Release URL**: Direct link to GitHub release page included.

2. **Release Notes**: Truncated snippet for notification preview.

3. **Browser Open**: Uses Wails BrowserOpenURL for cross-platform.

4. **Non-Intrusive**: Frontend controls notification display style.

### File List

Files implementing this story:
- `internal/version/version.go` - UpdateInfo with release info
- `app.go` - OpenReleaseURL binding
