# Contributing to PlexCord

Thank you for considering contributing to PlexCord! This document provides guidelines for contributing to the project.

## Code of Conduct

- Be respectful and inclusive
- Focus on constructive feedback
- Help create a welcoming environment for all contributors

## How Can I Contribute?

### Reporting Bugs

Before creating a bug report, please check existing issues to avoid duplicates.

**When reporting a bug, include:**

- PlexCord version
- Operating system and version
- Steps to reproduce the issue
- Expected vs actual behavior
- Error messages or screenshots
- Relevant log output

**Use this template:**

```markdown
**Environment:**
- PlexCord version: v1.0.0
- OS: Windows 11 / macOS 13 / Ubuntu 22.04
- Discord version: 1.0.9000
- Plex version: 1.40.0

**Steps to reproduce:**
1. Launch PlexCord
2. Start playing music in Plexamp
3. ...

**Expected behavior:**
Discord status should update

**Actual behavior:**
Discord status remains blank

**Error messages:**
[Paste any error messages]

**Additional context:**
[Any other relevant information]
```

### Suggesting Features

Feature requests are welcome! Please provide:

- Clear description of the feature
- Use case / problem it solves
- Mockups or examples (if applicable)
- Willingness to help implement

### Pull Requests

1. **Fork the repository** and create a branch from `main`
2. **Make your changes** following the coding guidelines
3. **Test your changes** thoroughly
4. **Update documentation** if needed
5. **Submit a pull request** with clear description

## Development Setup

See the [Development Guide](./development.md) for detailed setup instructions.

**Quick start:**

```bash
# Clone your fork
git clone https://github.com/yourusername/PlexCord.git
cd PlexCord

# Install dependencies
go mod download
cd frontend && npm install && cd ..

# Run in dev mode
wails dev
```

## Coding Guidelines

### General Rules

- Write clear, maintainable, and well-documented code
- Follow the existing project structure and naming conventions
- Do not generate documentation files unless explicitly needed
- Do not generate tests unless explicitly needed

### Backend (Go)

#### Style

- Follow [Effective Go](https://go.dev/doc/effective_go) guidelines
- Use `gofmt` for formatting
- Run `go vet` before committing
- Keep functions small and focused (< 50 lines ideal)
- Use meaningful variable names

#### Package Structure

```go
// Good: Clear interface with implementation
type PlexClient interface {
    GetSessions() ([]Session, error)
}

type plexClientImpl struct {
    baseURL string
    token   string
}

func NewPlexClient(baseURL, token string) PlexClient {
    return &plexClientImpl{baseURL, token}
}
```

#### Error Handling

```go
// Good: Return structured errors with context
if err != nil {
    return nil, fmt.Errorf("failed to fetch sessions: %w", err)
}

// Good: Use error codes for frontend
return errors.NewError(errors.ErrPlexUnreachable, "Server not responding")
```

#### Comments

```go
// Good: Document exported functions
// GetSessions retrieves all active playback sessions from the Plex server.
// Returns an error if the server is unreachable or authentication fails.
func GetSessions() ([]Session, error) {
    // ...
}
```

### Frontend (Vue.js)

#### Style

- Use Vue 3 Composition API
- Use `<script setup>` syntax
- Follow PrimeVue component patterns
- Use TailwindCSS utility classes
- Respect separation of concerns (logic vs presentation)

#### Component Structure

```vue
<template>
  <div class="component-name">
    <!-- Clear, semantic markup -->
  </div>
</template>

<script setup>
// Imports
import { ref, computed } from 'vue'

// Props
const props = defineProps({
  title: String
})

// State
const isActive = ref(false)

// Computed
const displayTitle = computed(() => props.title.toUpperCase())

// Methods
function handleClick() {
  isActive.value = !isActive.value
}
</script>

<style scoped>
/* Component-specific styles */
</style>
```

#### Store Pattern

```javascript
import { defineStore } from 'pinia'

export const useMyStore = defineStore('mystore', {
  state: () => ({
    items: [],
    loading: false,
    error: null
  }),

  getters: {
    itemCount: (state) => state.items.length
  },

  actions: {
    async fetchItems() {
      this.loading = true
      this.error = null
      try {
        this.items = await GetItems()
      } catch (error) {
        this.error = error.message
      } finally {
        this.loading = false
      }
    }
  }
})
```

#### Security

- Ensure all user inputs are sanitized to prevent XSS attacks
- Never log sensitive information (tokens, credentials)
- Validate data received from backend

### Database (MySQL)

If you're working on future features that involve databases:

- Follow normalization principles
- Use proper indexing for frequently queried fields
- Use transactions for multi-step operations
- Write efficient queries

## Testing

### Backend Tests

```go
func TestPlexClient_GetSessions(t *testing.T) {
    // Arrange
    client := NewPlexClient("http://localhost:32400", "test-token")

    // Act
    sessions, err := client.GetSessions()

    // Assert
    if err != nil {
        t.Fatalf("expected no error, got %v", err)
    }
    if len(sessions) == 0 {
        t.Error("expected sessions, got empty list")
    }
}
```

### Frontend Tests (Coming Soon)

```javascript
import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import MyComponent from './MyComponent.vue'

describe('MyComponent', () => {
  it('renders correctly', () => {
    const wrapper = mount(MyComponent, {
      props: { title: 'Test' }
    })
    expect(wrapper.text()).toContain('Test')
  })
})
```

## Commit Messages

Write clear, descriptive commit messages:

```
type(scope): brief description

Longer explanation if needed.

Fixes #123
```

**Types:**

- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Formatting, no code change
- `refactor`: Code restructuring
- `test`: Adding tests
- `chore`: Build process, dependencies

**Examples:**

```
feat(discord): add reconnection on client restart

fix(plex): handle empty session response correctly

docs(readme): update installation instructions

refactor(config): simplify paths logic
```

## Pull Request Process

### Before Submitting

1. **Update your branch** from `main`:

   ```bash
   git checkout main
   git pull upstream main
   git checkout your-branch
   git rebase main
   ```

2. **Test your changes:**

   ```bash
   # Backend tests
   go test ./...

   # Build for your platform
   wails build
   ```

3. **Check code quality:**

   ```bash
   # Go formatting
   go fmt ./...

   # Go linting
   go vet ./...
   ```

### Pull Request Template

```markdown
## Description

Brief description of the changes

## Type of Change

- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing

- [ ] Tested on Windows
- [ ] Tested on macOS
- [ ] Tested on Linux
- [ ] Added unit tests
- [ ] Manual testing performed

## Checklist

- [ ] Code follows project style guidelines
- [ ] Self-review completed
- [ ] Comments added for complex logic
- [ ] Documentation updated
- [ ] No new warnings generated

## Related Issues

Fixes #123
```

### Review Process

1. A maintainer will review your PR
2. Address any feedback or requested changes
3. Once approved, a maintainer will merge your PR
4. Your contribution will be included in the next release

## Getting Help

- **Questions?** Open a GitHub Discussion
- **Found a bug?** Open a GitHub Issue
- **Want to chat?** Join our Discord server (coming soon)

## Recognition

Contributors are recognized in:

- `CONTRIBUTORS.md` file
- Release notes
- Project README

## License

By contributing, you agree that your contributions will be licensed under the same license as the project (see [LICENSE](../LICENSE)).

## Thank You!

Your contributions make PlexCord better for everyone. We appreciate your time and effort! 🎵
