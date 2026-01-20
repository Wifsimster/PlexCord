# Story 3.3: Discord Application Client ID Configuration

Status: review

## Story

As a user,
I want to configure my own Discord Application Client ID,
So that I can customize my Rich Presence or use a personal application.

## Acceptance Criteria

1. **AC1: Default Client ID Pre-filled**
   - **Given** the user is configuring Discord integration
   - **When** the Discord settings are displayed
   - **Then** the default PlexCord Client ID is pre-filled
   - **And** `GetDefaultDiscordClientID()` returns the default

2. **AC2: Custom Client ID Entry**
   - **Given** the user wants to use a custom Discord application
   - **When** entering a custom Client ID
   - **Then** the user can optionally enter their own Discord Application Client ID
   - **And** the input accepts valid Client ID format (numeric, 17+ digits)

3. **AC3: Format Validation**
   - **Given** a Discord Client ID is entered
   - **When** validation is triggered
   - **Then** the Client ID is validated for correct format
   - **And** invalid format shows appropriate error message
   - **And** empty string is valid (means "use default")

4. **AC4: Save Configuration**
   - **Given** a valid Client ID is entered (or empty for default)
   - **When** the user saves the configuration
   - **Then** the Client ID is saved to settings
   - **And** subsequent `GetDiscordClientID()` returns the configured value

5. **AC5: Instructions for Custom Application** (Frontend)
   - **Given** the user wants to create their own Discord application
   - **When** viewing Discord settings
   - **Then** instructions explain how to create a Discord application
   - **And** link to Discord Developer Portal is provided

## Tasks / Subtasks

- [x] **Task 1: Backend Configuration Methods** (AC: 1, 2, 4)
  - [x] `GetDefaultDiscordClientID()` returns default Client ID
  - [x] `GetDiscordClientID()` returns configured or default
  - [x] `SaveDiscordClientID()` saves to config
  - [x] Config struct has `DiscordClientID` field

- [x] **Task 2: Client ID Validation** (AC: 3)
  - [x] `isValidClientID()` validates format (17+ digits, numeric only)
  - [x] Empty string validation (allowed - means use default)
  - [x] Add `ValidateDiscordClientID()` Wails binding for frontend
  - [x] Update `SaveDiscordClientID()` to validate before saving

- [x] **Task 3: Wails Bindings** (AC: 1, 2, 4)
  - [x] Generate TypeScript bindings for Discord config methods
  - [x] Bindings available in `frontend/wailsjs/go/main/App.js`

- [ ] **Task 4: Frontend Integration** (AC: 1, 2, 3, 5)
  - [ ] Discord settings UI with Client ID input (Epic 4/5)
  - [ ] Pre-fill default Client ID
  - [ ] Validate on blur/submit
  - [ ] Show validation errors
  - [ ] Add instructions/help text

- [x] **Task 5: Unit Tests** (AC: 3)
  - [x] Test `isValidClientID()` with various inputs
  - [x] Test empty string validation
  - [x] Test `ValidateClientID()` with 7 test cases

## Dev Notes

### Existing Implementation (from Story 3-1)

Backend methods already implemented in `app.go`:

```go
// Returns default PlexCord Client ID
func (a *App) GetDefaultDiscordClientID() string {
    return discord.DefaultClientID
}

// Returns configured or default Client ID
func (a *App) GetDiscordClientID() string {
    if a.config.DiscordClientID != "" {
        return a.config.DiscordClientID
    }
    return discord.DefaultClientID
}

// Saves custom Client ID to config
func (a *App) SaveDiscordClientID(clientID string) error {
    a.config.DiscordClientID = clientID
    return config.Save(a.config)
}
```

### Client ID Validation Rules

From `internal/discord/presence.go`:

```go
func isValidClientID(clientID string) bool {
    if clientID == "" {
        return false
    }
    // Client ID should be numeric and at least 17 characters
    if len(clientID) < 17 {
        return false
    }
    for _, c := range clientID {
        if c < '0' || c > '9' {
            return false
        }
    }
    return true
}
```

**Special case**: Empty string is valid for configuration (means "use default").

### Frontend Work (Deferred)

The UI for Discord configuration will be implemented in:
- Epic 4/5: Settings view with Discord configuration section

### References

- [Source: Story 3-1] - Backend implementation
- [Source: internal/discord/presence.go:211-228] - Validation function
- [Source: app.go:565-589] - Wails bindings
- [Discord Developer Portal: https://discord.com/developers/applications]

## Dev Agent Record

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Debug Log References

(To be filled during implementation)

### Completion Notes List

1. **Backend Complete from Story 3-1**: The core configuration methods were implemented in Story 3-1:
   - `GetDefaultDiscordClientID()` - Returns default Client ID
   - `GetDiscordClientID()` - Returns configured or default
   - `SaveDiscordClientID()` - Saves to config
   - Config `DiscordClientID` field

2. **New Exported Validation Function**: Added `ValidateClientID()` to `internal/discord/presence.go`:
   - Returns `nil` for valid Client IDs
   - Returns error with `DISCORD_CLIENT_ID_INVALID` code for invalid format
   - Empty string is valid (means "use default")
   - Checks for 17+ numeric digits

3. **New Wails Binding**: Added `ValidateDiscordClientID()` binding:
   - Calls `discord.ValidateClientID()` internally
   - TypeScript binding generated: `ValidateDiscordClientID(arg1:string):Promise<void>`

4. **SaveDiscordClientID Updated**: Now validates before saving:
   - Calls `ValidateClientID()` before persisting
   - Returns validation error if format is invalid
   - Prevents saving invalid Client IDs

5. **Unit Tests**: Added `TestValidateClientID` with 7 test cases:
   - Valid client IDs (17+ digits)
   - Empty string (use default)
   - Too short
   - Contains letters
   - Contains special characters

6. **Frontend Work Deferred**: UI components for Discord configuration deferred to Epic 4/5

### File List

Files modified in this story:
- `app.go` - Added `ValidateDiscordClientID()` Wails binding, updated `SaveDiscordClientID()` with validation
- `internal/discord/presence.go` - Added `ValidateClientID()` exported function
- `internal/discord/presence_test.go` - Added `TestValidateClientID` tests
- `frontend/wailsjs/go/main/App.js` - Generated TypeScript binding
- `frontend/wailsjs/go/main/App.d.ts` - Generated TypeScript types

Files from Story 3-1 (unchanged):
- `internal/discord/types.go` - DefaultClientID constant
- `internal/config/config.go` - DiscordClientID field
