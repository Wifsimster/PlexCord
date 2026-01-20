# PlexCord

A cross-platform desktop application that displays your Plex music playback in Discord Rich Presence.

Built with [Wails](https://wails.io) (Go + Vue 3), PlexCord bridges your Plex Media Server and Discord to show what you're listening to.

## Quick Links

- 🚀 [Getting Started](docs/getting-started.md) - Installation and setup guide
- 📚 [Documentation](#documentation) - Complete documentation
- 🛠️ [Development](docs/development.md) - Build from source and contribute
- 🐛 [Report an Issue](../../issues) - Found a bug?
- 💬 [Discussions](../../discussions) - Questions and feedback

## Features

- 🎵 Real-time music playback detection from Plex
- 💬 Discord Rich Presence integration
- 🖥️ Cross-platform support (Windows, macOS, Linux)
- ⚡ Lightweight and fast (<20MB binary)
- 🎨 Modern UI with dark mode support
- 🔐 Secure credential storage using OS keychain
- 📦 Single binary distribution - no dependencies

## Documentation

### Getting Started

- **[Getting Started Guide](docs/getting-started.md)** - First-time setup and usage
  - Installation instructions
  - Setup wizard walkthrough
  - Troubleshooting common issues

### Technical Documentation

- **[Architecture Overview](docs/architecture.md)** - System design and structure
  - High-level architecture diagram
  - Technology stack
  - Package structure
  - Design patterns
  - Data flow

- **[API Reference](docs/api.md)** - Backend API documentation
  - Plex API methods
  - Discord API methods
  - Configuration API
  - Events and error codes

- **[Development Guide](docs/development.md)** - Build and develop PlexCord
  - Development environment setup
  - Project structure
  - Development workflow
  - Building and debugging
  - Testing

### Contributing

- **[Contributing Guide](docs/contributing.md)** - How to contribute
  - Code of conduct
  - Reporting bugs
  - Suggesting features
  - Pull request process
  - Coding guidelines

- **[Project Roadmap](docs/roadmap.md)** - Future plans
  - Current status
  - Release plan
  - Feature requests
  - Version history

## Quick Start

### Installation

Download the latest release for your platform:

- **Windows**: `PlexCord.exe`
- **macOS**: `PlexCord.app`
- **Linux**: `plexcord`

👉 [Download from Releases](../../releases)

### First Launch

1. Launch PlexCord
2. Follow the Setup Wizard to connect your Plex server
3. Make sure Discord is running
4. Start playing music in Plex or Plexamp
5. Your Discord status updates automatically!

📖 See the [Getting Started Guide](docs/getting-started.md) for detailed instructions.

## System Requirements

- **Operating System**: Windows 10+, macOS 11+, or Linux (Ubuntu 20.04+)
- **Discord**: Desktop application (web version not supported)
- **Plex**: Active Plex Media Server with music library
- **No runtime dependencies** - just download and run!

## Platform-Specific Notes

### Windows

- No administrator privileges required
- Binary runs standalone, no installation needed
- Configuration stored in: `%APPDATA%\PlexCord\config.json`

### macOS

- Universal binary supports both Intel and Apple Silicon
- For unsigned builds, you may need to bypass Gatekeeper:

  ```bash
  xattr -cr PlexCord.app
  ```

- Configuration stored in: `~/Library/Application Support/PlexCord/config.json`

### Linux

- Make binary executable after download:

  ```bash
  chmod +x plexcord
  ```

- Requires X11 or Wayland display server
- Configuration stored in: `~/.config/plexcord/config.json`
- Works on Ubuntu, Fedora, Arch, and other major distributions

## Building from Source

Want to build PlexCord yourself or contribute to development?

👉 See the [Development Guide](docs/development.md) for complete instructions on:

- Setting up your development environment
- Running in development mode
- Building for different platforms
- Running tests
- Project structure overview

### Quick Build

```bash
# Clone and install dependencies
git clone https://github.com/yourusername/PlexCord.git
cd PlexCord
go mod download
cd frontend && npm install && cd ..

# Build for current platform
wails build

# Output: build/bin/
```

## Contributing

We welcome contributions! Please see:

- 📖 [Contributing Guide](docs/contributing.md) - Guidelines for contributing
- 🏗️ [Architecture Overview](docs/architecture.md) - Understand the codebase
- 🛠️ [Development Guide](docs/development.md) - Set up your environment

## Project Status

**Current Version:** v0.1.0 (In Development)

**Development Phase:** Epic 1-2 (Foundation & Setup Wizard)

See the [Project Roadmap](docs/roadmap.md) for upcoming features and release plans.

## License

MIT License - See LICENSE file for details

## Acknowledgments

- Built with [Wails](https://wails.io)
- UI components by [PrimeVue](https://primevue.org/)
- Styling with [TailwindCSS](https://tailwindcss.com/)
- Discord RPC integration via [rich-go](https://github.com/hugolgst/rich-go)

## Support

For issues, questions, or feature requests, please open an issue on GitHub.

---

**Note**: This project is not affiliated with or endorsed by Plex Inc. or Discord Inc.
