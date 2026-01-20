# Story 7.1: Version Display

Status: done

## Story

As a user,
I want to see the current PlexCord version,
So that I know which version I'm running for troubleshooting or support.

## Acceptance Criteria

1. **AC1: Version Number Display**
   - **Given** the user is in settings or about section
   - **When** viewing application information
   - **Then** the current version number is displayed (e.g., "v1.0.0")

2. **AC2: Semantic Versioning**
   - **Given** the version is displayed
   - **When** viewing the format
   - **Then** the version follows semantic versioning format

3. **AC3: Accessibility**
   - **Given** the version information exists
   - **When** viewing from different UI locations
   - **Then** the version is accessible from both settings and about dialog

4. **AC4: Build Consistency**
   - **Given** a built binary
   - **When** checking version info
   - **Then** the version matches the built binary version

## Tasks / Subtasks

- [x] **Task 1: Version Package** (AC: 1, 2, 4)
  - [x] Create `internal/version/version.go`
  - [x] Build-time version injection via ldflags
  - [x] Version, Commit, BuildDate variables

- [x] **Task 2: Version Info Struct** (AC: 1, 2)
  - [x] Info struct with Version, Commit, BuildDate
  - [x] GetInfo() function

- [x] **Task 3: Wails Binding** (AC: 3)
  - [x] GetVersion() method on App

## Dev Notes

### Implementation

Version package in `internal/version/version.go`:

```go
// Build-time variables - set via -ldflags
var (
    Version   = "v0.0.0-dev"
    Commit    = "unknown"
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
```

Build with version injection:

```bash
go build -ldflags "-X plexcord/internal/version.Version=v1.0.0 -X plexcord/internal/version.Commit=abc123 -X plexcord/internal/version.BuildDate=2024-01-15"
```

### Frontend Usage

```typescript
// Get current version
const version = await GetVersion();
console.log(`PlexCord ${version.version}`);
```

### References

- [Source: internal/version/version.go:11-32] - Version Info struct and GetInfo
- [Source: app.go:1029-1032] - GetVersion Wails binding

## Dev Agent Record

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Completion Notes List

1. **Build-time Injection**: Version set via ldflags at compile time.

2. **Default Values**: Dev builds show "v0.0.0-dev" for clarity.

3. **Full Info**: Includes version, git commit, and build date.

4. **Frontend Ready**: GetVersion() exposed as Wails binding.

### File List

Files implementing this story:
- `internal/version/version.go` - Version package
- `internal/version/version_test.go` - Tests
- `app.go` - GetVersion binding
