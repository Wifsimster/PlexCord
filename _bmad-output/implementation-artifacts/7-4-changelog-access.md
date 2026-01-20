# Story 7.4: Changelog Access

Status: done

## Story

As a user,
I want to view the release notes and changelog,
So that I can see what's new or changed in each version.

## Acceptance Criteria

1. **AC1: Link Available**
   - **Given** the user wants to see what changed
   - **When** accessing the changelog
   - **Then** a link to release notes is available in settings/about

2. **AC2: Browser Open**
   - **Given** the link is clicked
   - **When** the user clicks it
   - **Then** it opens the GitHub releases page in the default browser

3. **AC3: Current Version Highlight**
   - **Given** viewing releases
   - **When** navigating to GitHub
   - **Then** the current version's changes are highlighted (if available)

## Tasks / Subtasks

- [x] **Task 1: Releases URL** (AC: 1)
  - [x] GetReleasesURL() function
  - [x] Configurable GitHub repo

- [x] **Task 2: Browser Open** (AC: 2)
  - [x] OpenReleasesPage() method
  - [x] Uses Wails runtime.BrowserOpenURL

## Dev Notes

### Implementation

Releases URL in `internal/version/version.go`:

```go
const GitHubRepo = "your-username/PlexCord"

func GetReleasesURL() string {
    return fmt.Sprintf("https://github.com/%s/releases", GitHubRepo)
}
```

Browser open in `app.go`:

```go
func (a *App) OpenReleasesPage() error {
    url := version.GetReleasesURL()
    runtime.BrowserOpenURL(a.ctx, url)
    return nil
}
```

### Frontend Usage

```typescript
// View Changelog button
await OpenReleasesPage();
```

### GitHub Integration

The GitHub releases page automatically:
1. Shows all releases with their changelogs
2. Allows navigation to specific versions
3. Provides download links for each release

The GitHubRepo constant should be updated to the actual repository path before release.

### References

- [Source: internal/version/version.go:68] - GitHubRepo constant
- [Source: internal/version/version.go:166-168] - GetReleasesURL function
- [Source: app.go:1055-1062] - OpenReleasesPage binding

## Dev Agent Record

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Completion Notes List

1. **GitHub Releases**: Uses GitHub releases page for changelog.

2. **Browser Open**: Cross-platform via Wails BrowserOpenURL.

3. **Repo Config**: GitHubRepo constant needs update before release.

4. **Simple Implementation**: Leverages GitHub's built-in changelog UI.

### File List

Files implementing this story:
- `internal/version/version.go` - GetReleasesURL
- `app.go` - OpenReleasesPage binding
