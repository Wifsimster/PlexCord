# PlexCord

> Show off what you're listening to on Plex in your Discord status

PlexCord is a lightweight, cross-platform desktop application that seamlessly integrates your Plex music playback with Discord Rich Presence. Built with modern technologies (Go + Vue 3 + Wails), it runs efficiently in the background while displaying your currently playing music on Discord.

## Features

- **Real-Time Sync** - Automatically detects and displays your current Plex music playback
- **Rich Presence Integration** - Shows track name, artist, album, and playback state in Discord
- **Cross-Platform** - Native support for Windows, macOS, and Linux
- **Secure** - Credentials stored safely using OS keychain (Windows Credential Manager, macOS Keychain, Linux Secret Service)
- **Lightweight** - Small memory footprint, runs quietly in system tray
- **Easy Setup** - Intuitive setup wizard guides you through configuration
- **Auto-Recovery** - Automatically reconnects if Discord or Plex connection is lost
- **No Dependencies** - Single standalone binary, no runtime requirements


## Installation

### Download Pre-Built Binaries

Download the latest release for your platform from the [Releases](../../releases) page:

- **Windows** - `PlexCord-windows-amd64.exe`
- **macOS** - `PlexCord-darwin-universal.app` (Intel & Apple Silicon)
- **Linux** - `PlexCord-linux-amd64`

### Quick Setup

1. **Download** the appropriate binary for your operating system
2. **Launch** PlexCord - the setup wizard will guide you through:
   - Connecting to your Plex server (via auto-discovery or manual entry)
   - Entering your Plex authentication token
   - Verifying Discord connection
3. **Start listening** - Play music on Plex/Plexamp and watch your Discord status update!

For detailed setup instructions, see the [Getting Started Guide](docs/getting-started.md).

## How It Works

1. PlexCord connects to your Plex Media Server using your authentication token
2. It continuously monitors your active sessions for music playback
3. When music is detected, it extracts metadata (track, artist, album, artwork)
4. This information is sent to Discord via Rich Presence API
5. Your Discord status shows what you're listening to in real-time

## Requirements

- **Plex Media Server** with at least one music library
- **Plex Pass** (optional but recommended for best experience)
- **Discord Desktop App** running (Rich Presence only works with desktop client)
- **Supported OS:**
  - Windows 10 or later
  - macOS 11 (Big Sur) or later
  - Linux with X11/Wayland (Ubuntu 20.04+, Fedora, Arch, etc.)

## Documentation

- [Getting Started](docs/getting-started.md) - Installation and first-time setup
- [Architecture](docs/architecture.md) - Technical overview and design
- [API Reference](docs/api.md) - Backend API documentation
- [Development](docs/development.md) - Build from source and contribute
- [Contributing](docs/contributing.md) - Contribution guidelines
- [Roadmap](docs/roadmap.md) - Future plans and features

## Building from Source

### Prerequisites

- [Go 1.21+](https://golang.org/dl/)
- [Node.js 18+](https://nodejs.org/)
- [Wails CLI](https://wails.io/docs/gettingstarted/installation)

### Build Steps

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

# Output: build/bin/PlexCord (or PlexCord.exe on Windows)
```

For platform-specific builds and advanced options, see [Development Guide](docs/development.md).

## Configuration

PlexCord stores configuration in platform-specific locations:

- **Windows:** `%APPDATA%\PlexCord\config.json`
- **macOS:** `~/Library/Application Support/PlexCord/config.json`
- **Linux:** `~/.config/plexcord/config.json`

Credentials are securely stored in:
- **Windows:** Windows Credential Manager
- **macOS:** macOS Keychain
- **Linux:** Secret Service API (GNOME Keyring, KWallet, etc.)

## Troubleshooting

### Discord Not Showing Status

- Ensure Discord desktop app is running (not web version)
- Check that "Display current activity as a status message" is enabled in Discord settings
- Restart both PlexCord and Discord

### Can't Connect to Plex

- Verify your Plex token is correct
- Check that Plex Media Server is running and accessible
- For remote servers, ensure port forwarding/network access is configured
- Try manual server entry if auto-discovery fails

### Linux Specific Issues

- If binary won't run, make it executable: `chmod +x PlexCord`
- Ensure Secret Service is available for credential storage
- Check that D-Bus is running for Discord RPC

For more issues, see [Getting Started - Troubleshooting](docs/getting-started.md#troubleshooting) or [open an issue](../../issues).

## Contributing

Contributions are welcome! Here's how you can help:

- **Report Bugs** - [Open an issue](../../issues) with details and reproduction steps
- **Suggest Features** - [Start a discussion](../../discussions) with your ideas
- **Submit PRs** - See [Contributing Guide](docs/contributing.md) for guidelines
- **Improve Docs** - Help make documentation clearer and more comprehensive

Please read the [Contributing Guide](docs/contributing.md) before submitting pull requests.

## Project Status

**Current Version:** v0.1.0-dev

**Status:** Active Development

This project is in early development. Core features are functional but expect bugs and missing features. See [Roadmap](docs/roadmap.md) for planned features and timeline.

## Tech Stack

- **Backend:** Go 1.21+
- **Frontend:** Vue 3 (Composition API) + PrimeVue + TailwindCSS
- **Framework:** Wails v2 (Go + Web GUI)
- **Discord:** Rich Presence via [rich-go](https://github.com/hugolgst/rich-go)
- **Plex:** Official Plex API

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Acknowledgments

- [Wails](https://wails.io) - Amazing Go + Web framework
- [PrimeVue](https://primevue.org) - Beautiful Vue UI components
- [TailwindCSS](https://tailwindcss.com) - Utility-first CSS framework
- [rich-go](https://github.com/hugolgst/rich-go) - Discord Rich Presence for Go

## Disclaimer

This project is not affiliated with, endorsed by, or connected to Plex Inc. or Discord Inc. All product names, trademarks, and registered trademarks are property of their respective owners.

---

**Made with ❤️ for music lovers who want to share what they're listening to**
