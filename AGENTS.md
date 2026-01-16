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
