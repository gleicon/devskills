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
  python            Python 3.13+ (backend, APIs, CLIs, data)
  java              Java 25+ LTS (backend services, APIs, CLIs)
  zig               Zig 0.16 (systems, CLIs, embedded; Tiger Style native)

Options:
  --concise         Add a terse-response directive to AGENTS.md
  --phases          Add phase-aware Insight suggestions to AGENTS.md
  --uninstall       Remove devskills blocks from AGENTS.md/CLAUDE.md and the marker
  --dry-run         Show what would happen without writing files

Example:
  setup.sh                              # baseline only
  setup.sh --lang=go
  setup.sh --lang=typescript --concise --phases
  setup.sh --uninstall                  # back out devskills changes
EOF
}

LANG_PROFILE=""
DO_CONCISE=0
DO_PHASES=0
DO_UNINSTALL=0
DRY_RUN=0

for arg in "$@"; do
  case "$arg" in
    --lang=*) LANG_PROFILE="${arg#--lang=}" ;;
    --claude-dir=*|--skip-external) ;;  # install-only flags; ignored here
    --concise) DO_CONCISE=1 ;;
    --phases) DO_PHASES=1 ;;
    --uninstall) DO_UNINSTALL=1 ;;
    --dry-run) DRY_RUN=1 ;;
    --help|-h) usage; exit 0 ;;
    *) echo "Unknown argument: $arg"; usage; exit 1 ;;
  esac
done

# shellcheck source=lib/profile.sh
source "${DEVSKILLS_DIR}/scripts/lib/profile.sh"

if [ "$DO_UNINSTALL" -eq 1 ]; then
  echo "Removing devskills from ${PWD}"
  devskills_uninstall "$PWD" "$DRY_RUN"
  echo "Done."
  exit 0
fi

# Validate the language profile up front (if one was requested).
if [ -n "$LANG_PROFILE" ] && [ ! -f "${DEVSKILLS_DIR}/agents-md/language/${LANG_PROFILE}.md" ]; then
  echo "Error: no profile for '${LANG_PROFILE}'. Available: go, typescript, javascript, rust, python, java, zig"
  exit 1
fi

# AGENTS.md baseline (+ optional layers); CLAUDE.md imports it via @AGENTS.md.
echo "devskills baseline${LANG_PROFILE:+ + ${LANG_PROFILE} profile}"
devskills_apply "${DEVSKILLS_DIR}/agents-md" "$PWD" "$DRY_RUN" "$LANG_PROFILE" "$DO_CONCISE" "$DO_PHASES"

echo ""
echo "Done. AGENTS.md baseline${LANG_PROFILE:+ + ${LANG_PROFILE} profile} written; CLAUDE.md imports it."
echo "Activate in Claude Code: /ds-tiger-style-mode"
