#!/usr/bin/env sh
set -eu

usage() {
  cat <<'EOF'
Usage:
  scripts/check-conventional-commits.sh [<git-range>]
  scripts/check-conventional-commits.sh --message-file <path>

Examples:
  scripts/check-conventional-commits.sh HEAD~5..HEAD
  scripts/check-conventional-commits.sh --message-file .git/COMMIT_EDITMSG
EOF
}

regex='^(feat|fix|docs|style|refactor|perf|test|build|ci|chore|revert)(\([^)]+\))?(!)?: .+'

check_subject() {
  subject="$1"
  if printf '%s\n' "$subject" | grep -Eq "$regex"; then
    return 0
  fi

  printf '%s\n' "invalid conventional commit subject: $subject" >&2
  return 1
}

check_message_file() {
  message_file="$1"
  if [ ! -f "$message_file" ]; then
    printf '%s\n' "message file not found: $message_file" >&2
    return 1
  fi

  subject=$(sed -n '/^[^#[:space:]].*/{p;q;}' "$message_file")
  if [ -z "$subject" ]; then
    printf '%s\n' "empty commit message subject" >&2
    return 1
  fi

  check_subject "$subject"
}

check_git_range() {
  git_range="$1"

  subjects=$(git log --format=%s "$git_range")
  if [ -z "$subjects" ]; then
    printf '%s\n' "no commits found in range: $git_range" >&2
    return 1
  fi

  failed=0
  OLDIFS=$IFS
  IFS='
'
  for subject in $subjects; do
    if ! check_subject "$subject"; then
      failed=1
    fi
  done
  IFS=$OLDIFS

  if [ "$failed" -ne 0 ]; then
    printf '%s\n' "commit message check failed for range: $git_range" >&2
    return 1
  fi
}

if [ "$#" -eq 0 ]; then
  check_git_range "HEAD"
  exit 0
fi

if [ "$1" = "-h" ] || [ "$1" = "--help" ]; then
  usage
  exit 0
fi

if [ "$1" = "--message-file" ]; then
  if [ "$#" -ne 2 ]; then
    usage
    exit 1
  fi
  check_message_file "$2"
  exit 0
fi

if [ "$#" -ne 1 ]; then
  usage
  exit 1
fi

check_git_range "$1"
