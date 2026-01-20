# Story 2.7: Plex User Account Selection

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a user,
I want to select which Plex user account to monitor,
So that PlexCord tracks the correct user's playback on shared servers.

## Acceptance Criteria

1. **AC1: User List Retrieval**
   - **Given** the Plex connection is validated (from Story 2.6)
   - **When** the user account selection step loads
   - **Then** PlexCord queries the Plex server for available user accounts
   - **And** the query uses the authenticated token
   - **And** retrieval completes within 5 seconds

2. **AC2: Multiple Users Display**
   - **Given** multiple user accounts exist on the server
   - **When** the user list is displayed
   - **Then** each user is shown as a selectable card/item
   - **And** each card displays the user's name
   - **And** each card displays the user's avatar/thumbnail (if available)
   - **And** the currently selected user is visually highlighted
   - **And** the user can click to select a different account

3. **AC3: Single User Auto-Selection**
   - **Given** only one user account exists on the server
   - **When** the user selection step loads
   - **Then** that user is automatically selected
   - **And** the user sees a confirmation of auto-selection
   - **And** the user can proceed immediately to the next step
   - **And** a message explains "Only one user found - automatically selected"

4. **AC4: User Selection Persistence**
   - **Given** the user has selected an account
   - **When** the selection is made
   - **Then** the selected user ID is saved to configuration
   - **And** the selected user name is saved for display purposes
   - **And** the selection persists across application restarts
   - **And** the selection is used for session filtering in later stories

5. **AC5: Navigation and Proceed Control**
   - **Given** the user is on the account selection step
   - **When** no user is selected
   - **Then** the "Next" button is disabled
   - **And** when a user is selected, the "Next" button becomes enabled
   - **And** clicking "Next" proceeds to Discord setup (SetupDiscord.vue)
   - **And** clicking "Back" returns to server validation (preserves selection)

6. **AC6: Error Handling**
   - **Given** user retrieval fails
   - **When** the error is detected
   - **Then** a user-friendly error message is displayed
   - **And** a "Retry" button is available
   - **And** the user cannot proceed until users are loaded
   - **And** errors are logged for debugging

7. **AC7: Loading State**
   - **Given** user retrieval is in progress
   - **When** loading users
   - **Then** a loading indicator is displayed
   - **And** the message "Loading users..." is shown
   - **And** UI controls are appropriately disabled

## Tasks / Subtasks

- [x] **Task 1: Create User Data Types** (AC: 1, 2)
  - [x] Add `PlexUser` struct to `internal/plex/types.go`
  - [x] Include fields: `ID string`, `Name string`, `Thumb string` (avatar URL)
  - [x] Add JSON tags using camelCase for Wails serialization
  - [x] Add `UsersResponse` XML struct for parsing Plex API response

- [x] **Task 2: Implement GetUsers API Method** (AC: 1, 6, 7)
  - [x] Add `GetUsers() ([]PlexUser, error)` method to `Client` struct in `client.go`
  - [x] Query Plex endpoint `/accounts` for user info
  - [ ] For session-based user detection, check `/status/sessions` for unique users (deferred - /accounts works for setup)
  - [x] Parse XML response and extract user information
  - [x] Handle authentication and timeout (5 seconds)
  - [x] Map errors to appropriate error codes

- [x] **Task 3: Create Wails Binding for GetPlexUsers** (AC: 1, 6)
  - [x] Add method to `app.go`: `GetPlexUsers() ([]PlexUser, error)`
  - [x] Retrieve token from keychain
  - [x] Get server URL from config/setup store
  - [x] Create Plex client and call `GetUsers()`
  - [x] Log retrieval attempts and results
  - [x] Return user list or error to frontend

- [x] **Task 4: Update Setup Store for User Selection** (AC: 4, 5)
  - [x] Add `plexUsers: PlexUser[]` array to setup store state
  - [x] Add `selectedPlexUser: PlexUser | null` to store
  - [x] Add `setPlexUsers(users)` action
  - [x] Add `selectPlexUser(user)` action
  - [x] Add `isUserSelected` getter
  - [x] Update `saveState()` to persist selected user
  - [x] Update `loadState()` to restore selected user

- [x] **Task 5: Create SetupPlexUser.vue Component** (AC: 2, 3, 5, 7)
  - [x] Create new view file `frontend/src/views/SetupPlexUser.vue`
  - [x] Add loading state with spinner during user fetch
  - [x] Display user cards using PrimeVue Card or custom component
  - [x] Show user name and avatar thumbnail
  - [x] Implement selection highlighting with PrimeVue styles
  - [x] Add "Next" and "Back" navigation buttons
  - [x] Auto-fetch users on component mount

- [x] **Task 6: Implement Single User Auto-Selection** (AC: 3)
  - [x] Check if `users.length === 1` after fetch
  - [x] Automatically select the single user
  - [x] Show info message: "Only one user found - automatically selected"
  - [x] Enable "Next" button immediately
  - [x] Still allow user to proceed manually

- [x] **Task 7: Add Route to Vue Router** (AC: 5)
  - [x] Add `/setup/user` route to router configuration
  - [x] Update SetupPlex.vue "Next" to navigate to `/setup/user`
  - [x] Update SetupPlexUser.vue "Next" to navigate to `/setup/discord`
  - [x] Update SetupPlexUser.vue "Back" to navigate to `/setup/plex`

- [x] **Task 8: Update Config Types for User Storage** (AC: 4)
  - [x] Add `SelectedPlexUserID string` to config struct in `internal/config/types.go`
  - [x] Add `SelectedPlexUserName string` for display purposes
  - [x] Ensure config is saved when user selection changes

- [x] **Task 9: Create UserCard.vue Component** (AC: 2)
  - [x] Create reusable `frontend/src/components/UserCard.vue`
  - [x] Props: `user: PlexUser`, `selected: boolean`
  - [x] Display user avatar with fallback icon if no thumb
  - [x] Display user name prominently
  - [x] Emit `select` event on click
  - [x] Apply selected styling (border highlight, check mark)

- [x] **Task 10: Write Unit Tests** (AC: 1, 2, 3, 6)
  - [x] Add tests to `internal/plex/client_test.go`
  - [x] Test successful user retrieval with mock server
  - [x] Test handling of single user (auto-select scenario)
  - [x] Test handling of multiple users
  - [x] Test authentication failure
  - [x] Test timeout scenario
  - [x] Test empty user list handling

- [x] **Task 11: Integration Testing** (AC: 1-7)
  - [x] Test full flow: validation → user selection → proceed
  - [x] Test back navigation preserves selection
  - [x] Test state persistence across page refresh
  - [x] Test with real Plex server (if available)

## Dev Notes

### Context from Previous Stories

**Story 2.6 (Connection Validation) - Completed:**
- Created `internal/plex/client.go` with `Client` struct
- Established HTTP client pattern with 5-second timeout
- Uses `X-Plex-Token` header for authentication
- XML parsing with `encoding/xml` package
- Error mapping to error codes (`PLEX_AUTH_FAILED`, `PLEX_UNREACHABLE`, etc.)
- ValidationResult type with JSON tags for Wails

**Story 2.5 (Manual Server Entry) - Completed:**
- Server URL stored in setup store as `plexServerUrl`
- URL validation patterns established

**Story 2.3 (Secure Token Storage) - Completed:**
- Token stored in OS keychain
- Retrieved via `keychain.GetToken()`

**Current Story Dependencies:**
- REQUIRES: Validated connection from Story 2.6
- REQUIRES: Token from Story 2.3
- REQUIRES: Server URL from Story 2.4/2.5
- PROVIDES: Selected user for session monitoring (Story 2.8+)

### Architecture Requirements

**From Architecture.md:**

1. **Go HTTP Client Pattern:**
   - Use existing `Client` struct from `client.go`
   - Use `context.WithTimeout` for 5-second deadline
   - Set `X-Plex-Token` and `User-Agent` headers

2. **Wails Bindings:**
   - Method must be exported (capital first letter)
   - Return structs must have JSON tags (camelCase)
   - Follow existing `ValidatePlexConnection` pattern

3. **Frontend State Management:**
   - Use Pinia `setup.js` store
   - Persist to localStorage via `saveState()`
   - Follow camelCase naming convention

4. **Vue Component Naming:**
   - Views: `PascalCase.vue` in `/views/`
   - Components: `PascalCase.vue` in `/components/`

### Technical Requirements

**Plex API for Users:**

The Plex API provides user information through several endpoints. For a server-based context, the most relevant approach is:

1. **Server Accounts Endpoint:**
```
GET http://[server]:32400/accounts
Headers: X-Plex-Token: [token]

Response (XML):
<MediaContainer size="3">
  <Account id="1" name="Admin" />
  <Account id="2" name="FamilyMember" thumb="..." />
  <Account id="3" name="Guest" />
</MediaContainer>
```

2. **Alternative: Sessions Endpoint for Active Users:**
```
GET http://[server]:32400/status/sessions
Headers: X-Plex-Token: [token]

Response includes User elements with id, title, thumb attributes
```

**PlexUser Type Definition:**

```go
// internal/plex/types.go

// PlexUser represents a Plex account that can be monitored
type PlexUser struct {
    ID    string `json:"id"`    // Unique user identifier
    Name  string `json:"name"`  // Display name
    Thumb string `json:"thumb"` // Avatar URL (optional)
}

// AccountsResponse represents the XML response from /accounts endpoint
type AccountsResponse struct {
    XMLName  xml.Name       `xml:"MediaContainer"`
    Size     int            `xml:"size,attr"`
    Accounts []AccountEntry `xml:"Account"`
}

type AccountEntry struct {
    ID    string `xml:"id,attr"`
    Name  string `xml:"name,attr"`
    Thumb string `xml:"thumb,attr"`
}
```

**GetUsers Implementation Pattern:**

```go
// internal/plex/client.go

func (c *Client) GetUsers() ([]PlexUser, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    accountsURL := fmt.Sprintf("%s/accounts", c.serverURL)
    req, err := http.NewRequestWithContext(ctx, "GET", accountsURL, nil)
    if err != nil {
        return nil, errors.Wrap(err, errors.PLEX_CONN_FAILED, "failed to create accounts request")
    }

    req.Header.Set("X-Plex-Token", c.token)
    req.Header.Set("User-Agent", "PlexCord/1.0")

    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, mapHTTPError(err, ctx)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, mapHTTPStatusCode(resp.StatusCode)
    }

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, errors.Wrap(err, errors.PLEX_CONN_FAILED, "failed to read accounts response")
    }

    var accounts AccountsResponse
    if err := xml.Unmarshal(body, &accounts); err != nil {
        return nil, errors.Wrap(err, errors.PLEX_CONN_FAILED, "invalid accounts response format")
    }

    users := make([]PlexUser, len(accounts.Accounts))
    for i, acc := range accounts.Accounts {
        users[i] = PlexUser{
            ID:    acc.ID,
            Name:  acc.Name,
            Thumb: acc.Thumb,
        }
    }

    return users, nil
}
```

**Wails Binding:**

```go
// app.go

func (a *App) GetPlexUsers() ([]plex.PlexUser, error) {
    token, err := keychain.GetToken()
    if err != nil {
        return nil, errors.Wrap(err, errors.CONFIG_READ_FAILED, "failed to retrieve token")
    }

    serverURL := a.config.PlexServerURL
    if serverURL == "" {
        return nil, errors.New(errors.CONFIG_READ_FAILED, "no server URL configured")
    }

    client := plex.NewClient(token, serverURL)
    return client.GetUsers()
}
```

**Setup Store Updates:**

```javascript
// frontend/src/stores/setup.js

// Add to state
plexUsers: [],
selectedPlexUser: null,

// Add actions
setPlexUsers(users) {
    this.plexUsers = users;
},

selectPlexUser(user) {
    this.selectedPlexUser = user;
    this.saveState();
},

// Add getter
get isUserSelected() {
    return this.selectedPlexUser !== null;
}
```

**SetupPlexUser.vue Structure:**

```vue
<template>
  <div class="setup-step">
    <h2>Select User Account</h2>
    <p>Choose which Plex user to monitor for playback</p>

    <!-- Loading State -->
    <div v-if="isLoading" class="loading">
      <ProgressSpinner />
      <p>Loading users...</p>
    </div>

    <!-- Error State -->
    <div v-else-if="error" class="error">
      <Message severity="error">{{ error }}</Message>
      <Button label="Retry" @click="fetchUsers" />
    </div>

    <!-- User List -->
    <div v-else class="user-grid">
      <UserCard
        v-for="user in users"
        :key="user.id"
        :user="user"
        :selected="selectedUser?.id === user.id"
        @select="selectUser(user)"
      />
    </div>

    <!-- Auto-select message -->
    <Message v-if="autoSelected" severity="info">
      Only one user found - automatically selected
    </Message>

    <!-- Navigation -->
    <div class="navigation">
      <Button label="Back" @click="goBack" />
      <Button label="Next" :disabled="!isUserSelected" @click="goNext" />
    </div>
  </div>
</template>
```

### Testing Requirements

**Unit Test Scenarios:**

1. **Successful User Retrieval:**
   - Mock server returns valid XML with multiple users
   - Verify all users parsed correctly
   - Verify PlexUser fields populated

2. **Single User Scenario:**
   - Mock server returns one user
   - Verify single user returned
   - Frontend should auto-select

3. **Empty Users:**
   - Mock server returns empty list
   - Handle gracefully with message
   - Allow proceeding (owner-only scenario)

4. **Authentication Failure:**
   - Mock 401 response
   - Verify `PLEX_AUTH_FAILED` error

5. **Timeout:**
   - Mock slow response (>5s)
   - Verify timeout error

### File Structure

**Files to Create:**
- `frontend/src/views/SetupPlexUser.vue` - User selection step
- `frontend/src/components/UserCard.vue` - Reusable user display card

**Files to Modify:**
- `internal/plex/types.go` - Add PlexUser and AccountsResponse types
- `internal/plex/client.go` - Add GetUsers() method
- `internal/plex/client_test.go` - Add tests for GetUsers
- `app.go` - Add GetPlexUsers() Wails binding
- `frontend/src/stores/setup.js` - Add user state and actions
- `frontend/src/router/index.js` - Add /setup/user route
- `internal/config/types.go` - Add SelectedPlexUserID field

### Project Structure Notes

**Alignment with Project Structure:**

This story continues Epic 2's setup wizard flow:
- Story 2.1: Setup Wizard Navigation ✓
- Story 2.2: Plex Token Input ✓
- Story 2.3: Secure Token Storage ✓
- Story 2.4: Plex Server Auto-Discovery ✓
- Story 2.5: Manual Server Entry ✓
- Story 2.6: Plex Connection Validation ✓
- **Story 2.7: Plex User Account Selection** ← Current
- Story 2.8: Music Session Detection (next - uses selected user)

**Dependencies:**
- REQUIRES: Validated connection from Story 2.6
- PROVIDES: Selected user ID for session filtering (Story 2.8+)
- BLOCKS: Cannot proceed to Discord setup without user selection

### Edge Cases to Handle

1. **No users returned:** Display friendly message, possibly allow proceeding (admin-only server)
2. **Plex.tv vs local auth:** Handle both authentication contexts
3. **User avatar unavailable:** Show default user icon
4. **Very long user names:** Truncate with ellipsis in UI
5. **Server with 50+ users:** Consider pagination or scrollable list

### References

- [Source: epics.md#Story 2.7]
- [Source: architecture.md#Plex Integration]
- [Source: architecture.md#Go Backend Architecture]
- [Source: architecture.md#Frontend Architecture]
- [Source: 2-6-plex-connection-validation.md#Client Implementation]
- [Source: 2-6-plex-connection-validation.md#XML Parsing Notes]

## Dev Agent Record

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Debug Log References

- All 40 Go tests pass including 11 new GetUsers tests
- Frontend builds successfully with `npm run build`

### Completion Notes List

1. **Backend Implementation Complete:**
   - Created `PlexUser`, `AccountsResponse`, and `AccountEntry` types in `types.go`
   - Implemented `GetUsers()` method with full error handling, timeout, and XML parsing
   - Added `GetPlexUsers(serverURL string)` Wails binding with token retrieval from keychain
   - Added user storage fields to Config struct

2. **Frontend Implementation Complete:**
   - Created `SetupPlexUser.vue` with loading, error, and user selection states
   - Created reusable `UserCard.vue` component with avatar support and selection highlighting
   - Added `/setup/user` route between `/setup/plex` and `/setup/discord`
   - Updated setup store with user selection state management and persistence

3. **Single User Auto-Selection (AC3):**
   - Automatically selects user when only one account exists
   - Displays info message: "Only one user found - automatically selected"
   - Next button enabled immediately

4. **Testing Complete:**
   - 11 new unit tests for GetUsers covering success, single user, empty list, auth failures, server errors, invalid XML, timeout, and network errors
   - JSON serialization test for PlexUser struct
   - All tests pass (40 total in plex package)

5. **Code Review Fixes Applied:**
   - Fixed: User step now integrated into wizard flow (added to steps array in SetupWizard.vue)
   - Fixed: AC4 persistence now saves to Go config via SavePlexUserSelection() binding
   - Fixed: Removed duplicate navigation buttons from SetupPlexUser.vue
   - Fixed: Task 2 subtask about session-based detection marked as deferred (not needed for setup)

### File List

**Files Created:**
- `frontend/src/views/SetupPlexUser.vue` - User selection wizard step
- `frontend/src/components/UserCard.vue` - Reusable user display card

**Files Modified:**
- `internal/plex/types.go` - Added PlexUser, AccountsResponse, AccountEntry types
- `internal/plex/client.go` - Added GetUsers() method
- `internal/plex/client_test.go` - Added 11 GetUsers tests
- `app.go` - Added GetPlexUsers() and SavePlexUserSelection() Wails bindings
- `frontend/src/stores/setup.js` - Added user selection state and actions
- `frontend/src/router/index.js` - Added /setup/user route
- `frontend/src/views/SetupWizard.vue` - Added /setup/user step to wizard steps array
- `internal/config/config.go` - Added SelectedPlexUserID and SelectedPlexUserName fields
