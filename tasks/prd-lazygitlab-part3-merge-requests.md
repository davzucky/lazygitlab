# PRD: LazyGitLab - Part 3: Merge Request Management & Code Review

## Introduction

This PRD covers merge request (MR) management and code review functionality for LazyGitLab. This is a primary feature of the application, allowing users to view MRs, review code diffs, and add comments directly from the terminal - the key differentiator that makes lazygitlab valuable for developers who prefer command-line workflows.

**Dependency:** This PRD requires completion of Part 1 (Foundation) user stories US-001 through US-007.

## Goals

- View and browse merge requests with filtering and sorting
- View MR diffs with syntax highlighting
- Add line comments and general comments to MRs
- Approve and request changes on MRs
- Quick actions for common MR operations
- Support efficient code review workflow entirely from the terminal

## User Stories

### US-201: Display Merge Requests List View
**Description:** As a user, I want to see a list of merge requests so that I can browse and select MRs to review.

**Acceptance Criteria:**
- [ ] MR panel shows list of merge requests for current project
- [ ] Each row displays: MR ID (!123), title (truncated), state icon, author, source->target branches
- [ ] Shows approval status indicator (approved, needs approval, changes requested)
- [ ] Shows CI/CD pipeline status (passed, failed, running, pending)
- [ ] MRs sorted by updated_at descending by default
- [ ] Visual indicator for MRs where user is reviewer/assignee
- [ ] Visual indicator for draft/WIP MRs
- [ ] Pagination support for many MRs
- [ ] Empty state message when no MRs exist
- [ ] Typecheck/lint passes

### US-202: MR Detail View
**Description:** As a user, I want to view full MR details so that I understand the changes being proposed.

**Acceptance Criteria:**
- [ ] Pressing `Enter` on MR opens detail view
- [ ] Shows: title, description, author, created/updated dates
- [ ] Shows source branch -> target branch with commit count
- [ ] Shows reviewers and their approval status
- [ ] Shows assignees
- [ ] Shows labels
- [ ] Shows milestone if assigned
- [ ] Shows pipeline status with link
- [ ] Shows merge status (can merge, has conflicts, needs rebase)
- [ ] Shows changed files count (+additions, -deletions)
- [ ] `Esc` returns to list view
- [ ] Typecheck/lint passes

### US-203: Filter MRs by State
**Description:** As a user, I want to filter MRs by state so that I can focus on relevant MRs.

**Acceptance Criteria:**
- [ ] Filter options: All, Open, Merged, Closed
- [ ] Press `f` to open filter menu
- [ ] Current filter shown in panel header
- [ ] Keyboard shortcuts: `1` Open, `2` Merged, `3` Closed, `4` All
- [ ] Filter persists during session
- [ ] Typecheck/lint passes

### US-204: Filter MRs by Reviewer/Author
**Description:** As a user, I want to filter MRs by reviewer or author so that I can find MRs relevant to me.

**Acceptance Criteria:**
- [ ] Press `r` to filter by reviewer
- [ ] Quick option "Assigned to me as reviewer"
- [ ] Press `A` (shift+a) to filter by author
- [ ] Quick option "Authored by me"
- [ ] Shows project members list for selection
- [ ] Fuzzy search for names
- [ ] Clear filter option
- [ ] Typecheck/lint passes

### US-205: View MR Changed Files List
**Description:** As a user, I want to see the list of files changed in an MR so that I can navigate to specific files.

**Acceptance Criteria:**
- [ ] In MR detail view, press `Tab` to switch to files panel
- [ ] Shows list of all changed files
- [ ] Each file shows: filename, change type (added/modified/deleted/renamed), +/- line counts
- [ ] Files grouped by directory (collapsible)
- [ ] Total additions/deletions shown in header
- [ ] Indicator for files with existing comments
- [ ] Navigate files with j/k
- [ ] Typecheck/lint passes

### US-206: View File Diff
**Description:** As a user, I want to view the diff for a specific file so that I can review the code changes.

**Acceptance Criteria:**
- [ ] Press `Enter` on a file to view its diff
- [ ] Side-by-side or unified diff view (configurable)
- [ ] Syntax highlighting based on file extension
- [ ] Line numbers displayed for both old and new versions
- [ ] Added lines highlighted in green
- [ ] Removed lines highlighted in red
- [ ] Context lines (unchanged) shown around changes
- [ ] Scroll through diff with j/k
- [ ] Jump between hunks with `n` (next) and `N` (previous)
- [ ] Expand/collapse context lines
- [ ] `Esc` returns to file list
- [ ] Typecheck/lint passes

### US-207: Add Line Comment to Diff
**Description:** As a user, I want to add comments on specific lines so that I can provide targeted code review feedback.

**Acceptance Criteria:**
- [ ] In diff view, press `c` on a line to add comment
- [ ] Comment input appears below the selected line
- [ ] Multiline text input for comment
- [ ] Option to start a new thread or reply to existing
- [ ] Submit with `Ctrl+Enter`
- [ ] Cancel with `Esc`
- [ ] Comment appears inline after submission
- [ ] Line comment includes line number and file context
- [ ] Typecheck/lint passes

### US-208: View Existing Diff Comments
**Description:** As a user, I want to see existing comments in the diff so that I can follow the review discussion.

**Acceptance Criteria:**
- [ ] Existing comments shown inline in diff view
- [ ] Comments show: author, date, content
- [ ] Threaded replies shown nested under parent
- [ ] Resolved threads shown collapsed (expandable)
- [ ] Unresolved thread count shown in file list
- [ ] Navigate to next comment with `]`, previous with `[`
- [ ] Typecheck/lint passes

### US-209: Reply to Diff Comment
**Description:** As a user, I want to reply to existing comments so that I can participate in code review discussions.

**Acceptance Criteria:**
- [ ] When viewing a comment thread, press `r` to reply
- [ ] Reply input appears below the thread
- [ ] Submit with `Ctrl+Enter`
- [ ] Reply appears in thread after submission
- [ ] Typecheck/lint passes

### US-210: Resolve/Unresolve Discussion Thread
**Description:** As a user, I want to resolve discussions so that I can track what's been addressed.

**Acceptance Criteria:**
- [ ] Press `R` (shift+r) on a thread to toggle resolved status
- [ ] Resolved threads visually distinct (grayed out, collapsed)
- [ ] Shows "Resolved by @user" when resolved
- [ ] Can unresolve by pressing `R` again
- [ ] Only thread participants and project maintainers can resolve
- [ ] Typecheck/lint passes

### US-211: Add General MR Comment
**Description:** As a user, I want to add general comments to MRs so that I can provide overall feedback.

**Acceptance Criteria:**
- [ ] In MR detail view, press `C` (shift+c) to add general comment
- [ ] Comment not tied to specific line/file
- [ ] Appears in MR discussion/activity feed
- [ ] Same input UI as line comments
- [ ] Typecheck/lint passes

### US-212: Approve MR
**Description:** As a user, I want to approve an MR so that I can indicate the code is ready to merge.

**Acceptance Criteria:**
- [ ] Press `A` in MR detail/diff view to approve
- [ ] Confirmation popup showing approval action
- [ ] Option to add approval comment
- [ ] Approval reflected immediately in MR status
- [ ] Only available when user hasn't already approved
- [ ] Typecheck/lint passes

### US-213: Request Changes on MR
**Description:** As a user, I want to request changes on an MR so that I can indicate issues need addressing.

**Acceptance Criteria:**
- [ ] Press `X` to request changes
- [ ] Must provide a comment explaining requested changes
- [ ] Status changes to "Changes requested"
- [ ] Typecheck/lint passes

### US-214: Unapprove MR
**Description:** As a user, I want to remove my approval so that I can change my review decision.

**Acceptance Criteria:**
- [ ] Press `U` to unapprove (when already approved)
- [ ] Confirmation popup
- [ ] Removes user from approved reviewers list
- [ ] Only available when user has already approved
- [ ] Typecheck/lint passes

### US-215: Quick Actions on MRs
**Description:** As a user, I want quick keyboard shortcuts for common MR actions so that I can work efficiently.

**Acceptance Criteria:**
- [ ] Press `y` to copy MR URL to clipboard
- [ ] Press `b` to open MR in default browser
- [ ] Press `g` to view CI/CD pipeline status detail
- [ ] Press `m` to merge MR (if conditions met, with confirmation)
- [ ] Press `w` to toggle WIP/Draft status
- [ ] Press `L` to manage labels
- [ ] Typecheck/lint passes

### US-216: View MR Commits
**Description:** As a user, I want to view the commits in an MR so that I can understand the change history.

**Acceptance Criteria:**
- [ ] In MR detail, panel/tab for commits
- [ ] Shows commit list: hash (short), message, author, date
- [ ] Press `Enter` on commit to see its individual diff
- [ ] Shows total commit count
- [ ] Typecheck/lint passes

### US-217: Filter MRs by Labels
**Description:** As a user, I want to filter MRs by labels so that I can focus on specific categories.

**Acceptance Criteria:**
- [ ] Press `l` to open label filter
- [ ] Multi-select labels
- [ ] Same UI as issue label filtering
- [ ] Typecheck/lint passes

### US-218: Search MRs
**Description:** As a user, I want to search MRs by text so that I can find specific MRs quickly.

**Acceptance Criteria:**
- [ ] Press `/` to open search
- [ ] Searches MR titles and descriptions
- [ ] Same behavior as issue search
- [ ] Typecheck/lint passes

### US-219: View MR Pipelines
**Description:** As a user, I want to see pipeline status details so that I can understand CI/CD results.

**Acceptance Criteria:**
- [ ] Press `p` to open pipelines panel
- [ ] Shows list of pipelines for the MR
- [ ] Each pipeline: status, created time, duration
- [ ] Shows jobs within pipeline (expandable)
- [ ] Job status indicators (passed, failed, running, pending, skipped)
- [ ] Press `Enter` on job to view job log (truncated, last N lines)
- [ ] Option to open full log in browser
- [ ] Typecheck/lint passes

### US-220: Checkout MR Branch Locally
**Description:** As a user, I want to checkout the MR branch so that I can test changes locally.

**Acceptance Criteria:**
- [ ] Press `o` to checkout MR source branch
- [ ] Runs git fetch and checkout commands
- [ ] Works for MRs from forks (fetches from fork remote)
- [ ] Shows confirmation with branch name
- [ ] Error handling for conflicts/dirty working tree
- [ ] Typecheck/lint passes

## Functional Requirements

- FR-201: Diff view must handle files up to 10,000 lines
- FR-202: Syntax highlighting must support common languages (Go, Python, JS/TS, Ruby, Java, C/C++, Rust, YAML, JSON, Markdown)
- FR-203: Must support MRs from forked projects
- FR-204: Must handle merge conflicts display appropriately
- FR-205: Comments must support markdown rendering (as plain text with formatting hints)
- FR-206: Must respect GitLab approval rules (required approvals count)
- FR-207: Pipeline view must auto-refresh for running pipelines
- FR-208: Must handle large MRs with many files (100+ files)

## Non-Goals

- Merge conflict resolution within TUI
- Interactive rebase functionality
- Creating new MRs (use git push or glab CLI)
- MR templates
- Merge trains management
- Squash commit editing
- Cherry-picking commits
- Real-time notifications for MR updates

## Technical Considerations

### GitLab API Endpoints Used
- `GET /projects/:id/merge_requests` - List MRs
- `GET /projects/:id/merge_requests/:iid` - Get single MR
- `GET /projects/:id/merge_requests/:iid/changes` - Get MR diff
- `GET /projects/:id/merge_requests/:iid/diffs` - Get MR diffs (paginated)
- `GET /projects/:id/merge_requests/:iid/commits` - Get MR commits
- `GET /projects/:id/merge_requests/:iid/notes` - Get MR comments
- `POST /projects/:id/merge_requests/:iid/notes` - Add comment
- `GET /projects/:id/merge_requests/:iid/discussions` - Get discussion threads
- `POST /projects/:id/merge_requests/:iid/discussions` - Create discussion
- `PUT /projects/:id/merge_requests/:iid/discussions/:id` - Resolve discussion
- `POST /projects/:id/merge_requests/:iid/approve` - Approve MR
- `POST /projects/:id/merge_requests/:iid/unapprove` - Unapprove MR
- `GET /projects/:id/merge_requests/:iid/pipelines` - Get MR pipelines
- `PUT /projects/:id/merge_requests/:iid` - Update MR (draft status, etc.)
- `PUT /projects/:id/merge_requests/:iid/merge` - Merge MR

### Data Structures
```go
type MergeRequest struct {
    IID              int
    Title            string
    Description      string
    State            string  // "opened", "closed", "merged"
    SourceBranch     string
    TargetBranch     string
    Author           User
    Assignees        []User
    Reviewers        []User
    Labels           []string
    Draft            bool
    MergeStatus      string  // "can_be_merged", "cannot_be_merged", etc.
    Pipeline         *Pipeline
    ApprovalsRequired int
    ApprovalsLeft    int
    ChangesCount     int
    WebURL           string
    CreatedAt        time.Time
    UpdatedAt        time.Time
}

type DiffFile struct {
    OldPath     string
    NewPath     string
    Diff        string  // Raw diff content
    NewFile     bool
    RenamedFile bool
    DeletedFile bool
    Additions   int
    Deletions   int
}

type Discussion struct {
    ID        string
    Notes     []Note
    Resolved  bool
    Resolvable bool
}

type Note struct {
    ID        int
    Body      string
    Author    User
    CreatedAt time.Time
    System    bool    // True for system-generated notes
    Position  *DiffPosition  // For line comments
}

type DiffPosition struct {
    BaseSHA      string
    HeadSHA      string
    StartSHA     string
    OldPath      string
    NewPath      string
    OldLine      *int
    NewLine      *int
    LineRange    *LineRange
}
```

### UI Layout for MR Review
```
┌─────────────────────────────────────────────────────────────────────┐
│ LazyGitLab - mygroup/myproject                          Connected   │
├────────────┬────────────────────────────────────────────────────────┤
│            │ !456: Add user authentication                          │
│ Projects   │ feature/auth -> main  |  @john  |  Updated 1h ago     │
│   Issues   │ Pipeline: ✓ passed  |  Approvals: 1/2  |  +234 -56    │
│ > MRs      ├────────────────────────────────────────────────────────┤
│            │ Files (12 changed)                                     │
│            │ ├─ src/                                                │
│            │ │  ├─ auth/                                            │
│            │ │  │  ├─ login.go          +45  -12  💬 2              │
│            │ │  │  └─ middleware.go     +89  -0                     │
│            │ │  └─ api/                                             │
│            │ │     └─ handlers.go       +23  -8   💬 1              │
│            │ └─ tests/                                              │
│            │    └─ auth_test.go         +77  -36                    │
├────────────┼────────────────────────────────────────────────────────┤
│            │  src/auth/login.go                                     │
│ Commits    │ ─────────────────────────────────────────────────────  │
│  (3)       │  23  │ func Login(ctx context.Context) error {         │
│            │  24 +│     user, err := validateCredentials(ctx)       │
│ Pipeline   │  25 +│     if err != nil {                             │
│  ✓ passed  │  26 +│         return fmt.Errorf("auth failed: %w", err│
│            │      │                                                 │
│            │      │ 💬 @reviewer - 2h ago                           │
│            │      │ Should we add rate limiting here?               │
│            │      │   └─ @john - 1h ago                             │
│            │      │      Good idea, I'll add that.                  │
│            │      │                                                 │
│            │  27 +│     }                                           │
│            │  28 +│     return createSession(ctx, user)             │
│            │  29  │ }                                               │
├────────────┴────────────────────────────────────────────────────────┤
│ j/k:scroll  n/N:next/prev hunk  c:comment  A:approve  ?:help        │
└─────────────────────────────────────────────────────────────────────┘
```

### Diff Rendering Approach
1. Parse raw diff output from GitLab API
2. Split into hunks based on @@ markers
3. Apply syntax highlighting using [chroma](https://github.com/alecthomas/chroma) library
4. Render with line numbers and change indicators
5. Overlay comments at appropriate positions
6. Handle wide content with horizontal scrolling

### Syntax Highlighting
Use the chroma library for syntax highlighting:
```go
import "github.com/alecthomas/chroma/v2"

// Detect language from filename
lexer := lexers.Match(filename)
// Apply highlighting
iterator, _ := lexer.Tokenise(nil, sourceCode)
// Render to terminal colors
formatter := formatters.Get("terminal256")
```

## Success Metrics

- User can review an MR diff within 5 seconds of selection
- User can add a line comment in under 10 seconds
- User can approve an MR in under 5 seconds
- Diff rendering handles files up to 5000 lines smoothly
- Syntax highlighting works for 95% of common file types

## Open Questions

- Should we support suggestion comments (GitLab's code suggestion feature)?
  - Recommendation: Not for MVP, complex to implement in TUI
- How to handle binary file diffs?
  - Recommendation: Show "Binary file changed" message, no preview
- Should we support reviewing specific commits vs entire MR?
  - Recommendation: Yes, include in MVP (US-216)
- How to handle MRs with merge conflicts?
  - Recommendation: Show conflict markers, but resolution must happen externally
