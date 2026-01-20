# Story 3.8: Presence Recovery on Discord Restart

Status: review

## Story

As a user,
I want PlexCord to reconnect to Discord if it restarts,
So that my presence continues working without manual intervention.

## Acceptance Criteria

1. **AC1: Discord Restart Detection**
   - **Given** PlexCord was connected to Discord and Discord is closed
   - **When** Discord is reopened
   - **Then** PlexCord detects Discord availability on next playback update
   - **And** reconnection is attempted automatically

2. **AC2: Auto-Reconnection**
   - **Given** Discord was restarted
   - **When** music is playing and playback update occurs
   - **Then** PlexCord automatically reconnects to Discord
   - **And** uses the configured or default Client ID

3. **AC3: Presence Restoration**
   - **Given** Discord connection is re-established
   - **When** music is currently playing
   - **Then** presence is restored with current track info
   - **And** user sees their music activity in Discord

4. **AC4: Reconnection Timing**
   - **Given** Discord becomes available
   - **When** the next polling cycle occurs
   - **Then** reconnection occurs within one polling cycle (2 seconds)

## Tasks / Subtasks

- [x] **Task 1: Connection Lost Detection** (AC: 1)
  - [x] `isConnectionLostError()` detects broken pipe, connection reset, EOF
  - [x] `SetPresence()` marks connection as lost on error

- [x] **Task 2: Auto-Reconnect Logic** (AC: 2, 4)
  - [x] `updateDiscordFromSession()` checks connection status
  - [x] Added `tryDiscordReconnect()` helper method
  - [x] Reconnects silently without event emission

- [x] **Task 3: Presence Restoration** (AC: 3)
  - [x] After successful reconnect, presence update proceeds
  - [x] Current track info is sent to Discord

- [x] **Task 4: Timing** (AC: 4)
  - [x] Reconnection attempted on each playback update cycle
  - [x] 2-second polling interval ensures quick recovery

## Dev Notes

### Implementation

Modified `updateDiscordFromSession()` in `app.go` to attempt reconnection when Discord is not connected but music is playing:

```go
func (a *App) updateDiscordFromSession(session *plex.MusicSession) {
    a.discordMu.Lock()
    defer a.discordMu.Unlock()

    // If not connected, try to reconnect (auto-recovery for Discord restart)
    if !a.discord.IsConnected() {
        a.tryDiscordReconnect()
        // If still not connected after reconnect attempt, skip update
        if !a.discord.IsConnected() {
            return
        }
        log.Printf("Discord: Reconnected - restoring presence")
    }

    // Update presence with session data
    // ...
}

func (a *App) tryDiscordReconnect() {
    clientID := a.config.DiscordClientID
    if clientID == "" {
        clientID = discord.DefaultClientID
    }

    err := a.discord.Connect(clientID)
    if err != nil {
        // Failed to reconnect - Discord probably still not running
        return
    }

    log.Printf("Discord: Auto-reconnected to Discord")
}
```

### Recovery Flow

1. Discord closes while PlexCord is running
2. Next `SetPresence()` call fails with connection lost error
3. `connected` flag is set to false
4. On next playback update (within 2 seconds), `updateDiscordFromSession()` is called
5. Detects not connected, calls `tryDiscordReconnect()`
6. If Discord is now running, reconnection succeeds
7. Presence update proceeds with current track info
8. User sees restored presence in Discord

### Silent Reconnection

The reconnection is silent (no events emitted) because:
- This is background recovery, not user-initiated
- Frontend doesn't need to know about transient disconnections
- If user needs to know, they can check `IsDiscordConnected()`

### References

- [Source: app.go] - `updateDiscordFromSession()` and `tryDiscordReconnect()`
- [Source: internal/discord/presence.go:253-261] - `isConnectionLostError()`

## Dev Agent Record

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Completion Notes List

1. **Auto-Reconnect on Playback Update**: When music is playing and Discord is not connected, `updateDiscordFromSession()` attempts reconnection before updating presence.

2. **`tryDiscordReconnect()` Method**: New helper method that attempts silent reconnection using configured or default Client ID.

3. **Presence Restoration**: After successful reconnection, the current playback update proceeds, restoring presence with current track info.

4. **Silent Recovery**: Reconnection doesn't emit events - this is background recovery. Users can check status via `IsDiscordConnected()` if needed.

5. **Timing**: Recovery happens within one polling cycle (2 seconds) as reconnection is attempted on each playback update.

### File List

Files modified:
- `app.go` - Modified `updateDiscordFromSession()`, added `tryDiscordReconnect()`
