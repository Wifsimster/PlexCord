# Story 4.2: Connection Status Display

Status: done

## Story

As a user,
I want to see the connection status for both Plex and Discord,
So that I know if PlexCord is working correctly.

## Acceptance Criteria

1. **AC1: Plex Status**
   - **Given** the dashboard is displayed
   - **When** viewing the connection status component
   - **Then** Plex connection status is shown (Connected/Disconnected)

2. **AC2: Discord Status**
   - **Given** the dashboard is displayed
   - **When** viewing the connection status component
   - **Then** Discord connection status is shown (Connected/Disconnected)

3. **AC3: Server Info**
   - **Given** Plex is connected
   - **When** viewing the status
   - **Then** the connected server/user name is displayed

4. **AC4: Visual Cues**
   - **Given** the status is displayed
   - **When** viewing indicators
   - **Then** status indicators use clear visual cues (color, icons)

5. **AC5: Real-time Updates**
   - **Given** connection state changes
   - **When** viewing the dashboard
   - **Then** status updates in real-time without page refresh

## Tasks / Subtasks

- [x] **Task 1: Connection Store** (AC: 5)
  - [x] Create Pinia store for connection state
  - [x] Subscribe to Wails events
  - [x] Track Plex and Discord status

- [x] **Task 2: ConnectionStatus Component** (AC: 1, 2, 3, 4)
  - [x] Status indicators with colors
  - [x] Service icons (Plex orange, Discord indigo)
  - [x] User name display for Plex
  - [x] Last connected timestamps

- [x] **Task 3: Retry Integration** (AC: 5)
  - [x] Retry buttons when disconnected
  - [x] Retry state display
  - [x] Error message display

## Dev Notes

### Implementation

Connection store in `frontend/src/stores/connection.js`:

```javascript
export const useConnectionStore = defineStore('connection', {
    state: () => ({
        plex: {
            connected: false,
            polling: false,
            inErrorState: false,
            serverUrl: '',
            userName: '',
            lastConnected: null,
            retryState: null,
        },
        discord: {
            connected: false,
            lastConnected: null,
            retryState: null,
            error: null,
        },
    }),
    // ...
});
```

ConnectionStatus component features:
- Green/Yellow/Red status indicators
- Plex (orange icon) and Discord (indigo icon)
- Retry buttons with loading state
- Last connected relative time
- Retry state display with attempt count

### Events Handled

| Event | Action |
|-------|--------|
| PlexConnectionError | Mark disconnected, show error |
| PlexConnectionRestored | Mark connected |
| DiscordConnected | Mark connected |
| DiscordDisconnected | Mark disconnected |
| PlexRetryState | Update retry info |
| DiscordRetryState | Update retry info |

### References

- [Source: frontend/src/stores/connection.js] - Connection store
- [Source: frontend/src/components/ConnectionStatus.vue] - Status component

## Dev Agent Record

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Completion Notes List

1. **Pinia Store**: Centralized connection state management.

2. **Event Subscription**: Real-time updates via Wails events.

3. **Visual Indicators**: Color-coded status dots (green/yellow/red).

4. **Retry Integration**: Connected to backend retry managers.

5. **Relative Time**: Last connected shown as "5 minutes ago".

### File List

Files implementing this story:
- `frontend/src/stores/connection.js` - Connection store
- `frontend/src/components/ConnectionStatus.vue` - Status component
