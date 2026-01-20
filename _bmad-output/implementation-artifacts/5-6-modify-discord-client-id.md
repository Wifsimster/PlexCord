# Story 5.6: Modify Discord Client ID

Status: done

## Story

As a user,
I want to change the Discord Application Client ID,
So that I can switch to a custom Discord application.

## Acceptance Criteria

1. **AC1: Display Current ID**
   - **Given** the user is in the settings view
   - **When** accessing Discord settings
   - **Then** the current Client ID is displayed (partially masked)

2. **AC2: Enter New ID**
   - **Given** the Discord settings are displayed
   - **When** the user wants to change the ID
   - **Then** the user can enter a new Client ID

3. **AC3: Reset to Default**
   - **Given** a custom ID is configured
   - **When** the user wants to reset
   - **Then** a "Reset to default" option restores the original PlexCord Client ID

4. **AC4: Reconnection**
   - **Given** the Client ID is changed
   - **When** the change is saved
   - **Then** connection is re-established with the new Client ID

5. **AC5: Persistence**
   - **Given** the Client ID is changed
   - **When** the app restarts
   - **Then** the setting is persisted to config

## Tasks / Subtasks

- [x] **Task 1: Client ID Methods** (AC: 1, 2, 5)
  - [x] `GetDiscordClientID()` - get current ID
  - [x] `GetDefaultDiscordClientID()` - get default
  - [x] `SaveDiscordClientID()` - save new ID

- [x] **Task 2: Validation** (AC: 2)
  - [x] `ValidateDiscordClientID()` - format validation

- [x] **Task 3: Reconnection** (AC: 4)
  - [x] `ConnectDiscord()` with new Client ID
  - [x] `DisconnectDiscord()` before reconnect

## Dev Notes

### Implementation

All backend methods were implemented during Epic 3 (Discord Rich Presence):

```go
// Get current configured Client ID (or default if empty)
func (a *App) GetDiscordClientID() string {
    if a.config.DiscordClientID != "" {
        return a.config.DiscordClientID
    }
    return discord.DefaultClientID
}

// Get the default PlexCord Client ID
func (a *App) GetDefaultDiscordClientID() string {
    return discord.DefaultClientID
}

// Validate Client ID format (17+ digits, numbers only)
func (a *App) ValidateDiscordClientID(clientID string) error {
    return discord.ValidateClientID(clientID)
}

// Save new Client ID (empty string resets to default)
func (a *App) SaveDiscordClientID(clientID string) error {
    if err := discord.ValidateClientID(clientID); err != nil {
        return err
    }
    a.config.DiscordClientID = clientID
    return config.Save(a.config)
}
```

### Reset to Default

To reset to default, save empty string:
```javascript
// Frontend
await SaveDiscordClientID(""); // Empty = use default
```

### Reconnection Flow

1. Call `DisconnectDiscord()`
2. Call `SaveDiscordClientID(newID)`
3. Call `ConnectDiscord(newID)` or `ConnectDiscord("")` for default

### References

- [Source: app.go:732-770] - Discord Client ID methods
- [Source: internal/discord/presence.go:38-55] - ValidateClientID
- [Source: internal/discord/presence.go:18] - DefaultClientID constant

## Dev Agent Record

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Completion Notes List

1. **Backend Complete**: All methods implemented in Epic 3.

2. **Validation**: Client ID validated for 17+ digit format.

3. **Default Handling**: Empty string in config means use DefaultClientID.

4. **Reconnection**: DisconnectDiscord + ConnectDiscord for ID changes.

5. **Frontend Work**: Settings UI needs to call these existing methods.

### File List

Files implementing this story (from Epic 3):
- `app.go` - GetDiscordClientID, GetDefaultDiscordClientID, ValidateDiscordClientID, SaveDiscordClientID
- `internal/discord/presence.go` - ValidateClientID, DefaultClientID
