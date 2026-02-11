# AGENTS.md

Repository guidance for coding agents working on LazyGitLab.

## Stack and runtime

- Language: Go
- Minimum version: Go 1.23+
- TUI: Bubble Tea + Lipgloss (+ Bubbles components)
- GitLab API client: `gitlab.com/gitlab-org/api/client-go`

## Build, test, lint

- Build: `make build`
- Run: `make run`
- Test: `make test`
- Lint: `make lint`
- Format: `make fmt`

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
