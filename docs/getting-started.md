# Getting Started with PlexCord

PlexCord bridges your Plex Media Server and Discord, displaying your music playback as Discord Rich Presence.

## Quick Start

### 1. Download & Install

Download the latest release for your platform from the [Releases](https://github.com/Wifsimster/PlexCord/releases) page:

- **Windows (recommended)**: `PlexCord-windows-amd64-installer.exe` — the installer.
  It installs PlexCord for the current user only, under
  `%LOCALAPPDATA%\Programs\PlexCord`, so it never asks for administrator
  rights, and it adds Start menu and desktop shortcuts plus an entry in
  *Apps & features* for uninstalling. Installing over an existing copy closes
  the running PlexCord first and upgrades it in place, keeping your settings.
- **Windows (portable)**: `PlexCord-windows-amd64.exe` — the bare executable if
  you would rather not install anything. Put it wherever you like and run it.
- **macOS**: `PlexCord-darwin-universal.dmg`
- **Linux**: `PlexCord-linux-amd64.AppImage`

Windows may warn that the file is unrecognized: the binaries are not
code-signed, so SmartScreen shows "Windows protected your PC" — choose *More
info → Run anyway*. The SHA256 checksum published next to each file lets you
verify the download.

Both Windows options update themselves: PlexCord replaces its own executable in
the background and applies the update on the next restart (see
[Updates](#updates)). Because the installer puts PlexCord in your own profile
rather than in Program Files, that update needs no administrator prompt.

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

### Running in the background

Settings → App controls how PlexCord behaves around the window:

- **Start on login** launches PlexCord automatically when you log in.
- **Start minimized on login** (on by default, shown once "Start on login" is
  enabled) keeps that launch in the background — into the tray when "Minimize
  to tray" is on, otherwise minimized to the taskbar — so no window opens over
  your desktop at boot. Turn it off if you want the window at every login.
- **Minimize to tray** keeps PlexCord running in the tray when you close the
  window, instead of quitting.
- **Start minimized** does the same for the launches you start yourself:
  PlexCord opens without a window, into the tray when "Minimize to tray" is on
  and minimized to the taskbar otherwise. It takes effect on the next launch.

To bring the window back, click the tray icon (or its "Show PlexCord" menu
entry), or simply open PlexCord again: launching it while it is already running
in the background restores the running instance rather than starting a second
copy.

### Updates

PlexCord checks for a new release shortly after launch and every 6 hours after
that (Settings → About turns this off, and has a "Check for updates" button for
checking on demand). Where PlexCord can update itself, the new version
downloads in the background and all that is left is a restart.

You are told about it wherever you happen to be:

- **In the window**, a notification offers `Restart now` — or, on platforms that
  update manually, a `Download` link to the release page. "Later" dismisses it;
  Settings → About keeps the update, its release notes, and the same buttons.
- **In the tray**, the menu gains an entry for the pending update — `Restart to
  update to v1.5.0` once it is downloaded, `Update available: v1.5.0` before
  that — and the tray icon's tooltip says the same. This is what tells you
  about an update when PlexCord is running in the background with no window
  open. Clicking the entry restarts PlexCord when the update is ready to apply,
  and otherwise opens the window so you can read the release notes first.

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
