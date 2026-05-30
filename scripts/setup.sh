#!/usr/bin/env bash
# setup.sh: configure devskills for the current project directory
set -euo pipefail

DEVSKILLS_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

usage() {
  cat <<EOF
Usage: setup.sh [--lang=<profile>] [options]

Writes the devskills baseline (universal engineering principles) to AGENTS.md
and points CLAUDE.md at it. --lang stacks a language profile on top.

Profiles (optional):
  go                Go 1.22+ backend service
  typescript        TypeScript 5+ (Workers, Next.js, React)
  javascript        ES2022+ (Workers, vanilla frontend)
  rust              Rust stable (systems programming, large projects)

Options:
  --concise         Add a terse-response directive to AGENTS.md
  --hints           Add a devskills tooling reference to AGENTS.md
  --cursor          Install Cursor rules into current project
  --vscode          Install VSCode Copilot instructions into current project
  --uninstall       Remove devskills blocks from AGENTS.md/CLAUDE.md and the marker
  --dry-run         Show what would happen without writing files

Example:
  setup.sh                              # baseline only
  setup.sh --lang=go --cursor
  setup.sh --lang=typescript --concise --hints
  setup.sh --uninstall                  # back out devskills changes
EOF
}

LANG_PROFILE=""
DO_CURSOR=0
DO_VSCODE=0
DO_CONCISE=0
DO_HINTS=0
DO_UNINSTALL=0
DRY_RUN=0

for arg in "$@"; do
  case "$arg" in
    --lang=*) LANG_PROFILE="${arg#--lang=}" ;;
    --claude-dir=*|--skip-cursor|--skip-vscode|--skip-external) ;;  # install-only flags; ignored here
    --cursor) DO_CURSOR=1 ;;
    --vscode) DO_VSCODE=1 ;;
    --concise) DO_CONCISE=1 ;;
    --hints) DO_HINTS=1 ;;
    --uninstall) DO_UNINSTALL=1 ;;
    --dry-run) DRY_RUN=1 ;;
    --help|-h) usage; exit 0 ;;
    *) echo "Unknown argument: $arg"; usage; exit 1 ;;
  esac
done

# shellcheck source=lib/profile.sh
source "${DEVSKILLS_DIR}/scripts/lib/profile.sh"
# shellcheck source=lib/editors.sh
source "${DEVSKILLS_DIR}/scripts/lib/editors.sh"

if [ "$DO_UNINSTALL" -eq 1 ]; then
  echo "Removing devskills from ${PWD}"
  devskills_uninstall "$PWD" "$DRY_RUN"
  echo "Done."
  exit 0
fi

# Validate the language profile up front (if one was requested).
if [ -n "$LANG_PROFILE" ] && [ ! -f "${DEVSKILLS_DIR}/prompts/language/${LANG_PROFILE}.md" ]; then
  echo "Error: no profile for '${LANG_PROFILE}'. Available: go, typescript, javascript, rust"
  exit 1
fi

# AGENTS.md baseline (+ optional layers); CLAUDE.md imports it via @AGENTS.md.
echo "devskills baseline${LANG_PROFILE:+ + ${LANG_PROFILE} profile}"
devskills_apply "${DEVSKILLS_DIR}/prompts" "$PWD" "$DRY_RUN" "$LANG_PROFILE" "$DO_CONCISE" "$DO_HINTS"

# Cursor rules
if [ "$DO_CURSOR" -eq 1 ]; then
  echo "Cursor rules:"
  devskills_install_cursor "$PWD" "$LANG_PROFILE"
fi

# VSCode Copilot
if [ "$DO_VSCODE" -eq 1 ]; then
  echo "VSCode Copilot:"
  devskills_install_vscode "$PWD"
fi

echo ""
echo "Done. AGENTS.md baseline${LANG_PROFILE:+ + ${LANG_PROFILE} profile} written; CLAUDE.md imports it."
echo "Activate in Claude Code: /tiger-style"
