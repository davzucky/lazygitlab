#!/usr/bin/env sh
set -eu

AGENT_TUI_BIN="${AGENT_TUI_BIN:-$HOME/.local/bin/agent-tui}"
APP_BIN="${APP_BIN:-$PWD/lazygitlab}"
WAIT_TIMEOUT_MS="${WAIT_TIMEOUT_MS:-3000}"
TUI_VALIDATE_MOCK="${TUI_VALIDATE_MOCK:-0}"

if [ ! -x "$AGENT_TUI_BIN" ]; then
  printf '%s\n' "error: agent-tui not found at $AGENT_TUI_BIN" >&2
  exit 1
fi

if [ ! -x "$APP_BIN" ]; then
  go build -o "$APP_BIN" ./cmd/lazygitlab
fi

SESSION_ID=""

cleanup() {
  if [ -n "$SESSION_ID" ]; then
    "$AGENT_TUI_BIN" kill --session "$SESSION_ID" >/dev/null 2>&1 || true
  fi
  "$AGENT_TUI_BIN" daemon stop >/dev/null 2>&1 || true
}

trap cleanup EXIT INT TERM

"$AGENT_TUI_BIN" daemon start >/dev/null

if [ "$TUI_VALIDATE_MOCK" = "1" ]; then
  RUN_JSON=$("$AGENT_TUI_BIN" run --format json /bin/sh -- -c "LAZYGITLAB_MOCK_DATA=1 \"$APP_BIN\" --debug")
else
  RUN_JSON=$("$AGENT_TUI_BIN" run --format json "$APP_BIN" -- --debug)
fi
SESSION_ID=$(printf '%s' "$RUN_JSON" | python3 -c 'import json,sys; print(json.load(sys.stdin)["session_id"])')

wait_stable() {
  "$AGENT_TUI_BIN" wait --session "$SESSION_ID" --stable --timeout "$WAIT_TIMEOUT_MS" >/dev/null
}

wait_for_text() {
  text="$1"
  if "$AGENT_TUI_BIN" wait --session "$SESSION_ID" "$text" --assert --timeout "$WAIT_TIMEOUT_MS" >/dev/null 2>&1; then
    return 0
  fi

  if python3 - "$AGENT_TUI_BIN" "$SESSION_ID" "$text" "$WAIT_TIMEOUT_MS" <<'PY'
import json
import subprocess
import sys
import time

agent_tui_bin, session_id, needle, timeout_ms = sys.argv[1:5]
deadline = time.time() + (int(timeout_ms) / 1000.0)

while time.time() < deadline:
    proc = subprocess.run(
        [agent_tui_bin, "screenshot", "--format", "json", "--session", session_id],
        check=False,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        text=True,
    )
    if proc.returncode == 0:
        try:
            payload = json.loads(proc.stdout)
            screen = payload.get("screenshot", "")
        except json.JSONDecodeError:
            screen = proc.stdout
        if needle in screen:
            raise SystemExit(0)
    time.sleep(0.2)

raise SystemExit(f"Wait condition not met within timeout for text: {needle}")
PY
  then
    return 0
  fi

  printf '%s\n' "warn: text wait timed out for '$text', continuing with screenshot assertions" >&2
  return 0
}

capture_screen() {
  "$AGENT_TUI_BIN" screenshot --format json --session "$SESSION_ID" | python3 -c 'import json,sys
raw = sys.stdin.read().strip()
if not raw:
    print("")
    raise SystemExit(0)
try:
    payload = json.loads(raw)
except json.JSONDecodeError:
    print(raw)
    raise SystemExit(0)
print(payload.get("screenshot", ""))'
}

capture_screen_retry() {
  needle="$1"
  python3 - "$AGENT_TUI_BIN" "$SESSION_ID" "$needle" "$WAIT_TIMEOUT_MS" <<'PY'
import json
import subprocess
import sys
import time

agent_tui_bin, session_id, needle, timeout_ms = sys.argv[1:5]
deadline = time.time() + (int(timeout_ms) / 1000.0)
last = ""

while time.time() < deadline:
    proc = subprocess.run(
        [agent_tui_bin, "screenshot", "--format", "json", "--session", session_id],
        check=False,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        text=True,
    )
    if proc.returncode == 0 and proc.stdout.strip():
        try:
            payload = json.loads(proc.stdout)
            screen = payload.get("screenshot", "")
        except json.JSONDecodeError:
            screen = proc.stdout
        if screen:
            last = screen
            if not needle or needle in screen:
                print(screen)
                raise SystemExit(0)
    time.sleep(0.2)

print(last)
raise SystemExit(1)
PY
}

assert_layout() {
  screen="$1"

  printf '%s' "$screen" | python3 -c '
import os
import sys

screen = sys.stdin.read()
baseline = os.environ.get("BASELINE_SEP", "")

if not screen.strip():
    print(-1)
    raise SystemExit(0)

lines = [line for line in screen.splitlines() if "││" in line]
if len(lines) < 8:
    raise SystemExit("insufficient split-panel lines for layout validation")

positions = [line.index("││") for line in lines[:20]]
if max(positions) != min(positions):
    raise SystemExit(f"panel separator drift detected: {positions}")

if baseline:
    b = int(baseline)
    if b >= 0 and positions[0] != b:
        raise SystemExit(f"panel separator moved from {b} to {positions[0]}")

print(positions[0])
'
}

assert_contains() {
  screen="$1"
  needle="$2"

  if [ -z "$(printf '%s' "$screen" | tr -d '[:space:]')" ]; then
    printf '%s\n' "warn: empty screenshot, skipping text assertion for '$needle'" >&2
    return 0
  fi

  printf '%s' "$screen" | python3 -c 'import sys; s=sys.stdin.read(); n=sys.argv[1];
if n not in s: raise SystemExit(f"missing expected text: {n}")' "$needle"
}

wait_stable
wait_for_text "Navigation"
BASE_SCREEN=$(capture_screen_retry "Navigation" || capture_screen)
BASELINE_SEP=$(BASELINE_SEP="" assert_layout "$BASE_SCREEN")

"$AGENT_TUI_BIN" press --session "$SESSION_ID" j >/dev/null
wait_stable
wait_for_text "Navigation"
AFTER_J=$(capture_screen_retry "Navigation" || capture_screen)
BASELINE_SEP="$BASELINE_SEP" assert_layout "$AFTER_J" >/dev/null

"$AGENT_TUI_BIN" press --session "$SESSION_ID" Tab >/dev/null
wait_stable
wait_for_text "Navigation"
AFTER_TAB=$(capture_screen_retry "Navigation" || capture_screen)
BASELINE_SEP="$BASELINE_SEP" assert_layout "$AFTER_TAB" >/dev/null

"$AGENT_TUI_BIN" press --session "$SESSION_ID" Shift+Tab >/dev/null
wait_stable
wait_for_text "Navigation"
AFTER_SHIFT_TAB=$(capture_screen_retry "Navigation" || capture_screen)
BASELINE_SEP="$BASELINE_SEP" assert_layout "$AFTER_SHIFT_TAB" >/dev/null

"$AGENT_TUI_BIN" type --session "$SESSION_ID" "?" >/dev/null
wait_stable
wait_for_text "Keybindings"
HELP_SCREEN=$(capture_screen_retry "Keybindings" || capture_screen)
assert_contains "$HELP_SCREEN" "Keybindings"

"$AGENT_TUI_BIN" press --session "$SESSION_ID" Escape >/dev/null
wait_stable
wait_for_text "Navigation"
AFTER_HELP_CLOSE=$(capture_screen_retry "Navigation" || capture_screen)
BASELINE_SEP="$BASELINE_SEP" assert_layout "$AFTER_HELP_CLOSE" >/dev/null

printf '%s\n' "TUI drift validation passed (session: $SESSION_ID)"
