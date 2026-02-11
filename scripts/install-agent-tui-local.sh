#!/usr/bin/env sh
set -eu

INSTALL_DIR="${HOME}/.local/bin"
INSTALL_SCRIPT="/tmp/agent-tui-install.sh"

curl -fsSL "https://raw.githubusercontent.com/pproenca/agent-tui/master/install.sh" -o "${INSTALL_SCRIPT}"

AGENT_TUI_SKIP_PM=1 AGENT_TUI_INSTALL_DIR="${INSTALL_DIR}" sh "${INSTALL_SCRIPT}"

"${INSTALL_DIR}/agent-tui" --version
