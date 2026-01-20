# Story 6.3: Manual Retry Button

Status: done

## Story

As a user,
I want to manually retry failed connections,
So that I can immediately test after fixing an issue.

## Acceptance Criteria

1. **AC1: Retry Trigger**
   - **Given** a connection error is displayed
   - **When** the user clicks the retry button
   - **Then** PlexCord attempts to reconnect immediately

2. **AC2: Loading State**
   - **Given** the retry button is clicked
   - **When** the retry attempt is in progress
   - **Then** the retry button shows a loading state

3. **AC3: Success Behavior**
   - **Given** retry is in progress
   - **When** reconnection succeeds
   - **Then** success clears the error and restores normal operation

4. **AC4: Failure Update**
   - **Given** retry is in progress
   - **When** reconnection fails
   - **Then** failure updates the error message with latest status

## Tasks / Subtasks

- [x] **Task 1: Retry Methods** (AC: 1)
  - [x] RetryPlexConnection() Wails binding
  - [x] RetryDiscordConnection() Wails binding

- [x] **Task 2: Backoff Reset** (AC: 1)
  - [x] ManualRetry() resets backoff schedule
  - [x] Immediate retry attempt

- [x] **Task 3: State Feedback** (AC: 2, 3, 4)
  - [x] RetryState events emitted during retry
  - [x] Success clears retry state
  - [x] Failure updates error in state

## Dev Notes

### Implementation

Backend methods for manual retry in `app.go`:

```go
// RetryPlexConnection manually triggers a Plex connection retry.
// Resets the backoff schedule and attempts immediately.
func (a *App) RetryPlexConnection() {
    a.plexRetry.ManualRetry()
}

// RetryDiscordConnection manually triggers a Discord connection retry.
// Resets the backoff schedule and attempts immediately.
func (a *App) RetryDiscordConnection() {
    a.discordRetry.ManualRetry()
}
```

The `ManualRetry()` method in the retry manager:
1. Stops any pending automatic retry timer
2. Resets attempt number to 0 (fresh backoff schedule)
3. Immediately triggers retry callback
4. Emits state change events

### Frontend Integration

Frontend can:
1. Call `RetryPlexConnection()` or `RetryDiscordConnection()`
2. Listen for `PlexRetryState` or `DiscordRetryState` events
3. Show loading state based on `IsRetrying`
4. Display error from `LastError` on failure
5. Clear error display when `IsRetrying` becomes false with no error

### References

- [Source: app.go:991-1001] - RetryPlexConnection, RetryDiscordConnection
- [Source: internal/retry/retry.go:122-133] - ManualRetry method

## Dev Agent Record

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Completion Notes List

1. **Wails Bindings**: RetryPlexConnection and RetryDiscordConnection exposed.

2. **Backoff Reset**: ManualRetry resets backoff and retries immediately.

3. **State Events**: Frontend receives retry state updates via events.

4. **Frontend Work**: Loading state and error display are frontend responsibility.

### File List

Files implementing this story:
- `app.go` - RetryPlexConnection, RetryDiscordConnection
- `internal/retry/retry.go` - ManualRetry method
