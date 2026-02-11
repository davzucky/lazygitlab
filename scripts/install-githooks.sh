#!/usr/bin/env sh
set -eu

if [ ! -d ".git" ]; then
  printf '%s\n' "error: .git directory not found" >&2
  exit 1
fi

mkdir -p .git/hooks

cp .githooks/pre-commit .git/hooks/pre-commit
cp .githooks/commit-msg .git/hooks/commit-msg

chmod +x .git/hooks/pre-commit .git/hooks/commit-msg

printf '%s\n' "Installed git hooks: pre-commit, commit-msg"
