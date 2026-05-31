#!/usr/bin/env bash
set -euo pipefail

DEVSKILLS_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CLAUDE_CONFIG_DIR="${CLAUDE_CONFIG_DIR:-${HOME}/.claude}"
OPENCODE_COMMANDS_DIR="${HOME}/.opencode/commands"

log() { printf '[devskills] %s\n' "$1"; }
warn() { printf '[devskills] WARN: %s\n' "$1" >&2; }

# Shared GSD/RTK/tldt logic (depends on log/warn above and DRY_RUN below).
# shellcheck source=scripts/lib/external-tools.sh
source "${DEVSKILLS_DIR}/scripts/lib/external-tools.sh"
# shellcheck source=scripts/lib/editors.sh
source "${DEVSKILLS_DIR}/scripts/lib/editors.sh"

# ------------------------------------------------------------
# Arguments
# ------------------------------------------------------------

LANG_PROFILE=""
SKIP_EXTERNAL=0
SKIP_CURSOR=0
SKIP_VSCODE=0
CONCISE=0
HINTS=0
PHASES=0
DRY_RUN=0

for arg in "$@"; do
  case "$arg" in
    --lang=*) LANG_PROFILE="${arg#--lang=}" ;;
    --claude-dir=*) CLAUDE_CONFIG_DIR="${arg#--claude-dir=}" ;;
    --skip-external) SKIP_EXTERNAL=1 ;;
    --skip-cursor) SKIP_CURSOR=1 ;;
    --skip-vscode) SKIP_VSCODE=1 ;;
    --concise) CONCISE=1 ;;
    --hints) HINTS=1 ;;
    --phases) PHASES=1 ;;
    --dry-run) DRY_RUN=1 ;;
    --help|-h)
      echo "Usage: install.sh [--lang=go|typescript|javascript|rust|python] [--claude-dir=PATH] [--skip-external] [--skip-cursor] [--skip-vscode] [--concise] [--hints] [--phases] [--dry-run]"
      echo ""
      echo "  --lang=<profile>    Language profile to write: go|typescript|javascript|rust|python"
      echo "  --claude-dir=PATH   Claude config dir (default: \$CLAUDE_CONFIG_DIR or \$HOME/.claude)"
      echo "  --skip-external     Skip external tool installation (GSD, RTK, tldt)"
      echo "  --skip-cursor       Skip Cursor rules install into the current project"
      echo "  --skip-vscode       Skip VSCode Copilot instructions install into the current project"
      echo "  --concise           Add a terse-response directive to AGENTS.md (with --lang)"
      echo "  --hints             Add a devskills tooling reference to AGENTS.md (with --lang)"
      echo "  --phases            Add phase-aware Insight suggestions to AGENTS.md (with --lang)"
      echo "  --dry-run           Show what would happen, write nothing"
      exit 0
      ;;
  esac
done

# Expand leading ~ in --claude-dir value.
# Quote the strip pattern: an unquoted ~/ undergoes tilde expansion itself,
# strips nothing, and yields "$HOME/~/.claude".
case "$CLAUDE_CONFIG_DIR" in
  "~") CLAUDE_CONFIG_DIR="${HOME}" ;;
  "~/"*) CLAUDE_CONFIG_DIR="${HOME}/${CLAUDE_CONFIG_DIR#"~/"}" ;;
esac
CLAUDE_COMMANDS_DIR="${CLAUDE_CONFIG_DIR}/commands"

# Auto-skip project-local installers when run from inside the devskills
# source repo — otherwise they write contributor files into the repo itself.
case "${PWD}/" in
  "${DEVSKILLS_DIR}"/*)
    if [ "$SKIP_CURSOR" -eq 0 ] || [ "$SKIP_VSCODE" -eq 0 ]; then
      warn "Running inside the devskills source repo; skipping Cursor/VSCode install."
    fi
    SKIP_CURSOR=1
    SKIP_VSCODE=1
    if [ -n "$LANG_PROFILE" ]; then
      warn "Running inside the devskills source repo; ignoring --lang to avoid writing CLAUDE.md into the repo."
      LANG_PROFILE=""
    fi
    ;;
esac

# AGENTS.md is only written when --lang is given (see install_lang_profile).
# Flag --concise/--hints used without --lang so they aren't a silent no-op.
if [ -z "$LANG_PROFILE" ] && { [ "$CONCISE" -eq 1 ] || [ "$HINTS" -eq 1 ]; }; then
  warn "--concise/--hints apply with --lang; nothing written to AGENTS.md. Use scripts/setup.sh for a baseline-only project."
fi

# Validate --lang up front, before any install side effects: a bad profile
# should fail fast, not after GSD/RTK/tldt are already installed.
if [ -n "$LANG_PROFILE" ] && [ ! -f "${DEVSKILLS_DIR}/prompts/language/${LANG_PROFILE}.md" ]; then
  warn "No language profile for '${LANG_PROFILE}'. Available: go, typescript, javascript, rust, python"
  exit 1
fi

# ------------------------------------------------------------
# Helpers
# ------------------------------------------------------------

install_file() {
  local src="$1"
  local dst="$2"
  if [ "$DRY_RUN" -eq 1 ]; then
    log "[dry] would install $src -> $dst"
    return
  fi
  mkdir -p "$(dirname "$dst")"
  cp "$src" "$dst"
  log "installed $dst"
}

# Commands removed or renamed in past releases. install only ever copies, so
# without this the old name lingers next to its replacement forever (e.g. after
# update.sh). Remove the known stale files from a target commands dir; only
# touches names devskills itself shipped, never user-authored commands.
#   frontend.md     -> ui.md (now ds-ui-mode.md)
#   write-a-skill.md -> write-a-command.md (now ds-write-a-command.md)
# Every command was namespaced with a `ds-` prefix (modes also gain a `-mode`
# suffix); the pre-prefix filenames below are retired here. New names all carry
# the `ds-` prefix, so none collide with the stale names being removed.
RENAMED_COMMANDS=(
  frontend.md write-a-skill.md
  bug-review.md caveman-lite.md caveman-ultra.md code-quality-review.md
  debug.md deslop.md doc-quality-review.md explore.md go-review.md grill-me.md
  handoff.md project-checkpoint.md project-map.md project-plan.md
  project-resume.md python-review.md quality-gate.md rust-review.md
  security-review.md spec.md tdd.md test-quality-review.md test.md
  tiger-style.md tldt.md ts-review.md ui-quality-review.md ui.md
  verify-this.md workflow.md write-a-command.md zoom-out.md
)

purge_renamed_commands() {
  local dir="$1" name
  for name in "${RENAMED_COMMANDS[@]}"; do
    [ -f "${dir}/${name}" ] || continue
    if [ "$DRY_RUN" -eq 1 ]; then
      log "[dry] would remove renamed command ${dir}/${name}"
    else
      rm -f "${dir}/${name}"
      log "removed renamed command ${dir}/${name}"
    fi
  done
}

# ------------------------------------------------------------
# Claude Code skills
# ------------------------------------------------------------

install_claude() {
  if command -v claude &>/dev/null || [ -d "${CLAUDE_CONFIG_DIR}" ]; then
    log "Installing Claude Code commands to ${CLAUDE_COMMANDS_DIR}"
    # install_file makes the dir on real copies; guard so --dry-run leaves none.
    [ "$DRY_RUN" -eq 1 ] || mkdir -p "${CLAUDE_COMMANDS_DIR}"
    for f in "${DEVSKILLS_DIR}/commands/"*.md; do
      install_file "$f" "${CLAUDE_COMMANDS_DIR}/$(basename "$f")"
    done
    purge_renamed_commands "${CLAUDE_COMMANDS_DIR}"
  else
    warn "Claude Code not detected. Skipping. Install from https://claude.ai/code"
  fi
}

# ------------------------------------------------------------
# OpenCode skills
# ------------------------------------------------------------

install_opencode() {
  if command -v opencode &>/dev/null || [ -d "${HOME}/.opencode" ]; then
    log "Installing OpenCode commands to ${OPENCODE_COMMANDS_DIR}"
    # install_file makes the dir on real copies; guard so --dry-run leaves none.
    [ "$DRY_RUN" -eq 1 ] || mkdir -p "${OPENCODE_COMMANDS_DIR}"
    for f in "${DEVSKILLS_DIR}/commands/"*.md; do
      install_file "$f" "${OPENCODE_COMMANDS_DIR}/$(basename "$f")"
    done
    purge_renamed_commands "${OPENCODE_COMMANDS_DIR}"
  else
    warn "OpenCode not detected. Skipping."
  fi
}

# ------------------------------------------------------------
# Language profile
# ------------------------------------------------------------

install_lang_profile() {
  local lang="$1"
  log "Writing AGENTS.md baseline${lang:+ + ${lang} profile} to ${PWD}"

  # shellcheck source=scripts/lib/profile.sh
  source "${DEVSKILLS_DIR}/scripts/lib/profile.sh"
  devskills_apply "${DEVSKILLS_DIR}/prompts" "$PWD" "$DRY_RUN" "$lang" "$CONCISE" "$HINTS" "$PHASES"
}

# ------------------------------------------------------------
# Cursor rules
# ------------------------------------------------------------

install_cursor() {
  if [ -d "${PWD}/.cursor" ] || command -v cursor &>/dev/null; then
    log "Installing Cursor rules to ${PWD}/.cursor/rules/"
    devskills_install_cursor "$PWD" "$LANG_PROFILE"
  else
    warn "Cursor not detected in current project. Run from a project directory with .cursor/ or with Cursor installed."
  fi
}

# ------------------------------------------------------------
# VSCode Copilot
# ------------------------------------------------------------

install_vscode() {
  if [ -d "${PWD}/.vscode" ] || command -v code &>/dev/null; then
    log "Installing VSCode Copilot instructions to ${PWD}/.github/copilot-instructions.md"
    devskills_install_vscode "$PWD" "$LANG_PROFILE"
  else
    warn "VSCode not detected in current project."
  fi
}

# ------------------------------------------------------------
# Main
# ------------------------------------------------------------

log "devskills installer"
log "source: ${DEVSKILLS_DIR}"

install_claude
install_opencode

if [ "$SKIP_CURSOR" -eq 0 ]; then
  install_cursor
else
  log "Skipping Cursor rules (--skip-cursor)"
fi

if [ "$SKIP_VSCODE" -eq 0 ]; then
  install_vscode
else
  log "Skipping VSCode Copilot instructions (--skip-vscode)"
fi

if [ "$SKIP_EXTERNAL" -eq 0 ]; then
  log "Installing external tools..."
  devskills_gsd install
  devskills_rtk install
  devskills_tldt install
else
  log "Skipping external tools (--skip-external)"
fi

if [ -n "$LANG_PROFILE" ]; then
  install_lang_profile "$LANG_PROFILE"
fi

log ""
log "Done. Verify with:"
log "  claude /ds-tiger-style-mode   — in Claude Code"
log "  /ds-tiger-style-mode          — in Cursor or OpenCode"
log "  rtk --version                 — RTK token proxy"
log "  tldt --version                — text summarizer"
log ""
log "Set language profile in any project:"
log "  ./install.sh --lang=go"
log "  ./install.sh --lang=typescript"
