---
name: go-project-layout
description: "Apply strict internal-first Go project layout practices from golang-standards/project-layout."
---

# Go Project Layout Skill

Use this skill when creating or refactoring Go application repositories.

## Goals

- Keep executable entrypoints in `cmd/<app>/main.go` thin.
- Put private application code in `internal/...`.
- Use `pkg/...` only when intentionally exposing reusable public APIs.
- Avoid non-idiomatic top-level `src/` directories.

## Decision checklist

1. Is this code only for this app?
   - Put in `internal/<domain>`.
2. Is this code meant for external import and versioned as public API?
   - Put in `pkg/<name>`.
3. Is `main.go` doing orchestration logic?
   - Move orchestration into `internal/app` and keep `main` minimal.

## Recommended package map for LazyGitLab

- `internal/app` for startup flow and wiring.
- `internal/config` for load/save/precedence.
- `internal/gitlab` for GitLab client abstraction.
- `internal/project` for git remote parsing and context detection.
- `internal/tui` for Bubble Tea models and UI logic.
- `internal/logging` for debug log initialization.

## Validation before completion

- [ ] `just fmt`
- [ ] `just test`
- [ ] `just build`
- [ ] No new `pkg` package added unless justified.
- [ ] `cmd/lazygitlab/main.go` only handles flags + app startup.
