# Development Guide

This guide covers setting up your development environment, building PlexCord, and understanding the development workflow.

## Prerequisites

### Required Tools

- **Go 1.21 or later** - [Download](https://go.dev/dl/)
- **Node.js 18 or later** - [Download](https://nodejs.org/)
- **Wails CLI v2.11.0** - Install with:

  ```bash
  go install github.com/wailsapp/wails/v2/cmd/wails@latest
  ```

- **Git** - Version control

### Platform-Specific Requirements

**Windows:**

- No additional requirements

**macOS:**

- Xcode Command Line Tools: `xcode-select --install`

**Linux:**

- Build essentials: `sudo apt install build-essential libgtk-3-dev libwebkit2gtk-4.0-dev`

## Getting Started

### 1. Clone the Repository

```bash
git clone https://github.com/yourusername/PlexCord.git
cd PlexCord
```

### 2. Install Dependencies

```bash
# Install Go dependencies
go mod download

# Install frontend dependencies
cd frontend
npm install
cd ..
```

### 3. Run in Development Mode

```bash
wails dev
```

This starts the application with:

- Hot module reload for frontend changes
- Auto-restart on backend changes
- Developer console enabled
- Debug logging

## Project Structure

```
PlexCord/
├── app.go              # Wails app initialization
├── main.go             # Application entry point
├── go.mod              # Go dependencies
├── wails.json          # Wails configuration
│
├── internal/           # Backend Go code
│   ├── plex/          # Plex API integration
│   ├── discord/       # Discord RPC integration
│   ├── config/        # Configuration management
│   ├── keychain/      # Secure credential storage
│   ├── platform/      # OS-specific features
│   └── errors/        # Error handling
│
├── frontend/           # Vue.js frontend
│   ├── src/
│   │   ├── views/     # Page components
│   │   ├── components/# Reusable components
│   │   ├── stores/    # Pinia state stores
│   │   ├── router/    # Vue Router
│   │   └── service/   # Backend API calls
│   ├── package.json
│   ├── vite.config.js
│   └── tailwind.config.js
│
└── build/             # Build output directory
```

## Development Workflow

### Backend Development (Go)

#### Adding a New Package

1. Create package directory under `internal/`:

   ```bash
   mkdir internal/mypackage
   touch internal/mypackage/mypackage.go
   ```

2. Define package interface:

   ```go
   package mypackage

   type MyService interface {
       DoSomething() error
   }

   type myServiceImpl struct {
       // fields
   }

   func NewMyService() MyService {
       return &myServiceImpl{}
   }
   ```

3. Add tests:

   ```bash
   touch internal/mypackage/mypackage_test.go
   ```

#### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./internal/plex
```

#### Code Style

- Use `gofmt` for formatting
- Follow [Effective Go](https://go.dev/doc/effective_go) guidelines
- Add comments for exported functions and types
- Keep functions small and focused

### Frontend Development (Vue.js)

#### Adding a New View

1. Create view file:

   ```bash
   touch frontend/src/views/MyView.vue
   ```

2. Add route in `frontend/src/router/index.js`:

   ```javascript
   {
     path: '/my-view',
     name: 'MyView',
     component: () => import('../views/MyView.vue')
   }
   ```

#### Creating a Component

```vue
<template>
  <div class="my-component">
    <!-- Component template -->
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'

// Component logic
</script>

<style scoped>
/* Component styles */
</style>
```

#### Working with Stores

```javascript
// stores/mystore.ts
import { defineStore } from 'pinia'

export const useMyStore = defineStore('mystore', {
  state: () => ({
    data: null
  }),
  actions: {
    async fetchData() {
      // Fetch data logic
    }
  }
})
```

#### Code Style

- Use Vue 3 Composition API
- Use `<script setup>` syntax
- Follow PrimeVue component patterns
- Use TailwindCSS utility classes
- Keep components under 200 lines

### Calling Backend from Frontend

```javascript
import { GetSessions } from '../wailsjs/go/plex/Client'

try {
  const sessions = await GetSessions()
  console.log('Sessions:', sessions)
} catch (error) {
  console.error('Error:', error)
}
```

### Emitting Events from Backend

```go
import "github.com/wailsapp/wails/v2/pkg/runtime"

func (a *App) NotifyFrontend(ctx context.Context, message string) {
    runtime.EventsEmit(ctx, "notification", message)
}
```

### Listening to Events in Frontend

```javascript
import { EventsOn } from '../wailsjs/runtime/runtime'

EventsOn('notification', (message) => {
  console.log('Notification:', message)
})
```

## Building

### Development Build

```bash
# Build for current platform
wails build

# Output location: build/bin/
```

### Production Build

```bash
# Windows
wails build -platform windows/amd64 -clean

# macOS (Universal - Intel + Apple Silicon)
wails build -platform darwin/universal -clean

# Linux
wails build -platform linux/amd64 -clean
```

### Build Configuration

Edit `wails.json` to customize build settings:

```json
{
  "name": "PlexCord",
  "outputfilename": "plexcord",
  "frontend:install": "npm install",
  "frontend:build": "npm run build",
  "wailsjsdir": "./frontend"
}
```

## Debugging

### Backend Debugging

1. Add debug logging:

   ```go
   import "log"
   log.Printf("Debug: %+v", data)
   ```

2. Use VS Code debugger with this launch configuration:

   ```json
   {
     "version": "0.2.0",
     "configurations": [
       {
         "name": "Wails Dev",
         "type": "go",
         "request": "launch",
         "mode": "exec",
         "program": "${workspaceFolder}/build/bin/plexcord"
       }
     ]
   }
   ```

### Frontend Debugging

1. Open browser dev tools in running app: `Ctrl+Shift+I` (Windows/Linux) or `Cmd+Option+I` (macOS)
2. Use Vue DevTools browser extension
3. Add `console.log()` statements
4. Use `debugger;` to set breakpoints

## Testing

### Backend Tests

```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Run specific test
go test -run TestPlexClient ./internal/plex
```

### Frontend Tests (Coming Soon)

```bash
# Run unit tests
npm test

# Run with coverage
npm run test:coverage
```

## Common Tasks

### Adding a New Go Dependency

```bash
go get github.com/example/package
go mod tidy
```

### Adding a New Frontend Dependency

```bash
cd frontend
npm install package-name
```

### Updating Dependencies

```bash
# Update Go dependencies
go get -u ./...
go mod tidy

# Update frontend dependencies
cd frontend
npm update
```

### Cleaning Build Artifacts

```bash
# Clean Wails build
wails clean

# Remove node_modules
rm -rf frontend/node_modules

# Remove Go build cache
go clean -cache
```

## Platform-Specific Development

### Windows Development

- Use PowerShell or Command Prompt
- Build output: `build\bin\PlexCord.exe`
- Config location: `%APPDATA%\PlexCord\`

### macOS Development

- Universal builds require macOS
- Build output: `build/bin/PlexCord.app`
- Config location: `~/Library/Application Support/PlexCord/`
- For unsigned builds: `xattr -cr build/bin/PlexCord.app`

### Linux Development

- Build output: `build/bin/plexcord`
- Config location: `~/.config/plexcord/`
- Requires GTK3 and WebKit2GTK

## Troubleshooting

### Wails CLI not found

```bash
# Ensure GOPATH/bin is in PATH
export PATH=$PATH:$(go env GOPATH)/bin

# Or reinstall
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

### Frontend not building

```bash
# Clear npm cache and reinstall
cd frontend
rm -rf node_modules package-lock.json
npm install
```

### Hot reload not working

```bash
# Restart with clean build
wails clean
wails dev
```

### Build errors on macOS

```bash
# Install Xcode Command Line Tools
xcode-select --install

# Accept license
sudo xcodebuild -license accept
```

## Next Steps

- [Contributing Guide](./contributing.md) - How to contribute to PlexCord
- [Architecture Overview](./architecture.md) - Understand the codebase structure
- [API Documentation](./api.md) - Backend API reference
