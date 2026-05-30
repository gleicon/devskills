# profile.sh — shared AGENTS.md installer for devskills.
#
# Sourced by install.sh and scripts/setup.sh. Builds a project's AGENTS.md
# from stacked managed blocks and makes CLAUDE.md import it via `@AGENTS.md`
# so Claude Code and OpenCode read the same content with no duplication.
#
# Blocks (in order), each delimited by <!-- BEGIN/END devskills:<id> -->:
#   base      always — universal engineering principles
#   concise   optional — terse-response directive (--concise)
#   tooling   optional — devskills tooling reference (--hints)
#   language  optional — per-language profile (--lang=<x>)
#
# Behavior:
#   - Never clobbers: existing files are backed up (sibling .bak) before any
#     change, and only when the content actually changes.
#   - Idempotent: re-running is a no-op; changing options swaps blocks in
#     place. CLAUDE.md gets a single `@AGENTS.md` import.

[ -n "${DEVSKILLS_PROFILE_LIB:-}" ] && return 0
DEVSKILLS_PROFILE_LIB=1

_dsk_log()  { printf '[devskills] %s\n' "$1"; }
_dsk_warn() { printf '[devskills] WARN: %s\n' "$1" >&2; }

# Per-run bookkeeping: files we created this run (never back those up), and
# files already backed up this run (back a pre-existing file up at most once).
_dsk_mark_created()    { _DSK_CREATED="${_DSK_CREATED:-} $1"; }
_dsk_was_created()     { case " ${_DSK_CREATED:-} " in *" $1 "*) return 0 ;; esac; return 1; }

# Back up a pre-existing file once per run, to a sibling timestamped .bak.
# No-op for files this run created, or already backed up.
_dsk_backup_once() {
  local f="$1"
  _dsk_was_created "$f" && return 0
  case " ${_DSK_BACKED:-} " in *" $f "*) return 0 ;; esac
  : "${DEVSKILLS_STAMP:=$(date +%Y%m%d-%H%M%S)}"
  local bak="${f}.${DEVSKILLS_STAMP}.bak"
  cp "$f" "$bak"
  _DSK_BACKED="${_DSK_BACKED:-} $f"
  _dsk_log "backed up $(basename "$f") -> $(basename "$bak")"
}

# Insert or replace a managed block in a file.
#   $1 file  $2 block id  $3 file holding the full block (incl. markers)
#   $4 dry-run (1|0)  $5 display label
# Writes nothing when the result would be byte-identical.
_dsk_upsert_block() {
  local file="$1" id="$2" block_file="$3" dry="$4" label="${5:-$1}"
  local begin="<!-- BEGIN devskills:${id} -->"
  local end="<!-- END devskills:${id} -->"
  local tmp; tmp="$(mktemp)"

  if [ -f "$file" ] && grep -qF "$begin" "$file"; then
    # Replace the existing block in place.
    awk -v b="$begin" -v e="$end" -v bf="$block_file" '
      $0 == b { while ((getline line < bf) > 0) print line; close(bf); skip = 1; next }
      skip && $0 == e { skip = 0; next }
      skip { next }
      { print }
    ' "$file" > "$tmp"
  else
    # Append, preserving existing content with a blank-line separator.
    if [ -f "$file" ] && [ -s "$file" ]; then
      cat "$file" > "$tmp"
      [ -z "$(tail -c1 "$file")" ] || printf '\n' >> "$tmp"
      printf '\n' >> "$tmp"
    fi
    cat "$block_file" >> "$tmp"
  fi

  if [ -f "$file" ] && cmp -s "$tmp" "$file"; then
    _dsk_log "${label} already up to date."
    rm -f "$tmp"
    return 0
  fi

  if [ "$dry" = "1" ]; then
    if [ -f "$file" ]; then
      _dsk_log "[dry] would back up ${label} and update it."
    else
      _dsk_log "[dry] would create ${label}."
    fi
    rm -f "$tmp"
    return 0
  fi

  if [ -f "$file" ]; then
    _dsk_backup_once "$file"   # no-op for files we created earlier this run
    mv "$tmp" "$file"
    _dsk_log "updated ${label}"
  else
    mv "$tmp" "$file"
    _dsk_mark_created "$file"
    _dsk_log "created ${label}"
  fi
}

# Wrap a source file in a managed block and upsert it into <dir>/AGENTS.md.
#   $1 id  $2 source file  $3 dir  $4 dry-run  $5 optional note line
_dsk_inject() {
  local id="$1" src="$2" dir="$3" dry="$4" note="${5:-}"
  local agents="${dir}/AGENTS.md"
  local blk; blk="$(mktemp)"
  {
    echo "<!-- BEGIN devskills:${id} -->"
    [ -n "$note" ] && echo "$note"
    cat "$src"
  } > "$blk"
  [ -z "$(tail -c1 "$src")" ] || printf '\n' >> "$blk"
  echo "<!-- END devskills:${id} -->" >> "$blk"
  _dsk_upsert_block "$agents" "$id" "$blk" "$dry" "AGENTS.md (${id})"
  rm -f "$blk"
}

# Ensure CLAUDE.md imports AGENTS.md (skip if it already does manually).
_dsk_ensure_claude_import() {
  local dir="$1" dry="$2"
  local claude="${dir}/CLAUDE.md"
  if [ -f "$claude" ] && grep -qE '^[[:space:]]*@AGENTS\.md' "$claude" \
       && ! grep -qF "<!-- BEGIN devskills:import -->" "$claude"; then
    _dsk_log "CLAUDE.md already imports AGENTS.md; leaving as-is."
    return 0
  fi
  local imp; imp="$(mktemp)"
  {
    echo "<!-- BEGIN devskills:import -->"
    echo "@AGENTS.md"
    echo "<!-- END devskills:import -->"
  } > "$imp"
  _dsk_upsert_block "$claude" "import" "$imp" "$dry" "CLAUDE.md"
  rm -f "$imp"
}

# Remove the listed managed blocks from a file. If nothing but our blocks
# remains, the file is deleted. Backs up (once) before any change.
#   $1 file  $2 dry-run  $3 label  $4.. block ids
_dsk_remove_blocks() {
  local file="$1" dry="$2" label="$3"; shift 3
  [ -f "$file" ] || { _dsk_log "${label}: nothing to remove."; return 0; }

  local found=0 id
  for id in "$@"; do
    if grep -qF "<!-- BEGIN devskills:${id} -->" "$file"; then found=1; fi
  done
  if [ "$found" -eq 0 ]; then _dsk_log "${label}: no devskills blocks; left as-is."; return 0; fi

  local tmp; tmp="$(mktemp)"
  cp "$file" "$tmp"
  for id in "$@"; do
    awk -v b="<!-- BEGIN devskills:${id} -->" -v e="<!-- END devskills:${id} -->" '
      $0 == b { skip = 1; next }
      skip && $0 == e { skip = 0; next }
      skip { next }
      { print }
    ' "$tmp" > "${tmp}.2" && mv "${tmp}.2" "$tmp"
  done
  # Normalize whitespace: drop leading/trailing blanks, collapse blank runs.
  awk '
    /^[[:space:]]*$/ { pending = 1; next }
    { if (printed && pending) print ""; print; printed = 1; pending = 0 }
  ' "$tmp" > "${tmp}.2" && mv "${tmp}.2" "$tmp"

  if [ ! -s "$tmp" ]; then
    if [ "$dry" = "1" ]; then
      _dsk_log "[dry] would back up ${label} and remove it (only devskills content)."
    else
      _dsk_backup_once "$file"; rm -f "$file"
      _dsk_log "removed ${label} (held only devskills content)."
    fi
    rm -f "$tmp"; return 0
  fi

  if cmp -s "$tmp" "$file"; then _dsk_log "${label}: no change."; rm -f "$tmp"; return 0; fi

  if [ "$dry" = "1" ]; then
    _dsk_log "[dry] would back up ${label} and strip devskills blocks."
    rm -f "$tmp"; return 0
  fi
  _dsk_backup_once "$file"
  mv "$tmp" "$file"
  _dsk_log "stripped devskills blocks from ${label}."
}

# Remove everything devskills wrote into a project (managed blocks + marker).
# Preserves all user content; backs up before any change.
#   $1 target dir  $2 dry-run (1|0)
devskills_uninstall() {
  local dir="$1" dry="$2"
  _dsk_remove_blocks "${dir}/AGENTS.md" "$dry" "AGENTS.md" base concise tooling language
  _dsk_remove_blocks "${dir}/CLAUDE.md" "$dry" "CLAUDE.md" import

  # Legacy cleanup: older devskills versions recorded the profile in
  # .devskills/language. The profile now lives in AGENTS.md, but sweep up any
  # stray file/dir left behind by an earlier setup.
  if [ -f "${dir}/.devskills/language" ]; then
    if [ "$dry" = "1" ]; then
      _dsk_log "[dry] would remove legacy .devskills/language"
    else
      rm -f "${dir}/.devskills/language"
      rmdir "${dir}/.devskills" 2>/dev/null || true
      _dsk_log "removed legacy .devskills/language"
    fi
  fi

  if [ -d "${dir}/.cursor/rules" ] || [ -f "${dir}/.github/copilot-instructions.md" ]; then
    _dsk_warn "Cursor rules / VSCode Copilot instructions (if any) were left in place — remove manually if unwanted."
  fi
}

# Apply the devskills baseline (and optional layers) to a project.
#   $1 prompts dir  $2 target dir  $3 dry-run (1|0)
#   $4 lang ("" for none)  $5 concise (1|0)  $6 hints (1|0)
devskills_apply() {
  local pdir="$1" dir="$2" dry="$3" lang="$4" concise="$5" hints="$6"

  # 1. Base engineering principles (always).
  _dsk_inject base "${pdir}/system/agents-base.md" "$dir" "$dry"

  # 2. Concise-response directive (optional).
  if [ "$concise" = "1" ]; then
    _dsk_inject concise "${pdir}/system/concise.md" "$dir" "$dry"
  fi

  # 3. devskills tooling reference (optional).
  if [ "$hints" = "1" ]; then
    _dsk_inject tooling "${pdir}/system/tooling.md" "$dir" "$dry"
  fi

  # 4. Language profile (optional).
  if [ -n "$lang" ]; then
    _dsk_inject language "${pdir}/language/${lang}.md" "$dir" "$dry" \
      "<!-- profile: ${lang} — managed by devskills; edits between these markers are overwritten -->"
  fi

  # 5. CLAUDE.md imports AGENTS.md.
  _dsk_ensure_claude_import "$dir" "$dry"

  # 6. Flag a legacy inline profile from older devskills versions.
  if [ -f "${dir}/CLAUDE.md" ] && grep -qF "devskills language profile" "${dir}/CLAUDE.md"; then
    _dsk_warn "CLAUDE.md has a legacy inline devskills profile."
    _dsk_warn "Content now lives in AGENTS.md — remove the old inline block from CLAUDE.md."
  fi

}
