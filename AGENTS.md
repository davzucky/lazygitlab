# AGENTS.md

This guide summarizes how to build, lint, and test LazyDocker, plus the
expected Go coding style in this repo. Use it as a default playbook for
agentic edits.

## Quick facts
- Language: Go (module `github.com/jesseduffield/lazydocker`).
- go.mod: `go 1.22`, toolchain `go1.23.6`.
- CI uses Go 1.24.x and enforces `GOFLAGS=-mod=vendor`.
- Vendored dependencies are required; `vendor/` must stay in sync.

## Build commands
- Build locally: `go build`.
- Build for OS targets (as in CI):
  - `GOOS=linux go build`
  - `GOOS=windows go build`
  - `GOOS=darwin go build`
- Run locally (source): `go run main.go`.

## Test commands
- Full test suite (CI): `bash ./test.sh`
  - Uses `GOFLAGS=-mod=vendor`.
  - Runs `go test -race -coverprofile=profile.out -covermode=atomic` per package.
  - If `gotest` is installed, it uses `gotest` instead of `go test`.
- Quick full test (manual): `GOFLAGS=-mod=vendor go test ./...`

### Single test or single package
- Single package: `GOFLAGS=-mod=vendor go test ./pkg/commands`
- Single test by name:
  - `GOFLAGS=-mod=vendor go test ./pkg/commands -run TestName`
- Single subtest:
  - `GOFLAGS=-mod=vendor go test ./pkg/commands -run TestName/Subcase`
- Re-run without cache when debugging:
  - `GOFLAGS=-mod=vendor go test ./pkg/commands -run TestName -count=1 -v`

## Lint and formatting
- Lint (CI): `golangci-lint run`
  - Config: `.golangci.yml`.
- Format check (CI):
  - `gofmt -s` on all `*.go` files outside `vendor/`.
- Formatting expectations (by lint config):
  - `gofumpt` and `goimports` are enabled. Use them if available.

## Other CI checks
- Cheatsheet validation:
  - `go run scripts/cheatsheet/main.go check`
- Vendor consistency check:
  - `go mod vendor && git diff --exit-code`

## Dependency and vendor workflow
- Vendor directory is the source of truth.
- Typical dependency bump:
  - `go get -u github.com/jesseduffield/gocui@master`
  - `go mod tidy`
  - `go mod vendor`
- Local development often sets: `GOFLAGS=-mod=vendor`.

## Repository layout
- `main.go`: entry point and top-level error handling.
- `pkg/`: application code (commands, gui, config, utils, etc.).
- `scripts/`: helper scripts (cheatsheet generation/checks).
- `vendor/`: vendored deps; do not edit unless updating deps.

## Code style guidelines

### Imports
- Let `goimports` group and order imports.
- Standard library first, then third-party, then local (`github.com/jesseduffield/lazydocker/...`).
- Avoid unused imports; `goimports` will remove them.

### Formatting
- Run `gofmt -s` on all Go files.
- Prefer `gofumpt` formatting for consistency (enforced by lint).
- Keep lines readable; split long parameter lists and chained calls.

### Naming conventions
- Use idiomatic Go naming: CamelCase for exports, lowerCamelCase for locals.
- Avoid stutter in exported types and functions.
- Keep acronyms consistent (e.g., `URL`, `ID`, `OS`).
- File names match package purpose; platform-specific files use `_windows.go`, `_unix.go`.

### Types and interfaces
- Prefer concrete types unless an interface improves testability or decoupling.
- Keep interfaces small and focused; prefer single-purpose interfaces.
- Use `struct` fields to group related data; avoid oversized parameter lists.

### Error handling
- Return errors to the caller; handle at the appropriate layer.
- Use contextual errors and wrap where stack traces are needed.
  - `commands.WrapError` wraps with `github.com/go-errors/errors`.
- For command execution, prefer sanitized error messages (see `sanitisedCommandOutput`).
- Avoid naked returns (enforced by `nakedret`).

### Logging
- Use `logrus` (`pkg/log`) for structured logging.
- Production logger discards output; debug uses JSON logs.
- Prefer `logrus.Entry` with contextual fields rather than global logging.

### Concurrency and tests
- When appropriate, use `t.Parallel()` in tests (lint `tparallel`).
- Mark helper test functions with `t.Helper()` (lint `thelper`).
- Avoid data races; the test runner uses `-race` in CI.

### Switches and enums
- Exhaustive checks are enabled (`exhaustive`).
- Use explicit `default` only when it is a true fallback.

### Performance and correctness
- Lints enforce: `wastedassign`, `unparam`, `prealloc`, `unconvert`, `makezero`.
- Avoid unnecessary allocations; preallocate slices when size is known.

## Build/test environment expectations
- Set `GOFLAGS=-mod=vendor` for reproducible builds.
- Ensure `vendor/` is up to date before committing changes.
- CI runs on Linux and Windows for tests; keep platform differences in mind.

## Cursor/Copilot rules
- No `.cursorrules`, `.cursor/rules/`, or `.github/copilot-instructions.md` detected in this repo.

## When in doubt
- Follow Effective Go: https://golang.org/doc/effective_go.html
- Mirror existing patterns in `pkg/commands`, `pkg/gui`, and `pkg/utils`.
