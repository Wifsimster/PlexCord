# Story 4.1: Dashboard Main View

Status: done

## Story

As a user,
I want a dashboard showing PlexCord's current state at a glance,
So that I can quickly see if everything is working.

## Acceptance Criteria

1. **AC1: Dashboard Display**
   - **Given** the user has completed setup
   - **When** the main application window is opened
   - **Then** the dashboard view is displayed

2. **AC2: Theme Support**
   - **Given** the dashboard is displayed
   - **When** viewing in different system themes
   - **Then** the dashboard respects system dark/light mode (NFR26)

3. **AC3: Responsiveness**
   - **Given** the dashboard is displayed
   - **When** interacting with UI elements
   - **Then** UI interactions respond within 100ms (NFR6)

4. **AC4: Clean Layout**
   - **Given** the dashboard is displayed
   - **When** viewing the layout
   - **Then** the layout is clean and uncluttered

## Tasks / Subtasks

- [x] **Task 1: Dashboard Route** (AC: 1)
  - [x] Set dashboard as default route ('/')
  - [x] Clean up template routes
  - [x] Add navigation guard for setup check

- [x] **Task 2: Dashboard Layout** (AC: 4)
  - [x] Header with title and version
  - [x] Two-column grid for widgets
  - [x] NowPlaying widget
  - [x] ConnectionStatus widget
  - [x] Quick actions section

- [x] **Task 3: Theme Support** (AC: 2)
  - [x] Use PrimeVue theme tokens
  - [x] Dark mode support via Tailwind classes

## Dev Notes

### Implementation

Dashboard layout in `frontend/src/views/pages/Dashboard.vue`:

```vue
<template>
    <div class="grid grid-cols-12 gap-6">
        <!-- Header with version and settings button -->
        <div class="col-span-12 flex items-center justify-between">
            <div>
                <h1>Dashboard</h1>
                <p>PlexCord status at a glance</p>
            </div>
            <Button icon="pi pi-cog" @click="goToSettings" />
        </div>

        <!-- Now Playing Widget -->
        <div class="col-span-12 lg:col-span-6">
            <NowPlaying />
        </div>

        <!-- Connection Status Widget -->
        <div class="col-span-12 lg:col-span-6">
            <ConnectionStatus />
        </div>

        <!-- Quick Actions -->
        <div class="col-span-12">
            <div class="card">
                <Button label="Settings" @click="goToSettings" />
                <Button label="View Changelog" @click="OpenReleasesPage" />
            </div>
        </div>
    </div>
</template>
```

Router updated in `frontend/src/router/index.js`:
- Dashboard is now the default route ('/')
- Removed template demo routes
- Added settings route

### References

- [Source: frontend/src/views/pages/Dashboard.vue] - Dashboard view
- [Source: frontend/src/router/index.js] - Router configuration

## Dev Agent Record

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Completion Notes List

1. **Default Route**: Dashboard is now at '/' instead of '/pages/dashboard'.

2. **Clean Layout**: Removed template demo widgets, kept only PlexCord components.

3. **Version Display**: Shows current version in header.

4. **Settings Access**: Settings button in header and quick actions.

5. **Responsive Grid**: Uses 12-column grid with lg breakpoint.

### File List

Files implementing this story:
- `frontend/src/views/pages/Dashboard.vue` - Dashboard view
- `frontend/src/router/index.js` - Router configuration
