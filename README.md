# LazyGitLab

LazyGitLab is a keyboard-first terminal UI for common GitLab workflows.

This repository currently implements the Part 1 foundation: project setup,
configuration handling, GitLab API client wrapper, context detection, first-run
setup wizard, and a multi-panel Bubble Tea interface.

## Requirements

- Go 1.23+
- A GitLab personal access token

## Quick start

```bash
make run
```

On first run, LazyGitLab opens an interactive setup wizard to collect your host
and token, then saves config to `~/.config/lazygitlab/config.yml`.

## Configuration precedence

1. `GITLAB_TOKEN` / `GITLAB_HOST`
2. `~/.config/lazygitlab/config.yml`
3. `~/.config/glab-cli/config.yml`

## Flags

- `--project group/subgroup/name`: manually set project context
- `--debug`: write verbose logs to `~/.local/share/lazygitlab/debug.log`

## Keybindings

- `j`/`k` or arrows: move selection
- `h`/`l` or arrows: switch panel view
- `tab` / `shift+tab`: cycle view
- `1`, `2`, `3`: jump to Projects / Issues / Merge Requests
- `?`: help popup
- `q`: quit

## TUI regression check

Use `agent-tui` to validate panel stability and navigation after TUI changes:

```bash
make tui-validate
```

This runs a scripted flow (`j`, `Tab`, `Shift+Tab`, `?`, `Esc`) and fails if
the sidebar/main separator drifts during updates.

## Project layout

The code follows a strict internal-first layout inspired by
`golang-standards/project-layout`:

- `cmd/lazygitlab`: executable entrypoint
- `internal/app`: startup orchestration
- `internal/config`: config loading and persistence
- `internal/gitlab`: GitLab API client wrapper
- `internal/project`: git remote project detection
- `internal/tui`: Bubble Tea models and views
- `internal/logging`: debug logger
