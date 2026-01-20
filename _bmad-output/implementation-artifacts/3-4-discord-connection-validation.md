# Story 3.4: Discord Connection Validation

Status: review

## Story

As a user,
I want PlexCord to verify my Discord connection works during setup,
So that I know Rich Presence will function correctly.

## Acceptance Criteria

1. **AC1: Trigger Connection Validation**
   - **Given** the user has configured Discord settings
   - **When** connection validation is triggered
   - **Then** PlexCord attempts to connect to Discord using the Client ID
   - **And** `ConnectDiscord()` is called with the configured Client ID

2. **AC2: Successful Connection**
   - **Given** Discord is running and Client ID is valid
   - **When** connection succeeds
   - **Then** "Connected to Discord" is displayed (frontend)
   - **And** `DiscordConnected` Wails event is emitted
   - **And** `IsDiscordConnected()` returns true

3. **AC3: Failed Connection - Specific Errors**
   - **Given** Discord connection fails
   - **When** the error is returned
   - **Then** specific error codes indicate the failure reason:
     - `DISCORD_NOT_RUNNING` - Discord client not running
     - `DISCORD_CLIENT_ID_INVALID` - Invalid Client ID format
     - `DISCORD_CONN_FAILED` - Generic connection failure
   - **And** `DiscordDisconnected` event includes error details

4. **AC4: Retry Capability**
   - **Given** Discord connection failed
   - **When** the user starts Discord and retries
   - **Then** `ConnectDiscord()` can be called again
   - **And** connection succeeds if Discord is now running

5. **AC5: Setup Proceeds with Warning**
   - **Given** Discord validation fails
   - **When** continuing setup
   - **Then** setup can proceed without Discord connection
   - **And** user is warned that Discord features won't work
   - **And** Discord can be connected later from settings

## Tasks / Subtasks

- [x] **Task 1: Connection API** (AC: 1, 2)
  - [x] `ConnectDiscord(clientID string) error` Wails binding
  - [x] Uses configured or default Client ID
  - [x] Emits `DiscordConnected` event on success

- [x] **Task 2: Error Handling** (AC: 3)
  - [x] `DISCORD_NOT_RUNNING` for Discord not running
  - [x] `DISCORD_CLIENT_ID_INVALID` for invalid format
  - [x] `DISCORD_CONN_FAILED` for other failures
  - [x] `DiscordDisconnected` event includes error details

- [x] **Task 3: Connection Status API** (AC: 2, 4)
  - [x] `IsDiscordConnected()` returns connection status
  - [x] `DisconnectDiscord()` for cleanup
  - [x] Connection can be retried by calling `ConnectDiscord()` again

- [ ] **Task 4: Frontend Setup Integration** (AC: 1, 4, 5)
  - [ ] Discord validation step in setup wizard
  - [ ] Display connection status (success/failure)
  - [ ] Retry button for failed connections
  - [ ] Continue with warning option

- [x] **Task 5: Unit Tests** (AC: 1, 2, 3)
  - [x] Test connection with invalid Client ID
  - [x] Test error code mapping
  - [x] Test connection status tracking

## Dev Notes

### Backend Implementation (Complete from Stories 3-1, 3-2, 3-3)

All backend functionality is implemented:

**Connection Methods (app.go):**
```go
// Attempt connection with Client ID
func (a *App) ConnectDiscord(clientID string) error

// Check if connected
func (a *App) IsDiscordConnected() bool

// Disconnect
func (a *App) DisconnectDiscord() error
```

**Error Codes:**
- `DISCORD_NOT_RUNNING` - Discord client not running
- `DISCORD_CLIENT_ID_INVALID` - Invalid Client ID
- `DISCORD_CONN_FAILED` - Connection failure

**Wails Events:**
- `DiscordConnected` - Emitted on successful connection
- `DiscordDisconnected` - Emitted on failure with error details

### Frontend Work (Deferred)

The setup wizard Discord validation UI will be implemented in:
- Epic 2/4/5: Setup wizard with Discord step
- Shows "Test Connection" button
- Displays success/error messages
- Allows retry or continue with warning

### References

- [Source: Story 3-1] - Core connection implementation
- [Source: Story 3-2] - Error detection and handling
- [Source: Story 3-3] - Client ID validation
- [Source: app.go:521-562] - Wails bindings

## Dev Agent Record

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Debug Log References

N/A - Backend complete from previous stories

### Completion Notes List

1. **Backend Complete**: All backend functionality was implemented in Stories 3-1, 3-2, and 3-3:
   - `ConnectDiscord()` - Attempts connection, emits events
   - `DisconnectDiscord()` - Cleans up connection
   - `IsDiscordConnected()` - Returns status
   - Error codes for specific failure reasons
   - Wails events for frontend notification

2. **Error Handling**: Complete error detection and mapping:
   - Connection refused → `DISCORD_NOT_RUNNING`
   - No socket/pipe → `DISCORD_NOT_RUNNING`
   - Invalid format → `DISCORD_CLIENT_ID_INVALID`
   - Other errors → `DISCORD_CONN_FAILED`

3. **Retry Support**: Built-in retry capability:
   - `ConnectDiscord()` can be called multiple times
   - Disconnects existing connection if Client ID changes
   - Returns immediately if already connected with same ID

4. **Frontend Work Deferred**: UI for Discord validation step in setup deferred to Epic 4/5

### File List

Files from previous stories (unchanged):
- `app.go` - Discord Wails bindings
- `internal/discord/presence.go` - PresenceManager with Connect/Disconnect
- `internal/errors/codes.go` - Discord error codes
