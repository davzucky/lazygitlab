# GitLab API Client

## pkg/gitlab/ Package

The gitlab package provides an abstraction layer over GitLab's Go SDK (gitlab.com/gitlab-org/api/client-go).

### API Patterns

- Use the `Client` interface for all GitLab interactions (enables mocking in tests)
- Call `client.Close()` when done to clean up resources
- Pagination is automatic for list endpoints (issues, merge requests) unless you specify a specific page
- All API errors are wrapped with context using `fmt.Errorf("context: %w", err)`

### Type Notes

- `ListProjectMergeRequests` returns `[]*gitlab.BasicMergeRequest`, not `[]*gitlab.MergeRequest`
- Pagination fields (Page, PerPage) are `int64`, not `int`
- String options need pointer to string (`&str`), not a helper function like `gitlab.String()`
- Issue IID (internal ID) is `int64`, used for getting single issues via `GetProjectIssue`

### Testing

- Use the `mockClient` struct in client_test.go as a template for testing code that uses the GitLab API
- The mock implements the same interface as the real client for easy swapping

# GUI Framework

## pkg/gui/ Package

The gui package provides the TUI (Terminal User Interface) using the bubbletea framework.

### Architecture

- The main layout consists of three panels: sidebar (navigation), main panel (list view), and details panel (preview)
- Use lipgloss for styling and layout management
- Keyboard navigation: `j`/`k` for list navigation, `1`/`2`/`3` for view switching, `q` to quit
- Status bar displays project context and connection status
- Help popup shows all keybindings (press `?` to open, `Esc` to close)

### Component Structure

- `styles.go`: Centralized styling with lipgloss (colors, borders, padding)
- `model.go`: Bubbletea Model implementation with Update and View methods
- ViewMode enum tracks current panel (ProjectsView, IssuesView, MergeRequestsView)

### Patterns

- Initialize styles once with `NewStyle()` constructor
- Render methods (`renderSidebar`, `renderMainPanel`, etc.) should return styled strings
- Always handle empty states in list views
- Use `tea.WindowSizeMsg` to handle terminal resizing

### Keyboard Navigation

- Use `tea.KeyMsg` type in Update method switch statement for keyboard input
- Key strings: `j`, `k`, `up`, `down`, `h`, `l`, `left`, `right`, `tab`, `shift+tab`, `enter`, `esc`, `q`, `?`, `ctrl+c`
- For popup/modals: add a state field (e.g., `showHelp bool`) and check it first in Update method
- Popup rendering: return early from View method if popup is showing, calling a dedicated render function
- View cycling: use modulo or bounds checking for Tab/Shift+Tab to wrap around views

# Project Context

## pkg/project/ Package

The project package handles automatic detection of GitLab project context from git remotes.

### Functionality

- Auto-detects project from git remote URL in current directory
- Supports both SSH (`git@gitlab.com:group/project.git`) and HTTPS (`https://gitlab.com/group/project.git`) formats
- Handles subgroups in project path (e.g., `group/subgroup/project`)
- Supports manual project override via command-line flag `--project`
- Validates project path format (must have at least 2 segments)

### Patterns

- Use `exec.Command("git", "remote", "-v")` to get git remote URLs
- Parse remote URLs using regex: look for `git@gitlab.com:` for SSH and `https://gitlab.com/` for HTTPS
- Extract project path by removing protocol, host, and `.git` suffix
- Validate project path by ensuring it has at least 2 non-empty segments
- Always provide graceful fallback for non-git directories or missing remotes

### Integration

- Import in main.go along with config and gui packages
- Call `project.DetectProjectPath(*projectFlag)` after config validation
- Pass detected project path to GUI model via `gui.NewModel(projectPath, connection)`
- Display project path in status bar for user awareness

# Utils & Logging

## pkg/utils/ Package

The utils package provides shared utility functions, including logging functionality for debugging.

### Logger Implementation

- Singleton pattern using `sync.Once` for one-time initialization
- Logs to `~/.local/share/lazygitlab/debug.log` (creates directory if needed)
- Thread-safe writes using mutex lock
- Three log levels: Debug, Info, Error (all use same format with prefix)
- Debug messages only logged when debug mode is enabled

### Patterns

- Initialize logger with `utils.InitLogger(debug bool)` in main.go
- Always defer `utils.Close()` after initialization to clean up resources
- Use format strings with placeholders: `Debug("message: %s", value)` not `Debug(message)`
- Log directory path: `filepath.Join(os.Getenv("HOME"), ".local", "share", "lazygitlab")`
- Use `os.MkdirAll(logDir, 0755)` to create nested directories
- Open log file with `os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)`
- Sync file after each write: `file.Sync()`

### Integration

- Add `--debug` flag to main.go for enabling verbose logging
- Import utils package in main.go and call `utils.InitLogger(*debugFlag)`
- Use `utils.Debug()`, `utils.Info()`, and `utils.Error()` throughout the codebase
- Log important events: config loading, validation, API calls, errors
- Debug mode helps diagnose issues in production without cluttering normal operation

# Loading and Error States

## GUI Model Patterns

The GUI model includes built-in support for loading indicators and error popups.

### Loading State

- Add `isLoading bool` field to Model struct
- Add `spinner spin.Model` field for loading animation
- Initialize spinner with `spin.New()` in `NewModel()`
- Set spinner style: `spinner.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("62"))`
- Return `m.spinner.Tick` from `Init()` method
- Handle `spin.TickMsg` in `Update()` method: `m.spinner, cmd = m.spinner.Update(msg)`
- Show spinner in View: `m.spinner.View() + " Loading..."`
- Add `SetLoading(bool)` helper method to control loading state

### Error Popup

- Add `showError bool` and `errorMessage string` fields to Model struct
- Check `showError` first in both `Update()` and `View()` methods
- In `View()`: return `m.renderErrorPopup()` if showing, before other popups
- In `Update()`: handle popup-specific keys (e.g., `r` for retry, `q`/`esc` to close)
- Create `renderErrorPopup()` method with styled error message
- Error popup styling: use red border (`lipgloss.Color("196")`) and dark red background
- Add helper methods: `SetError(message string)`, `ClearError()`
