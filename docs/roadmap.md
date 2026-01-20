# Project Roadmap

This document outlines the development roadmap for PlexCord, organized by releases.

## Current Status

**Current Version:** v0.1.0 (In Development)

**Development Phase:** Epic 1-2 (Foundation & Setup Wizard)

## Release Plan

### v1.0.0 - MVP Release (Target: Q1 2026)

**Goal:** Deliver core functionality for music playback detection and Discord presence.

**Epic 1: Application Foundation** ✅ (Complete)

- Project initialization with Wails
- Go backend package structure
- Configuration file management
- Error code system
- Cross-platform build verification

**Epic 2: Setup Wizard & Plex Connection** 🚧 (In Progress)

- Setup wizard navigation framework
- Plex token input with instructions
- Secure token storage
- Plex server auto-discovery
- Manual server entry
- Connection validation
- Music session detection
- Track metadata extraction
- Playback state detection

**Epic 3: Discord Rich Presence**

- Discord RPC connection
- Client detection
- Application Client ID configuration
- Rich Presence updates
- Pause/stop state handling
- Connection recovery

**Epic 4: Dashboard & System Tray**

- Dashboard main view
- Connection status display
- Now playing display
- System tray integration
- Minimize to tray
- Tray context menu

**Epic 5: Settings & Preferences**

- Settings view layout
- Polling interval configuration
- Auto-start on login
- Modify connection settings
- Reset application

**Epic 6: Error Recovery & Resilience**

- Error banner component
- Actionable error messages
- Manual retry functionality
- Automatic retry with backoff
- Graceful unavailability handling
- Long-running stability

**Epic 7: Updates & Maintenance**

- Version display
- Manual update check
- Update notification
- Changelog access

---

### v1.1.0 - Webhook Support (Target: Q2 2026)

**Goal:** Replace polling with real-time webhooks for instant updates.

**Features:**

- Plex webhook endpoint configuration
- Webhook server in PlexCord
- Real-time playback updates (<1s latency)
- Fallback to polling if webhooks unavailable
- Webhook authentication/security
- Reduced API calls and network traffic

**Benefits:**

- Instant Discord updates (vs 5-second polling delay)
- Lower CPU usage (event-driven vs continuous polling)
- Reduced Plex server load

---

### v1.2.0 - Headless Mode (Target: Q3 2026)

**Goal:** Enable server/Docker deployment for always-on operation.

**Features:**

- JSON configuration file support
- CLI-only operation (no GUI required)
- Docker container image
- systemd service file (Linux)
- launchd service file (macOS)
- Windows service support
- Health check endpoint
- Metrics/status API

**Use Cases:**

- Run on home server alongside Plex
- Deploy in Docker Compose with Plex
- Set-and-forget background operation
- Multiple user monitoring (future)

---

### v1.3.0 - Multi-Account Support (Target: Q4 2026)

**Goal:** Monitor multiple Plex accounts simultaneously.

**Features:**

- Multiple account configuration
- Per-account Discord presence
- Account switching UI
- Separate credentials per account
- Activity aggregation dashboard
- Account priority settings

**Use Cases:**

- Household with multiple Plex users
- Shared Plex servers
- Family Discord presence

---

### v2.0.0 - Advanced Features (Target: 2027)

**Goal:** Enhanced user experience and customization.

**Features:**

- **Custom Presence Templates**
  - User-defined presence formats
  - Dynamic field insertion
  - Multiple template presets

- **Playback History**
  - Local SQLite database
  - Listening statistics
  - Most played tracks/artists
  - Weekly/monthly summaries

- **Smart Presence**
  - Show/hide based on privacy rules
  - Time-based presence (e.g., work hours)
  - Playlist-specific presence
  - Genre-based custom messages

- **Plugin System**
  - Third-party integrations
  - Custom presence providers
  - Event hooks for automation

- **Mobile App (Companion)**
  - iOS/Android companion app
  - Remote control
  - Push notifications
  - Mobile presence control

---

## Feature Requests

Features under consideration (not yet scheduled):

- **Last.fm Scrobbling** - Automatic scrobbling to Last.fm
- **Statistics Dashboard** - Detailed listening analytics
- **Friend Activity** - See what friends are listening to
- **Playlist Integration** - Show current playlist in presence
- **Lyrics Display** - Show current lyrics in app
- **Audio Visualization** - Real-time audio waveform
- **Cross-Platform Sync** - Sync settings across devices
- **Browser Extension** - Control from browser toolbar

Vote for features on [GitHub Discussions](https://github.com/yourusername/PlexCord/discussions).

---

## Development Milestones

### Q1 2026

- ✅ Project initialization
- ✅ Backend package structure
- 🚧 Setup wizard implementation
- ⏳ Plex integration
- ⏳ Discord integration
- ⏳ v1.0.0 Beta release

### Q2 2026

- ⏳ v1.0.0 Stable release
- ⏳ Community feedback integration
- ⏳ Bug fixes and stability improvements
- ⏳ Webhook implementation
- ⏳ v1.1.0 Beta release

### Q3 2026

- ⏳ v1.1.0 Stable release
- ⏳ Headless mode development
- ⏳ Docker container
- ⏳ v1.2.0 Beta release

### Q4 2026

- ⏳ v1.2.0 Stable release
- ⏳ Multi-account support
- ⏳ v1.3.0 planning

---

## Contributing to Roadmap

Have ideas for PlexCord's future? We'd love to hear them!

- **Feature Requests:** Open an issue with the `enhancement` label
- **Discussions:** Join [GitHub Discussions](https://github.com/yourusername/PlexCord/discussions)
- **Vote:** React to issues to show interest in features

---

## Version History

| Version | Release Date | Highlights |
|---------|--------------|------------|
| v0.1.0 | In Development | Foundation, Setup Wizard |

---

**Legend:**

- ✅ Complete
- 🚧 In Progress
- ⏳ Planned
- ❓ Under Consideration
