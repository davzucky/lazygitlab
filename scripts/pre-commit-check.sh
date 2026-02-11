#!/usr/bin/env sh
set -eu

printf '%s\n' "Running formatting checks..."
just fmt

if ! git diff --quiet -- cmd internal; then
  printf '%s\n' "go formatting changed files. Please review and re-stage changes." >&2
  exit 1
fi

printf '%s\n' "Running lint checks..."
just lint

printf '%s\n' "Running tests..."
just test

printf '%s\n' "Pre-commit checks passed."
