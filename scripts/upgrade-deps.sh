#!/usr/bin/env bash
# upgrade-deps.sh: force-upgrade all external tools to their latest versions
#
# Unlike install.sh (idempotent, skips already-installed binaries), this
# script forces reinstall of every tool regardless of current state.
# Run it when tools feel stale or after a major version bump upstream.
set -euo pipefail

log()  { printf '[devskills:upgrade] %s\n' "$1"; }
warn() { printf '[devskills:upgrade] WARN: %s\n' "$1" >&2; }

DRY_RUN=0
for arg in "$@"; do
  case "$arg" in
    --dry-run) DRY_RUN=1 ;;
    --help|-h)
      echo "Usage: upgrade-deps.sh [--dry-run]"
      echo "Force-reinstalls GSD, RTK, and tldt to their latest published versions."
      exit 0
      ;;
  esac
done

# ------------------------------------------------------------
# GSD — npx always fetches from registry; @latest resolves fresh
# ------------------------------------------------------------
upgrade_gsd() {
  if command -v npx &>/dev/null; then
    log "Upgrading GSD — interactive, follow prompts..."
    if [ "$DRY_RUN" -eq 0 ]; then
      npx get-shit-done-cc@latest
    else
      log "DRY: would run npx get-shit-done-cc@latest"
    fi
  else
    warn "npx not found. Skip GSD upgrade."
  fi
}

# ------------------------------------------------------------
# RTK — macOS: brew upgrade. Linux: re-download binary.
# Never use cargo install: name collision with reachingforthejack/rtk.
# Verify correct binary first via 'rtk gain'.
# ------------------------------------------------------------
upgrade_rtk() {
  if command -v rtk &>/dev/null && ! rtk gain &>/dev/null; then
    warn "Wrong 'rtk' binary detected (not rtk-ai). Fix first:"
    warn "  cargo uninstall rtk   # remove wrong binary"
    warn "  Then re-run upgrade."
    return 1
  fi

  if [[ "$(uname)" == "Darwin" ]] && command -v brew &>/dev/null; then
    log "Upgrading RTK via Homebrew..."
    if [ "$DRY_RUN" -eq 0 ]; then
      brew upgrade rtk-ai/tap/rtk || brew install rtk-ai/tap/rtk || warn "RTK brew upgrade failed."
    else
      log "DRY: would run brew upgrade rtk-ai/tap/rtk"
    fi
  elif [[ "$(uname)" == "Linux" ]]; then
    log "Upgrading RTK via GitHub release (Linux)..."
    if [ "$DRY_RUN" -eq 0 ]; then
      local bin_dir="${HOME}/.local/bin"
      local arch; arch="$(uname -m)"
      local url="https://github.com/rtk-ai/rtk/releases/latest/download/rtk-${arch}-unknown-linux-musl"
      curl -fsSL "$url" -o "${bin_dir}/rtk" && chmod +x "${bin_dir}/rtk" \
        || warn "RTK binary download failed."
    else
      log "DRY: would download latest rtk-ai binary to ~/.local/bin/rtk"
    fi
  else
    warn "RTK: unsupported OS. Upgrade manually: https://github.com/rtk-ai/rtk"
  fi
}

# ------------------------------------------------------------
# tldt — go install @latest always fetches and installs the latest
# published module version, even if the binary already exists.
# ------------------------------------------------------------
upgrade_tldt() {
  if command -v go &>/dev/null; then
    log "Upgrading tldt..."
    if [ "$DRY_RUN" -eq 0 ]; then
      go install github.com/gleicon/tldt/cmd/tldt@latest
      log "tldt upgraded to $(go env GOPATH)/bin/tldt"
    else
      log "DRY: would run go install github.com/gleicon/tldt/cmd/tldt@latest"
    fi
  else
    warn "Go not found. Upgrade tldt manually: https://github.com/gleicon/tldt"
  fi
}

# ------------------------------------------------------------
# Caveman — managed by its own installer; re-run to upgrade
# ------------------------------------------------------------
upgrade_caveman() {
  if command -v node &>/dev/null; then
    log "Upgrading Caveman..."
    if [ "$DRY_RUN" -eq 0 ]; then
      curl -fsSL https://raw.githubusercontent.com/JuliusBrussee/caveman/main/install.sh | bash
    else
      log "DRY: would curl-reinstall caveman"
    fi
  else
    warn "Node not found. Upgrade Caveman manually: https://github.com/juliusbrussee/caveman"
  fi
}

log "Force-upgrading all external tools..."
upgrade_gsd
upgrade_rtk
upgrade_tldt
upgrade_caveman
log "Done."
