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
# GSD Redux — purge old gsd-build hooks, then npx @latest
# Old GSD installed hooks with a gsd-hook-version: header; remove before upgrading.
# ------------------------------------------------------------
purge_old_gsd_hooks() {
  local hooks_dir="${HOME}/.claude/hooks"
  [ -d "$hooks_dir" ] || return 0
  local found=0
  for f in "$hooks_dir"/*.sh "$hooks_dir"/*.js; do
    [ -f "$f" ] || continue
    if grep -q "gsd-hook-version:" "$f" 2>/dev/null; then
      if [ "$DRY_RUN" -eq 0 ]; then
        rm "$f"
        log "removed old GSD hook: $(basename "$f")"
      else
        log "DRY: would remove old GSD hook: $f"
      fi
      found=1
    fi
  done
  [ "$found" -eq 1 ] && log "Old GSD hooks removed. Redux will reinstall fresh hooks."
  return 0
}

upgrade_gsd() {
  purge_old_gsd_hooks
  if command -v npx &>/dev/null; then
    log "Upgrading GSD Redux — interactive, follow prompts..."
    if [ "$DRY_RUN" -eq 0 ]; then
      npx @opengsd/get-shit-done-redux@latest
    else
      log "DRY: would run npx @opengsd/get-shit-done-redux@latest"
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
      mkdir -p "$bin_dir"
      local arch target; arch="$(uname -m)"
      case "$arch" in
        x86_64)        target="x86_64-unknown-linux-musl" ;;
        aarch64|arm64) target="aarch64-unknown-linux-gnu" ;;
        *) warn "RTK: unsupported Linux arch '${arch}'. Upgrade manually: https://github.com/rtk-ai/rtk"; return ;;
      esac
      local url="https://github.com/rtk-ai/rtk/releases/latest/download/rtk-${target}.tar.gz"
      local tmp; tmp="$(mktemp -d)"
      if curl -fsSL "$url" -o "${tmp}/rtk.tar.gz" && tar -xzf "${tmp}/rtk.tar.gz" -C "$bin_dir"; then
        chmod +x "${bin_dir}/rtk"
      else
        warn "RTK download failed."
      fi
      rm -rf "$tmp"
    else
      log "DRY: would download and extract latest rtk-ai release to ~/.local/bin/rtk"
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

log "Force-upgrading all external tools..."
upgrade_gsd
upgrade_rtk
upgrade_tldt
log "Caveman: bundled in devskills prompt files. Run scripts/update.sh to update."
log "Done."
