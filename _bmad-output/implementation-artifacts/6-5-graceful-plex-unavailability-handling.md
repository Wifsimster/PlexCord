# Story 6.5: Graceful Plex Unavailability Handling

Status: done

## Story

As a user,
I want PlexCord to handle Plex server outages gracefully,
So that the application remains stable and recovers automatically.

## Acceptance Criteria

1. **AC1: Detect Outage**
   - **Given** the Plex server becomes unavailable
   - **When** PlexCord detects the outage
   - **Then** the error is detected on next poll

2. **AC2: Clear Presence**
   - **Given** Plex is unavailable
   - **When** the error is detected
   - **Then** Discord presence is cleared (not stale)

3. **AC3: Status Display**
   - **Given** Plex is unavailable
   - **When** viewing the dashboard
   - **Then** the dashboard shows Plex disconnected status

4. **AC4: Error Banner**
   - **Given** Plex is unavailable
   - **When** the error is displayed
   - **Then** error banner explains the issue

5. **AC5: Automatic Retry**
   - **Given** Plex is unavailable
   - **When** the error is detected
   - **Then** automatic retry begins with backoff

6. **AC6: Auto Recovery**
   - **Given** Plex becomes available again
   - **When** the next successful poll occurs
   - **Then** connection is restored automatically

7. **AC7: Presence Restoration**
   - **Given** Plex recovers and music was playing
   - **When** playback is detected
   - **Then** presence is restored

## Tasks / Subtasks

- [x] **Task 1: Poller Error Callbacks** (AC: 1, 2, 5)
  - [x] Add onError callback to Poller
  - [x] Add onRecovered callback to Poller
  - [x] Track error state in Poller

- [x] **Task 2: Error State Handling** (AC: 1, 2)
  - [x] Detect first error (not repeated)
  - [x] Clear Discord presence on error
  - [x] Emit PlexConnectionError event

- [x] **Task 3: Recovery Handling** (AC: 6, 7)
  - [x] Detect recovery from error state
  - [x] Stop retry manager
  - [x] Update connection history
  - [x] Emit PlexConnectionRestored event

- [x] **Task 4: Status API** (AC: 3)
  - [x] PlexConnectionStatus struct
  - [x] GetPlexConnectionStatus() method
  - [x] Includes Connected, Polling, InErrorState

## Dev Notes

### Implementation

Poller error callbacks in `internal/plex/poller.go`:

```go
type Poller struct {
    // ... existing fields ...

    // Error handling (Story 6.5)
    onError       func(err error)
    onRecovered   func()
    lastErrorTime time.Time
    inErrorState  bool
}

func (p *Poller) SetErrorCallbacks(onError func(err error), onRecovered func()) {
    p.onError = onError
    p.onRecovered = onRecovered
}
```

Error handling in app.go StartSessionPolling:

```go
a.poller.SetErrorCallbacks(
    // onError: Clear Discord, emit event, start retry
    func(err error) {
        a.clearDiscordOnStop()
        runtime.EventsEmit(a.ctx, "PlexConnectionError", ...)
        a.startPlexRetry(err)
    },
    // onRecovered: Stop retry, update history, emit event
    func() {
        a.stopPlexRetry()
        a.updatePlexConnectionTime()
        runtime.EventsEmit(a.ctx, "PlexConnectionRestored", nil)
    },
)
```

Status API:

```go
type PlexConnectionStatus struct {
    Connected    bool   `json:"connected"`
    Polling      bool   `json:"polling"`
    InErrorState bool   `json:"inErrorState"`
    ServerURL    string `json:"serverUrl"`
    UserID       string `json:"userId"`
    UserName     string `json:"userName"`
}

func (a *App) GetPlexConnectionStatus() PlexConnectionStatus
```

### Frontend Events

| Event | Payload | Description |
|-------|---------|-------------|
| PlexConnectionError | `{error, errorCode}` | Plex connection failed |
| PlexConnectionRestored | `null` | Plex connection recovered |

### Flow

1. Plex server goes down
2. Next poll fails → onError callback
3. Discord presence cleared
4. PlexConnectionError event emitted
5. Retry manager starts with backoff
6. Polling continues (may succeed or fail)
7. When poll succeeds → onRecovered callback
8. PlexConnectionRestored event emitted
9. If music is playing, presence updates normally

### References

- [Source: internal/plex/poller.go:24-28] - Error state fields
- [Source: internal/plex/poller.go:119-134] - SetErrorCallbacks, IsInErrorState
- [Source: internal/plex/poller.go:207-249] - doPoll with error handling
- [Source: app.go:479-507] - Error callbacks setup
- [Source: app.go:665-693] - PlexConnectionStatus, GetPlexConnectionStatus

## Dev Agent Record

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Completion Notes List

1. **Error State Tracking**: Poller tracks error state to avoid repeated callbacks.

2. **First Error Only**: onError called only on first error, not every poll.

3. **Automatic Recovery**: onRecovered called when poll succeeds after errors.

4. **Discord Cleared**: Presence cleared immediately on Plex failure.

5. **Retry Integration**: Connected to existing retry manager (Story 6.4).

6. **Status API**: Frontend can query current Plex connection status.

### File List

Files modified:
- `internal/plex/poller.go` - Error callbacks, state tracking
- `app.go` - Error callback setup, GetPlexConnectionStatus
