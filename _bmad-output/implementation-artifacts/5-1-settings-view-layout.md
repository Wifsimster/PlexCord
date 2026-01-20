# Story 5.1: Settings View Layout

Status: done

## Story

As a user,
I want a well-organized settings page,
So that I can easily find and adjust application options.

## Acceptance Criteria

1. **AC1: Organized Layout**
   - **Given** the user is in settings
   - **When** viewing the page
   - **Then** settings are organized into logical sections

2. **AC2: Clear Labels**
   - **Given** settings are displayed
   - **When** viewing each option
   - **Then** each setting has a clear label and description

3. **AC3: Immediate Feedback**
   - **Given** a setting is changed
   - **When** the change is saved
   - **Then** feedback is provided (success/error)

## Tasks / Subtasks

- [x] **Task 1: Settings Sections** (AC: 1)
  - [x] Connection section (polling, Discord Client ID)
  - [x] Behavior section (auto-start, minimize to tray)
  - [x] About section (version, updates, changelog)
  - [x] Danger Zone section (reset application)

- [x] **Task 2: Setting Controls** (AC: 2)
  - [x] Labels and descriptions for each setting
  - [x] Appropriate input types (number, switch, text)
  - [x] Save buttons where needed

- [x] **Task 3: Feedback** (AC: 3)
  - [x] Toast notifications for success/error
  - [x] Loading states on buttons
  - [x] Confirmation dialog for reset

## Dev Notes

### Implementation

Settings view in `frontend/src/views/pages/Settings.vue`:

**Sections:**
1. **Connection**
   - Polling Interval (InputNumber, 1-60 seconds)
   - Discord Client ID (InputText with default fallback)

2. **Behavior**
   - Start on Login (InputSwitch)
   - Minimize to Tray (InputSwitch)

3. **About**
   - Version display with commit hash
   - Check for Updates button
   - Update available banner
   - View Changelog link

4. **Danger Zone**
   - Reset Application button
   - Confirmation dialog

**Feedback:**
- PrimeVue Toast for success/error messages
- Loading states on all async operations
- Confirmation dialog for destructive actions

### UI Components

| Component | PrimeVue |
|-----------|----------|
| Number input | InputNumber |
| Toggle | InputSwitch |
| Text input | InputText |
| Buttons | Button |
| Dialog | Dialog |
| Notifications | Toast |

### References

- [Source: frontend/src/views/pages/Settings.vue] - Settings view

## Dev Agent Record

### Agent Model Used

Claude Opus 4.5 (claude-opus-4-5-20251101)

### Completion Notes List

1. **Logical Sections**: Settings grouped by function.

2. **Clear Descriptions**: Each setting has label and helper text.

3. **Toast Notifications**: Immediate feedback on save.

4. **Loading States**: Visual feedback during async operations.

5. **Confirmation Dialog**: Prevents accidental reset.

### File List

Files implementing this story:
- `frontend/src/views/pages/Settings.vue` - Settings view
