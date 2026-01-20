# Story 2.8: Music Session Detection

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a user,
I want PlexCord to detect when I'm playing music on Plex,
So that my listening activity can be shared.

## Acceptance Criteria

1. **AC1: Session Detection Speed**
   - **Given** PlexCord is connected to a Plex server
   - **When** the monitored user starts playing music
   - **Then** PlexCord detects the active music session within 2 seconds
   - **And** detection occurs within one polling cycle

2. **AC2: Music Type Identification**
   - **Given** the monitored user has an active session
   - **When** PlexCord polls for sessions
   - **Then** sessions are filtered to identify music only (type="track")
   - **And** video sessions (type="episode", "movie") are ignored
   - **And** photo sessions (type="photo") are ignored
   - **And** only the selected user's sessions are considered

3. **AC3: Configurable Polling Interval**
   - **Given** PlexCord is monitoring for sessions
   - **When** polling is active
   - **Then** polling occurs at the configured interval
   - **And** the default interval is 5 seconds
   - **And** the interval can be configured (min 1s, max 60s)
   - **And** interval changes take effect without restart

4. **AC4: Polling Performance (NFR5)**
   - **Given** PlexCord is polling the Plex server
   - **When** each poll request is made
   - **Then** the request completes within 500ms
   - **And** timeout handling prevents blocking on slow responses
   - **And** failed polls log errors and continue polling

5. **AC5: Memory Efficiency (NFR2)**
   - **Given** PlexCord is continuously polling
   - **When** polling runs for extended periods
   - **Then** memory usage remains below 50MB
   - **And** no memory leaks occur from session data
   - **And** old session data is properly garbage collected

6. **AC6: CPU Efficiency (NFR3)**
   - **Given** PlexCord is continuously polling
   - **When** polling runs for extended periods
   - **Then** CPU usage averages below 1%
   - **And** polling sleeps between intervals (not busy-waiting)

7. **AC7: Session State Events**
   - **Given** a music session state changes
   - **When** the change is detected
   - **Then** a Wails event `PlaybackUpdated` is emitted
   - **And** the event contains session/track information
   - **And** the frontend can subscribe to updates

8. **AC8: Poller Lifecycle Management**
   - **Given** PlexCord needs to manage polling
   - **When** the poller is started
   - **Then** polling begins immediately with first poll
   - **And** the poller can be stopped gracefully
   - **And** stopping the poller cleans up resources
   - **And** the poller can be restarted after stopping

## Tasks / Subtasks

- [x] **Task 1: Create Session Data Types** (AC: 1, 2, 7)
  - [x] Add `Session` struct to `internal/plex/types.go`
    - Fields: `SessionKey`, `UserID`, `User`, `Type`, `Player`, `State`
  - [x] Add `SessionsResponse` XML struct for parsing `/status/sessions` response
  - [x] Add `MusicSession` struct for music-specific session data
    - Fields: `Session` (embedded), `Track`, `Artist`, `Album`, `Thumb`, `Duration`, `ViewOffset`
  - [x] Add JSON tags using camelCase per architecture conventions

- [x] **Task 2: Implement GetSessions API Method** (AC: 1, 2, 4)
  - [x] Add `GetSessions() ([]Session, error)` method to `Client` struct in `client.go`
  - [x] Query Plex endpoint `/status/sessions` for active sessions
  - [x] Use context with 500ms timeout (NFR5)
  - [x] Parse XML response and extract session information
  - [x] Filter sessions by the selected user ID
  - [x] Map errors to appropriate error codes (`PLEX_UNREACHABLE`, `PLEX_AUTH_FAILED`)
  - [x] Return empty slice (not error) when no sessions active

- [x] **Task 3: Implement Music Session Filtering** (AC: 2)
  - [x] Add `GetMusicSessions(userID string) ([]MusicSession, error)` method
  - [x] Filter sessions where Type == "track" (Plex's music type identifier)
  - [x] Extract music-specific metadata from session
  - [x] Handle sessions with missing metadata gracefully

- [x] **Task 4: Create Session Poller** (AC: 3, 5, 6, 8)
  - [x] Create `internal/plex/poller.go` with `Poller` struct
  - [x] Implement `NewPoller(client *Client, userID string, interval time.Duration) *Poller`
  - [x] Implement `Start(ctx context.Context) <-chan MusicSession` to begin polling
  - [x] Use `time.Ticker` for interval-based polling (not busy-waiting)
  - [x] Implement `Stop()` method for graceful shutdown
  - [x] Implement `SetInterval(interval time.Duration)` for dynamic interval changes
  - [x] Use channels for session state communication
  - [x] Ensure goroutine cleanup on stop (prevent leaks)

- [x] **Task 5: Add Poller Tests** (AC: 4, 5, 6)
  - [x] Create `internal/plex/poller_test.go`
  - [x] Test polling starts and emits sessions
  - [x] Test polling respects interval timing
  - [x] Test poller stops cleanly
  - [x] Test interval changes take effect
  - [x] Test error handling (network failures continue polling)
  - [x] Use mock HTTP server for deterministic testing

- [x] **Task 6: Create Wails Bindings for Session Polling** (AC: 7, 8)
  - [x] Add `StartSessionPolling() error` method to `app.go`
  - [x] Add `StopSessionPolling() error` method to `app.go`
  - [x] Store poller instance in App struct
  - [x] Emit `PlaybackUpdated` Wails event when session changes
  - [x] Emit `PlaybackStopped` Wails event when no music session detected
  - [x] Log polling lifecycle events

- [x] **Task 7: Add Session Types for Frontend** (AC: 7)
  - [x] Update Wails bindings generation (`wails generate`)
  - [x] TypeScript types generated for exported functions
  - [x] Created JSDoc types for MusicSession event payload in `frontend/src/types/events.js`

- [x] **Task 8: Update Config for Polling Interval** (AC: 3)
  - [x] `PollingInterval int` field already exists in config struct (seconds)
  - [x] Default value of 5 seconds
  - [x] Validation for min (1) and max (60) bounds in poller and app.go
  - [x] Save/load interval from config.json

## Dev Notes

### Architecture Patterns

- **Go Package Location**: `internal/plex/` per architecture document
- **Poller Pattern**: Separation of concerns - Client handles API calls, Poller handles timing/lifecycle
- **Event Naming**: Use `PascalCase` for Wails events (`PlaybackUpdated`, `PlaybackStopped`)
- **Error Handling**: Use structured error codes from `internal/errors/`
- **Logging**: Use `slog` for structured logging per architecture

### Performance Requirements (Critical)

From PRD/Architecture:
- **NFR2**: Memory usage below 50MB during idle operation
- **NFR3**: CPU usage averages below 1% during normal polling
- **NFR5**: Plex session polling completes within 500ms per request

Implementation guidance:
- Use `time.Ticker` not `time.Sleep` for accurate intervals
- Context with 500ms timeout on HTTP requests
- Avoid allocating new slices each poll when session unchanged
- Use comparison to detect actual changes before emitting events

### Plex API Details

**Endpoint**: `/status/sessions`

**Response Format** (XML):
```xml
<MediaContainer size="1">
  <Track key="/library/metadata/12345"
         title="Song Name"
         grandparentTitle="Artist Name"
         parentTitle="Album Name"
         type="track"
         thumb="/library/metadata/12345/thumb/1234567890"
         duration="180000"
         viewOffset="45000">
    <User id="1" title="username" thumb="..." />
    <Player state="playing" title="Chrome" product="Plex Web" />
  </Track>
</MediaContainer>
```

**Session Type Values**:
- `track` = Music
- `episode` = TV Episode
- `movie` = Movie
- `photo` = Photo

### Source Tree Components

**Files to Create**:
- `internal/plex/poller.go` - Session poller implementation
- `internal/plex/poller_test.go` - Poller tests

**Files to Modify**:
- `internal/plex/types.go` - Add Session, MusicSession types
- `internal/plex/client.go` - Add GetSessions() method
- `internal/plex/client_test.go` - Add GetSessions tests
- `internal/config/types.go` - Add PollingInterval field
- `app.go` - Add poller management and Wails events

### Testing Standards

- Unit tests for session type filtering
- Unit tests for poller lifecycle (start/stop/interval change)
- Mock HTTP server for API response testing
- Verify no goroutine leaks after poller stop
- Test with various session response formats (empty, single, multiple)

### Project Structure Notes

- Aligns with architecture: `internal/plex/poller.go` matches documented structure
- Follows existing patterns in `client.go` for HTTP requests with context/timeout
- Uses same error code pattern as existing methods

### References

- [Source: _bmad-output/planning-artifacts/architecture.md#Go Backend Architecture] - Package structure
- [Source: _bmad-output/planning-artifacts/architecture.md#Communication Patterns] - Wails events naming
- [Source: _bmad-output/planning-artifacts/epics.md#Story 2.8] - Acceptance criteria
- [Source: internal/plex/client.go] - Existing HTTP request patterns
- [Source: internal/plex/types.go] - Existing type definitions

## Dev Agent Record

### Agent Model Used
Claude Opus 4.5 (claude-opus-4-5-20251101)

### Debug Log References
N/A

### Completion Notes List
- All 8 tasks completed successfully
- Session types with JSON tags added to types.go
- GetSessions and GetMusicSessions methods with 500ms timeout per NFR5
- Poller uses time.Ticker for CPU efficiency (not busy-waiting)
- Poller emits initial session immediately on Start() per AC8
- sessionChanged() helper avoids duplicate event emissions
- Wails events: PlaybackUpdated and PlaybackStopped
- Frontend JSDoc types created for event payloads since MusicSession passed via events
- Config already had PollingInterval with default 5 seconds
- All tests pass (65 tests across all packages)

**Code Review Fixes Applied:**
- HIGH-1: Fixed poller `running` state not updating on context cancellation (added defer in pollLoop)
- HIGH-2: Fixed session channel not closed on stop (channel now closed in pollLoop defer, prevents goroutine leak)
- MEDIUM-3: Added Artist and Album comparison to sessionChanged() for metadata refresh detection
- Added 2 new tests: TestPollerContextCancellation (updated), TestPollerChannelClosedOnStop
- Added 2 new test cases in TestSessionChangedLogic for artist/album changes

### File List

**Created:**
- `internal/plex/poller.go` - Session poller with start/stop/interval management (221 lines)
- `internal/plex/poller_test.go` - Comprehensive poller tests (480 lines)
- `frontend/src/types/events.js` - JSDoc type definitions for event payloads

**Modified:**
- `internal/plex/types.go` - Added SessionsResponse, SessionEntry, SessionUser, SessionPlayer, Session, MusicSession types
- `internal/plex/client.go` - Added GetSessions() and GetMusicSessions() methods
- `internal/plex/client_test.go` - Added tests for session retrieval methods
- `app.go` - Added poller management, StartSessionPolling, StopSessionPolling, IsPollingActive, SetPollingInterval, GetPollingInterval methods and handleSessionUpdates goroutine
- `frontend/wailsjs/go/main/App.d.ts` - Auto-generated TypeScript types for new bindings
- `frontend/wailsjs/go/models.ts` - Auto-generated models (unchanged exports)

