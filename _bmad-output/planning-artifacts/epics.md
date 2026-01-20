---
stepsCompleted: [1, 2, 3, 4]
status: complete
inputDocuments:
  - _bmad-output/planning-artifacts/prd.md
  - _bmad-output/planning-artifacts/architecture.md
---

# PlexCord - Epic Breakdown

## Overview

This document provides the complete epic and story breakdown for PlexCord, decomposing the requirements from the PRD and Architecture into implementable stories.

## Requirements Inventory

### Functional Requirements

**Plex Integration (FR1-FR7):**
- FR1: User can connect to a Plex Media Server using authentication token
- FR2: User can discover Plex servers on the local network automatically
- FR3: User can manually enter a Plex server URL if auto-discovery fails
- FR4: System can detect current music playback session from Plex server
- FR5: System can extract track metadata (title, artist, album) from Plex session
- FR6: System can detect playback state changes (play, pause, stop)
- FR7: User can select which Plex user account to monitor (if multiple exist)

**Discord Integration (FR8-FR12):**
- FR8: System can establish connection to local Discord client via RPC
- FR9: System can update Discord Rich Presence with current track information
- FR10: System can clear Discord Rich Presence when playback stops
- FR11: System can detect when Discord client is not running
- FR12: User can configure Discord Application Client ID

**Setup & Onboarding (FR13-FR18):**
- FR13: User can complete initial setup through a guided wizard
- FR14: User can view instructions for obtaining Plex authentication token
- FR15: User can see live preview of Discord presence during setup
- FR16: User can skip optional setup steps and configure later
- FR17: System can validate Plex connection before completing setup
- FR18: System can validate Discord connection before completing setup

**User Interface (FR19-FR26):**
- FR19: User can view current connection status (Plex and Discord)
- FR20: User can view currently playing track information
- FR21: User can access application settings from main window
- FR22: User can minimize application to system tray
- FR23: User can restore application from system tray
- FR24: User can view tray icon indicating connection status
- FR25: User can access quick actions from tray context menu
- FR26: User can quit application from tray menu

**Configuration & Settings (FR27-FR32):**
- FR27: User can configure polling interval for Plex session checks
- FR28: User can enable/disable auto-start on system login
- FR29: User can configure minimize-to-tray behavior
- FR30: User can modify Plex server connection settings
- FR31: User can modify Discord client ID
- FR32: User can reset application to initial state

**Error Handling & Recovery (FR33-FR38):**
- FR33: User can view clear error messages when connections fail
- FR34: User can manually retry failed connections
- FR35: System can automatically retry connections with backoff
- FR36: User can re-authenticate with Plex when token expires
- FR37: System can detect and report specific error conditions
- FR38: User can view connection history/last successful connection time

**Cross-Platform Support (FR39-FR43):**
- FR39: User can install and run application on Windows 10+
- FR40: User can install and run application on macOS 11+
- FR41: User can install and run application on Linux (Ubuntu 20.04+ equivalent)
- FR42: System can integrate with platform-native system tray
- FR43: System can store credentials securely using platform keychain

**Updates & Maintenance (FR44-FR46):**
- FR44: User can check for available updates manually
- FR45: User can view current application version
- FR46: User can access release notes/changelog

### NonFunctional Requirements

**Performance (NFR1-NFR6):**
- NFR1: Application startup time shall be less than 3 seconds
- NFR2: Memory usage shall remain below 50MB during idle operation
- NFR3: CPU usage shall average less than 1% during normal polling
- NFR4: Discord presence updates shall occur within 2 seconds of playback state change
- NFR5: Plex session polling shall complete within 500ms per request
- NFR6: UI interactions shall respond within 100ms

**Security (NFR7-NFR12):**
- NFR7: Plex tokens shall be stored using OS-native secure storage where available
- NFR8: Plex tokens shall be encrypted at rest when secure storage unavailable
- NFR9: All Plex API communication shall use HTTPS/TLS
- NFR10: No credentials shall be written to log files
- NFR11: Application shall not collect or transmit telemetry
- NFR12: Configuration files shall have appropriate file permissions

**Reliability (NFR13-NFR18):**
- NFR13: Application shall maintain operation for 30+ days without restart
- NFR14: Application shall achieve 99.9% crash-free session rate
- NFR15: Application shall gracefully handle Plex server unavailability
- NFR16: Application shall gracefully handle Discord client unavailability
- NFR17: Application shall automatically reconnect after transient failures
- NFR18: Reconnection shall use exponential backoff (5s → 10s → 30s → 60s max)

**Integration (NFR19-NFR23):**
- NFR19: Application shall support Plex Media Server API v1.x
- NFR20: Application shall support Discord RPC protocol (local IPC)
- NFR21: Application shall support mDNS/GDM for server discovery
- NFR22: Application shall function with local Plex server without internet
- NFR23: Discord integration shall function without internet (local IPC only)

**Usability (NFR24-NFR27):**
- NFR24: Setup wizard shall be completable in under 2 minutes
- NFR25: Application shall provide clear, actionable error messages
- NFR26: Application shall respect system dark/light mode preferences
- NFR27: System tray icon shall clearly indicate connection status

**Maintainability (NFR28-NFR29):**
- NFR28: Application binary shall be less than 20MB per platform
- NFR29: Application shall be distributable as single file (no dependencies)

### Additional Requirements

**From Architecture - Starter Template (CRITICAL for Epic 1):**
- Initialize project using: `wails init -n plexcord -t https://github.com/TekWizely/wails-template-primevue-sakai`
- Pre-configured stack: Vue 3, PrimeVue, TailwindCSS, Vite, Dark Mode

**From Architecture - Go Package Structure:**
- `/internal/plex/` - Plex API client, session polling, server discovery
- `/internal/discord/` - Discord RPC connection, presence management
- `/internal/config/` - Settings management, JSON persistence
- `/internal/keychain/` - Secure credential storage wrapper
- `/internal/platform/` - OS-specific abstractions (tray, autostart)
- `/internal/errors/` - Structured error types and codes

**From Architecture - Go Dependencies:**
- `github.com/zalando/go-keyring` - Cross-platform secure storage
- `github.com/emersion/go-autostart` - Auto-start on login
- `github.com/hashicorp/mdns` - Plex server discovery
- `github.com/hugolgst/rich-go` - Discord RPC protocol

**From Architecture - Frontend Structure:**
- Pinia stores: connection.ts, playback.ts, settings.ts, setup.ts
- Vue components: NowPlaying.vue, ConnectionStatus.vue, ServerCard.vue, ErrorBanner.vue
- Views: Dashboard.vue, Settings.vue, SetupWizard.vue, SetupPlex.vue, SetupDiscord.vue, SetupComplete.vue

**From Architecture - Error Code System:**
- `PLEX_UNREACHABLE` - Server not responding
- `PLEX_AUTH_FAILED` - Invalid token
- `DISCORD_NOT_RUNNING` - Discord client not detected
- `DISCORD_CONN_FAILED` - RPC connection error

**From Architecture - Platform-Specific Paths:**
- Windows: `%APPDATA%\PlexCord\config.json`
- macOS: `~/Library/Application Support/PlexCord/config.json`
- Linux: `~/.config/plexcord/config.json`

### FR Coverage Map

| FR | Epic | Description |
|----|------|-------------|
| FR1 | Epic 2 | Connect to Plex using auth token |
| FR2 | Epic 2 | Auto-discover Plex servers |
| FR3 | Epic 2 | Manual server URL entry |
| FR4 | Epic 2 | Detect music playback session |
| FR5 | Epic 2 | Extract track metadata |
| FR6 | Epic 2 | Detect playback state changes |
| FR7 | Epic 2 | Select Plex user account |
| FR8 | Epic 3 | Establish Discord RPC connection |
| FR9 | Epic 3 | Update Discord Rich Presence |
| FR10 | Epic 3 | Clear presence when playback stops |
| FR11 | Epic 3 | Detect Discord not running |
| FR12 | Epic 3 | Configure Discord Client ID |
| FR13 | Epic 2 | Setup wizard |
| FR14 | Epic 2 | Token instructions |
| FR15 | Epic 2 | Live presence preview |
| FR16 | Epic 2 | Skip optional setup steps |
| FR17 | Epic 2 | Validate Plex connection |
| FR18 | Epic 3 | Validate Discord connection |
| FR19 | Epic 4 | View connection status |
| FR20 | Epic 4 | View currently playing track |
| FR21 | Epic 4 | Access settings from main window |
| FR22 | Epic 4 | Minimize to system tray |
| FR23 | Epic 4 | Restore from system tray |
| FR24 | Epic 4 | Tray icon status indicator |
| FR25 | Epic 4 | Tray context menu quick actions |
| FR26 | Epic 4 | Quit from tray menu |
| FR27 | Epic 5 | Configure polling interval |
| FR28 | Epic 5 | Enable/disable auto-start |
| FR29 | Epic 5 | Configure minimize-to-tray |
| FR30 | Epic 5 | Modify Plex connection settings |
| FR31 | Epic 5 | Modify Discord client ID |
| FR32 | Epic 5 | Reset application |
| FR33 | Epic 6 | Clear error messages |
| FR34 | Epic 6 | Manual retry |
| FR35 | Epic 6 | Automatic retry with backoff |
| FR36 | Epic 6 | Re-authenticate expired token |
| FR37 | Epic 6 | Detect/report error conditions |
| FR38 | Epic 6 | Connection history |
| FR39 | Epic 1 | Windows 10+ support |
| FR40 | Epic 1 | macOS 11+ support |
| FR41 | Epic 1 | Linux support |
| FR42 | Epic 4 | Platform-native system tray |
| FR43 | Epic 2 | Secure keychain storage |
| FR44 | Epic 7 | Check for updates |
| FR45 | Epic 7 | View current version |
| FR46 | Epic 7 | Access changelog |

## Epic List

### Epic 1: Application Foundation & First Launch
User can download, install, and launch PlexCord on any supported platform with a functional application window.

**FRs covered:** FR39, FR40, FR41
**Architecture:** Starter template (wails-template-primevue-sakai), Go package scaffolding, error codes, config paths

---

### Epic 2: Setup Wizard & Plex Connection
User can complete the setup wizard, connect to their Plex server (via auto-discovery or manual entry), and see their music activity detected.

**FRs covered:** FR1, FR2, FR3, FR4, FR5, FR6, FR7, FR13, FR14, FR15, FR16, FR17, FR43
**Architecture:** `internal/plex/` package, `internal/keychain/`, Setup*.vue views, connection.ts store

---

### Epic 3: Discord Rich Presence
User's Discord status automatically updates to show what they're listening to on Plex.

**FRs covered:** FR8, FR9, FR10, FR11, FR12, FR18
**Architecture:** `internal/discord/` package, playback.ts store, PresenceManager

---

### Epic 4: Dashboard & System Tray
User can monitor their connection status, view currently playing track, and control PlexCord from the system tray.

**FRs covered:** FR19, FR20, FR21, FR22, FR23, FR24, FR25, FR26, FR42
**Architecture:** Dashboard.vue, NowPlaying.vue, ConnectionStatus.vue, `internal/platform/tray.go`

---

### Epic 5: Settings & Preferences
User can customize PlexCord behavior including polling interval, auto-start, and tray behavior.

**FRs covered:** FR27, FR28, FR29, FR30, FR31, FR32
**Architecture:** Settings.vue, settings.ts store, `internal/platform/autostart.go`

---

### Epic 6: Error Recovery & Resilience
User can easily identify and recover from connection issues with clear error messages and automatic retry.

**FRs covered:** FR33, FR34, FR35, FR36, FR37, FR38
**Architecture:** ErrorBanner.vue, error codes system, exponential backoff (NFR13-NFR18)

---

### Epic 7: Updates & Maintenance
User can check for updates, view current version, and access release notes.

**FRs covered:** FR44, FR45, FR46

---

## Epic 1: Application Foundation & First Launch

User can download, install, and launch PlexCord on any supported platform with a functional application window.

### Story 1.1: Project Initialization with Wails Starter Template

As a developer,
I want the PlexCord project initialized with the Wails PrimeVue Sakai starter template,
So that I have a working cross-platform desktop application foundation with Vue 3, PrimeVue, and TailwindCSS pre-configured.

**Acceptance Criteria:**

**Given** a fresh development environment with Go 1.21+ and Node.js installed
**When** the project is initialized using `wails init -n plexcord -t https://github.com/TekWizely/wails-template-primevue-sakai`
**Then** the project compiles successfully with `wails build`
**And** the application window launches with the default template UI
**And** the project includes Vue 3, PrimeVue components, and TailwindCSS styling
**And** dark mode toggle functions correctly

---

### Story 1.2: Go Backend Package Structure

As a developer,
I want the Go backend organized into internal packages,
So that code is modular and maintainable with clear separation of concerns.

**Acceptance Criteria:**

**Given** the initialized Wails project
**When** the Go package structure is created
**Then** the following packages exist with placeholder interfaces:
  - `/internal/plex/` - with `client.go` stub
  - `/internal/discord/` - with `presence.go` stub
  - `/internal/config/` - with `config.go` stub
  - `/internal/keychain/` - with `keychain.go` stub
  - `/internal/platform/` - with `platform.go` stub
  - `/internal/errors/` - with `errors.go` stub
**And** the project still compiles successfully
**And** the main `app.go` can import all internal packages

---

### Story 1.3: Configuration File Management

As a user,
I want my settings persisted to a configuration file in the appropriate OS location,
So that my preferences are preserved between application restarts.

**Acceptance Criteria:**

**Given** PlexCord is running on any supported platform
**When** settings are saved
**Then** the configuration file is created at:
  - Windows: `%APPDATA%\PlexCord\config.json`
  - macOS: `~/Library/Application Support/PlexCord/config.json`
  - Linux: `~/.config/plexcord/config.json`
**And** the config directory is created if it doesn't exist
**And** the JSON file has appropriate file permissions (readable only by user)
**And** subsequent application launches read settings from this file

---

### Story 1.4: Error Code System Foundation

As a developer,
I want a structured error code system,
So that errors can be consistently identified and handled across the application.

**Acceptance Criteria:**

**Given** the error code system is implemented in `/internal/errors/`
**When** an error occurs in any module
**Then** errors include a code from the defined set:
  - `PLEX_UNREACHABLE` - Server not responding
  - `PLEX_AUTH_FAILED` - Invalid token
  - `DISCORD_NOT_RUNNING` - Discord client not detected
  - `DISCORD_CONN_FAILED` - RPC connection error
  - `CONFIG_READ_FAILED` - Cannot read config file
  - `CONFIG_WRITE_FAILED` - Cannot write config file
**And** errors include a human-readable message
**And** errors can be serialized for frontend display
**And** no sensitive data (tokens, credentials) appears in error messages (NFR10)

---

### Story 1.5: Cross-Platform Build Verification

As a user,
I want to install PlexCord on Windows, macOS, or Linux,
So that I can use the application on my preferred operating system.

**Acceptance Criteria:**

**Given** the PlexCord project with all Epic 1 components
**When** the project is built for each platform
**Then** Windows build produces a `.exe` file that runs on Windows 10+
**And** macOS build produces an `.app` bundle that runs on macOS 11+
**And** Linux build produces a binary that runs on Ubuntu 20.04+
**And** each build is under 20MB (NFR28)
**And** each build requires no external dependencies (NFR29)
**And** each build starts within 3 seconds (NFR1)
**And** the application window displays correctly on each platform

---

## Epic 2: Setup Wizard & Plex Connection

User can complete the setup wizard, connect to their Plex server (via auto-discovery or manual entry), and see their music activity detected.

### Story 2.1: Setup Wizard Navigation Framework

As a user,
I want a guided setup wizard when I first launch PlexCord,
So that I can easily configure the application step by step.

**Acceptance Criteria:**

**Given** the user launches PlexCord for the first time (no config exists)
**When** the application starts
**Then** the setup wizard is displayed automatically
**And** the wizard shows step indicators (Plex → Discord → Complete)
**And** the user can navigate forward through wizard steps
**And** the user can navigate back to previous steps
**And** the wizard state persists if the app is closed mid-setup

---

### Story 2.2: Plex Token Input with Instructions

As a user,
I want clear instructions for obtaining my Plex authentication token,
So that I can connect PlexCord to my Plex account.

**Acceptance Criteria:**

**Given** the user is on the Plex setup step of the wizard
**When** the step is displayed
**Then** instructions explain how to obtain the Plex token from plex.tv
**And** a text input field accepts the Plex token
**And** a link opens the Plex token page in the default browser
**And** the token field masks the input for security
**And** the user can proceed only after entering a token

---

### Story 2.3: Secure Token Storage

As a user,
I want my Plex token stored securely,
So that my credentials are protected from unauthorized access.

**Acceptance Criteria:**

**Given** the user has entered a valid Plex token
**When** the token is saved
**Then** the token is stored in the OS keychain (Windows Credential Manager, macOS Keychain, Linux Secret Service)
**And** the token is never written to the config.json file
**And** the token is never written to log files (NFR10)
**And** if OS keychain is unavailable, the token is encrypted before storing locally
**And** the application can retrieve the token on subsequent launches

---

### Story 2.4: Plex Server Auto-Discovery

As a user,
I want PlexCord to automatically find my Plex server on the network,
So that I don't have to manually enter server details.

**Acceptance Criteria:**

**Given** the user has entered a valid Plex token
**When** server discovery is initiated
**Then** PlexCord uses mDNS/GDM to scan the local network
**And** discovered servers are displayed as selectable cards
**And** each server card shows the server name and address
**And** the user can select a discovered server to connect
**And** discovery completes within 5 seconds
**And** a "searching..." indicator is shown during discovery

---

### Story 2.5: Manual Server Entry

As a user,
I want to manually enter my Plex server URL,
So that I can connect when auto-discovery fails or my server is remote.

**Acceptance Criteria:**

**Given** the user is on the server selection step
**When** auto-discovery finds no servers OR the user wants to enter manually
**Then** a manual entry option is available
**And** the user can enter a server URL (e.g., http://192.168.1.100:32400)
**And** HTTPS URLs are accepted
**And** the URL is validated for correct format
**And** the user receives clear feedback if the URL format is invalid

---

### Story 2.6: Plex Connection Validation

As a user,
I want PlexCord to verify my Plex connection works,
So that I know the setup is correct before proceeding.

**Acceptance Criteria:**

**Given** the user has selected or entered a Plex server
**When** connection validation is triggered
**Then** PlexCord attempts to connect using the token and server URL
**And** successful connection shows server name and library count
**And** failed connection shows a specific error message (invalid token, server unreachable)
**And** validation completes within 5 seconds
**And** the user can retry validation after fixing issues
**And** only validated connections can proceed to the next step

---

### Story 2.7: Plex User Account Selection

As a user,
I want to select which Plex user account to monitor,
So that PlexCord tracks the correct user's playback on shared servers.

**Acceptance Criteria:**

**Given** the Plex connection is validated
**When** multiple user accounts exist on the server
**Then** a list of user accounts is displayed
**And** the user can select which account to monitor
**And** the selected account is saved to configuration
**And** if only one account exists, it is auto-selected
**And** the user can change the monitored account later in settings

---

### Story 2.8: Music Session Detection

As a user,
I want PlexCord to detect when I'm playing music on Plex,
So that my listening activity can be shared.

**Acceptance Criteria:**

**Given** PlexCord is connected to a Plex server
**When** the monitored user plays music
**Then** PlexCord detects the active music session within 2 seconds
**And** the session is identified as music (not video/photo)
**And** polling occurs at the configured interval (default 5 seconds)
**And** each poll completes within 500ms (NFR5)
**And** memory usage remains below 50MB during polling (NFR2)
**And** CPU usage averages below 1% (NFR3)

---

### Story 2.9: Track Metadata Extraction

As a user,
I want PlexCord to extract track information from my Plex session,
So that accurate details can be displayed.

**Acceptance Criteria:**

**Given** an active music session is detected
**When** metadata is extracted
**Then** the track title is captured
**And** the artist name is captured
**And** the album name is captured
**And** the album artwork URL is captured (if available)
**And** the track duration is captured
**And** the current playback position is captured
**And** missing metadata fields show appropriate fallback values

---

### Story 2.10: Playback State Detection

As a user,
I want PlexCord to detect play, pause, and stop states,
So that my Discord presence accurately reflects my activity.

**Acceptance Criteria:**

**Given** PlexCord is monitoring a Plex session
**When** the playback state changes
**Then** play state is detected within 2 seconds (NFR4)
**And** pause state is detected within 2 seconds
**And** stop/session end is detected within 2 seconds
**And** track changes are detected within 2 seconds
**And** the frontend receives state updates via Wails events

---

### Story 2.11: Setup Completion with Live Preview

As a user,
I want to see a preview of my Discord presence during setup,
So that I can verify everything is working correctly.

**Acceptance Criteria:**

**Given** the user has completed Plex configuration
**When** music is playing on Plex
**Then** a preview shows what the Discord status will look like
**And** the preview updates in real-time as track changes
**And** the preview shows track title, artist, and album
**And** the user can complete setup even if no music is currently playing
**And** the user can skip to complete setup later (FR16)

---

## Epic 3: Discord Rich Presence

User's Discord status automatically updates to show what they're listening to on Plex.

### Story 3.1: Discord RPC Connection

As a user,
I want PlexCord to connect to my local Discord client,
So that it can update my Discord status.

**Acceptance Criteria:**

**Given** Discord desktop client is running locally
**When** PlexCord attempts to connect
**Then** a connection is established via local IPC (Discord RPC protocol)
**And** the connection works without internet access (NFR23)
**And** connection status is reported to the frontend
**And** the connection uses the configured Discord Application Client ID

---

### Story 3.2: Discord Client Detection

As a user,
I want PlexCord to detect when Discord is not running,
So that I receive helpful feedback instead of silent failures.

**Acceptance Criteria:**

**Given** PlexCord attempts to connect to Discord
**When** the Discord client is not running
**Then** the error `DISCORD_NOT_RUNNING` is returned
**And** the user sees a clear message: "Discord is not running"
**And** the UI indicates Discord is disconnected
**And** PlexCord continues to function (Plex monitoring continues)

---

### Story 3.3: Discord Application Client ID Configuration

As a user,
I want to configure my own Discord Application Client ID,
So that I can customize my Rich Presence or use a personal application.

**Acceptance Criteria:**

**Given** the user is in the Discord setup step (or settings)
**When** configuring Discord integration
**Then** a default Client ID is pre-filled for PlexCord
**And** the user can optionally enter a custom Discord Application Client ID
**And** instructions explain how to create a Discord application (if desired)
**And** the Client ID is validated for correct format
**And** the configured Client ID is saved to settings

---

### Story 3.4: Discord Connection Validation

As a user,
I want PlexCord to verify my Discord connection works during setup,
So that I know Rich Presence will function correctly.

**Acceptance Criteria:**

**Given** the user has configured Discord settings
**When** connection validation is triggered
**Then** PlexCord attempts to connect to Discord using the Client ID
**And** successful connection displays "Connected to Discord"
**And** failed connection shows specific error (not running, connection failed)
**And** the user can retry validation after starting Discord
**And** setup can proceed even if Discord validation fails (with warning)

---

### Story 3.5: Rich Presence Update with Track Info

As a user,
I want my Discord status to show what I'm listening to on Plex,
So that my friends can see my music activity.

**Acceptance Criteria:**

**Given** PlexCord is connected to both Plex and Discord
**When** music is playing on Plex
**Then** Discord Rich Presence shows the track title
**And** Discord Rich Presence shows the artist name
**And** Discord Rich Presence shows the album name
**And** Discord Rich Presence shows album artwork (if available)
**And** Discord Rich Presence shows playback progress/elapsed time
**And** updates occur within 2 seconds of track changes (NFR4)

---

### Story 3.6: Presence State for Paused Playback

As a user,
I want my Discord status to reflect when music is paused,
So that my status accurately shows my activity.

**Acceptance Criteria:**

**Given** music was playing and Discord presence was active
**When** the user pauses playback on Plex
**Then** Discord Rich Presence updates to show "Paused" state
**And** the track information remains visible
**And** the elapsed time stops incrementing
**And** the update occurs within 2 seconds (NFR4)

---

### Story 3.7: Clear Presence on Playback Stop

As a user,
I want my Discord status cleared when I stop listening to music,
So that my presence doesn't show stale information.

**Acceptance Criteria:**

**Given** music was playing and Discord presence was active
**When** the user stops playback or the session ends
**Then** Discord Rich Presence is cleared completely
**And** Discord status returns to normal (no PlexCord presence)
**And** the clear occurs within 2 seconds (NFR4)
**And** no stale track information remains visible

---

### Story 3.8: Presence Recovery on Discord Restart

As a user,
I want PlexCord to reconnect to Discord if it restarts,
So that my presence continues working without manual intervention.

**Acceptance Criteria:**

**Given** PlexCord was connected to Discord and Discord is closed
**When** Discord is reopened
**Then** PlexCord detects Discord availability
**And** PlexCord automatically reconnects
**And** if music is currently playing, presence is restored
**And** reconnection occurs within one polling cycle

---

## Epic 4: Dashboard & System Tray

User can monitor their connection status, view currently playing track, and control PlexCord from the system tray.

### Story 4.1: Dashboard Main View

As a user,
I want a dashboard showing PlexCord's current state at a glance,
So that I can quickly see if everything is working.

**Acceptance Criteria:**

**Given** the user has completed setup
**When** the main application window is opened
**Then** the dashboard view is displayed
**And** the dashboard respects system dark/light mode (NFR26)
**And** UI interactions respond within 100ms (NFR6)
**And** the layout is clean and uncluttered

---

### Story 4.2: Connection Status Display

As a user,
I want to see the connection status for both Plex and Discord,
So that I know if PlexCord is working correctly.

**Acceptance Criteria:**

**Given** the dashboard is displayed
**When** viewing the connection status component
**Then** Plex connection status is shown (Connected/Disconnected)
**And** Discord connection status is shown (Connected/Disconnected)
**And** the connected server name is displayed for Plex
**And** status indicators use clear visual cues (color, icons)
**And** status updates in real-time without page refresh

---

### Story 4.3: Now Playing Display

As a user,
I want to see what track is currently playing,
So that I can verify PlexCord is detecting my music.

**Acceptance Criteria:**

**Given** music is playing on Plex
**When** viewing the dashboard
**Then** the current track title is displayed
**And** the artist name is displayed
**And** the album name is displayed
**And** album artwork is displayed (if available)
**And** playback state is indicated (playing/paused)
**And** the display updates in real-time as tracks change
**And** when no music is playing, a "Not playing" state is shown

---

### Story 4.4: Settings Navigation

As a user,
I want to access application settings from the main window,
So that I can adjust PlexCord's behavior.

**Acceptance Criteria:**

**Given** the user is on the dashboard
**When** the settings button/link is clicked
**Then** the settings view is displayed
**And** navigation back to dashboard is available
**And** the transition is smooth and responsive (under 100ms)

---

### Story 4.5: System Tray Integration

As a user,
I want PlexCord to appear in my system tray,
So that it can run in the background without cluttering my taskbar.

**Acceptance Criteria:**

**Given** PlexCord is running
**When** the application starts
**Then** a tray icon appears in the system tray
**And** the tray integrates with the native platform (Windows, macOS, Linux)
**And** the tray icon is visible and appropriately sized for each platform

---

### Story 4.6: Tray Icon Status Indicator

As a user,
I want the tray icon to indicate connection status,
So that I can see PlexCord's state without opening the window.

**Acceptance Criteria:**

**Given** PlexCord is running in the system tray
**When** connection status changes
**Then** the tray icon updates to reflect the status
**And** different states are visually distinguishable (connected, disconnected, error)
**And** the icon clearly indicates connection status (NFR27)

---

### Story 4.7: Minimize to System Tray

As a user,
I want to minimize PlexCord to the system tray,
So that it runs in the background without a visible window.

**Acceptance Criteria:**

**Given** the PlexCord window is open
**When** the user clicks the minimize or close button (based on settings)
**Then** the window is hidden
**And** the application continues running in the system tray
**And** Plex monitoring and Discord presence continue working
**And** the behavior follows the user's minimize-to-tray preference

---

### Story 4.8: Restore from System Tray

As a user,
I want to restore the PlexCord window from the system tray,
So that I can view the dashboard and adjust settings.

**Acceptance Criteria:**

**Given** PlexCord is minimized to the system tray
**When** the user clicks the tray icon
**Then** the main window is restored and brought to focus
**And** the window appears in its previous position
**And** the dashboard shows current status

---

### Story 4.9: Tray Context Menu

As a user,
I want quick actions available from the tray icon menu,
So that I can control PlexCord without opening the window.

**Acceptance Criteria:**

**Given** PlexCord is running in the system tray
**When** the user right-clicks the tray icon
**Then** a context menu is displayed
**And** the menu shows current status (connected/disconnected)
**And** the menu includes "Open PlexCord" option
**And** the menu includes "Settings" option
**And** the menu includes "Quit" option
**And** menu items respond immediately when clicked

---

### Story 4.10: Quit from Tray Menu

As a user,
I want to quit PlexCord from the tray menu,
So that I can fully close the application when needed.

**Acceptance Criteria:**

**Given** the tray context menu is open
**When** the user clicks "Quit"
**Then** Discord presence is cleared
**And** all connections are closed gracefully
**And** the application exits completely
**And** no background processes remain running

---

## Epic 5: Settings & Preferences

User can customize PlexCord behavior including polling interval, auto-start, and tray behavior.

### Story 5.1: Settings View Layout

As a user,
I want a well-organized settings page,
So that I can easily find and adjust application options.

**Acceptance Criteria:**

**Given** the user navigates to settings
**When** the settings view is displayed
**Then** settings are organized into logical sections
**And** each section has a clear heading
**And** the layout respects system dark/light mode
**And** changes are saved automatically or with a clear save action

---

### Story 5.2: Polling Interval Configuration

As a user,
I want to configure how often PlexCord checks for playback updates,
So that I can balance responsiveness with resource usage.

**Acceptance Criteria:**

**Given** the user is in the settings view
**When** adjusting the polling interval
**Then** a slider or input allows selecting intervals (1-30 seconds)
**And** the default value is 5 seconds
**And** the current value is clearly displayed
**And** changes take effect immediately
**And** the setting is persisted to config

---

### Story 5.3: Auto-Start on Login

As a user,
I want PlexCord to start automatically when I log in,
So that my Discord presence works without manual intervention.

**Acceptance Criteria:**

**Given** the user is in the settings view
**When** toggling the auto-start option
**Then** a toggle/checkbox enables or disables auto-start
**And** enabling adds PlexCord to system startup (Windows registry, macOS Login Items, Linux autostart)
**And** disabling removes PlexCord from system startup
**And** the current state is accurately reflected
**And** the setting works correctly on all three platforms

---

### Story 5.4: Minimize-to-Tray Behavior

As a user,
I want to configure how PlexCord behaves when minimized or closed,
So that I can control whether it runs in the background.

**Acceptance Criteria:**

**Given** the user is in the settings view
**When** configuring minimize-to-tray options
**Then** options include: "Minimize to tray on close", "Minimize to tray on minimize"
**And** each option can be toggled independently
**And** the current state is clearly shown
**And** changes take effect immediately
**And** settings are persisted to config

---

### Story 5.5: Modify Plex Connection Settings

As a user,
I want to change my Plex server connection after initial setup,
So that I can switch servers or update my token.

**Acceptance Criteria:**

**Given** the user is in the settings view
**When** accessing Plex connection settings
**Then** the current server name/URL is displayed
**And** the user can change the Plex token
**And** the user can change the server URL or re-run discovery
**And** the user can change the monitored user account
**And** connection validation is required before saving changes
**And** existing connection is maintained until new one is validated

---

### Story 5.6: Modify Discord Client ID

As a user,
I want to change the Discord Application Client ID,
So that I can switch to a custom Discord application.

**Acceptance Criteria:**

**Given** the user is in the settings view
**When** accessing Discord settings
**Then** the current Client ID is displayed (partially masked)
**And** the user can enter a new Client ID
**And** a "Reset to default" option restores the original PlexCord Client ID
**And** connection is re-established with the new Client ID
**And** the setting is persisted to config

---

### Story 5.7: Reset Application

As a user,
I want to reset PlexCord to its initial state,
So that I can start fresh or troubleshoot issues.

**Acceptance Criteria:**

**Given** the user is in the settings view
**When** selecting the reset option
**Then** a confirmation dialog warns about data loss
**And** reset clears all configuration settings
**And** reset removes the Plex token from secure storage
**And** reset removes auto-start registration
**And** after reset, the setup wizard is shown on next launch
**And** the application does not exit automatically (user decides)

---

## Epic 6: Error Recovery & Resilience

User can easily identify and recover from connection issues with clear error messages and automatic retry.

### Story 6.1: Error Banner Component

As a user,
I want errors displayed prominently but non-intrusively,
So that I'm aware of issues without losing context.

**Acceptance Criteria:**

**Given** an error occurs (Plex or Discord connection issue)
**When** the error is detected
**Then** an error banner appears at the top of the dashboard
**And** the banner shows the error message in plain language (NFR25)
**And** the banner includes the error code for troubleshooting
**And** the banner is dismissible
**And** multiple errors can be shown if needed

---

### Story 6.2: Actionable Error Messages

As a user,
I want error messages that tell me what went wrong and how to fix it,
So that I can resolve issues without guessing.

**Acceptance Criteria:**

**Given** an error occurs
**When** the error is displayed
**Then** the message explains the problem clearly
**And** the message suggests corrective action where applicable:
  - `PLEX_UNREACHABLE`: "Cannot reach Plex server. Check if server is running and network is connected."
  - `PLEX_AUTH_FAILED`: "Plex authentication failed. Your token may have expired."
  - `DISCORD_NOT_RUNNING`: "Discord is not running. Start Discord to enable Rich Presence."
  - `DISCORD_CONN_FAILED`: "Cannot connect to Discord. Try restarting Discord."
**And** no technical jargon is used in user-facing messages

---

### Story 6.3: Manual Retry Button

As a user,
I want to manually retry failed connections,
So that I can immediately test after fixing an issue.

**Acceptance Criteria:**

**Given** a connection error is displayed
**When** the user clicks the retry button
**Then** PlexCord attempts to reconnect immediately
**And** the retry button shows a loading state during attempt
**And** success clears the error and restores normal operation
**And** failure updates the error message with latest status

---

### Story 6.4: Automatic Retry with Exponential Backoff

As a user,
I want PlexCord to automatically retry failed connections,
So that temporary issues resolve without my intervention.

**Acceptance Criteria:**

**Given** a connection fails (Plex or Discord)
**When** automatic retry is triggered
**Then** retry attempts follow exponential backoff: 5s → 10s → 30s → 60s max (NFR18)
**And** the UI shows "Retrying in X seconds..."
**And** automatic retry continues indefinitely at 60s intervals
**And** successful reconnection resets the backoff timer
**And** manual retry resets the backoff timer
**And** automatic retry does not block UI responsiveness

---

### Story 6.5: Graceful Plex Unavailability Handling

As a user,
I want PlexCord to handle Plex server outages gracefully,
So that the application remains stable and recovers automatically.

**Acceptance Criteria:**

**Given** the Plex server becomes unavailable
**When** PlexCord detects the outage
**Then** Discord presence is cleared (not stale)
**And** the dashboard shows Plex disconnected status
**And** error banner explains the issue
**And** automatic retry begins with backoff
**And** when Plex becomes available, connection is restored automatically
**And** if music was playing, presence is restored

---

### Story 6.6: Graceful Discord Unavailability Handling

As a user,
I want PlexCord to handle Discord being closed or unavailable,
So that Plex monitoring continues and presence resumes when Discord returns.

**Acceptance Criteria:**

**Given** Discord is closed or becomes unavailable
**When** PlexCord detects the disconnect
**Then** Plex monitoring continues normally
**And** the dashboard shows Discord disconnected status
**And** error banner explains Discord is not running
**And** PlexCord periodically checks for Discord availability
**And** when Discord becomes available, connection is restored automatically
**And** if music is playing, presence is immediately updated

---

### Story 6.7: Token Expiration Detection and Re-authentication

As a user,
I want to be prompted when my Plex token expires,
So that I can re-authenticate and continue using PlexCord.

**Acceptance Criteria:**

**Given** PlexCord detects a `PLEX_AUTH_FAILED` error
**When** the error indicates token expiration
**Then** the error message explains the token has expired
**And** a "Re-authenticate" button is shown
**And** clicking the button opens the Plex token input dialog
**And** entering a new valid token restores connection
**And** the new token is securely stored

---

### Story 6.8: Connection History Display

As a user,
I want to see when connections were last successful,
So that I can understand the reliability of my setup.

**Acceptance Criteria:**

**Given** the user is viewing connection status
**When** connections have been established
**Then** "Last connected" timestamp is shown for Plex
**And** "Last connected" timestamp is shown for Discord
**And** timestamps update when connections are restored
**And** timestamps persist across application restarts
**And** timestamps show relative time (e.g., "5 minutes ago")

---

### Story 6.9: Long-Running Stability

As a user,
I want PlexCord to run reliably for extended periods,
So that I don't need to restart it regularly.

**Acceptance Criteria:**

**Given** PlexCord is running continuously
**When** 30+ days have passed (NFR13)
**Then** the application continues operating normally
**And** memory usage remains stable (no memory leaks)
**And** CPU usage remains low during idle
**And** all connections recover from transient failures
**And** no manual intervention is required for normal operation

---

## Epic 7: Updates & Maintenance

User can check for updates, view current version, and access release notes.

### Story 7.1: Version Display

As a user,
I want to see the current PlexCord version,
So that I know which version I'm running for troubleshooting or support.

**Acceptance Criteria:**

**Given** the user is in settings or about section
**When** viewing application information
**Then** the current version number is displayed (e.g., "v1.0.0")
**And** the version follows semantic versioning format
**And** the version is accessible from both settings and about dialog
**And** the version matches the built binary version

---

### Story 7.2: Manual Update Check

As a user,
I want to check if a newer version of PlexCord is available,
So that I can stay up to date with improvements and fixes.

**Acceptance Criteria:**

**Given** the user is in settings
**When** clicking "Check for Updates"
**Then** PlexCord checks for newer versions (via GitHub releases or similar)
**And** a loading indicator is shown during the check
**And** if an update is available, the new version number is displayed
**And** if no update is available, "You're up to date" is shown
**And** if the check fails (no network), a clear error message is shown

---

### Story 7.3: Update Notification

As a user,
I want to be notified when a new version is available,
So that I can update when convenient.

**Acceptance Criteria:**

**Given** a new version is detected
**When** the update check completes
**Then** the update information includes the new version number
**And** a "Download" or "View Release" button is available
**And** clicking the button opens the download page in the default browser
**And** the notification is non-intrusive (doesn't block usage)

---

### Story 7.4: Changelog Access

As a user,
I want to view the release notes and changelog,
So that I can see what's new or changed in each version.

**Acceptance Criteria:**

**Given** the user wants to see what changed
**When** accessing the changelog
**Then** a link to release notes is available in settings/about
**And** clicking opens the GitHub releases page in the default browser
**And** the current version's changes are highlighted (if available)

