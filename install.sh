#!/usr/bin/env bash
set -euo pipefail

DEVSKILLS_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CLAUDE_CONFIG_DIR="${CLAUDE_CONFIG_DIR:-${HOME}/.claude}"
OPENCODE_COMMANDS_DIR="${HOME}/.opencode/commands"

log() { printf '[devskills] %s\n' "$1"; }
warn() { printf '[devskills] WARN: %s\n' "$1" >&2; }

# ------------------------------------------------------------
# Arguments
# ------------------------------------------------------------

LANG_PROFILE=""
SKIP_EXTERNAL=0
SKIP_CURSOR=0
SKIP_VSCODE=0
DRY_RUN=0

for arg in "$@"; do
  case "$arg" in
    --lang=*) LANG_PROFILE="${arg#--lang=}" ;;
    --claude-dir=*) CLAUDE_CONFIG_DIR="${arg#--claude-dir=}" ;;
    --skip-external) SKIP_EXTERNAL=1 ;;
    --skip-cursor) SKIP_CURSOR=1 ;;
    --skip-vscode) SKIP_VSCODE=1 ;;
    --dry-run) DRY_RUN=1 ;;
    --help|-h)
      echo "Usage: install.sh [--lang=go|typescript|javascript|rust] [--claude-dir=PATH] [--skip-external] [--skip-cursor] [--skip-vscode] [--dry-run]"
      echo ""
      echo "  --claude-dir=PATH   Claude config dir (default: \$CLAUDE_CONFIG_DIR or \$HOME/.claude)"
      echo "  --skip-external     Skip external tool installation (GSD, RTK, tldt)"
      echo "  --skip-cursor       Skip Cursor rules install into the current project"
      echo "  --skip-vscode       Skip VSCode Copilot instructions install into the current project"
      exit 0
      ;;
  esac
done

# Expand leading ~ in --claude-dir value
case "$CLAUDE_CONFIG_DIR" in
  "~") CLAUDE_CONFIG_DIR="${HOME}" ;;
  "~/"*) CLAUDE_CONFIG_DIR="${HOME}/${CLAUDE_CONFIG_DIR#~/}" ;;
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

# ------------------------------------------------------------
# Helpers
# ------------------------------------------------------------

install_file() {
  local src="$1"
  local dst="$2"
  if [ "$DRY_RUN" -eq 1 ]; then
    log "DRY: would install $src -> $dst"
    return
  fi
  mkdir -p "$(dirname "$dst")"
  cp "$src" "$dst"
  log "installed $dst"
}

# ------------------------------------------------------------
# Claude Code skills
# ------------------------------------------------------------

install_claude() {
  if command -v claude &>/dev/null || [ -d "${CLAUDE_CONFIG_DIR}" ]; then
    log "Installing Claude Code commands to ${CLAUDE_COMMANDS_DIR}"
    mkdir -p "${CLAUDE_COMMANDS_DIR}"
    for f in "${DEVSKILLS_DIR}/claude/commands/"*.md; do
      install_file "$f" "${CLAUDE_COMMANDS_DIR}/$(basename "$f")"
    done
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
    mkdir -p "${OPENCODE_COMMANDS_DIR}"
    for f in "${DEVSKILLS_DIR}/opencode/commands/"*.md; do
      install_file "$f" "${OPENCODE_COMMANDS_DIR}/$(basename "$f")"
    done
  else
    warn "OpenCode not detected. Skipping."
  fi
}

# ------------------------------------------------------------
# External tools
# ------------------------------------------------------------

install_gsd() {
  if command -v npx &>/dev/null; then
    log "Installing GSD (Get Shit Done) — interactive, follow prompts..."
    if [ "$DRY_RUN" -eq 0 ]; then
      npx get-shit-done-cc@latest
    else
      log "DRY: would run npx get-shit-done-cc@latest"
    fi
  else
    warn "npx not found. Install GSD manually: https://github.com/gsd-build/get-shit-done"
  fi
}

install_rtk() {
  # Detect name collision: reachingforthejack/rtk installs a binary also called 'rtk'
  # but has no 'gain' subcommand. Test for rtk-ai by probing 'rtk gain'.
  if command -v rtk &>/dev/null; then
    if rtk gain &>/dev/null; then
      log "RTK (rtk-ai) already installed at $(command -v rtk). Skipping."
      return 0
    else
      warn "A binary named 'rtk' exists at $(command -v rtk) but is NOT rtk-ai (token proxy)."
      warn "This is likely reachingforthejack/rtk (Rust toolkit) — a known name collision."
      warn "To fix:"
      warn "  1. Remove the wrong binary:  cargo uninstall rtk"
      warn "  2. Re-run install:            ./install.sh (rtk-ai will then install via brew)"
      warn "Skipping RTK install to avoid shadowing."
      return 1
    fi
  fi

  # macOS: Homebrew is the preferred path — avoids cargo name collision
  if [[ "$(uname)" == "Darwin" ]] && command -v brew &>/dev/null; then
    log "Installing RTK via Homebrew (macOS)..."
    if [ "$DRY_RUN" -eq 0 ]; then
      brew install rtk-ai/tap/rtk || warn "RTK brew install failed. See: https://github.com/rtk-ai/rtk"
    else
      log "DRY: would run brew install rtk-ai/tap/rtk"
    fi
    return

  # Linux: download prebuilt binary from GitHub releases
  elif [[ "$(uname)" == "Linux" ]]; then
    log "Installing RTK via GitHub release (Linux)..."
    if [ "$DRY_RUN" -eq 0 ]; then
      local bin_dir="${HOME}/.local/bin"
      mkdir -p "$bin_dir"
      local arch
      arch="$(uname -m)"
      local asset="rtk-${arch}-unknown-linux-musl"
      local url="https://github.com/rtk-ai/rtk/releases/latest/download/${asset}"
      if curl -fsSL "$url" -o "${bin_dir}/rtk" 2>/dev/null; then
        chmod +x "${bin_dir}/rtk"
        log "RTK installed to ${bin_dir}/rtk"
        log "Ensure ${bin_dir} is on your PATH."
      else
        warn "RTK binary download failed. Install manually: https://github.com/rtk-ai/rtk"
      fi
    else
      log "DRY: would download rtk-ai binary to ~/.local/bin/rtk"
    fi
    return
  fi

  warn "RTK: unsupported OS or no package manager. Install manually: https://github.com/rtk-ai/rtk"
}

install_tldt() {
  if command -v go &>/dev/null; then
    log "Installing tldt..."
    if [ "$DRY_RUN" -eq 0 ]; then
      go install github.com/gleicon/tldt/cmd/tldt@latest
      log "tldt installed to $(go env GOPATH)/bin/tldt"
    else
      log "DRY: would run go install github.com/gleicon/tldt/cmd/tldt@latest"
    fi
  else
    warn "Go not found. Install tldt manually: https://github.com/gleicon/tldt"
  fi
}

# ------------------------------------------------------------
# Language profile
# ------------------------------------------------------------

install_lang_profile() {
  local lang="$1"
  local profile_file="${DEVSKILLS_DIR}/prompts/language/${lang}.md"

  if [ ! -f "$profile_file" ]; then
    warn "No language profile for '${lang}'. Available: go, typescript, javascript, rust"
    return 1
  fi

  log "Setting up language profile: ${lang}"

  # Write profile marker for Claude Code CLAUDE.md
  local claude_md="${PWD}/CLAUDE.md"
  if [ "$DRY_RUN" -eq 0 ]; then
    if [ -f "$claude_md" ]; then
      # Append if not already present
      if ! grep -q "devskills language profile" "$claude_md" 2>/dev/null; then
        {
          echo ""
          echo "<!-- devskills language profile: ${lang} -->"
          cat "$profile_file"
        } >> "$claude_md"
        log "Appended ${lang} profile to ${claude_md}"
      else
        log "Language profile already in ${claude_md}"
      fi
    else
      {
        echo "<!-- devskills language profile: ${lang} -->"
        cat "$profile_file"
      } > "$claude_md"
      log "Created ${claude_md} with ${lang} profile"
    fi

    # Write .devskills/language marker
    mkdir -p "${PWD}/.devskills"
    echo "${lang}" > "${PWD}/.devskills/language"
    log "Wrote .devskills/language: ${lang}"
  else
    log "DRY: would write ${lang} profile to CLAUDE.md and .devskills/language"
  fi
}

# ------------------------------------------------------------
# Cursor rules
# ------------------------------------------------------------

install_cursor() {
  if [ -d "${PWD}/.cursor" ] || command -v cursor &>/dev/null; then
    log "Installing Cursor rules to ${PWD}/.cursor/rules/"
    mkdir -p "${PWD}/.cursor/rules"
    for f in "${DEVSKILLS_DIR}/cursor/rules/"*.mdc; do
      install_file "$f" "${PWD}/.cursor/rules/$(basename "$f")"
    done
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
    install_file \
      "${DEVSKILLS_DIR}/vscode/copilot-instructions.md" \
      "${PWD}/.github/copilot-instructions.md"
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
  install_gsd
  install_rtk
  install_tldt
else
  log "Skipping external tools (--skip-external)"
fi

if [ -n "$LANG_PROFILE" ]; then
  install_lang_profile "$LANG_PROFILE"
fi

log ""
log "Done. Verify with:"
log "  claude /tiger-style    — in Claude Code"
log "  /tiger-style           — in Cursor or OpenCode"
log "  rtk --version          — RTK token proxy"
log "  tldt --version         — text summarizer"
log ""
log "Set language profile in any project:"
log "  ./install.sh --lang=go"
log "  ./install.sh --lang=typescript"
