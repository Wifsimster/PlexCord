# Story 6.4: Automatic Retry with Exponential Backoff

Status: done

## Story

As a user,
I want PlexCord to automatically retry failed connections,
So that temporary issues resolve without my intervention.

## Acceptance Criteria

1. **AC1: Exponential Backoff**
   - **Given** a connection fails (Plex or Discord)
   - **When** automatic retry is triggered
   - **Then** retry attempts follow exponential backoff: 5s → 10s → 30s → 60s max (NFR18)

2. **AC2: UI Updates**
   - **Given** retry is in progress
   - **When** the UI displays status
   - **Then** the UI shows "Retrying in X seconds..."

3. **AC3: Indefinite Retry**
   - **Given** connection continues to fail
   - **When** max interval is reached
   - **Then** automatic retry continues indefinitely at 60s intervals

4. **AC4: Reset on Success**
   - **Given** retry is in progress
   - **When** connection succeeds
   - **Then** successful reconnection resets the backoff timer

5. **AC5: Manual Reset**
   - **Given** automatic retry is active
   - **When** user clicks manual retry
   - **Then** manual retry resets the backoff timer

6. **AC6: Non-Blocking**
   - **Given** retry is in progress
   - **When** interacting with the app
   - **Then** automatic retry does not block UI responsiveness

## Tasks / Subtasks

- [x] **Task 1: Retry Manager** (AC: 1, 3, 6)
  - [x] Create `internal/retry/retry.go` package
  - [x] BackoffSchedule: 5s, 10s, 30s, 60s
  - [x] RetryState struct for UI
  - [x] Non-blocking timer-based retries

- [x] **Task 2: State Events** (AC: 2)
  - [x] StateChangeCallback for UI updates
  - [x] Emit PlexRetryState event
  - [x] Emit DiscordRetryState event

- [x] **Task 3: Reset Logic** (AC: 4, 5)
  - [x] Reset() method for success
  - [x] ManualRetry() resets backoff and retries immediately

- [x] **Task 4: App Integration** (AC: 1, 4, 5)
  - [x] Add plexRetry and discordRetry managers to App
  - [x] Wire up callbacks in setupRetryCallbacks()
  - [x] Call stopPlexRetry/stopDiscordRetry on success

- [x] **Task 5: Wails Bindings** (AC: 5)
  - [x] GetPlexRetryState() method
  - [x] GetDiscordRetryState() method
  - [x] RetryPlexConnection() method
  - [x] RetryDiscordConnection() method

## Dev Notes

### Implementation

Created `internal/retry/retry.go` with exponential backoff manager:

```go
var BackoffSchedule = []time.Duration{
    5 * time.Second,
    10 * time.Second,
    30 * time.Second,
    60 * time.Second, // Max interval
}

type RetryState struct {
    AttemptNumber      int           `json:"attemptNumber"`
    NextRetryIn        time.Duration `json:"nextRetryIn"`
    NextRetryAt        time.Time     `json:"nextRetryAt"`
    LastError          string        `json:"lastError,omitempty"`
    LastErrorCode      string        `json:"lastErrorCode,omitempty"`
    IsRetrying         bool          `json:"isRetrying"`
    MaxIntervalReached bool          `json:"maxIntervalReached"`
}

type Manager struct {
    // Thread-safe retry management
    // Timer-based non-blocking retries
}

func (m *Manager) Start(err error, code string)  // Begin retry cycle
func (m *Manager) Stop()                          // Cancel retries
func (m *Manager) Reset()                         // Clear on success
func (m *Manager) ManualRetry()                   // Immediate retry, reset backoff
func (m *Manager) GetState() RetryState           // Current state for UI
```

### Events Emitted

- `PlexRetryState`: RetryState when Plex retry state changes
- `DiscordRetryState`: RetryState when Discord retry state changes

### Wails Bindings

```go
func (a *App) GetPlexRetryState() retry.RetryState
func (a *App) GetDiscordRetryState() retry.RetryState
func (a *App) RetryPlexConnection()    // Manual retry
func (a *App) RetryDiscordConnection() // Manual retry
```

### References

- [Source: internal/retry/retry.go] - Retry manager package
- [Source: app.go:948-1027] - Retry integration and bindings

## Dev Agent Record

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Completion Notes List

1. **Retry Manager Package**: Complete implementation with exponential backoff.

2. **NFR18 Compliance**: Backoff schedule 5s → 10s → 30s → 60s max.

3. **Non-Blocking**: Timer-based retries don't block UI.

4. **State Events**: Frontend receives retry state updates via Wails events.

5. **Manual Override**: ManualRetry() resets backoff and retries immediately.

6. **Auth Error Handling**: Auth errors don't trigger auto-retry (require user action).

### File List

Files created/modified:
- `internal/retry/retry.go` - Retry manager package
- `app.go` - plexRetry, discordRetry managers and bindings
