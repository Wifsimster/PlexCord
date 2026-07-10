# Contributing to PlexCord

Thank you for your interest in contributing to PlexCord! This document provides guidelines and instructions for contributing.

## Code of Conduct

Please be respectful and constructive in all interactions. We aim to foster an inclusive and welcoming community.

## Getting Started

### Prerequisites

- **Go 1.24+** - [Download](https://go.dev/dl/)
- **Node.js 20+** - [Download](https://nodejs.org/)
- **Wails CLI** - Install with `go install github.com/wailsapp/wails/v2/cmd/wails@latest`

**Platform-specific dependencies:**

- **Linux**: `sudo apt-get install libgtk-3-dev libwebkit2gtk-4.0-dev`
- **macOS**: Xcode Command Line Tools
- **Windows**: No additional dependencies

### Development Setup

1. **Fork and clone the repository**
   ```bash
   git clone https://github.com/YOUR_USERNAME/PlexCord.git
   cd PlexCord
   ```

2. **Install dependencies**
   ```bash
   # Install Go dependencies
   go mod download
   
   # Install frontend dependencies
   cd frontend
   npm install
   cd ..
   ```

3. **Run in development mode**
   ```bash
   wails dev
   ```

## Development Workflow

### Branch Naming

- `feature/description` - New features
- `fix/description` - Bug fixes
- `docs/description` - Documentation updates
- `refactor/description` - Code refactoring
- `test/description` - Test additions/updates

### Making Changes

1. **Create a branch**
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes**
   - Write clean, readable code
   - Follow existing code style
   - Add comments for complex logic
   - Update documentation as needed

3. **Test your changes**
   ```bash
   # Run Go tests
   go test ./...
   
   # Run with race detector
   go test -race ./...
   
   # Lint Go code
   golangci-lint run
   
   # Lint frontend code
   cd frontend && npm run lint
   ```

4. **Commit your changes** (see [Commit Message Format](#commit-message-format) below)

5. **Push and create PR**
   ```bash
   git push origin feature/your-feature-name
   ```
   Then create a Pull Request on GitHub.

## Commit Message Format

This project uses [Conventional Commits](https://www.conventionalcommits.org/) for automated versioning and changelog generation. Your commit messages are validated by commitlint when you commit.

### Format

```
type(scope): description

[optional body]

[optional footer(s)]
```

### Supported Types

| Type | Description | Version Bump |
|------|-------------|--------------|
| `feat` | A new feature | Minor (0.x.0) |
| `fix` | A bug fix | Patch (0.0.x) |
| `docs` | Documentation only changes | None |
| `style` | Code style/formatting (no logic changes) | None |
| `refactor` | Code change that neither fixes a bug nor adds a feature | None |
| `perf` | A code change that improves performance | Patch (0.0.x) |
| `test` | Adding missing tests or correcting existing tests | None |
| `build` | Changes to build system or external dependencies | None |
| `ci` | Changes to CI configuration files and scripts | None |
| `chore` | Other changes that don't modify src or test files | None |

### Scope (Optional)

The scope provides additional context about what area of the codebase is affected:

```
feat(discord): add custom activity support
fix(plex): handle connection timeout
docs(readme): update installation instructions
```

Common scopes: `discord`, `plex`, `frontend`, `config`, `keychain`, `platform`

### Examples

**Feature:**
```
feat(discord): add support for custom activity images

Add ability to configure custom images for Discord Rich Presence
activity using Plex poster artwork.
```

**Bug fix:**
```
fix(plex): handle reconnection after network timeout

Previously, the connection would not recover after a network
interruption. Now retries with exponential backoff.

Closes #123
```

**Breaking change (using `!`):**
```
feat(api)!: change configuration file format to YAML

BREAKING CHANGE: The config file is now YAML instead of JSON.
Run `plexcord migrate-config` to convert your existing config.
```

**Breaking change (using footer):**
```
refactor(config): restructure configuration schema

BREAKING CHANGE: Configuration keys have been renamed.
- `plex_server` is now `plex.server_url`
- `discord_id` is now `discord.client_id`
```

### Breaking Changes

To indicate a breaking change, you can either:

1. Add `!` after the type/scope: `feat(api)!: description`
2. Add a `BREAKING CHANGE:` footer in the commit body

Breaking changes trigger a major version bump (x.0.0).

### Rules

- **Type** must be lowercase
- **Scope** must be lowercase (if provided)
- **Description** must not be empty
- **Description** must not end with a period
- **Header** (type + scope + description) must be 100 characters or less

### Why This Matters

Conventional commits enable:
- Automated semantic versioning (version bumps based on commit types)
- Automated changelog generation
- Easier navigation through commit history
- Triggering CI/CD workflows based on change types

## Code Style

### Go Code

- Follow [Effective Go](https://go.dev/doc/effective_go)
- Run `gofmt` and `goimports`
- Use meaningful variable names
- Add comments for exported functions
- Handle errors explicitly
- Write table-driven tests

**Example:**
```go
// GetSession retrieves the current music session for the specified user.
// Returns nil if no music session is active.
func (c *Client) GetSession(userID string) (*Session, error) {
    if userID == "" {
        return nil, errors.New(errors.INVALID_INPUT, "userID cannot be empty")
    }
    
    // Implementation
}
```

### Frontend Code

- Follow Vue 3 Composition API patterns
- Use PrimeVue components
- Keep components focused and reusable
- Use Tailwind CSS for styling
- Run ESLint before committing

**Example:**
```vue
<script setup>
import { ref, computed, onMounted } from 'vue';
import Button from 'primevue/button';

const data = ref(null);
const isLoading = computed(() => data.value === null);

onMounted(async () => {
    data.value = await fetchData();
});
</script>
```

## Testing

### Go Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run specific package
go test ./internal/plex

# Run with race detector
go test -race ./...
```

### Writing Tests

- Use table-driven tests
- Test error cases
- Mock external dependencies
- Keep tests isolated and deterministic

**Example:**
```go
func TestGetSession(t *testing.T) {
    tests := []struct {
        name    string
        userID  string
        want    *Session
        wantErr bool
    }{
        {
            name:    "valid user",
            userID:  "12345",
            want:    &Session{/* ... */},
            wantErr: false,
        },
        {
            name:    "empty user ID",
            userID:  "",
            want:    nil,
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := GetSession(tt.userID)
            if (err != nil) != tt.wantErr {
                t.Errorf("GetSession() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            // Assert results
        })
    }
}
```

## Pull Request Process

1. **Update documentation** - README, comments, etc.
2. **Add tests** - For new features or bug fixes
3. **Run tests** - Ensure all tests pass
4. **Run linters** - Fix any linting issues
5. **Update CHANGELOG** - If applicable
6. **Fill out PR template** - Provide clear description
7. **Request review** - Tag maintainers if needed

### PR Checklist

Before submitting:

- [ ] Code follows project style guidelines
- [ ] Self-reviewed the code
- [ ] Commented complex code sections
- [ ] Updated relevant documentation
- [ ] Added/updated tests
- [ ] All tests pass locally
- [ ] No new warnings or errors
- [ ] Tested on target platform(s)

## Building

### Development Build

```bash
wails dev
```

### Production Build

```bash
# Current platform
wails build

# Specific platform
wails build -platform windows/amd64
wails build -platform darwin/universal
wails build -platform linux/amd64

# With version info
wails build -ldflags "-X plexcord/internal/version.Version=v1.0.0"
```

### Cross-Platform Building

Use GitHub Actions for official releases. See `.github/workflows/release.yml`.

## Project Structure

```
PlexCord/
├── internal/           # Go backend code
│   ├── config/        # Configuration management
│   ├── discord/       # Discord RPC integration
│   ├── errors/        # Error handling
│   ├── keychain/      # Secure credential storage
│   ├── plex/          # Plex API client
│   └── platform/      # Platform-specific code
├── frontend/          # Vue.js frontend
│   ├── src/
│   │   ├── components/ # Reusable components
│   │   ├── stores/    # Pinia state management
│   │   ├── views/     # Page components
│   │   └── router/    # Vue Router config
│   └── public/        # Static assets
├── docs/              # Documentation
└── .github/           # GitHub config and workflows
```

## Architecture

- **Backend**: Go with Wails framework
- **Frontend**: Vue 3 + PrimeVue + Tailwind CSS
- **State Management**: Pinia
- **Routing**: Vue Router
- **Build**: Wails (bundles Go + Vue into single binary)

## Need Help?

- **Questions**: Open a [Discussion](https://github.com/YOUR_USERNAME/PlexCord/discussions)
- **Bugs**: Open an [Issue](https://github.com/YOUR_USERNAME/PlexCord/issues)
- **Features**: Open a [Feature Request](https://github.com/YOUR_USERNAME/PlexCord/issues/new?template=feature_request.md)

## License

By contributing, you agree that your contributions will be licensed under the same license as the project (see LICENSE file).

---

Thank you for contributing to PlexCord! 🎵
