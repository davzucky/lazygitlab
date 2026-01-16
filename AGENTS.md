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
