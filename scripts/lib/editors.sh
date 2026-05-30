# editors.sh — shared editor-rule installer for devskills (Cursor + VSCode).
#
# Sourced by install.sh and scripts/setup.sh. Each script keeps its own
# decision of *whether* to install (install.sh auto-detects and is opt-out via
# --skip-cursor; setup.sh is opt-in via --cursor); this lib owns *how*, so the
# mechanics live in one place instead of drifting between the two scripts.
#
# Cursor rules are curated, not dumped: always tiger-style, plus the single
# rule matching the language profile (javascript reuses the typescript rule).
# A project's .cursor/rules/ stays scoped to what it actually uses.
#
# Contract: DEVSKILLS_DIR points at the devskills source root. DRY_RUN (0|1)
# is honored; defaults to 0.

[ -n "${DEVSKILLS_EDITORS_LIB:-}" ] && return 0
DEVSKILLS_EDITORS_LIB=1

_dske_log() { printf '[devskills] %s\n' "$1"; }

# Copy one file into place, creating its parent dir. Honors DRY_RUN.
#   $1 src  $2 dst
_dske_copy() {
  local src="$1" dst="$2"
  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    _dske_log "[dry] would write ${dst}"
    return 0
  fi
  mkdir -p "$(dirname "$dst")"
  cp "$src" "$dst"
  _dske_log "wrote ${dst}"
}

# Install Cursor rules into <dir>/.cursor/rules: tiger-style plus the rule
# matching <lang> (an empty lang installs tiger-style only).
#   $1 target dir  $2 lang ("" for none)
devskills_install_cursor() {
  local dir="$1" lang="$2"
  local rules="${DEVSKILLS_DIR}/cursor/rules"
  _dske_copy "${rules}/tiger-style.mdc" "${dir}/.cursor/rules/tiger-style.mdc"
  case "$lang" in
    go)                    _dske_copy "${rules}/go.mdc"         "${dir}/.cursor/rules/go.mdc" ;;
    typescript|javascript) _dske_copy "${rules}/typescript.mdc" "${dir}/.cursor/rules/typescript.mdc" ;;
    rust)                  _dske_copy "${rules}/rust.mdc"       "${dir}/.cursor/rules/rust.mdc" ;;
  esac
}

# Install VSCode Copilot instructions into <dir>/.github.
#   $1 target dir
devskills_install_vscode() {
  _dske_copy \
    "${DEVSKILLS_DIR}/vscode/copilot-instructions.md" \
    "${1}/.github/copilot-instructions.md"
}
