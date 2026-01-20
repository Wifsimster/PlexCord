# Story 4.4: Settings Navigation

Status: done

## Story

As a user,
I want to access application settings from the main window,
So that I can adjust PlexCord's behavior.

## Acceptance Criteria

1. **AC1: Settings Access**
   - **Given** the user is on the dashboard
   - **When** the settings button/link is clicked
   - **Then** the settings view is displayed

2. **AC2: Back Navigation**
   - **Given** the user is in settings
   - **When** navigation back is used
   - **Then** dashboard is displayed

3. **AC3: Responsive Transition**
   - **Given** navigation is triggered
   - **When** switching views
   - **Then** the transition is smooth and responsive (under 100ms)

## Tasks / Subtasks

- [x] **Task 1: Settings Route** (AC: 1)
  - [x] Add /settings route
  - [x] Create Settings.vue view
  - [x] Add settings button to dashboard

- [x] **Task 2: Navigation** (AC: 2, 3)
  - [x] Settings icon button in dashboard header
  - [x] Back arrow button in settings header
  - [x] Vue router navigation

- [x] **Task 3: Quick Actions** (AC: 1)
  - [x] Settings button in quick actions section
  - [x] Changelog button

## Dev Notes

### Implementation

Router configuration in `frontend/src/router/index.js`:

```javascript
{
    path: '/',
    component: AppLayout,
    children: [
        {
            path: '',
            name: 'dashboard',
            component: () => import('@/views/pages/Dashboard.vue')
        },
        {
            path: 'settings',
            name: 'settings',
            component: () => import('@/views/pages/Settings.vue')
        }
    ]
}
```

Dashboard settings access:
- Header: Settings icon button (pi-cog)
- Quick Actions: "Settings" button

Settings back navigation:
- Header: Back arrow button (pi-arrow-left)
- Uses router.push('/')

### Settings View Features

The Settings view (Story 5.1) includes:
- Polling interval configuration
- Discord Client ID
- Auto-start on login toggle
- Minimize to tray toggle
- Version display
- Update check
- Reset application

### References

- [Source: frontend/src/router/index.js] - Router configuration
- [Source: frontend/src/views/pages/Dashboard.vue] - Dashboard navigation
- [Source: frontend/src/views/pages/Settings.vue] - Settings view

## Dev Agent Record

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Completion Notes List

1. **Clean Routes**: Simplified router with only PlexCord routes.

2. **Multiple Access Points**: Header button and quick actions.

3. **Back Navigation**: Clear back button in settings header.

4. **Lazy Loading**: Settings view is lazy-loaded for performance.

### File List

Files implementing this story:
- `frontend/src/router/index.js` - Route configuration
- `frontend/src/views/pages/Dashboard.vue` - Settings navigation
- `frontend/src/views/pages/Settings.vue` - Settings view
