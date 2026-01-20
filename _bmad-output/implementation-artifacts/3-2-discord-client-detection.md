# Story 3.2: Discord Client Detection

Status: review

## Story

As a user,
I want PlexCord to detect when Discord is not running,
So that I receive helpful feedback instead of silent failures.

## Acceptance Criteria

1. **AC1: Discord Not Running Detection**
   - **Given** PlexCord attempts to connect to Discord
   - **When** the Discord client is not running
   - **Then** the error `DISCORD_NOT_RUNNING` is returned
   - **And** the connection attempt fails gracefully

2. **AC2: Clear Error Message**
   - **Given** Discord connection fails because Discord is not running
   - **When** the error is returned to the frontend
   - **Then** the user sees a clear message: "Discord is not running"
   - **And** the error code is `DISCORD_NOT_RUNNING`

3. **AC3: UI Disconnected State**
   - **Given** Discord connection fails
   - **When** the `DiscordDisconnected` event is emitted
   - **Then** the UI indicates Discord is disconnected
   - **And** the event includes error details (code and message)

4. **AC4: Plex Monitoring Continues**
   - **Given** Discord connection fails
   - **When** PlexCord handles the error
   - **Then** Plex session monitoring continues unaffected
   - **And** the user can still see their playback status
   - **And** Discord connection can be retried later

## Tasks / Subtasks

- [x] **Task 1: Error Code Definition** (AC: 1, 2)
  - [x] `DISCORD_NOT_RUNNING` error code exists in `internal/errors/codes.go`
  - [x] Error code has descriptive documentation
  - [x] Error code is used in Discord connection handling

- [x] **Task 2: Connection Error Detection** (AC: 1)
  - [x] `mapDiscordError()` detects "connection refused" errors
  - [x] `mapDiscordError()` detects "no such file" errors (Unix socket)
  - [x] `mapDiscordError()` detects "pipe" errors (Windows named pipe)
  - [x] All detection patterns map to `DISCORD_NOT_RUNNING`

- [x] **Task 3: Error Event Emission** (AC: 2, 3)
  - [x] `ConnectDiscord()` emits `DiscordDisconnected` event on failure
  - [x] Event payload includes error code and message
  - [x] Event uses `discord.Error` struct for frontend consumption

- [x] **Task 4: Independent Plex Monitoring** (AC: 4)
  - [x] Plex poller has no dependency on Discord connection
  - [x] Plex poller continues running regardless of Discord state
  - [x] Discord and Plex have separate state management

- [ ] **Task 5: Frontend Integration** (AC: 3)
  - [ ] Dashboard shows Discord disconnected state (Epic 4)
  - [ ] Error message displayed to user (Epic 4)
  - [ ] Retry connection option available (Epic 4)

- [x] **Task 6: Unit Tests** (AC: 1, 2, 4)
  - [x] Test `mapDiscordError()` with connection refused error
  - [x] Test error code mapping returns `DISCORD_NOT_RUNNING`
  - [x] Verify Plex and Discord independence in architecture

## Dev Notes

### Implementation Status

Most backend work was completed in Story 3-1: Discord RPC Connection. This story focuses on verifying the error detection and ensuring graceful handling when Discord is not running.

### Error Detection Patterns

The `mapDiscordError()` function in `internal/discord/presence.go` already handles:

```go
// Check for common error patterns
if strings.Contains(errStr, "connection refused") ||
    strings.Contains(errStr, "no such file") ||
    strings.Contains(errStr, "pipe") {
    return errors.New(errors.DISCORD_NOT_RUNNING, "Discord is not running")
}
```

### Event Payload Structure

```javascript
// DiscordDisconnected event when Discord not running
{
  connected: false,
  error: {
    code: "DISCORD_NOT_RUNNING",
    message: "Discord is not running"
  }
}
```

### Architecture Independence

The Plex poller (`a.poller`) and Discord (`a.discord`) are completely independent:
- Separate fields in App struct
- No shared state or coupling
- Plex monitoring continues regardless of Discord status
- Each can fail independently without affecting the other

### Frontend Work (Deferred to Epic 4)

The UI indication of Discord disconnected state will be implemented in:
- Story 4-2: Connection Status Display
- Story 4-3: Now Playing Display

### References

- [Source: Story 3-1] - Discord RPC Connection implementation
- [Source: internal/discord/presence.go:231-243] - Error mapping function
- [Source: app.go:17-31] - Independent Plex/Discord architecture

## Dev Agent Record

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Debug Log References

- Verified `mapDiscordError()` implementation from Story 3-1
- Confirmed Plex/Discord independence in app.go architecture
- Error detection patterns cover Windows (pipe) and Unix (socket) cases

### Completion Notes List

1. **Backend Complete**: All backend implementation was completed in Story 3-1. The error detection, mapping, and event emission are fully functional.

2. **Error Detection**: `mapDiscordError()` correctly identifies when Discord is not running by detecting:
   - "connection refused" - Socket/pipe connection rejected
   - "no such file" - Unix socket doesn't exist
   - "pipe" - Windows named pipe not available

3. **Event System**: `DiscordDisconnected` event is emitted with full error details including code (`DISCORD_NOT_RUNNING`) and human-readable message.

4. **Architecture Verification**: Confirmed that Plex poller and Discord manager are completely independent - no shared state, no coupling. Discord failure does not affect Plex monitoring.

5. **Frontend Deferred**: UI work for showing Discord disconnected state is deferred to Epic 4 (Dashboard & System Tray) stories where the status displays are implemented.

### File List

Files from Story 3-1 that implement this story's requirements:
- `internal/discord/presence.go` - Error detection in `mapDiscordError()`
- `internal/errors/codes.go` - `DISCORD_NOT_RUNNING` error code
- `app.go` - Event emission in `ConnectDiscord()`, independent architecture
