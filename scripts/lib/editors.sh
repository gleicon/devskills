# editors.sh — shared editor-rule installer for devskills (Cursor + VSCode).
#
# Sourced by install.sh and scripts/setup.sh. Each script keeps its own
# decision of *whether* to install (install.sh auto-detects and is opt-out via
# --skip-cursor; setup.sh is opt-in via --cursor); this lib owns *how*, so the
# mechanics live in one place instead of drifting between the two scripts.
#
# Cursor rules are curated, not dumped: always tiger-style and context, plus the
# single rule matching the language profile (javascript reuses the typescript
# rule). A project's .cursor/rules/ stays scoped to what it actually uses.
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

# Install Cursor rules into <dir>/.cursor/rules: tiger-style and context, plus
# the rule matching <lang> (an empty lang installs the always-apply rules only).
#   $1 target dir  $2 lang ("" for none)
devskills_install_cursor() {
  local dir="$1" lang="$2"
  local rules="${DEVSKILLS_DIR}/cursor/rules"
  _dske_copy "${rules}/tiger-style.mdc" "${dir}/.cursor/rules/tiger-style.mdc"
  _dske_copy "${rules}/context.mdc"     "${dir}/.cursor/rules/context.mdc"
  case "$lang" in
    go)                    _dske_copy "${rules}/go.mdc"         "${dir}/.cursor/rules/go.mdc" ;;
    typescript|javascript) _dske_copy "${rules}/typescript.mdc" "${dir}/.cursor/rules/typescript.mdc" ;;
    rust)                  _dske_copy "${rules}/rust.mdc"       "${dir}/.cursor/rules/rust.mdc" ;;
    python)                _dske_copy "${rules}/python.mdc"     "${dir}/.cursor/rules/python.mdc" ;;
  esac
}

# Install VSCode Copilot instructions into <dir>/.github: the base instructions
# plus the Language-Specific Notes for <lang> only (an empty/unknown lang writes
# the base alone). Parallels devskills_install_cursor — one source per language,
# all editor targets follow when a profile is added.
#   $1 target dir  $2 lang ("" for none)
devskills_install_vscode() {
  local dir="$1" lang="$2"
  local src="${DEVSKILLS_DIR}/vscode"
  local dst="${dir}/.github/copilot-instructions.md"
  local note="${src}/lang/${lang}.md"
  local has_note=0
  [ -n "$lang" ] && [ -f "$note" ] && has_note=1

  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    _dske_log "[dry] would write ${dst}$( [ "$has_note" -eq 1 ] && printf ' (+ %s notes)' "$lang" )"
    return 0
  fi
  mkdir -p "$(dirname "$dst")"
  cat "${src}/copilot-base.md" > "$dst"
  if [ "$has_note" -eq 1 ]; then
    { printf '\n## Language-Specific Notes\n\n'; cat "$note"; } >> "$dst"
  fi
  _dske_log "wrote ${dst}$( [ "$has_note" -eq 1 ] && printf ' (+ %s notes)' "$lang" )"
}
