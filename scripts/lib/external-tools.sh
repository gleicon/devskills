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
# Logging: install.sh and upgrade-deps.sh each define log()/warn() with their
# own prefix and this lib uses theirs (resolved at call time). DRY_RUN (0|1) is
# honored; defaults to 0.

[ -n "${DEVSKILLS_EXTERNAL_LIB:-}" ] && return 0
DEVSKILLS_EXTERNAL_LIB=1

# Run a command, or just describe it under --dry-run.
#   $1 human-readable description  $2.. command + args
run_or_dry() {
  local desc="$1"; shift
  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log "[dry] would ${desc}"
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
      log "[dry] would remove old GSD hook: $f"
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

# True (0) if any archive member would escape the extract dir — an absolute
# path or a '..' component. Input: a newline-separated member list (from
# `tar -tzf`). Pure: status only, no output. Kept as a here-string match rather
# than `tar | grep -q`, whose SIGPIPE under `set -o pipefail` inverts the test.
#   $1 member list
_rtk_archive_has_unsafe_paths() {
  grep -Eq '(^|/)\.\.(/|$)|^/' <<<"$1"
}

# True (0) if <dir> holds a real, top-level 'rtk' binary — a regular file, not
# a symlink that could point elsewhere. Pure: status only.
#   $1 staged dir
_rtk_staged_binary_ok() {
  [ -f "${1}/rtk" ] && [ ! -L "${1}/rtk" ]
}

# Print the sha256 recorded for <asset> in a checksums.txt body (coreutils
# format: "<hash>  <name>" per line), or nothing if the asset isn't listed.
# Pure: stdout only.
#   $1 checksums body  $2 asset name
_rtk_expected_sha256() {
  awk -v a="$2" '$2 == a { print $1; exit }' <<<"$1"
}

# Download, verify, and safely extract the latest rtk-ai release into
# ~/.local/bin. Identical for install and upgrade. Honors DRY_RUN.
#
# The archive is untrusted until validated (see the step comments below), so a
# tampered release can't write outside ~/.local/bin.
_rtk_linux_download() {
  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log "[dry] would download, verify, and extract the latest rtk-ai release to ~/.local/bin/rtk"
    return 0
  fi
  local arch target
  arch="$(uname -m)"
  case "$arch" in
    x86_64)        target="x86_64-unknown-linux-musl" ;;
    aarch64|arm64) target="aarch64-unknown-linux-gnu" ;;
    *) warn "RTK: unsupported Linux arch '${arch}'. Install manually: https://github.com/rtk-ai/rtk"; return 1 ;;
  esac
  local base_url="https://github.com/rtk-ai/rtk/releases/latest/download"
  local asset="rtk-${target}.tar.gz"
  local tmp; tmp="$(mktemp -d)"
  trap 'rm -rf "$tmp"' RETURN

  if ! curl -fsSL "${base_url}/${asset}" -o "${tmp}/${asset}"; then
    warn "RTK download failed. Install manually: https://github.com/rtk-ai/rtk"
    return 1
  fi

  # Integrity: verify against the release's checksums.txt. This binary is the
  # only thing the script writes into $HOME, so refuse anything unverifiable —
  # a missing checksums.txt, an unlisted asset, or a mismatch all abort.
  # (Upstream ships one checksums.txt, not a per-asset .sha256.)
  if ! curl -fsSL "${base_url}/checksums.txt" -o "${tmp}/checksums.txt" 2>/dev/null; then
    warn "RTK: no published checksums.txt — refusing to install an unverifiable binary. Install manually: https://github.com/rtk-ai/rtk"
    return 1
  fi
  local want got
  want="$(_rtk_expected_sha256 "$(cat "${tmp}/checksums.txt")" "$asset")"
  got="$(sha256sum "${tmp}/${asset}" | awk '{print $1}')"
  if [ -z "$want" ]; then
    warn "RTK: ${asset} not listed in checksums.txt — refusing to install an unverifiable binary. Install manually: https://github.com/rtk-ai/rtk"
    return 1
  fi
  if [ "$want" != "$got" ]; then
    warn "RTK checksum mismatch — refusing to install (expected ${want}, got ${got})."
    return 1
  fi

  # Refuse any member that could escape the extract dir before extracting a
  # single byte.
  local members
  if ! members="$(tar -tzf "${tmp}/${asset}")"; then
    warn "RTK: could not read archive. Install manually: https://github.com/rtk-ai/rtk"
    return 1
  fi
  if _rtk_archive_has_unsafe_paths "$members"; then
    warn "RTK archive contains unsafe paths — refusing to extract. Install manually: https://github.com/rtk-ai/rtk"
    return 1
  fi

  # Extract into an isolated dir, then take only the rtk binary — and only if
  # it's a real file, not a symlink pointing elsewhere.
  local stage="${tmp}/extract"
  mkdir -p "$stage"
  if ! tar -xzf "${tmp}/${asset}" -C "$stage" --no-same-owner; then
    warn "RTK extract failed. Install manually: https://github.com/rtk-ai/rtk"
    return 1
  fi
  if ! _rtk_staged_binary_ok "$stage"; then
    warn "RTK archive did not contain the expected 'rtk' binary. Install manually: https://github.com/rtk-ai/rtk"
    return 1
  fi

  mkdir -p "${HOME}/.local/bin"
  install -m 0755 "${stage}/rtk" "${HOME}/.local/bin/rtk"
  log "RTK installed to ${HOME}/.local/bin/rtk"
  log "Ensure ${HOME}/.local/bin is on your PATH."
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
        log "[dry] would run brew upgrade rtk-ai/tap/rtk"
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
