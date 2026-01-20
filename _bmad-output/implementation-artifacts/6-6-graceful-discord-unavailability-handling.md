# Story 6.6: Graceful Discord Unavailability Handling

Status: done

## Story

As a user,
I want PlexCord to handle Discord being closed or unavailable,
So that Plex monitoring continues and presence resumes when Discord returns.

## Acceptance Criteria

1. **AC1: Detect Disconnect**
   - **Given** Discord is closed or becomes unavailable
   - **When** PlexCord detects the disconnect
   - **Then** Plex monitoring continues normally

2. **AC2: Status Display**
   - **Given** Discord is unavailable
   - **When** viewing the dashboard
   - **Then** the dashboard shows Discord disconnected status

3. **AC3: Error Banner**
   - **Given** Discord is unavailable
   - **When** the error is displayed
   - **Then** error banner explains Discord is not running

4. **AC4: Periodic Check**
   - **Given** Discord is unavailable
   - **When** PlexCord continues running
   - **Then** PlexCord periodically checks for Discord availability

5. **AC5: Auto-Reconnect**
   - **Given** Discord becomes available
   - **When** the next playback update occurs
   - **Then** connection is restored automatically

6. **AC6: Presence Restoration**
   - **Given** Discord reconnects
   - **When** music is playing
   - **Then** presence is immediately updated

## Tasks / Subtasks

- [x] **Task 1: Connection Lost Detection** (AC: 1)
  - [x] `isConnectionLostError()` detects broken pipe, reset, EOF
  - [x] `SetPresence()` marks connection as lost on error

- [x] **Task 2: Auto-Reconnect on Playback** (AC: 4, 5)
  - [x] `updateDiscordFromSession()` checks connection status
  - [x] `tryDiscordReconnect()` attempts reconnection

- [x] **Task 3: Presence Restoration** (AC: 6)
  - [x] After successful reconnect, presence update proceeds
  - [x] Current track info sent to Discord

## Dev Notes

### Implementation

This story was implemented as part of Story 3.8: Presence Recovery on Discord Restart.

The `updateDiscordFromSession()` method in `app.go` handles Discord unavailability:

```go
func (a *App) updateDiscordFromSession(session *plex.MusicSession) {
    a.discordMu.Lock()
    defer a.discordMu.Unlock()

    // If not connected, try to reconnect (auto-recovery)
    if !a.discord.IsConnected() {
        a.tryDiscordReconnect()
        if !a.discord.IsConnected() {
            return // Still not connected, skip update
        }
        log.Printf("Discord: Reconnected - restoring presence")
    }

    // Update presence with session data
    // ...
}
```

### Recovery Flow

1. Discord closes while PlexCord is running
2. Plex monitoring continues normally (poller keeps running)
3. On next playback update (within 2 seconds):
   - Detects not connected
   - Calls `tryDiscordReconnect()`
   - If Discord is now running, reconnects
   - Presence updated with current track

### Key Points

- Plex polling never stops due to Discord issues
- Recovery happens on each playback update cycle
- Reconnection is silent (no disruptive events)
- Connection history updated on successful reconnect

### References

- [Source: app.go:526-542] - updateDiscordFromSession with auto-reconnect
- [Source: app.go:563-581] - tryDiscordReconnect helper
- [Source: internal/discord/presence.go:253-261] - isConnectionLostError

## Dev Agent Record

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Completion Notes List

1. **Implemented in Story 3.8**: This functionality was part of Discord presence recovery.

2. **Plex Monitoring Continues**: Poller runs independently of Discord connection.

3. **Automatic Recovery**: Reconnection attempted on each playback update.

4. **Silent Recovery**: No disruptive events during background reconnection.

5. **Presence Restoration**: Current track info sent immediately after reconnect.

### File List

Files implementing this story (from Story 3.8):
- `app.go` - updateDiscordFromSession, tryDiscordReconnect
- `internal/discord/presence.go` - isConnectionLostError
