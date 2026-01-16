# PRD: LazyGitLab - Part 1: Foundation & Core Architecture

## Introduction

LazyGitLab is a terminal user interface (TUI) application for managing GitLab operations, inspired by [lazygit](https://github.com/jesseduffield/lazygit) and [lazydocker](https://github.com/jesseduffield/lazydocker). The goal is to simplify and streamline GitLab workflows by providing a fast, keyboard-driven interface for managing issues, merge requests, and code reviews directly from the command line.

This PRD covers the foundational architecture, project setup, and core components that will be used across all features.

## Goals

- Create a Go-based TUI application using the same patterns as lazygit/lazydocker
- Support both gitlab.com and self-hosted GitLab instances
- Implement flexible authentication (glab config, custom config, environment variables)
- Establish a modular architecture that supports future feature expansion
- Provide a responsive, keyboard-driven interface with optional mouse support

## User Stories

### US-001: Initialize Go Project with TUI Framework
**Description:** As a developer, I need the project scaffolding with proper Go module setup and TUI library integration so that I can build the application.

**Acceptance Criteria:**
- [ ] Go module initialized with `github.com/lazygitlab/lazygitlab` (or appropriate path)
- [ ] [gocui](https://github.com/jroimartin/gocui) or [tcell](https://github.com/gdamore/tcell)/[tview](https://github.com/rivo/tview) library integrated
- [ ] Basic main.go with application entry point
- [ ] Makefile with `build`, `run`, `test`, and `lint` targets
- [ ] `.gitignore` configured for Go projects
- [ ] Project compiles and runs showing empty TUI window
- [ ] Typecheck/lint passes

### US-002: Implement Configuration Management
**Description:** As a user, I want lazygitlab to read my GitLab configuration so that I don't have to reconfigure authentication.

**Acceptance Criteria:**
- [ ] Read glab CLI config from `~/.config/glab-cli/config.yml` if it exists
- [ ] Support custom config file at `~/.config/lazygitlab/config.yml`
- [ ] Support `GITLAB_TOKEN` and `GITLAB_HOST` environment variables
- [ ] Config precedence: env vars > lazygitlab config > glab config
- [ ] Configuration struct with GitLab host URL and access token
- [ ] Validate token on startup with API call
- [ ] Clear error message when no valid configuration found
- [ ] Typecheck/lint passes

### US-003: Create GitLab API Client
**Description:** As a developer, I need a GitLab API client abstraction so that all features can interact with GitLab consistently.

**Acceptance Criteria:**
- [ ] Wrapper around [go-gitlab](https://github.com/xanzy/go-gitlab) library
- [ ] Support for custom GitLab host URLs (self-hosted instances)
- [ ] Automatic pagination handling for list endpoints
- [ ] Rate limiting awareness and retry logic
- [ ] Error handling with user-friendly messages
- [ ] Interface-based design for testing/mocking
- [ ] Unit tests for client initialization
- [ ] Typecheck/lint passes

### US-004: Build Main Application Layout
**Description:** As a user, I want to see a structured TUI layout so that I can navigate between different GitLab entities.

**Acceptance Criteria:**
- [ ] Multi-panel layout similar to lazygit (sidebar + main content + details)
- [ ] Left sidebar with navigation options (Projects, Issues, Merge Requests)
- [ ] Main panel showing list of selected entity type
- [ ] Bottom panel showing details/preview of selected item
- [ ] Status bar showing current project/context and connection status
- [ ] Title bar showing application name and version
- [ ] Typecheck/lint passes

### US-005: Implement Keyboard Navigation System
**Description:** As a user, I want to navigate the TUI using keyboard shortcuts so that I can work efficiently without a mouse.

**Acceptance Criteria:**
- [ ] `j`/`k` or arrow keys for up/down navigation in lists
- [ ] `h`/`l` or arrow keys for panel switching
- [ ] `Enter` to select/open item
- [ ] `Esc` to go back/close popup
- [ ] `q` to quit application
- [ ] `?` to show help/keybindings popup
- [ ] Tab/Shift+Tab for panel cycling
- [ ] Keybindings displayed in status bar contextually
- [ ] Typecheck/lint passes

### US-006: Implement Project Context Management
**Description:** As a user, I want lazygitlab to automatically detect my current project context so that I see relevant issues and MRs.

**Acceptance Criteria:**
- [ ] Auto-detect GitLab project from git remote URL in current directory
- [ ] Support for SSH and HTTPS remote URL formats
- [ ] Support for subgroups in project path (e.g., `group/subgroup/project`)
- [ ] Manual project selection via command line flag `--project`
- [ ] Display current project in status bar
- [ ] Graceful handling when not in a git repository
- [ ] Typecheck/lint passes

### US-007: Create Loading and Error States
**Description:** As a user, I want to see loading indicators and clear error messages so that I know what's happening.

**Acceptance Criteria:**
- [ ] Loading spinner/indicator when fetching data from API
- [ ] Error popup for API failures with retry option
- [ ] Network connectivity check on startup
- [ ] Graceful degradation when certain endpoints fail
- [ ] Log file for debugging (`~/.local/share/lazygitlab/debug.log`)
- [ ] `--debug` flag to enable verbose logging
- [ ] Typecheck/lint passes

### US-008: Implement First-Run Setup Flow
**Description:** As a new user, I want to be guided through initial setup so that I can start using lazygitlab quickly.

**Acceptance Criteria:**
- [ ] Detect if no valid configuration exists on startup
- [ ] Interactive prompt to enter GitLab host URL
- [ ] Interactive prompt to enter personal access token
- [ ] Link to GitLab documentation for creating tokens
- [ ] Validate credentials before saving
- [ ] Save configuration to `~/.config/lazygitlab/config.yml`
- [ ] Option to save token to system keyring (optional, not MVP)
- [ ] Typecheck/lint passes

## Functional Requirements

- FR-1: Application must be a single statically-linked binary for easy distribution
- FR-2: Must support gitlab.com and any self-hosted GitLab instance (v12.0+)
- FR-3: Must handle GitLab API rate limiting gracefully
- FR-4: Configuration file must use YAML format for consistency with glab
- FR-5: All keyboard shortcuts must be configurable via config file
- FR-6: Application must start and show UI within 2 seconds on typical hardware
- FR-7: Must work in terminals with minimum 80x24 character size
- FR-8: Must support 256-color and true-color terminals

## Non-Goals

- OAuth authentication flow (out of scope for MVP)
- GitLab CI/CD pipeline management
- Repository browsing/file viewing
- Wiki management
- Snippets management
- Package registry management
- Multiple simultaneous project contexts
- Offline mode/caching

## Technical Considerations

### Dependencies
- Go 1.21+ for generics support and modern features
- [go-gitlab](https://github.com/xanzy/go-gitlab) - Official GitLab API client
- [gocui](https://github.com/jroimartin/gocui) or [bubbletea](https://github.com/charmbracelet/bubbletea) - TUI framework (recommend bubbletea for modern approach)
- [viper](https://github.com/spf13/viper) - Configuration management
- [cobra](https://github.com/spf13/cobra) - CLI argument parsing

### Project Structure (following lazygit patterns)
```
lazygitlab/
├── cmd/
│   └── lazygitlab/
│       └── main.go
├── pkg/
│   ├── app/           # Main application logic
│   ├── config/        # Configuration management
│   ├── gitlab/        # GitLab API client wrapper
│   ├── gui/           # TUI components and views
│   │   ├── views/     # Individual view implementations
│   │   ├── widgets/   # Reusable UI widgets
│   │   └── keybindings/
│   ├── commands/      # Command handlers
│   └── utils/         # Shared utilities
├── docs/
│   └── keybindings/
├── scripts/
├── test/
├── Makefile
├── go.mod
├── go.sum
└── README.md
```

### Configuration File Format
```yaml
# ~/.config/lazygitlab/config.yml
gitlab:
  host: https://gitlab.com  # or self-hosted URL
  token: glpat-xxxxxxxxxxxx

gui:
  theme: default
  mouse_enabled: true
  
keybindings:
  quit: q
  help: "?"
  # ... customizable keybindings
```

## Success Metrics

- Application launches successfully with valid configuration
- Can connect to GitLab API and fetch user information
- TUI renders correctly in common terminal emulators (iTerm2, Alacritty, kitty, gnome-terminal)
- Keyboard navigation works intuitively between panels
- Startup time under 2 seconds

## Open Questions

- Should we use bubbletea (modern, Elm-architecture) or gocui (what lazygit uses)?
  - Recommendation: Start with bubbletea for modern patterns, but gocui is proven in lazygit
- Should we support the `glab` binary for operations or implement everything via API?
  - Recommendation: Pure API implementation for control and consistency
- What minimum GitLab version should we support?
  - Recommendation: GitLab 12.0+ for reasonable API compatibility
