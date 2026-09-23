![PlexCord](docs/images/banner.svg)

# PlexCord

> Display your Plex playback — music, films and TV — as Discord Rich Presence

PlexCord is a lightweight, cross-platform desktop application that bridges your Plex Media Server and Discord, showing what you're listening to or watching in real-time. Built with Go + Vue 3 + Wails.

![PlexCord live on air with real Plex artwork](docs/images/now-playing.png)

## Features

- 🎵 Real-time Plex playback detection — music, movies and TV shows
- 💬 Discord Rich Presence integration (track/artist/album, film and year, show with season & episode, artwork)
- 🎚️ Pick what gets shared: each kind of media has its own switch
- 🖥️ Cross-platform support (Windows, macOS, Linux)
- 🔐 Secure credential storage (OS keychain integration)
- ⚡ Lightweight single binary (<20MB)
- 🎨 "ON AIR" interface: a tally lamp that tells you when you are live, and a stage washed in your artwork
- 🔄 Auto-recovery on connection loss

## Screenshots

### On air

Your artwork floods the stage while the track goes live on your Discord profile (French UI shown).

![PlexCord on air](docs/images/now-playing.png)

### Dashboard

Live presence preview and connection health at a glance.

![Dashboard](docs/images/dashboard.png)

### Settings

Customize the polling interval, presence format, startup behavior, and more.

![Settings](docs/images/settings.png)

### Setup wizard

A guided, two-minute setup connects Plex and Discord.

| Welcome | Complete |
| --- | --- |
| ![Setup — Welcome](docs/images/setup-welcome.png) | ![Setup — Complete](docs/images/setup-complete.png) |


## Quick Start

**[Download the latest release](../../releases)** for your platform, run the app, and follow the setup wizard to connect Plex and Discord.

On Windows, `PlexCord-windows-amd64-installer.exe` installs PlexCord for the
current user (no administrator rights, shortcuts and an uninstall entry
included); `PlexCord-windows-amd64.exe` is the same app as a portable
executable. Both keep themselves up to date.

**Requirements:** Plex Media Server + Discord Desktop App

📖 **Detailed instructions:** [Getting Started Guide](docs/getting-started.md)

## Documentation

- **[Getting Started](docs/getting-started.md)** - Installation, setup wizard, and troubleshooting
- **[Brand](docs/brand.md)** - The mark, the wordmark, and where the brand assets live
- **[Architecture](docs/architecture.md)** - System design, technology stack, and project structure
- **[SOLID Review](docs/solid-review.md)** - Design audit against the SOLID principles, scored
- **[API Reference](docs/api.md)** - Backend API documentation and error codes
- **[Development](docs/development.md)** - Build from source, development workflow, and testing
- **[Contributing](docs/contributing.md)** - How to contribute, coding guidelines, and PR process
- **[Roadmap](docs/roadmap.md)** - Feature roadmap, release plan, and version history

## Building from Source

```bash
# Clone repository
git clone https://github.com/Wifsimster/PlexCord.git
cd PlexCord

# Install dependencies
go mod download
cd frontend && npm install && cd ..

# Development mode with hot reload
wails dev

# Build production binary
wails build
```

**Prerequisites:** Go 1.21+, Node.js 18+, Wails CLI

For detailed build instructions and development setup, see the [Development Guide](docs/development.md).

## Contributing

Contributions are welcome! Please read the [Contributing Guide](docs/contributing.md) for guidelines.

- Report bugs via [Issues](../../issues)
- Suggest features via [Discussions](../../discussions)
- Submit pull requests for improvements

## Tech Stack

- **Backend:** Go 1.21+
- **Frontend:** Vue 3 (Composition API) + PrimeVue + TailwindCSS
- **Framework:** Wails v2
- **Discord Integration:** [rich-go](https://github.com/hugolgst/rich-go)

## License

MIT License - see [LICENSE](LICENSE) for details.

## Disclaimer

This project is not affiliated with or endorsed by Plex Inc. or Discord Inc.

---

**Made with ❤️ for music and movie lovers**
