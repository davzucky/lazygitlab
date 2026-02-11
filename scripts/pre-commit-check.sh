#!/usr/bin/env sh
set -eu

printf '%s\n' "Running formatting checks..."
make fmt

if ! git diff --quiet -- cmd internal; then
  printf '%s\n' "go formatting changed files. Please review and re-stage changes." >&2
  exit 1
fi

printf '%s\n' "Running lint checks..."
make lint

printf '%s\n' "Running tests..."
make test

printf '%s\n' "Pre-commit checks passed."
