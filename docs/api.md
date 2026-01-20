# API Reference

This document describes the backend API exposed to the frontend via Wails.

## Plex API

### `ConnectToPlex(token string, serverURL string) error`

Establishes connection to a Plex server.

**Parameters:**

- `token` - Plex authentication token
- `serverURL` - URL of the Plex server (e.g., `http://192.168.1.100:32400`)

**Returns:**

- `error` - Error if connection fails

**Errors:**

- `PLEX_UNREACHABLE` - Server not responding
- `PLEX_AUTH_FAILED` - Invalid token
- `PLEX_INVALID_URL` - Malformed server URL

**Example:**

```javascript
import { ConnectToPlex } from '../wailsjs/go/main/App'

try {
  await ConnectToPlex('abc123token', 'http://192.168.1.100:32400')
  console.log('Connected successfully')
} catch (error) {
  console.error('Connection failed:', error)
}
```

---

### `DiscoverServers() ([]Server, error)`

Discovers Plex servers on the local network using mDNS/GDM.

**Returns:**

- `Server[]` - Array of discovered servers
- `error` - Error if discovery fails

**Server Type:**

```go
type Server struct {
    Name     string `json:"name"`     // Server name
    URL      string `json:"url"`      // Server URL
    Version  string `json:"version"`  // Plex version
    Local    bool   `json:"local"`    // Is local server
}
```

**Example:**

```javascript
import { DiscoverServers } from '../wailsjs/go/main/App'

try {
  const servers = await DiscoverServers()
  servers.forEach(server => {
    console.log(`Found: ${server.name} at ${server.url}`)
  })
} catch (error) {
  console.error('Discovery failed:', error)
}
```

---

### `GetCurrentSession() (Session, error)`

Gets the current music playback session.

**Returns:**

- `Session` - Current playback session
- `error` - Error if no session or fetch fails

**Session Type:**

```go
type Session struct {
    SessionKey  string `json:"sessionKey"`   // Unique session ID
    Title       string `json:"title"`        // Track title
    Artist      string `json:"artist"`       // Track artist
    Album       string `json:"album"`        // Album name
    AlbumArt    string `json:"albumArt"`     // Album art URL
    Duration    int    `json:"duration"`     // Track duration (ms)
    Position    int    `json:"position"`     // Current position (ms)
    State       string `json:"state"`        // playing, paused, stopped
    UserID      string `json:"userId"`       // Plex user ID
}
```

**Example:**

```javascript
import { GetCurrentSession } from '../wailsjs/go/main/App'

try {
  const session = await GetCurrentSession()
  console.log(`Now playing: ${session.artist} - ${session.title}`)
} catch (error) {
  console.log('No active session')
}
```

---

### `StartSessionPolling(intervalSeconds int) error`

Starts polling Plex for session updates at the specified interval.

**Parameters:**

- `intervalSeconds` - Polling interval in seconds (default: 5)

**Returns:**

- `error` - Error if polling cannot start

**Example:**

```javascript
import { StartSessionPolling } from '../wailsjs/go/main/App'

await StartSessionPolling(5) // Poll every 5 seconds
```

---

### `StopSessionPolling() error`

Stops the session polling loop.

**Returns:**

- `error` - Error if polling cannot stop

---

## Discord API

### `ConnectToDiscord(clientID string) error`

Establishes connection to local Discord client via RPC.

**Parameters:**

- `clientID` - Discord Application Client ID

**Returns:**

- `error` - Error if connection fails

**Errors:**

- `DISCORD_NOT_RUNNING` - Discord client not detected
- `DISCORD_CONN_FAILED` - RPC connection error
- `DISCORD_INVALID_CLIENT_ID` - Invalid client ID format

**Example:**

```javascript
import { ConnectToDiscord } from '../wailsjs/go/main/App'

try {
  await ConnectToDiscord('123456789012345678')
  console.log('Discord connected')
} catch (error) {
  console.error('Discord connection failed:', error)
}
```

---

### `UpdatePresence(session Session) error`

Updates Discord Rich Presence with current track info.

**Parameters:**

- `session` - Current playback session

**Returns:**

- `error` - Error if update fails

**Example:**

```javascript
import { UpdatePresence } from '../wailsjs/go/main/App'

const session = {
  title: 'Bohemian Rhapsody',
  artist: 'Queen',
  album: 'A Night at the Opera',
  state: 'playing'
}

await UpdatePresence(session)
```

---

### `ClearPresence() error`

Clears Discord Rich Presence.

**Returns:**

- `error` - Error if clear fails

---

### `IsDiscordRunning() (bool, error)`

Checks if Discord client is running.

**Returns:**

- `bool` - True if Discord is running
- `error` - Error if check fails

---

## Config API

### `GetConfig() (Config, error)`

Retrieves the current application configuration.

**Returns:**

- `Config` - Current configuration
- `error` - Error if config cannot be read

**Config Type:**

```go
type Config struct {
    PlexServerURL    string `json:"plexServerUrl"`
    PlexToken        string `json:"plexToken"`
    DiscordClientID  string `json:"discordClientId"`
    PollingInterval  int    `json:"pollingInterval"`
    AutoStart        bool   `json:"autoStart"`
    MinimizeToTray   bool   `json:"minimizeToTray"`
    SetupComplete    bool   `json:"setupComplete"`
}
```

---

### `SaveConfig(config Config) error`

Saves the application configuration.

**Parameters:**

- `config` - Configuration to save

**Returns:**

- `error` - Error if save fails

---

### `ResetConfig() error`

Resets configuration to defaults.

**Returns:**

- `error` - Error if reset fails

---

## Keychain API

### `StoreToken(service string, account string, token string) error`

Stores a token securely in OS keychain.

**Parameters:**

- `service` - Service name (e.g., "PlexCord")
- `account` - Account identifier (e.g., "plex_token")
- `token` - Token to store

**Returns:**

- `error` - Error if storage fails

---

### `GetToken(service string, account string) (string, error)`

Retrieves a token from OS keychain.

**Parameters:**

- `service` - Service name
- `account` - Account identifier

**Returns:**

- `string` - Retrieved token
- `error` - Error if retrieval fails

---

### `DeleteToken(service string, account string) error`

Deletes a token from OS keychain.

**Parameters:**

- `service` - Service name
- `account` - Account identifier

**Returns:**

- `error` - Error if deletion fails

---

## Events

Events are emitted from the backend to the frontend using Wails runtime events.

### `session:update`

Emitted when playback session changes.

**Payload:** `Session` object

**Example:**

```javascript
import { EventsOn } from '../wailsjs/runtime/runtime'

EventsOn('session:update', (session) => {
  console.log('Session updated:', session)
})
```

---

### `connection:status`

Emitted when connection status changes.

**Payload:**

```javascript
{
  plexConnected: boolean,
  discordConnected: boolean,
  lastUpdate: string
}
```

---

### `error`

Emitted when an error occurs.

**Payload:**

```javascript
{
  code: string,      // Error code (e.g., "PLEX_UNREACHABLE")
  message: string,   // Human-readable message
  timestamp: string  // ISO 8601 timestamp
}
```

**Example:**

```javascript
EventsOn('error', (error) => {
  console.error(`Error ${error.code}: ${error.message}`)
})
```

---

## Error Codes

| Code | Description |
|------|-------------|
| `PLEX_UNREACHABLE` | Plex server not responding |
| `PLEX_AUTH_FAILED` | Invalid Plex token |
| `PLEX_INVALID_URL` | Malformed server URL |
| `PLEX_NO_SESSION` | No active music session |
| `DISCORD_NOT_RUNNING` | Discord client not detected |
| `DISCORD_CONN_FAILED` | Discord RPC connection error |
| `DISCORD_INVALID_CLIENT_ID` | Invalid Discord client ID |
| `CONFIG_READ_FAILED` | Cannot read config file |
| `CONFIG_WRITE_FAILED` | Cannot write config file |
| `KEYCHAIN_UNAVAILABLE` | OS keychain not accessible |
