---
stepsCompleted: [step-01-init, step-02-discovery, step-03-success, step-04-journeys, step-05-domain, step-06-innovation, step-07-project-type, step-08-scoping, step-09-functional, step-10-nonfunctional, step-11-polish]
inputDocuments:
  - _bmad-output/planning-artifacts/product-brief-PlexCord-2026-01-16.md
workflowType: 'prd'
documentCounts:
  briefs: 1
  research: 0
  brainstorming: 0
  projectDocs: 0
classification:
  projectType: desktop_app
  domain: general
  complexity: low
  projectContext: greenfield
---

# Product Requirements Document - PlexCord

**Author:** Batti
**Date:** 2026-01-16

## Executive Summary

PlexCord is a cross-platform desktop application that bridges Plex Media Server and Discord Rich Presence. Built with Go and Wails, featuring a Vue.js interface powered by PrimeVue and TailwindCSS, it delivers the seamless music-sharing experience that Plex users have been missing.

### Problem

Plex users who have invested in curating their own music libraries are invisible on Discord. While Spotify users enjoy native Discord integration, Plex/Plexamp users have no out-of-the-box solution for sharing music activity with friends.

### Solution

PlexCord delivers Discord Rich Presence for Plexamp users with emphasis on universal accessibility, effortless setup, and premium user experience:

- **Cross-platform**: Windows, Mac, Linux single-binary distribution
- **Zero dependencies**: Single executable, no runtime required
- **Auto-discovery**: Detects Plex servers on local network automatically
- **Modern UI**: Vue.js + PrimeVue + TailwindCSS, dark mode, system tray integration

### Key Differentiators vs. PlexampRPC

| Capability | PlexampRPC | PlexCord |
|------------|------------|----------|
| Platform | Windows-only | Windows, Mac, Linux |
| Runtime | .NET required | None (single binary) |
| Setup | Manual token entry | Auto-discovery + guided wizard |
| UI/UX | Basic .NET GUI | Modern Vue.js + Tailwind |

### Target Users

1. **Self-Hosting Enthusiast** - Technical users on Linux/Mac wanting cross-platform solution
2. **Music Curator** - Non-technical users wanting one-click setup experience
3. **Headless Operator** - Power users preferring Docker/JSON configuration (v1.2)

## Success Criteria

### User Success

| Metric | Target | Measurement Method |
|--------|--------|-------------------|
| Setup Completion Rate | >95% | Wizard completion tracking |
| Time to First Presence | <2 minutes | Download → Discord status |
| Crash-Free Sessions | >99.9% | Error reporting / GitHub issues |
| Zero-Touch Operation | No daily interaction | Absence of usage questions |

**Success Indicators:**
- "How did you do that?" moments from Discord friends
- Set-and-forget operation after initial setup
- Mac/Linux users finally have a working solution

### Business Success

As an open-source project, success focuses on adoption and community growth:

| Metric | 3-Month Target | 12-Month Target |
|--------|----------------|-----------------|
| GitHub Stars | 100+ | 500+ |
| Total Downloads | 500+ | 5,000+ |
| Active Installations | 200+ | 1,000+ |
| External Contributors | 2+ | 10+ |

**Strategic Goal:** Become the default recommendation for Plex-Discord integration, surpassing PlexampRPC.

### Technical Success

| Metric | Target |
|--------|--------|
| Binary Size | <20MB |
| Memory Usage (Idle) | <50MB |
| CPU Usage (Average) | <1% |
| Startup Time | <3 seconds |
| Discord Update Latency | <2 seconds (webhooks) |

### Measurable Outcomes

**North Star Metric:** Active Weekly Users - users with at least one Discord presence update per week.

**Supporting KPIs:**
1. Onboarding Success Rate (>95%)
2. Platform Coverage (no platform <20% of users)
3. Stability Score (<1 crash per 1,000 sessions)
4. Community Sentiment (positive GitHub/Reddit mentions)
5. Webhook Adoption Rate (% using real-time vs. polling - post v1.1)

## User Journeys

### Journey 1: Marcus - The Self-Hosting Enthusiast

**Persona:** Marcus, 32, software developer who runs a home lab with Proxmox, manages his own Plex server on Ubuntu, and is active in r/selfhosted. He's frustrated that PlexampRPC only works on Windows while his daily driver is Pop!_OS.

**Opening Scene:**
Marcus is listening to his meticulously organized music library through Plexamp on his Linux desktop. His Discord friends are chatting about music, sharing what they're listening to via Spotify integration. Marcus's status shows nothing—he's invisible despite actively listening. He's tried PlexampRPC but it requires Windows and .NET. "There has to be a better way," he mutters.

**Rising Action:**
Marcus discovers PlexCord on GitHub while browsing r/Plex. He downloads the Linux binary—a single 15MB file. No dependencies, no runtime installations. He runs the executable and a modern, dark-themed UI appears. The setup wizard immediately detects his local Plex server via mDNS. He clicks it, enters his Plex token (the wizard shows him exactly where to find it), and connects.

**Climax:**
Within 90 seconds of downloading, Marcus sees his Discord status update: "Listening to 'Lateralus' by Tool on Plexamp." His friend Jake immediately messages: "Wait, how did you get Plex working with Discord? I thought that was Windows-only!"

**Resolution:**
Marcus minimizes PlexCord to his system tray and forgets about it. It auto-starts with his system, silently syncing his music activity. He recommends it on r/selfhosted, and contributes a Docker compose file to the project. His music library finally gets the visibility it deserves.

**Requirements Revealed:**
- Linux binary distribution
- mDNS/GDM server auto-discovery
- System tray integration with auto-start
- Dark theme UI
- Guided token entry with instructions

---

### Journey 2: Sofia - The Music Curator

**Persona:** Sofia, 28, graphic designer on macOS who chose Plex because she wanted control over her music library and hated Spotify's algorithm. She's not technical—she just wants things to work beautifully.

**Opening Scene:**
Sofia is working on a design project, listening to her carefully curated playlist on Plexamp. Her design team's Discord server has a music channel where everyone shares what they're listening to. Sofia feels left out—her Spotify-using colleagues have automatic presence, but she has to manually type what she's playing. It's tedious and she often doesn't bother.

**Rising Action:**
A colleague shares a link to PlexCord. Sofia downloads the Mac app and drags it to Applications. On first launch, a beautiful setup wizard greets her. It automatically finds her Plex server (running on her NAS). She clicks "Connect with Plex" and a browser window opens for authentication. No tokens to copy, no config files to edit.

**Climax:**
Sofia returns to the app and sees a live preview: her current track displayed exactly as it will appear on Discord. She clicks "Finish Setup" and checks Discord—her status shows the album art and track info. It's beautiful. She takes a screenshot and posts it to her team's channel.

**Resolution:**
PlexCord lives in Sofia's menu bar, showing a tiny music note icon. She never thinks about it again. When she plays music, Discord shows it. When she stops, it clears. Her colleagues start asking how to set up Plex themselves.

**Requirements Revealed:**
- macOS native app experience (menu bar)
- OAuth authentication flow (no manual tokens)
- Live presence preview in UI
- Elegant, polished visual design
- Zero-configuration ideal

---

### Journey 3: Derek - The Headless Operator (v1.2)

**Persona:** Derek, 41, sysadmin who runs Plex on an Unraid server. He wants PlexCord running 24/7 on his server, not his desktop. He prefers Docker containers and JSON configuration over GUIs.

**Opening Scene:**
Derek's Plex server runs headless in his basement. He listens to music from various devices—phone, laptop, work computer. He wants his Discord presence to reflect his listening regardless of which device he uses, which means PlexCord needs to run on the server, not any particular client.

**Rising Action:**
Derek pulls the PlexCord Docker image. He creates a `config.json` with his Plex token, Discord client ID, and his Plex username to track. He adds the container to his docker-compose stack and starts it.

**Climax:**
Derek opens Discord on his phone while listening to music through Plexamp on his work laptop. His Discord status updates within seconds. He didn't install anything on either device—the server handles everything.

**Resolution:**
Derek adds PlexCord to his monitoring stack. It runs for months without intervention, using minimal resources. He contributes documentation for Unraid users.

**Requirements Revealed:**
- Headless/daemon mode
- Docker support with official image
- JSON configuration file
- Server-side session monitoring
- Multi-device user tracking

---

### Journey 4: Error Recovery - Connection Lost

**Persona:** Any user experiencing a disconnection scenario.

**Opening Scene:**
User has been running PlexCord successfully for weeks. One day, they notice their Discord status isn't updating. They check PlexCord and see a warning indicator.

**Rising Action:**
The user opens PlexCord from the system tray. The UI clearly shows: "Plex server unreachable" with a yellow warning icon. The status panel shows the last successful connection time and a "Retry" button.

**Climax:**
The user realizes their Plex server rebooted for updates. They click "Retry"—connection restored. Alternatively, if their Plex token expired, the app shows "Re-authenticate with Plex" button that launches the auth flow.

**Resolution:**
Normal operation resumes. The user appreciates clear error messaging instead of silent failures. They didn't need to check logs or restart the app manually.

**Requirements Revealed:**
- Clear connection status indicators
- Graceful error handling with user-friendly messages
- Manual retry capability
- Re-authentication flow for expired tokens
- No silent failures

## Project Scoping & Phased Development

### MVP Strategy & Philosophy

**MVP Approach:** Problem-Solving MVP

PlexCord's MVP focuses on delivering the core value proposition—Discord Rich Presence for Plex users—with enough polish to be credible and usable. The goal is validated learning: proving that cross-platform Plex-Discord integration fills a real gap.

**Resource Requirements:**
- Solo developer or small team (1-2)
- Go/Wails experience preferred

### MVP Feature Set (Phase 1 - v1.0)

**Core User Journeys Supported:**
- Marcus (Self-Hosting Enthusiast) - Linux setup, guided auth
- Sofia (Music Curator) - Mac setup, polished experience
- Error Recovery - Connection handling for all users

**Must-Have Capabilities:**

| Feature | Rationale |
|---------|-----------|
| Discord Rich Presence | Core value - the entire point of the product |
| Plex Session Polling | Required to detect playback |
| Setup Wizard | Critical for <2min onboarding target |
| Server Auto-Discovery | Removes friction, key differentiator |
| System Tray | Enables set-and-forget operation |
| Cross-Platform Binaries | Key differentiator vs. PlexampRPC |
| Connection Status UI | Essential for error recovery journey |
| Modern UI | Wails + Vue.js + PrimeVue + TailwindCSS |

**Authentication:** Guided token/Client ID entry with validation (full OAuth deferred to v1.1)

**MVP Success Gate:** 50+ stars, 100+ downloads, positive community sentiment

### Post-MVP Features

**Phase 2 - v1.1 (Polish Release):**
- Webhook support for instant (<2s) updates
- Album art in Discord presence
- Progress bar display
- Auto-update mechanism
- Full OAuth flows (eliminates manual token entry)

**Phase 3 - v1.2 (Power User Release):**
- Headless/daemon mode (Derek's journey)
- Docker support with official image
- Custom presence templates
- JSON configuration for automation

**Phase 4 - v2.0 (Platform Release):**
- Multiple Plex server support
- Last.fm scrobbling integration
- Listening statistics dashboard
- Discord bot companion

### Risk Mitigation Strategy

**Technical Risks:**

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Plex API changes | Low | High | Abstract API layer, version detection, community monitoring |
| Discord RPC deprecation | Very Low | Critical | Follow official spec, monitor Discord developer announcements |
| Wails framework issues | Low | Medium | Active community, fallback to Electron if needed |
| mDNS discovery failures | Medium | Low | Manual server entry fallback always available |

**Market Risks:**

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Low adoption | Medium | High | Target r/Plex, r/selfhosted; demonstrate clear value vs. PlexampRPC |
| Competition | Low | Medium | First-mover advantage on cross-platform; execution excellence |

**Resource Risks:**

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Solo developer burnout | Medium | High | Keep MVP minimal; leverage community contributions |
| Scope creep | Medium | Medium | Strict phase boundaries; say "v1.1" liberally |

**Minimum Viable Team:** 1 developer with Go experience can ship MVP. Community contributions expected for platform-specific testing and Docker/packaging support.

## Desktop Application Requirements

### Platform Support

| Platform | Format | Min Version | Notes |
|----------|--------|-------------|-------|
| Windows | Portable EXE / MSI installer | Windows 10+ | Single binary, no runtime required |
| macOS | DMG with app bundle | macOS 11+ (Big Sur) | Code-signed, menu bar integration |
| Linux | AppImage + .deb + .rpm | Ubuntu 20.04+ equiv | Single binary, desktop integration |

**Build Strategy:**
- Wails framework compiles Go backend + Vue.js frontend into single native binary
- Platform-specific builds via GitHub Actions CI/CD
- Target binary size: <20MB per platform

### System Integration

**System Tray / Menu Bar:**
- Minimize to tray on close (configurable)
- Tray icon shows connection status (green/yellow/red)
- Tray tooltip displays current track info
- Right-click menu: Show/Hide, Settings, Quit
- Left-click: Toggle main window

**Auto-Start:**
- Option to launch on system startup (default: enabled after setup)
- Platform-specific implementation:
  - Windows: Registry entry or Startup folder
  - macOS: Login Items via LaunchAgent
  - Linux: XDG autostart desktop entry

**Notifications:**
- Connection status changes (optional)
- Error conditions requiring user action
- Respect system Do Not Disturb / Focus modes

### Update Strategy

**MVP (v1.0):** Manual updates
- Check for updates button in Settings
- Link to GitHub releases page
- Display current vs. latest version

**v1.1:** Auto-update mechanism
- Background update check on startup (configurable)
- Download and prompt to install
- Platform-appropriate update flow:
  - Windows: Download installer, prompt restart
  - macOS: Sparkle framework or similar
  - Linux: Notify user, link to package manager / AppImage

### Offline & Network Capabilities

**Network Requirements:**
- Local network access for Plex server discovery (mDNS/GDM)
- HTTPS to Plex server for session polling
- Local IPC to Discord client (no internet required)

**Offline Behavior:**
- Plex server unreachable: Clear presence, show status, retry periodically
- Discord not running: Queue updates, apply when available
- No internet: Functions normally if Plex server is local

**Reconnection Logic:**
- Exponential backoff for retries (5s → 10s → 30s → 60s max)
- Immediate retry on network change detection
- User-triggered manual retry always available

### Data Storage

**Configuration Location:**
- Windows: `%APPDATA%\PlexCord\`
- macOS: `~/Library/Application Support/PlexCord/`
- Linux: `~/.config/plexcord/`

**Stored Data:**
- `config.json`: Server URL, polling interval, preferences
- Plex token: OS keychain when available, encrypted file fallback
- Discord Client ID: Plain text (not sensitive)
- Window state: Position, size, minimized state

**Security Considerations:**
- Plex tokens stored securely (OS keychain preferred)
- No analytics or telemetry in MVP
- All communication over HTTPS/TLS
- No credentials logged (even in debug mode)

### Permissions Required

| Permission | Platform | Reason |
|------------|----------|--------|
| Network access | All | Plex API communication |
| mDNS/Bonjour | All | Server auto-discovery |
| System tray | All | Background operation |
| Autostart | All | Launch on login (optional) |
| Keychain access | macOS | Secure token storage |

**No elevated/admin permissions required for normal operation.**

## Functional Requirements

### Plex Integration

- FR1: User can connect to a Plex Media Server using authentication token
- FR2: User can discover Plex servers on the local network automatically
- FR3: User can manually enter a Plex server URL if auto-discovery fails
- FR4: System can detect current music playback session from Plex server
- FR5: System can extract track metadata (title, artist, album) from Plex session
- FR6: System can detect playback state changes (play, pause, stop)
- FR7: User can select which Plex user account to monitor (if multiple exist)

### Discord Integration

- FR8: System can establish connection to local Discord client via RPC
- FR9: System can update Discord Rich Presence with current track information
- FR10: System can clear Discord Rich Presence when playback stops
- FR11: System can detect when Discord client is not running
- FR12: User can configure Discord Application Client ID

### Setup & Onboarding

- FR13: User can complete initial setup through a guided wizard
- FR14: User can view instructions for obtaining Plex authentication token
- FR15: User can see live preview of Discord presence during setup
- FR16: User can skip optional setup steps and configure later
- FR17: System can validate Plex connection before completing setup
- FR18: System can validate Discord connection before completing setup

### User Interface

- FR19: User can view current connection status (Plex and Discord)
- FR20: User can view currently playing track information
- FR21: User can access application settings from main window
- FR22: User can minimize application to system tray
- FR23: User can restore application from system tray
- FR24: User can view tray icon indicating connection status
- FR25: User can access quick actions from tray context menu
- FR26: User can quit application from tray menu

### Configuration & Settings

- FR27: User can configure polling interval for Plex session checks
- FR28: User can enable/disable auto-start on system login
- FR29: User can configure minimize-to-tray behavior
- FR30: User can modify Plex server connection settings
- FR31: User can modify Discord client ID
- FR32: User can reset application to initial state

### Error Handling & Recovery

- FR33: User can view clear error messages when connections fail
- FR34: User can manually retry failed connections
- FR35: System can automatically retry connections with backoff
- FR36: User can re-authenticate with Plex when token expires
- FR37: System can detect and report specific error conditions
- FR38: User can view connection history/last successful connection time

### Cross-Platform Support

- FR39: User can install and run application on Windows 10+
- FR40: User can install and run application on macOS 11+
- FR41: User can install and run application on Linux (Ubuntu 20.04+ equivalent)
- FR42: System can integrate with platform-native system tray
- FR43: System can store credentials securely using platform keychain (where available)

### Updates & Maintenance

- FR44: User can check for available updates manually
- FR45: User can view current application version
- FR46: User can access release notes/changelog

## Non-Functional Requirements

### Performance

- NFR1: Application startup time shall be less than 3 seconds on supported platforms
- NFR2: Memory usage shall remain below 50MB during idle operation
- NFR3: CPU usage shall average less than 1% during normal polling operation
- NFR4: Discord presence updates shall occur within 2 seconds of Plex playback state change
- NFR5: Plex session polling shall complete within 500ms per request
- NFR6: UI interactions shall respond within 100ms

### Security

- NFR7: Plex authentication tokens shall be stored using OS-native secure storage (Keychain/Credential Manager) where available
- NFR8: Plex authentication tokens shall be encrypted at rest when secure storage is unavailable
- NFR9: All Plex API communication shall use HTTPS/TLS encryption
- NFR10: No credentials shall be written to log files, including in debug mode
- NFR11: Application shall not collect or transmit usage analytics or telemetry
- NFR12: Configuration files shall have appropriate file permissions (user-only read/write)

### Reliability

- NFR13: Application shall maintain operation for 30+ days without requiring restart
- NFR14: Application shall achieve 99.9% crash-free session rate
- NFR15: Application shall gracefully handle Plex server unavailability without crashing
- NFR16: Application shall gracefully handle Discord client unavailability without crashing
- NFR17: Application shall automatically reconnect after transient network failures
- NFR18: Reconnection shall use exponential backoff (5s → 10s → 30s → 60s max)

### Integration

- NFR19: Application shall support Plex Media Server API v1.x
- NFR20: Application shall support Discord RPC protocol (local IPC)
- NFR21: Application shall support mDNS/GDM for server discovery on all platforms
- NFR22: Application shall function with Plex server on local network without internet access
- NFR23: Discord integration shall function without internet access (local IPC only)

### Usability

- NFR24: Setup wizard shall be completable in under 2 minutes by non-technical users
- NFR25: Application shall provide clear, actionable error messages (no technical jargon)
- NFR26: Application shall respect system dark/light mode preferences where supported
- NFR27: System tray icon shall clearly indicate connection status at a glance

### Maintainability

- NFR28: Application binary shall be less than 20MB per platform
- NFR29: Application shall be distributable as a single file (no runtime dependencies)
