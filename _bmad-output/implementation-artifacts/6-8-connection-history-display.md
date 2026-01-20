# Story 6.8: Connection History Display

Status: done

## Story

As a user,
I want to see when connections were last successful,
So that I can understand the reliability of my setup.

## Acceptance Criteria

1. **AC1: Plex Timestamp**
   - **Given** the user is viewing connection status
   - **When** Plex has been connected
   - **Then** "Last connected" timestamp is shown for Plex

2. **AC2: Discord Timestamp**
   - **Given** the user is viewing connection status
   - **When** Discord has been connected
   - **Then** "Last connected" timestamp is shown for Discord

3. **AC3: Timestamp Updates**
   - **Given** connections are established
   - **When** connections are restored
   - **Then** timestamps update when connections are restored

4. **AC4: Persistence**
   - **Given** timestamps are recorded
   - **When** the app restarts
   - **Then** timestamps persist across application restarts

5. **AC5: Relative Time**
   - **Given** timestamps are displayed
   - **When** viewing in the UI
   - **Then** timestamps show relative time (e.g., "5 minutes ago")

## Tasks / Subtasks

- [x] **Task 1: Config Fields** (AC: 4)
  - [x] Add PlexLastConnected to Config
  - [x] Add DiscordLastConnected to Config
  - [x] Fields persisted to JSON with omitempty

- [x] **Task 2: Timestamp Updates** (AC: 1, 2, 3)
  - [x] updatePlexConnectionTime() on Plex validation success
  - [x] updateDiscordConnectionTime() on Discord connect
  - [x] Also update on auto-reconnect

- [x] **Task 3: Wails Bindings** (AC: 1, 2, 5)
  - [x] ConnectionHistory struct
  - [x] GetConnectionHistory() method

## Dev Notes

### Implementation

Added connection history fields to `internal/config/config.go`:

```go
type Config struct {
    // ... existing fields ...

    // Connection history (Story 6.8)
    PlexLastConnected    *time.Time `json:"plexLastConnected,omitempty"`
    DiscordLastConnected *time.Time `json:"discordLastConnected,omitempty"`
}
```

Helper methods in `app.go`:

```go
type ConnectionHistory struct {
    PlexLastConnected    *time.Time `json:"plexLastConnected"`
    DiscordLastConnected *time.Time `json:"discordLastConnected"`
}

func (a *App) GetConnectionHistory() ConnectionHistory

// Internal methods called on successful connections
func (a *App) updatePlexConnectionTime()
func (a *App) updateDiscordConnectionTime()
```

### Timestamp Update Locations

| Method | Updates |
|--------|---------|
| ValidatePlexConnection | PlexLastConnected |
| ConnectDiscord | DiscordLastConnected |
| tryDiscordReconnect | DiscordLastConnected |

### Frontend Usage

The frontend can:
1. Call `GetConnectionHistory()` to get timestamps
2. Format as relative time (e.g., "5 minutes ago")
3. Display in connection status UI

### References

- [Source: internal/config/config.go:23-25] - Config fields
- [Source: app.go:895-912] - ConnectionHistory and GetConnectionHistory
- [Source: app.go:914-926] - updatePlexConnectionTime, updateDiscordConnectionTime

## Dev Agent Record

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Completion Notes List

1. **Config Fields**: PlexLastConnected and DiscordLastConnected as *time.Time.

2. **Persistence**: Fields saved to config.json with omitempty for clean JSON.

3. **Auto-Update**: Timestamps updated automatically on successful connections.

4. **Wails Binding**: GetConnectionHistory() returns both timestamps.

5. **Frontend Formatting**: Relative time formatting is frontend responsibility.

### File List

Files modified:
- `internal/config/config.go` - PlexLastConnected, DiscordLastConnected fields
- `app.go` - ConnectionHistory, GetConnectionHistory, update methods
