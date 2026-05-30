# external-tools.sh — shared GSD/RTK/tldt logic for devskills.
#
# Sourced by install.sh and scripts/upgrade-deps.sh. Each public function takes
# a mode (install|upgrade). The underlying commands are identical for GSD and
# tldt (both fetch @latest); they differ only for RTK, where:
#   - install short-circuits when a correct rtk-ai binary is already present;
#   - install uses `brew install`, upgrade uses `brew upgrade || brew install`.
# The Linux download path and the old-hook purge are byte-identical either way,
# so they live here once instead of being copy-pasted into both scripts.
#
# Contract: the sourcing script MUST define log() and warn(), and MAY set
# DRY_RUN (0|1; defaults to 0). They are resolved dynamically at call time, so
# each caller keeps its own log prefix.

[ -n "${DEVSKILLS_EXTERNAL_LIB:-}" ] && return 0
DEVSKILLS_EXTERNAL_LIB=1

# Run a command, or just describe it under --dry-run.
#   $1 human-readable description  $2.. command + args
run_or_dry() {
  local desc="$1"; shift
  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log "DRY: would ${desc}"
    return 0
  fi
  "$@"
}

# ------------------------------------------------------------
# GSD Redux
# ------------------------------------------------------------

# Remove hooks left by pre-Redux GSD, identified by a gsd-hook-version: header.
# CLAUDE_CONFIG_DIR is honored when set, else falls back to ~/.claude.
purge_old_gsd_hooks() {
  local hooks_dir="${CLAUDE_CONFIG_DIR:-${HOME}/.claude}/hooks"
  [ -d "$hooks_dir" ] || return 0
  local found=0 f
  for f in "$hooks_dir"/*.sh "$hooks_dir"/*.js; do
    [ -f "$f" ] || continue
    grep -q "gsd-hook-version:" "$f" 2>/dev/null || continue
    if [ "${DRY_RUN:-0}" -eq 0 ]; then
      rm "$f"
      log "removed old GSD hook: $(basename "$f")"
    else
      log "DRY: would remove old GSD hook: $f"
    fi
    found=1
  done
  [ "$found" -eq 1 ] && log "Old GSD hooks removed. Redux will reinstall fresh hooks."
  return 0
}

# Install or upgrade GSD Redux. The npx command is the same for both modes.
#   $1 mode (install|upgrade)
devskills_gsd() {
  local mode="$1" verbing="Installing" verb="Install"
  [ "$mode" = upgrade ] && { verbing="Upgrading"; verb="Upgrade"; }

  purge_old_gsd_hooks
  if command -v npx &>/dev/null; then
    log "${verbing} GSD Redux — interactive, follow prompts..."
    run_or_dry "run npx @opengsd/get-shit-done-redux@latest" \
      npx @opengsd/get-shit-done-redux@latest
  else
    warn "npx not found. ${verb} GSD manually: https://github.com/open-gsd/get-shit-done-redux"
  fi
}

# ------------------------------------------------------------
# RTK (rtk-ai token proxy)
# ------------------------------------------------------------

# Download + extract the latest rtk-ai release into ~/.local/bin. Identical for
# install and upgrade. Honors DRY_RUN.
_rtk_linux_download() {
  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log "DRY: would download and extract latest rtk-ai release to ~/.local/bin/rtk"
    return 0
  fi
  local bin_dir="${HOME}/.local/bin"
  mkdir -p "$bin_dir"
  local arch target
  arch="$(uname -m)"
  case "$arch" in
    x86_64)        target="x86_64-unknown-linux-musl" ;;
    aarch64|arm64) target="aarch64-unknown-linux-gnu" ;;
    *) warn "RTK: unsupported Linux arch '${arch}'. Install manually: https://github.com/rtk-ai/rtk"; return 1 ;;
  esac
  local url="https://github.com/rtk-ai/rtk/releases/latest/download/rtk-${target}.tar.gz"
  local tmp; tmp="$(mktemp -d)"
  if curl -fsSL "$url" -o "${tmp}/rtk.tar.gz" && tar -xzf "${tmp}/rtk.tar.gz" -C "$bin_dir"; then
    chmod +x "${bin_dir}/rtk"
    log "RTK installed to ${bin_dir}/rtk"
    log "Ensure ${bin_dir} is on your PATH."
  else
    warn "RTK download failed. Install manually: https://github.com/rtk-ai/rtk"
  fi
  rm -rf "$tmp"
}

# Install or upgrade RTK. Avoids `cargo install` deliberately: the
# reachingforthejack/rtk crate ships a binary also named 'rtk', so we detect
# that collision via `rtk gain` (only rtk-ai has it) before doing anything.
#   $1 mode (install|upgrade)
devskills_rtk() {
  local mode="$1" verbing="Installing" verb="Install"
  [ "$mode" = upgrade ] && { verbing="Upgrading"; verb="Upgrade"; }

  if command -v rtk &>/dev/null; then
    if rtk gain &>/dev/null; then
      # Correct binary already present. Install is done; upgrade proceeds.
      if [ "$mode" = install ]; then
        log "RTK (rtk-ai) already installed at $(command -v rtk). Skipping."
        return 0
      fi
    else
      warn "A binary named 'rtk' exists at $(command -v rtk) but is NOT rtk-ai (token proxy)."
      warn "This is likely reachingforthejack/rtk (Rust toolkit) — a known name collision."
      warn "Fix it, then re-run:"
      warn "  cargo uninstall rtk"
      return 1
    fi
  fi

  # macOS: Homebrew avoids the cargo name collision entirely.
  if [[ "$(uname)" == "Darwin" ]] && command -v brew &>/dev/null; then
    log "${verbing} RTK via Homebrew (macOS)..."
    if [ "$mode" = upgrade ]; then
      if [ "${DRY_RUN:-0}" -eq 1 ]; then
        log "DRY: would run brew upgrade rtk-ai/tap/rtk"
      else
        brew upgrade rtk-ai/tap/rtk || brew install rtk-ai/tap/rtk || warn "RTK brew upgrade failed."
      fi
    else
      run_or_dry "run brew install rtk-ai/tap/rtk" \
        brew install rtk-ai/tap/rtk || warn "RTK brew install failed. See: https://github.com/rtk-ai/rtk"
    fi
    return 0
  elif [[ "$(uname)" == "Linux" ]]; then
    log "${verbing} RTK via GitHub release (Linux)..."
    _rtk_linux_download
    return 0
  fi

  warn "RTK: unsupported OS or no package manager. ${verb} manually: https://github.com/rtk-ai/rtk"
}

# ------------------------------------------------------------
# tldt
# ------------------------------------------------------------

# Install or upgrade tldt. `go install @latest` always fetches the newest
# published version, so the command is the same for both modes.
#   $1 mode (install|upgrade)
devskills_tldt() {
  local mode="$1" verbing="Installing" verb="Install"
  [ "$mode" = upgrade ] && { verbing="Upgrading"; verb="Upgrade"; }

  if command -v go &>/dev/null; then
    log "${verbing} tldt..."
    run_or_dry "run go install github.com/gleicon/tldt/cmd/tldt@latest" \
      go install github.com/gleicon/tldt/cmd/tldt@latest
    if [ "${DRY_RUN:-0}" -eq 0 ]; then
      log "tldt ready at $(go env GOPATH)/bin/tldt"
    fi
  else
    warn "Go not found. ${verb} tldt manually: https://github.com/gleicon/tldt"
  fi
}
