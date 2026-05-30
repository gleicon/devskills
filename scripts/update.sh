#!/usr/bin/env bash
# update.sh: pull latest devskills and reinstall
set -euo pipefail

DEVSKILLS_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

echo "[devskills] Updating..."

if [ -d "${DEVSKILLS_DIR}/.git" ]; then
  git -C "${DEVSKILLS_DIR}" pull --ff-only
  echo "[devskills] Pulled latest from git"
elif command -v npm &>/dev/null && [ -f "${DEVSKILLS_DIR}/package.json" ]; then
  npm -g update devskills
  echo "[devskills] Updated via npm"
else
  echo "[devskills] WARN: not a git repo and npm not available. Update manually."
  exit 1
fi

# Reinstall skills only (no external tools, no Cursor/VSCode, no language profile changes)
"${DEVSKILLS_DIR}/install.sh" --skip-external --skip-cursor --skip-vscode

# Optionally force-upgrade external tools
if [[ " $* " == *" --upgrade-deps "* ]]; then
  echo "[devskills] Upgrading external tools..."
  bash "${DEVSKILLS_DIR}/scripts/upgrade-deps.sh"
fi

echo "[devskills] Update complete."
echo "[devskills] To also upgrade external tools: ./scripts/update.sh --upgrade-deps"
echo "[devskills] To refresh a project's AGENTS.md blocks, re-run setup.sh in that project."
