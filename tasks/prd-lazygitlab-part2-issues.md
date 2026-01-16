# PRD: LazyGitLab - Part 2: Issue Management

## Introduction

This PRD covers the issue management functionality for LazyGitLab. Issues are a core GitLab feature for tracking work, bugs, and feature requests. This module will allow users to view, create, update, and manage issues directly from the terminal.

**Dependency:** This PRD requires completion of Part 1 (Foundation) user stories US-001 through US-007.

## Goals

- View and browse project issues with filtering and sorting
- Create new issues with labels and assignees
- Update existing issues (title, description, status, labels)
- Quick actions for common operations (close, reopen, assign)
- Search issues across the project

## User Stories

### US-101: Display Issues List View
**Description:** As a user, I want to see a list of issues in the current project so that I can browse and select issues to work with.

**Acceptance Criteria:**
- [ ] Issues panel shows list of issues for current project
- [ ] Each issue row displays: ID, title (truncated), state (open/closed icon), labels (as colored badges)
- [ ] Issues sorted by updated_at descending by default
- [ ] Pagination support for projects with many issues (load more on scroll)
- [ ] Visual indicator for issues assigned to current user
- [ ] Empty state message when no issues exist
- [ ] Loading indicator while fetching issues
- [ ] Typecheck/lint passes

### US-102: Issue Detail View
**Description:** As a user, I want to view full issue details so that I can understand the issue context.

**Acceptance Criteria:**
- [ ] Pressing `Enter` on an issue opens detail view in main panel
- [ ] Detail view shows: title, description (markdown rendered as plain text), author, created date, updated date
- [ ] Shows assignees with usernames
- [ ] Shows labels with colors
- [ ] Shows milestone if assigned
- [ ] Shows time tracking info if present (estimate, spent)
- [ ] Shows related merge requests if any
- [ ] Shows issue weight if set
- [ ] Scroll support for long descriptions
- [ ] `Esc` returns to list view
- [ ] Typecheck/lint passes

### US-103: Filter Issues by State
**Description:** As a user, I want to filter issues by state so that I can focus on relevant issues.

**Acceptance Criteria:**
- [ ] Filter dropdown/toggle for: All, Open, Closed
- [ ] Press `f` to open filter menu
- [ ] Current filter displayed in panel header
- [ ] Filter persists during session
- [ ] Issue count updates based on filter
- [ ] Keyboard shortcut `1` for Open, `2` for Closed, `3` for All
- [ ] Typecheck/lint passes

### US-104: Filter Issues by Labels
**Description:** As a user, I want to filter issues by labels so that I can focus on specific categories.

**Acceptance Criteria:**
- [ ] Press `l` to open label filter popup
- [ ] Shows list of all project labels with colors
- [ ] Multi-select support (filter by multiple labels - AND logic)
- [ ] Clear label filter option
- [ ] Selected labels shown in panel header
- [ ] Fuzzy search within label list
- [ ] Typecheck/lint passes

### US-105: Filter Issues by Assignee
**Description:** As a user, I want to filter issues by assignee so that I can see my assigned work.

**Acceptance Criteria:**
- [ ] Press `a` to open assignee filter popup
- [ ] Quick option "Assigned to me"
- [ ] Quick option "Unassigned"
- [ ] Shows list of project members
- [ ] Fuzzy search for assignee name
- [ ] Current assignee filter shown in panel header
- [ ] Typecheck/lint passes

### US-106: Search Issues
**Description:** As a user, I want to search issues by text so that I can find specific issues quickly.

**Acceptance Criteria:**
- [ ] Press `/` to open search input
- [ ] Search queries issue titles and descriptions
- [ ] Search results update list in real-time (debounced)
- [ ] Clear search with `Esc` or empty input
- [ ] Search term highlighted in results
- [ ] Minimum 2 characters to trigger search
- [ ] Typecheck/lint passes

### US-107: Create New Issue
**Description:** As a user, I want to create new issues so that I can track new work items.

**Acceptance Criteria:**
- [ ] Press `c` to open create issue form
- [ ] Form fields: title (required), description (optional, multiline)
- [ ] Label selector (optional, multi-select)
- [ ] Assignee selector (optional)
- [ ] Milestone selector (optional)
- [ ] Submit with `Ctrl+Enter` or submit button
- [ ] Cancel with `Esc`
- [ ] Validation: title required, minimum 3 characters
- [ ] Success message and refresh list on creation
- [ ] Error handling for API failures
- [ ] Typecheck/lint passes

### US-108: Edit Issue
**Description:** As a user, I want to edit existing issues so that I can update information.

**Acceptance Criteria:**
- [ ] Press `e` on selected issue to open edit form
- [ ] Pre-populated form with current values
- [ ] Editable fields: title, description, labels, assignees, milestone
- [ ] Same form layout as create
- [ ] Shows "Editing Issue #123" in form header
- [ ] Submit saves changes via API
- [ ] Success message and refresh on save
- [ ] Typecheck/lint passes

### US-109: Quick Actions on Issues
**Description:** As a user, I want quick keyboard shortcuts for common actions so that I can work efficiently.

**Acceptance Criteria:**
- [ ] Press `o` to close an open issue (with confirmation)
- [ ] Press `o` to reopen a closed issue (with confirmation)
- [ ] Press `m` to assign issue to self
- [ ] Press `M` (shift+m) to unassign from self
- [ ] Press `y` to copy issue URL to clipboard
- [ ] Press `b` to open issue in default browser
- [ ] Confirmation popup for destructive actions
- [ ] Typecheck/lint passes

### US-110: View Issue Comments
**Description:** As a user, I want to view issue comments so that I can see the discussion.

**Acceptance Criteria:**
- [ ] In issue detail view, press `Tab` to switch to comments section
- [ ] Comments displayed chronologically
- [ ] Each comment shows: author, date, content
- [ ] System notes (label changes, assignments) shown differently
- [ ] Pagination for issues with many comments
- [ ] Typecheck/lint passes

### US-111: Add Issue Comment
**Description:** As a user, I want to add comments to issues so that I can participate in discussions.

**Acceptance Criteria:**
- [ ] In issue detail/comments view, press `r` to reply
- [ ] Multiline text input for comment
- [ ] Support for basic markdown (rendered as plain text)
- [ ] Submit with `Ctrl+Enter`
- [ ] Cancel with `Esc`
- [ ] Comment appears in list after submission
- [ ] Typecheck/lint passes

### US-112: Sort Issues
**Description:** As a user, I want to sort issues by different criteria so that I can organize my view.

**Acceptance Criteria:**
- [ ] Press `s` to open sort menu
- [ ] Sort options: Created date, Updated date, Priority, Due date, Title
- [ ] Each option can be ascending or descending
- [ ] Current sort shown in panel header
- [ ] Sort persists during session
- [ ] Typecheck/lint passes

## Functional Requirements

- FR-101: Issue list must load within 3 seconds for projects with up to 1000 issues
- FR-102: Must support GitLab issue weight feature
- FR-103: Must handle confidential issues appropriately (show indicator)
- FR-104: Must preserve issue state/position when returning from detail view
- FR-105: All filters must be combinable (state + labels + assignee + search)
- FR-106: Must support issues in group-level views (future enhancement noted)

## Non-Goals

- Issue boards/kanban view
- Issue templates
- Issue due date management (beyond display)
- Issue time tracking entry
- Issue relations/dependencies management
- Issue move between projects
- Bulk operations on multiple issues

## Technical Considerations

### GitLab API Endpoints Used
- `GET /projects/:id/issues` - List issues
- `GET /projects/:id/issues/:iid` - Get single issue
- `POST /projects/:id/issues` - Create issue
- `PUT /projects/:id/issues/:iid` - Update issue
- `GET /projects/:id/issues/:iid/notes` - List comments
- `POST /projects/:id/issues/:iid/notes` - Add comment
- `GET /projects/:id/labels` - List labels for filtering
- `GET /projects/:id/members` - List members for assignee filtering

### Data Structures
```go
type Issue struct {
    IID         int
    Title       string
    Description string
    State       string      // "opened" or "closed"
    Author      User
    Assignees   []User
    Labels      []string
    Milestone   *Milestone
    CreatedAt   time.Time
    UpdatedAt   time.Time
    WebURL      string
    Confidential bool
    Weight      *int
}

type IssueFilter struct {
    State     string   // "opened", "closed", "all"
    Labels    []string
    Assignee  *int     // User ID or nil
    Search    string
    Sort      string
    Order     string   // "asc" or "desc"
}
```

### UI Layout for Issues
```
┌─────────────────────────────────────────────────────────────────┐
│ LazyGitLab - mygroup/myproject                      Connected   │
├──────────────┬──────────────────────────────────────────────────┤
│              │ Issues (42 open) [Filter: Open] [Sort: Updated]  │
│  Projects    │──────────────────────────────────────────────────│
│  > Issues    │ #123 Fix login bug                    [bug] [P1] │
│    MRs       │ #122 Add user profile page        [feature] [P2] │
│              │ #121 Update documentation                  [docs]│
│              │ #120 Refactor auth module           [refactor]   │
│              │ ...                                              │
├──────────────┼──────────────────────────────────────────────────┤
│              │ #123: Fix login bug                              │
│  Labels      │ Author: @john  |  Created: 2 days ago            │
│  Members     │ Assignees: @jane, @bob                           │
│              │ Labels: bug, priority::1                         │
│              │                                                  │
│              │ The login form fails when...                     │
├──────────────┴──────────────────────────────────────────────────┤
│ j/k:navigate  Enter:view  c:create  e:edit  o:close  ?:help     │
└─────────────────────────────────────────────────────────────────┘
```

## Success Metrics

- User can view all project issues within 3 seconds
- User can create a new issue in under 30 seconds
- User can find a specific issue using search in under 10 seconds
- All issue CRUD operations complete successfully via API

## Open Questions

- Should we support GitLab's quick actions in comments (e.g., `/close`, `/assign`)?
  - Recommendation: Not for MVP, but nice to have
- Should we show related issues/epics?
  - Recommendation: Display only, no management for MVP
- How should we handle issues with very long descriptions?
  - Recommendation: Truncate with "show more" option
