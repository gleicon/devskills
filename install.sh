#!/usr/bin/env bash
set -euo pipefail

DEVSKILLS_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CLAUDE_CONFIG_DIR="${CLAUDE_CONFIG_DIR:-${HOME}/.claude}"
OPENCODE_COMMANDS_DIR="${HOME}/.opencode/commands"
# Codex honors CODEX_HOME (default ~/.codex); custom prompts live in prompts/.
CODEX_HOME_DIR="${CODEX_HOME:-${HOME}/.codex}"
CODEX_COMMANDS_DIR="${CODEX_HOME_DIR}/prompts"
# Gemini CLI reads custom commands from ~/.gemini/commands as TOML (converted
# from commands/*.md at install time — see devskills_install_gemini).
GEMINI_DIR="${HOME}/.gemini"
GEMINI_COMMANDS_DIR="${GEMINI_DIR}/commands"

log() { printf '[devskills] %s\n' "$1"; }
warn() { printf '[devskills] WARN: %s\n' "$1" >&2; }

# Shared tldt logic (depends on log/warn above and DRY_RUN below).
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
    --phases) PHASES=1 ;;
    --dry-run) DRY_RUN=1 ;;
    --help|-h)
      echo "Usage: install.sh [--lang=go|typescript|javascript|rust|python|java|zig] [--claude-dir=PATH] [--skip-external] [--skip-cursor] [--skip-vscode] [--concise] [--phases] [--dry-run]"
      echo ""
      echo "  --lang=<profile>    Language profile to write: go|typescript|javascript|rust|python|java|zig"
      echo "  --claude-dir=PATH   Claude config dir (default: \$CLAUDE_CONFIG_DIR or \$HOME/.claude)"
      echo "  --skip-external     Skip external tool installation (tldt)"
      echo "  --skip-cursor       Skip Cursor rules install into the current project"
      echo "  --skip-vscode       Skip VSCode Copilot instructions install into the current project"
      echo "  --concise           Add a terse-response directive to AGENTS.md (with --lang)"
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
# Flag --concise used without --lang so it isn't a silent no-op.
if [ -z "$LANG_PROFILE" ] && [ "$CONCISE" -eq 1 ]; then
  warn "--concise applies with --lang; nothing written to AGENTS.md. Use scripts/setup.sh for a baseline-only project."
fi

# Validate --lang up front, before any install side effects: a bad profile
# should fail fast, not after tldt is already installed.
if [ -n "$LANG_PROFILE" ] && [ ! -f "${DEVSKILLS_DIR}/prompts/language/${LANG_PROFILE}.md" ]; then
  warn "No language profile for '${LANG_PROFILE}'. Available: go, typescript, javascript, rust, python, java, zig"
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
#   ds-project-plan.md -> ds-roadmap.md (a plan-generator, not `.project` memory —
#                         so it left the project-* family; a post-prefix rename)
#   ds-modes.md, ds-review.md -> removed (no replacement): the host question UI
#                         caps a picker at four options, but these launchers had
#                         ten modes / eight reviews, so the picker could never
#                         render — deleted rather than degraded to a prose menu.
# Every command was namespaced with a `ds-` prefix (modes also gain a `-mode`
# suffix); the pre-prefix filenames below are retired here, plus the one
# post-prefix rename and the two removed launchers above. New names all carry
# the `ds-` prefix and none collide with the stale names being removed.
RENAMED_COMMANDS=(
  frontend.md write-a-skill.md
  bug-review.md caveman-lite.md caveman-ultra.md code-quality-review.md
  debug.md deslop.md doc-quality-review.md explore.md go-review.md grill-me.md
  handoff.md project-checkpoint.md project-map.md project-plan.md
  project-resume.md python-review.md quality-gate.md rust-review.md
  security-review.md spec.md tdd.md test-quality-review.md test.md
  tiger-style.md tldt.md ts-review.md ui-quality-review.md ui.md
  verify-this.md workflow.md write-a-command.md zoom-out.md
  ds-project-plan.md ds-modes.md ds-review.md
)

# $1 dir  $2 extension (default .md) — the file extension this target ships.
# RENAMED_COMMANDS lists .md names; the Gemini target ships .toml, so it passes
# .toml and the stale name is matched with its extension swapped. One source
# list serves every target.
purge_renamed_commands() {
  local dir="$1" ext="${2:-.md}" name file
  for name in "${RENAMED_COMMANDS[@]}"; do
    file="${name%.md}${ext}"
    [ -f "${dir}/${file}" ] || continue
    if [ "$DRY_RUN" -eq 1 ]; then
      log "[dry] would remove renamed command ${dir}/${file}"
    else
      rm -f "${dir}/${file}"
      log "removed renamed command ${dir}/${file}"
    fi
  done
}

# ------------------------------------------------------------
# Claude Code commands
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
# OpenCode commands
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
# OpenAI Codex prompts
# ------------------------------------------------------------

# Codex reads project AGENTS.md natively (built by setup.sh/profile.sh), so only
# the command surface needs installing. Custom prompts are plain .md files in
# ${CODEX_HOME}/prompts, invoked namespaced as /prompts:<filename>. Like Claude
# and OpenCode, this is a global target with no opt-out — detection gates it.
install_codex() {
  if command -v codex &>/dev/null || [ -d "${CODEX_HOME_DIR}" ]; then
    log "Installing Codex prompts to ${CODEX_COMMANDS_DIR}"
    # install_file makes the dir on real copies; guard so --dry-run leaves none.
    [ "$DRY_RUN" -eq 1 ] || mkdir -p "${CODEX_COMMANDS_DIR}"
    for f in "${DEVSKILLS_DIR}/commands/"*.md; do
      install_file "$f" "${CODEX_COMMANDS_DIR}/$(basename "$f")"
    done
    purge_renamed_commands "${CODEX_COMMANDS_DIR}"
  else
    warn "Codex not detected. Skipping. Install from https://developers.openai.com/codex"
  fi
}

# ------------------------------------------------------------
# Gemini CLI commands
# ------------------------------------------------------------

# Gemini commands are TOML, not markdown, so this target converts rather than
# copies (devskills_install_gemini in editors.sh). Like Claude/OpenCode/Codex
# it is a global install with no opt-out; detection gates it. The conversion
# writes nothing under --dry-run and creates the dir per-file on real runs.
install_gemini() {
  if command -v gemini &>/dev/null || [ -d "${GEMINI_DIR}" ]; then
    log "Installing Gemini CLI commands to ${GEMINI_COMMANDS_DIR}"
    devskills_install_gemini "${GEMINI_COMMANDS_DIR}"
    purge_renamed_commands "${GEMINI_COMMANDS_DIR}" .toml
  else
    warn "Gemini CLI not detected. Skipping. Install from https://github.com/google-gemini/gemini-cli"
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
  devskills_apply "${DEVSKILLS_DIR}/prompts" "$PWD" "$DRY_RUN" "$lang" "$CONCISE" "$PHASES"
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
install_codex
install_gemini

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
  devskills_osv install
  devskills_tldt install
else
  log "Skipping external tools (--skip-external)"
fi

# RTK was removed after an upstream supply-chain compromise. Warn (never delete)
# if an earlier devskills install left it behind. Runs even with --skip-external,
# since skipping the install does not make an already-present compromised binary safe.
devskills_rtk_remediate

if [ -n "$LANG_PROFILE" ]; then
  install_lang_profile "$LANG_PROFILE"
fi

log ""
log "Done. Verify with:"
log "  claude /ds-tiger-style-mode   — in Claude Code"
log "  /ds-tiger-style-mode          — in Cursor or OpenCode"
log "  /prompts:ds-tiger-style-mode  — in Codex"
log "  osv-scanner --version         — supply-chain vulnerability scanner"
log "  tldt --version                — text summarizer"
log ""
log "Set language profile in any project:"
log "  ./install.sh --lang=go"
log "  ./install.sh --lang=typescript"
