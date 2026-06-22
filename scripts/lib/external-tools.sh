# external-tools.sh — shared tldt installer for devskills.
#
# Sourced by install.sh and scripts/upgrade-deps.sh. `go install @latest`
# always fetches the newest published version, so the command is the same
# for both install and upgrade modes.
#
# Logging: install.sh and upgrade-deps.sh each define log()/warn() with their
# own prefix; this lib uses theirs (resolved at call time). DRY_RUN (0|1) is
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
# osv-scanner (Google OSV vulnerability scanner)
# ------------------------------------------------------------

# Install or upgrade osv-scanner. Prefers Homebrew on macOS; falls back to
# `go install` when Go is present.
#   $1 mode (install|upgrade)
devskills_osv() {
  local mode="$1" verbing="Installing" verb="Install"
  [ "$mode" = upgrade ] && { verbing="Upgrading"; verb="Upgrade"; }

  if [ "$mode" = install ] && command -v osv-scanner &>/dev/null; then
    log "osv-scanner already installed at $(command -v osv-scanner). Skipping."
    return 0
  fi

  if [[ "$(uname)" == "Darwin" ]] && command -v brew &>/dev/null; then
    log "${verbing} osv-scanner via Homebrew..."
    if [ "$mode" = upgrade ]; then
      run_or_dry "run brew upgrade osv-scanner" \
        brew upgrade osv-scanner || brew install osv-scanner || warn "osv-scanner brew upgrade failed."
    else
      run_or_dry "run brew install osv-scanner" \
        brew install osv-scanner || warn "osv-scanner brew install failed. See: https://github.com/google/osv-scanner"
    fi
    return 0
  fi

  if command -v go &>/dev/null; then
    log "${verbing} osv-scanner via go install..."
    run_or_dry "run go install github.com/google/osv-scanner/cmd/osv-scanner@latest" \
      go install github.com/google/osv-scanner/cmd/osv-scanner@latest
    if [ "${DRY_RUN:-0}" -eq 0 ]; then
      log "osv-scanner ready at $(go env GOPATH)/bin/osv-scanner"
    fi
    return 0
  fi

  warn "osv-scanner: needs Homebrew (macOS) or Go. ${verb} from releases: https://github.com/google/osv-scanner/releases"
}

# ------------------------------------------------------------
# RTK removal remediation (warn-only)
# ------------------------------------------------------------

# RTK (rtk-ai token proxy) was removed from devskills after an upstream
# supply-chain compromise. Earlier devskills versions installed it via the
# Homebrew tap (macOS) or a binary in ~/.local/bin (Linux). We never auto-delete
# a binary we can't prove we own, so this detects and warns only — it stays
# silent when there is nothing to remediate. Runs regardless of --skip-external.
devskills_rtk_remediate() {
  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    log "[dry] would check for a leftover RTK install and warn if found"
    return 0
  fi

  # Detection is read-only: `command -v` and `readlink` only. We deliberately do
  # NOT invoke `brew` — it mutates its cache under $HOME, which would surprise
  # the user (and breaks dry-run purity).
  local rtk_path
  rtk_path="$(command -v rtk 2>/dev/null)" || return 0   # nothing to remediate; stay silent

  warn "SECURITY: RTK was removed from devskills after an upstream supply-chain compromise."

  # If the binary resolves into a Homebrew Cellar it is brew-managed, so the
  # correct removal is `brew uninstall <formula>`, not `rm` on brew's symlink.
  # The single-level symlink target (e.g. ../Cellar/rtk/0.42.0/bin/rtk) is enough
  # to both detect this and recover the formula name, portably (no readlink -f).
  local target
  target="$(readlink "$rtk_path" 2>/dev/null || true)"
  case "${target:-$rtk_path}" in
    */Cellar/*)
      local formula="${target##*/Cellar/}"; formula="${formula%%/*}"
      warn "  A Homebrew-managed RTK is still present (${rtk_path}). Remove it manually:"
      warn "    brew uninstall ${formula}"
      ;;
    *)
      warn "  A binary named 'rtk' is on your PATH at ${rtk_path}. If devskills installed it, remove it manually:"
      warn "    rm -f ${rtk_path}"
      ;;
  esac
  return 0
}

# ------------------------------------------------------------
# ast-grep (structural code search and rewrite)
# ------------------------------------------------------------

# Install or upgrade ast-grep. Prefers Homebrew on macOS; falls back to
# npm when Node is present (works on Linux too).
#   $1 mode (install|upgrade)
devskills_astgrep() {
  local mode="$1" verbing="Installing" verb="Install"
  [ "$mode" = upgrade ] && { verbing="Upgrading"; verb="Upgrade"; }

  if [ "$mode" = install ] && command -v ast-grep &>/dev/null; then
    log "ast-grep already installed at $(command -v ast-grep). Skipping."
    return 0
  fi

  if [[ "$(uname)" == "Darwin" ]] && command -v brew &>/dev/null; then
    log "${verbing} ast-grep via Homebrew..."
    if [ "$mode" = upgrade ]; then
      run_or_dry "run brew upgrade ast-grep" \
        brew upgrade ast-grep || brew install ast-grep || warn "ast-grep brew upgrade failed."
    else
      run_or_dry "run brew install ast-grep" \
        brew install ast-grep || warn "ast-grep brew install failed. See: https://github.com/ast-grep/ast-grep"
    fi
    return 0
  fi

  if command -v npm &>/dev/null; then
    log "${verbing} ast-grep via npm..."
    run_or_dry "run npm install -g @ast-grep/cli" \
      npm install -g @ast-grep/cli || warn "ast-grep npm install failed."
    return 0
  fi

  warn "ast-grep: needs Homebrew (macOS) or npm. ${verb} from releases: https://github.com/ast-grep/ast-grep/releases"
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
