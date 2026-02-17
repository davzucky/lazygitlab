---
name: tui-good-practices
description: Use when designing, reviewing, or refactoring interactive terminal interfaces and keyboard-driven flows where focus behavior, state transitions, loading/error handling, and layout stability must be predictable.
---

# TUI Good Practices

## Overview

Build TUIs as deterministic state machines with explicit focus, predictable key handling, and resilient views for loading, empty, and error states.

## Interaction Contract

- Keep a stable keymap and show it in help/status (`j/k`, arrows, `tab`, `shift+tab`, `enter`, `esc`, `q`, `?`).
- Make focus visible at all times; never leave the user unsure which panel consumes input.
- Define `esc` behavior by level: close local overlay first, then return to parent screen.
- Keep destructive actions confirmable and reversible when possible.

## Layout and Rendering

- Treat terminal size as dynamic input; recompute dimensions on resize.
- Use bounded panels with clipping/wrapping to prevent drift and overflow.
- Avoid full-screen flicker: update only from model state changes.
- Prefer stable placeholders/skeletons while data loads.

## State and Data Flow

- Track explicit UI state: active screen, focused panel, selected row, page, scroll offset, filters, async status.
- Separate view state from data-fetch side effects.
- Use pagination for large sets; keep selection visible when moving pages.
- Refresh data after mutating actions and preserve user context where safe.

## Errors and Recovery

- Render actionable errors in-line with retry affordance.
- Fail soft on partial data: keep navigation usable while one panel retries.
- Detect non-interactive terminals and provide a non-TUI fallback when needed.

## Bubble Tea / Lipgloss Notes

- Keep shell routing/chrome in one model and move screen-specific logic to dedicated screen modules.
- Centralize key handling per screen and avoid duplicating bindings across components.
- Use deterministic `Update` transitions and table/list viewport helpers for long content.

## Validation Checklist

- Verify keyboard paths: `j/k`, arrows, `tab`, `shift+tab`, `enter`, `esc`, `q`, `?`.
- Verify loading, empty, error, and retry states for each major screen.
- Verify no panel drift/overflow during async loads and terminal resizes.
- Run mock regression when available: `TUI_VALIDATE_MOCK=1 just tui-validate`.
- For interactive regression, automate a PTY flow with `agent-tui` and assert expected text after each action.

## Common Mistakes

- Coupling data fetching directly to rendering with hidden side effects.
- Overloading keys differently across screens without discoverable hints.
- Treating resize as cosmetic instead of a first-class state transition.
- Showing fatal errors for recoverable operations.
