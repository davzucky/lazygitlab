# AGENTS.md

Repository guidance for coding agents working on LazyGitLab.

## Work intake policy

- Every code change must map to a GitHub issue.
- If work is requested without an existing issue, create one first (or ask the user to create it) before implementation.
- Reference the issue number in branches, commits, and PR descriptions.
- Keep scope aligned to the issue acceptance criteria; open a follow-up issue for extra work.

## Branch and MR workflow

- Use an MR-based workflow for all feature work.
- One issue maps to one branch and one merge request.
- Branch naming should include the issue number (e.g., `feature/11-ci-automation`).
- Do not implement multiple issues in a single MR.
- Open the MR as soon as the issue scope is implemented and local checks pass.

## Stack and runtime

- Language: Go
- Minimum version: Go 1.23+
- TUI: Bubble Tea + Lipgloss (+ Bubbles components)
- GitLab API client: `gitlab.com/gitlab-org/api/client-go`

## Build, test, lint

- Build: `just build`
- Run: `just run`
- Test: `just test`
- Lint: `just lint`
- Format: `just fmt`
- Local quality gate: `just pre-commit`
- TUI regression (mock mode): `TUI_VALIDATE_MOCK=1 just tui-validate`

## Automation and commit policy

- CI is defined in `.github/workflows/ci.yml` and must pass before merge.
- Conventional commits are required and validated in CI and local hooks.
- Install local hooks with `just hooks`.

## Layout rules (strict internal-first)

Follow `golang-standards/project-layout` with a strict internal strategy.

- Keep `cmd/lazygitlab/main.go` thin; only parse flags and call app bootstrap.
- Put app implementation under `internal/...`.
- Use `pkg/...` only when a package is intentionally reusable externally.
- Do not add a `/src` directory.

Current package map:

- `internal/app`: lifecycle orchestration and wiring
- `internal/config`: config load/save and precedence
- `internal/gitlab`: API wrapper with retries and pagination
- `internal/project`: Git remote parsing and project detection
- `internal/tui`: all Bubble Tea models and views
- `internal/logging`: debug log setup

## Configuration behavior

Load configuration in this order (highest priority first):

1. Environment: `GITLAB_TOKEN`, `GITLAB_HOST`
2. `~/.config/lazygitlab/config.yml`
3. `~/.config/glab-cli/config.yml`

If configuration is missing/invalid, start interactive first-run setup wizard.

## TUI expectations

- Keyboard-driven navigation:
  - `j/k`, arrow keys, `h/l`, `tab`, `shift+tab`, `enter`, `esc`, `q`, `?`
- Multi-panel layout:
  - Sidebar navigation
  - Main list panel
  - Details panel
  - Status bar
- Include loading and retryable error states.

## TUI screen architecture

- Treat each main TUI area as its own screen module:
  - `Primary` (home/lookup)
  - `Issue`
  - `Merge Request`
- Keep `internal/tui/dashboard.go` as the shell/router:
  - app-level chrome (sidebar/status/help/error)
  - active screen switching
  - shared context/provider wiring
- Keep screen-specific behavior in dedicated files (`screen_<name>.go`) and avoid expanding shell conditionals.
- New screen workflow:
  1. Add/extend a `ViewMode` entry for the screen.
  2. Create `internal/tui/screen_<name>.go` with key handling and rendering helpers.
  3. Register navigation from `Primary` so the screen is discoverable.
  4. Add tests for routing, key behavior, and loading/error handling.
  5. Validate in mock mode with `TUI_VALIDATE_MOCK=1 just tui-validate`.
- Esc behavior for top-level screens should return to `Primary` (screen-local Esc behavior can still close local overlays/details first).

## TUI validation workflow

- Prefer validating UI changes with `agent-tui` when available.
- Skill file: `.opencode/skill/agent-tui/SKILL.md`.
- Minimum validation pass:
  - launch app in virtual PTY
  - navigate with `j/k`, `tab`, `shift+tab`, `?`, `esc`
  - verify no panel drift/overflow while data loads

## Quality bar

- Keep package names short and domain-focused.
- Prefer explicit error wrapping with context.
- Add unit tests for parsing/config precedence/state transitions.
- Run `gofmt` and tests before finishing.
