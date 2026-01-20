---
stepsCompleted: [1, 2, 3, 4, 5]
inputDocuments:
  - README.md
date: 2026-01-16
author: Batti
project_name: PlexCord
---

# Product Brief: PlexCord

## Executive Summary

PlexCord is a modern, cross-platform desktop application that bridges Plex Media Server and Discord Rich Presence. Built with Go and Wails, featuring a polished Vue.js interface powered by PrimeVue and TailwindCSS, it delivers the seamless music-sharing experience that Plex users have been missing.

Unlike existing solutions that are Windows-only and require manual configuration, PlexCord offers automatic Plex server discovery, streamlined OAuth authentication, real-time webhook-based updates, and a beautiful modern UI—all packaged as a single executable with zero dependencies.

PlexCord isn't just another Discord-Plex bridge. It's the premium, polished experience that Plex users deserve.

---

## Core Vision

### Problem Statement

Plex users who have invested in building and curating their own music libraries are invisible on Discord. While Spotify users enjoy native Discord integration that automatically displays what they're listening to, Plex/Plexamp users have no out-of-the-box solution for sharing their music activity with friends.

### Problem Impact

- **Social disconnection**: Plex users can't participate in the casual music discovery that happens naturally when friends see each other's listening activity
- **Visibility gap**: Despite running sophisticated media servers, Plex users appear "offline" or inactive on Discord while listening to music
- **Community exclusion**: Discord communities built around music sharing inadvertently exclude Plex users

### Why Existing Solutions Fall Short

The primary existing solution, PlexampRPC, addresses core functionality but has significant limitations:

| Limitation | Impact |
|------------|--------|
| Windows-only | Mac and Linux users completely excluded |
| .NET runtime required | Extra installation step, dependency management |
| Manual token configuration | Friction-heavy setup process |
| Polling-only updates | 15-second delay, continuous API requests |
| Basic GUI | Dated appearance, minimal polish |

### Proposed Solution

PlexCord delivers Discord Rich Presence for Plexamp users with an emphasis on universal accessibility, effortless setup, and premium user experience:

**Architecture:**
- **Backend**: Go for performance, single-binary distribution
- **Frontend**: Wails + Vue.js + PrimeVue + TailwindCSS for modern, polished UI
- **Communication**: Webhooks (instant updates) with polling fallback
- **Distribution**: Single executable, zero dependencies, cross-platform

**User Experience:**
- **Auto-discovery**: Detects Plex servers on local network automatically
- **Streamlined auth**: OAuth flows for both Plex and Discord—no manual token hunting
- **Setup wizard**: Guided first-run experience, configured in under 2 minutes
- **Modern UI**: Dark mode, system tray integration, real-time now-playing display
- **Headless support**: JSON configuration for servers and power users

### Key Differentiators

| Feature | PlexampRPC | PlexCord |
|---------|------------|----------|
| Platform | Windows-only | Windows, Mac, Linux |
| Runtime | .NET 10 required | None (single binary) |
| Setup | Manual token entry | Auto-discovery + OAuth wizard |
| Updates | Polling only | Webhooks + polling fallback |
| UI/UX | Basic .NET GUI | Modern Vue.js + PrimeVue + Tailwind |
| Auth | Manual configuration | Automatic OAuth flows |

### Tech Stack

| Layer | Technology |
|-------|------------|
| Backend | Go |
| Framework | Wails |
| Frontend | Vue.js 3 (Composition API) |
| Components | PrimeVue |
| Styling | TailwindCSS |
| Plex Integration | Webhooks + REST API |
| Discord Integration | RPC Protocol |

---

## Target Users

### Primary Users

#### The Self-Hosting Enthusiast

Technical users running their own Plex servers, often on Linux, who want a cross-platform solution that "just works." They value control, open-source ethos, and showing off their curated libraries. PlexampRPC's Windows-only limitation has left them without options.

**Key Characteristics:**

- Runs home lab / self-hosted infrastructure
- Linux or Mac primary OS
- Active in technical Discord communities
- Comfortable with CLI but appreciates good UX

#### The Music Curator

Non-technical users who chose Plex/Plexamp for the experience and their curated music library. They want the same social visibility that Spotify users enjoy without needing to understand APIs or configuration files.

**Key Characteristics:**

- Mac or Windows desktop user
- Values aesthetic and polish
- Not comfortable editing config files
- Wants one-click setup experience

#### The Headless Operator

Power users who want PlexCord running on their server 24/7, not on their desktop. They prefer JSON configuration, Docker containers, and systemd services over GUI applications.

**Key Characteristics:**

- Runs Plex on NAS or dedicated server
- Prefers headless/CLI operation
- Wants set-and-forget automation
- Comfortable with JSON/environment variables

### Secondary Users

**Discord Friends** - Don't use PlexCord directly but benefit from seeing their friends' music activity. They represent a word-of-mouth growth vector through "how did you do that?" moments.

### User Journey

1. **Discovery**: Word of mouth, Reddit (r/Plex, r/selfhosted), GitHub search, PlexampRPC alternative seekers
2. **Onboarding**: Download binary → Wizard auto-discovers Plex → OAuth authentication → Connected in <2 minutes
3. **Core Usage**: Runs invisibly in system tray or as background service; zero daily interaction
4. **Success Moment**: First Discord status update; first friend asking "how did you do that?"
5. **Long-term**: Becomes invisible infrastructure they forget is running

---

## Success Metrics

### User Success Metrics

| Metric | Target | Measurement |
|--------|--------|-------------|
| Setup Completion Rate | >95% | Wizard completion tracking |
| Time to First Presence | <2 minutes | Download to Discord status |
| Crash-Free Sessions | >99.9% | Error reporting / issues |
| Zero-Touch Operation | No daily interaction needed | Absence of usage questions |

**User Value Indicators:**

- "How did you do that?" moments from Discord friends
- Set-and-forget operation after initial setup
- Mac/Linux users finally have a working solution

### Business Objectives

As an open-source project, business objectives focus on adoption and community:

| Objective | 3-Month | 12-Month |
|-----------|---------|----------|
| GitHub Stars | 100+ | 500+ |
| Total Downloads | 500+ | 5,000+ |
| Active Installations | 200+ | 1,000+ |
| External Contributors | 2+ | 10+ |

**Strategic Goal:** Become the default recommendation for Plex-Discord integration, surpassing PlexampRPC as the go-to solution.

### Key Performance Indicators

**North Star Metric:** Active Weekly Users - users with at least one Discord presence update per week.

**Supporting KPIs:**

1. Onboarding Success Rate (>95%)
2. Platform Coverage (no platform <20% of users)
3. Stability Score (<1 crash per 1,000 sessions)
4. Community Sentiment (positive GitHub/Reddit mentions)
5. Webhook Adoption Rate (% using real-time vs. polling)

**Technical Quality:**

- Binary size <20MB
- Memory usage <50MB idle
- CPU usage <1% average
- Startup time <3 seconds
- Discord update latency <2 seconds (webhooks)

---

## MVP Scope

### Core Features (v1.0)

| Feature | Description |
|---------|-------------|
| Discord Rich Presence | Track title, artist, album, play/pause state |
| Plex Playback Detection | Polling `/status/sessions` API |
| Setup Wizard | Guided first-run with server discovery |
| Plex Server Discovery | Auto-detect via mDNS/GDM |
| Modern UI | Wails + Vue.js + PrimeVue + TailwindCSS |
| System Tray | Minimize to tray, current track display |
| Cross-Platform | Windows, Mac, Linux single binaries |

**Authentication:** Guided token/Client ID entry with validation (full OAuth deferred to v1.1)

### Out of Scope for MVP

- Webhooks (polling sufficient for v1)
- Full OAuth flows (guided entry works)
- Headless/JSON config mode
- Docker support
- Album art in Discord presence
- Custom presence templates
- Auto-updates
- Multiple Plex server support
- Last.fm scrobbling

### MVP Success Criteria

1. Works on all three platforms (Windows, Mac, Linux)
2. Download → Discord presence in <5 minutes
3. Wizard completion rate >90%
4. Stable 24+ hour operation
5. Polished, modern UI

**Proceed to v1.1 when:** 50+ stars, 100+ downloads, positive community sentiment

### Future Vision

**v1.1 - Polish Release:** Webhooks, album art, progress bar, auto-updates, full OAuth

**v1.2 - Power User Release:** Headless mode, Docker, custom templates, CLI automation

**v2.0 - Platform Release:** Multi-server, Last.fm, listening stats, Discord bot companion
