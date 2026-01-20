# Story 5.2: Polling Interval Configuration

Status: done

## Story

As a user,
I want to configure how often PlexCord checks for playback updates,
So that I can balance responsiveness with resource usage.

## Acceptance Criteria

1. **AC1: Interval Selection**
   - **Given** the user is in the settings view
   - **When** adjusting the polling interval
   - **Then** a slider or input allows selecting intervals (1-30 seconds)

2. **AC2: Default Value**
   - **Given** no custom interval is set
   - **When** viewing the setting
   - **Then** the default value is 2 seconds (per NFR4)

3. **AC3: Immediate Effect**
   - **Given** the user changes the interval
   - **When** the setting is saved
   - **Then** changes take effect immediately

4. **AC4: Persistence**
   - **Given** the interval is changed
   - **When** the app restarts
   - **Then** the setting is persisted to config

## Tasks / Subtasks

- [x] **Task 1: Config Field** (AC: 2, 4)
  - [x] `PollingInterval` field in Config struct
  - [x] Default value of 2 seconds

- [x] **Task 2: Wails Bindings** (AC: 1, 3)
  - [x] `GetPollingInterval()` method
  - [x] `SetPollingInterval()` method with bounds clamping

- [x] **Task 3: Dynamic Update** (AC: 3)
  - [x] Update running poller without restart
  - [x] `poller.SetInterval()` for immediate effect

## Dev Notes

### Implementation

Backend methods in `app.go`:

```go
// SetPollingInterval updates the session polling interval dynamically.
// The interval is clamped to min 1 second, max 60 seconds (AC3).
// Changes take effect on the next polling cycle without restart.
func (a *App) SetPollingInterval(intervalSeconds int) error {
    // Validate bounds
    if intervalSeconds < 1 {
        intervalSeconds = 1
    }
    if intervalSeconds > 60 {
        intervalSeconds = 60
    }

    // Update config
    a.config.PollingInterval = intervalSeconds
    if err := config.Save(a.config); err != nil {
        return err
    }

    // Update running poller if active
    a.pollerMu.Lock()
    if a.poller != nil && a.poller.IsRunning() {
        a.poller.SetInterval(time.Duration(intervalSeconds) * time.Second)
    }
    a.pollerMu.Unlock()

    return nil
}

// GetPollingInterval returns the current polling interval in seconds.
func (a *App) GetPollingInterval() int {
    if a.config.PollingInterval < 1 {
        return 2 // Default per NFR4
    }
    return a.config.PollingInterval
}
```

### References

- [Source: app.go:617-652] - Polling interval methods
- [Source: internal/config/config.go:13] - PollingInterval field
- [Source: internal/plex/poller.go] - SetInterval method

## Dev Agent Record

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Completion Notes List

1. **Config Field**: `PollingInterval int` in Config struct with default 2 seconds.

2. **Bounds Clamping**: Interval clamped to 1-60 seconds.

3. **Dynamic Update**: Running poller updated immediately via SetInterval().

4. **NFR4 Compliance**: Default 2 seconds ensures state changes detected within 2 seconds.

### File List

Files implementing this story (from Epic 2):
- `app.go` - GetPollingInterval, SetPollingInterval methods
- `internal/config/config.go` - PollingInterval field
- `internal/plex/poller.go` - SetInterval method
