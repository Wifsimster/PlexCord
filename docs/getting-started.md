# Getting Started with PlexCord

PlexCord bridges your Plex Media Server and Discord, displaying your music playback as Discord Rich Presence.

## Quick Start

### 1. Download & Install

Download the latest release for your platform from the [Releases](https://github.com/yourusername/PlexCord/releases) page:

- **Windows**: `PlexCord.exe`
- **macOS**: `PlexCord.app`
- **Linux**: `plexcord`

### 2. First Launch

When you first run PlexCord, the Setup Wizard will guide you through configuration:

#### Step 1: Plex Authentication

1. You'll need a Plex authentication token
2. Click the "How to get token" link for instructions
3. Paste your token into the input field

**Getting Your Plex Token:**
1. Log into [Plex Web App](https://app.plex.tv)
2. Play any media item
3. Click the ⋮ menu → "Get Info"
4. Click "View XML"
5. Look for `X-Plex-Token=` in the URL
6. Copy everything after the `=` sign

#### Step 2: Select Your Plex Server

PlexCord will automatically discover Plex servers on your local network. Select your server from the list.

**If auto-discovery doesn't work:**
- Click "Enter manually"
- Enter your server URL (e.g., `http://192.168.1.100:32400`)

#### Step 3: Verify Connection

PlexCord will test the connection to your Plex server and confirm that it can detect your music playback.

#### Step 4: Discord Setup (Automatic)

PlexCord automatically connects to your local Discord client. Just make sure Discord is running.

### 3. Using PlexCord

Once setup is complete:

1. PlexCord runs in your system tray
2. Start playing music in Plex or Plexamp
3. Your Discord status will update automatically
4. When you stop playing, the status clears

## System Requirements

- **Operating System**: Windows 10+, macOS 11+, or Linux (Ubuntu 20.04+)
- **Discord**: Desktop application must be installed and running
- **Plex**: Active Plex Media Server with music library

## Troubleshooting

### Discord status not showing

1. Make sure Discord Desktop is running (web version won't work)
2. Check that Discord Rich Presence is enabled in Discord Settings → Activity Privacy
3. Try restarting both PlexCord and Discord

### Can't connect to Plex

1. Verify your token is correct (tokens don't expire but can be revoked)
2. Check that your Plex server is running
3. If using remote connection, ensure your server is accessible
4. Try entering the server URL manually

### PlexCord not detecting playback

1. Make sure you're playing music (not videos or other media)
2. Check that PlexCord shows "Connected" status
3. Try stopping and starting playback
4. Check the polling interval in Settings (default: 5 seconds)

## Next Steps

- [Configure Settings](./settings.md) - Customize polling intervals, auto-start, and more
- [Architecture Overview](./architecture.md) - Learn how PlexCord works
- [Contributing](./contributing.md) - Help improve PlexCord
