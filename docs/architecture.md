# Architecture Overview

PlexCord is built with a clean separation between backend (Go) and frontend (Vue.js), connected through the Wails framework.

## High-Level Architecture

```
┌─────────────────────────────────────────────────┐
│                  Frontend (Vue.js)              │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐     │
│  │Dashboard │  │ Settings │  │  Setup   │     │
│  │   View   │  │   View   │  │  Wizard  │     │
│  └──────────┘  └──────────┘  └──────────┘     │
│         │              │              │         │
│  ┌─────────────────────────────────────┐       │
│  │       Pinia Stores (State)          │       │
│  └─────────────────────────────────────┘       │
└─────────────────┬───────────────────────────────┘
                  │ Wails Bridge
┌─────────────────┴───────────────────────────────┐
│              Backend (Go)                       │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐     │
│  │   Plex   │  │ Discord  │  │  Config  │     │
│  │  Client  │  │   RPC    │  │ Manager  │     │
│  └──────────┘  └──────────┘  └──────────┘     │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐     │
│  │ Keychain │  │ Platform │  │  Error   │     │
│  │ Manager  │  │ Services │  │ Handler  │     │
│  └──────────┘  └──────────┘  └──────────┘     │
└─────────────────────────────────────────────────┘
         │                    │
    ┌────┴────┐         ┌────┴────┐
    │  Plex   │         │ Discord │
    │  API    │         │ Client  │
    └─────────┘         └─────────┘
```

## Technology Stack

### Frontend
- **Vue.js 3** (Composition API) - Reactive UI framework
- **PrimeVue** - Pre-built UI components
- **TailwindCSS** - Utility-first styling
- **Pinia** - State management
- **Vite** - Build tool and dev server

### Backend
- **Go 1.21+** - Backend runtime
- **Wails v2** - Desktop app framework
- **go-keyring** - Secure credential storage
- **go-autostart** - Auto-start on login
- **hashicorp/mdns** - Plex server discovery
- **rich-go** - Discord RPC protocol

### Build & Distribution
- **Wails CLI** - Cross-platform compilation
- **Single binary** - No runtime dependencies
- **Platform-specific builds** - Windows, macOS (Universal), Linux

## Backend Package Structure

```
internal/
├── plex/          # Plex Media Server integration
│   ├── client.go      # HTTP client, session polling
│   ├── discovery.go   # mDNS/GDM server discovery
│   └── types.go       # Session, track metadata structs
│
├── discord/       # Discord Rich Presence
│   ├── presence.go    # RPC connection, presence updates
│   └── types.go       # Presence payload structures
│
├── config/        # Configuration management
│   ├── config.go      # Settings load/save, JSON persistence
│   └── paths.go       # Platform-specific config paths
│
├── keychain/      # Secure credential storage
│   └── keychain.go    # Wrapper for go-keyring
│
├── platform/      # OS-specific features
│   ├── platform.go    # Auto-start, system tray abstractions
│   └── tray.go        # System tray implementation
│
└── errors/        # Error handling
    ├── codes.go       # Error code constants
    └── errors.go      # Structured error types
```

## Frontend Structure

```
frontend/src/
├── views/             # Page components
│   ├── Dashboard.vue      # Main application view
│   ├── Settings.vue       # Configuration interface
│   └── setup/             # Setup wizard pages
│       ├── SetupWizard.vue
│       ├── SetupPlex.vue
│       ├── SetupDiscord.vue
│       └── SetupComplete.vue
│
├── components/        # Reusable components
│   ├── NowPlaying.vue     # Current track display
│   ├── ConnectionStatus.vue # Plex/Discord status
│   ├── ServerCard.vue     # Plex server selection
│   └── ErrorBanner.vue    # Error messages
│
├── stores/            # Pinia state management
│   ├── connection.ts      # Plex/Discord connection state
│   ├── playback.ts        # Current track, playback state
│   ├── settings.ts        # User preferences
│   └── setup.ts           # Setup wizard state
│
├── router/            # Vue Router configuration
│   └── index.js
│
└── service/           # Backend API calls
    └── wails.js           # Wails bridge functions
```

## Key Design Patterns

### Backend Patterns

**Repository Pattern** - Plex and Discord clients encapsulate API interactions
```go
type PlexClient interface {
    GetSessions() ([]Session, error)
    DiscoverServers() ([]Server, error)
}
```

**Error Codes** - Structured errors with unique codes for frontend handling
```go
const (
    ErrPlexUnreachable  = "PLEX_UNREACHABLE"
    ErrPlexAuthFailed   = "PLEX_AUTH_FAILED"
    ErrDiscordNotRunning = "DISCORD_NOT_RUNNING"
)
```

**Platform Abstraction** - OS-specific features behind common interfaces
```go
type Platform interface {
    ConfigPath() string
    SecureStore() KeychainManager
    AutoStart() AutoStartManager
}
```

### Frontend Patterns

**Component Composition** - Small, focused components composed into views
- `NowPlaying.vue` displays track info
- `ConnectionStatus.vue` shows Plex/Discord status
- `Dashboard.vue` composes these together

**Store-Driven UI** - Components react to store state changes
```javascript
const playbackStore = usePlaybackStore()
watch(() => playbackStore.currentTrack, (track) => {
  // Update UI when track changes
})
```

**Error-First Design** - Every backend call includes error handling
```javascript
try {
  await ConnectToPlex(token)
} catch (error) {
  errorStore.showError(error.code, error.message)
}
```

## Data Flow

### Playback Detection Flow

1. **Polling Timer** (backend) triggers every 5 seconds (configurable)
2. **Plex Client** fetches `/status/sessions` from Plex API
3. **Session Parser** extracts music sessions and track metadata
4. **Event Emitter** sends update to frontend via Wails runtime
5. **Playback Store** updates current track state
6. **UI Components** reactively update to show new track
7. **Discord Manager** (backend) updates Rich Presence

### Error Handling Flow

1. **Backend** detects error (e.g., Plex unreachable)
2. **Error Handler** creates structured error with code
3. **Wails Bridge** emits error event to frontend
4. **Error Store** receives error and stores it
5. **Error Banner** component displays message with retry action
6. **Retry Logic** uses exponential backoff (5s → 10s → 30s → 60s)

## Security Considerations

### Credential Storage
- **Plex tokens** stored in OS keychain (macOS Keychain, Windows Credential Manager, Linux Secret Service)
- **Fallback encryption** when secure storage unavailable
- **No tokens in logs** - all sensitive data redacted

### API Communication
- **HTTPS only** for Plex API (local servers may use HTTP if explicitly configured)
- **Local IPC** for Discord (no network communication)
- **No telemetry** - application never phones home

### File Permissions
- Config files: `0600` (owner read/write only)
- Log files: `0600` (if logging enabled)

## Performance Characteristics

### Resource Usage
- **Startup**: <3 seconds
- **Memory (idle)**: <50MB
- **CPU (polling)**: <1% average
- **Binary size**: <20MB per platform

### Latency
- **Discord update**: <2 seconds from playback change
- **Plex polling**: 500ms per request
- **UI response**: <100ms for interactions

### Reliability
- **Crash-free rate**: 99.9% target
- **Uptime**: 30+ days without restart
- **Auto-reconnect**: Exponential backoff after failures

## Build & Distribution

### Build Process
1. Frontend compiled with Vite to static assets
2. Assets embedded in Go binary
3. Wails CLI compiles to native executable
4. Platform-specific packaging (MSI, DMG, AppImage)

### Cross-Platform Compilation
```bash
# Windows (from any OS)
wails build -platform windows/amd64

# macOS Universal (from macOS only)
wails build -platform darwin/universal

# Linux (from any OS)
wails build -platform linux/amd64
```

### Distribution
- **GitHub Releases** - Primary distribution channel
- **Single file** - No installers or dependencies
- **Auto-update** - Planned for v1.1 (not MVP)

## Future Architecture Considerations

### Webhook Support (v1.1)
- Replace polling with Plex webhooks for instant updates
- Fallback to polling if webhooks unavailable
- Requires webhook endpoint configuration

### Headless Mode (v1.2)
- JSON configuration file instead of UI
- Docker container support
- systemd/launchd service integration

### Multi-Account Support (v1.3)
- Monitor multiple Plex accounts
- Per-account Discord presence
- Account switching UI
