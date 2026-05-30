#!/usr/bin/env bash
# profile.test.sh — behavior tests for scripts/lib/profile.sh, which mutates
# and deletes a user's real AGENTS.md/CLAUDE.md. These pin its three promises:
# never clobber, stay idempotent, and delete a file only when it held nothing
# but devskills content.
#
# Plain bash (no bats). Each test runs against a throwaway mktemp dir and drives
# the public interface (devskills_apply / devskills_uninstall). Run via
# `npm test`; exits non-zero on any failure.

# NOT set -e: we want every test to run and report all failures, not stop at
# the first one.
set -uo pipefail

REPO="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PROMPTS="${REPO}/prompts"

# shellcheck source=../scripts/lib/profile.sh
source "${REPO}/scripts/lib/profile.sh"

PASS=0
FAIL=0
pass() { PASS=$((PASS + 1)); printf '  ok   %s\n' "$1"; }
fail() { FAIL=$((FAIL + 1)); printf '  FAIL %s\n' "$1"; }

assert_exists()    { [ -f "$1" ] && pass "$2" || fail "$2 (missing: $1)"; }
assert_absent()    { [ ! -e "$1" ] && pass "$2" || fail "$2 (still present: $1)"; }
assert_grep()      { grep -qF "$2" "$1" 2>/dev/null && pass "$3" || fail "$3 (not found in $1: $2)"; }
assert_no_grep()   { grep -qF "$2" "$1" 2>/dev/null && fail "$3 (found in $1: $2)" || pass "$3"; }
assert_identical() { cmp -s "$1" "$2" && pass "$3" || fail "$3 ($1 differs from $2)"; }
assert_count()     { [ "$1" = "$2" ] && pass "$3" || fail "$3 (expected $2, got $1)"; }

count_baks() { find "$1" -name '*.bak' 2>/dev/null | wc -l | tr -d ' '; }
workspace()  { mktemp -d "${TMPDIR:-/tmp}/dsk-profile.XXXXXX"; }

# profile.sh keeps per-run bookkeeping in shell globals (which files it created
# / backed up this run). Reset them between tests so each one is hermetic.
reset_state() { unset _DSK_CREATED _DSK_BACKED DEVSKILLS_STAMP; }

USER_CONTENT=$'# My Project\n\nHand-written notes the user owns.\n'

test_create_new() {
  echo "test: create new project files"
  reset_state
  local ws; ws="$(workspace)"
  devskills_apply "$PROMPTS" "$ws" 0 go 0 0 >/dev/null 2>&1
  assert_exists "$ws/AGENTS.md" "AGENTS.md created"
  assert_exists "$ws/CLAUDE.md" "CLAUDE.md created"
  assert_grep "$ws/AGENTS.md" "<!-- BEGIN devskills:base -->" "base block present"
  assert_grep "$ws/AGENTS.md" "## 1. Think Before Coding" "base content present"
  assert_grep "$ws/AGENTS.md" "<!-- BEGIN devskills:language -->" "language block present"
  assert_grep "$ws/AGENTS.md" "## Language Profile — Go" "go profile content present"
  assert_grep "$ws/CLAUDE.md" "@AGENTS.md" "CLAUDE.md imports AGENTS.md"
  assert_grep "$ws/CLAUDE.md" "<!-- BEGIN devskills:import -->" "import block present"
  assert_count "$(count_baks "$ws")" 0 "no backups when files are new"
  rm -rf "$ws"
}

test_idempotent() {
  echo "test: re-run is byte-identical and writes no new backups"
  reset_state
  local ws; ws="$(workspace)"
  devskills_apply "$PROMPTS" "$ws" 0 go 0 0 >/dev/null 2>&1
  cp "$ws/AGENTS.md" "$ws/agents.snapshot"
  cp "$ws/CLAUDE.md" "$ws/claude.snapshot"
  devskills_apply "$PROMPTS" "$ws" 0 go 0 0 >/dev/null 2>&1
  assert_identical "$ws/AGENTS.md" "$ws/agents.snapshot" "AGENTS.md byte-identical after re-run"
  assert_identical "$ws/CLAUDE.md" "$ws/claude.snapshot" "CLAUDE.md byte-identical after re-run"
  assert_count "$(count_baks "$ws")" 0 "no .bak spam on idempotent re-run"
  rm -rf "$ws"
}

test_option_stacking() {
  echo "test: --concise/--hints stack blocks without duplicating base"
  reset_state
  local ws; ws="$(workspace)"
  devskills_apply "$PROMPTS" "$ws" 0 go 0 0 >/dev/null 2>&1
  devskills_apply "$PROMPTS" "$ws" 0 go 1 1 >/dev/null 2>&1
  assert_grep "$ws/AGENTS.md" "<!-- BEGIN devskills:concise -->" "concise block added on re-run"
  assert_grep "$ws/AGENTS.md" "<!-- BEGIN devskills:tooling -->" "tooling block added on re-run"
  assert_count "$(grep -cF "<!-- BEGIN devskills:base -->" "$ws/AGENTS.md")" 1 "base block not duplicated"
  rm -rf "$ws"
}

test_preserve_and_backup_once() {
  echo "test: existing user content is preserved and backed up exactly once"
  reset_state
  local ws; ws="$(workspace)"
  printf '%s' "$USER_CONTENT" > "$ws/AGENTS.md"
  printf '%s' "$USER_CONTENT" > "$ws/expected.orig"
  devskills_apply "$PROMPTS" "$ws" 0 "" 0 0 >/dev/null 2>&1   # base only
  assert_grep "$ws/AGENTS.md" "# My Project" "user heading preserved"
  assert_grep "$ws/AGENTS.md" "Hand-written notes the user owns." "user body preserved"
  assert_grep "$ws/AGENTS.md" "<!-- BEGIN devskills:base -->" "devskills block appended"
  assert_count "$(count_baks "$ws")" 1 "pre-existing file backed up exactly once"
  local bak; bak="$(find "$ws" -name '*.bak' | head -1)"
  assert_identical "$bak" "$ws/expected.orig" "backup matches the original content"
  rm -rf "$ws"
}

test_uninstall_preserves_user() {
  echo "test: uninstall strips devskills blocks but keeps user content and file"
  reset_state
  local ws; ws="$(workspace)"
  printf '%s' "$USER_CONTENT" > "$ws/AGENTS.md"
  devskills_apply "$PROMPTS" "$ws" 0 go 0 0 >/dev/null 2>&1
  devskills_uninstall "$ws" 0 >/dev/null 2>&1
  assert_exists "$ws/AGENTS.md" "AGENTS.md survives uninstall (had user content)"
  assert_grep "$ws/AGENTS.md" "# My Project" "user heading survives uninstall"
  assert_grep "$ws/AGENTS.md" "Hand-written notes the user owns." "user body survives uninstall"
  assert_no_grep "$ws/AGENTS.md" "<!-- BEGIN devskills:base -->" "base block removed"
  assert_no_grep "$ws/AGENTS.md" "<!-- BEGIN devskills:language -->" "language block removed"
  rm -rf "$ws"
}

test_uninstall_pure_devskills() {
  echo "test: uninstall deletes files that held only devskills content"
  reset_state
  local ws; ws="$(workspace)"
  devskills_apply "$PROMPTS" "$ws" 0 go 0 0 >/dev/null 2>&1
  devskills_uninstall "$ws" 0 >/dev/null 2>&1
  assert_absent "$ws/AGENTS.md" "pure-devskills AGENTS.md removed"
  assert_absent "$ws/CLAUDE.md" "pure-devskills CLAUDE.md removed"
  rm -rf "$ws"
}

test_uninstall_dry_run_deletes_nothing() {
  echo "test: dry-run uninstall touches nothing (the deletion safety rail)"
  reset_state
  local ws; ws="$(workspace)"
  devskills_apply "$PROMPTS" "$ws" 0 go 0 0 >/dev/null 2>&1
  cp "$ws/AGENTS.md" "$ws/agents.snapshot"
  cp "$ws/CLAUDE.md" "$ws/claude.snapshot"
  devskills_uninstall "$ws" 1 >/dev/null 2>&1
  assert_exists "$ws/AGENTS.md" "dry-run uninstall leaves AGENTS.md in place"
  assert_exists "$ws/CLAUDE.md" "dry-run uninstall leaves CLAUDE.md in place"
  assert_identical "$ws/AGENTS.md" "$ws/agents.snapshot" "dry-run uninstall does not modify AGENTS.md"
  assert_identical "$ws/CLAUDE.md" "$ws/claude.snapshot" "dry-run uninstall does not modify CLAUDE.md"
  assert_count "$(count_baks "$ws")" 0 "dry-run uninstall writes no backups"
  rm -rf "$ws"
}

test_dry_run_writes_nothing() {
  echo "test: --dry-run writes nothing to disk"
  reset_state
  local ws; ws="$(workspace)"
  devskills_apply "$PROMPTS" "$ws" 1 go 0 0 >/dev/null 2>&1
  assert_absent "$ws/AGENTS.md" "dry-run does not create AGENTS.md"
  assert_absent "$ws/CLAUDE.md" "dry-run does not create CLAUDE.md"
  rm -rf "$ws"
}

test_manual_import_untouched() {
  echo "test: a hand-written @AGENTS.md import in CLAUDE.md is left as-is"
  reset_state
  local ws; ws="$(workspace)"
  printf '# Notes\n\n@AGENTS.md\n' > "$ws/CLAUDE.md"
  cp "$ws/CLAUDE.md" "$ws/claude.orig"
  devskills_apply "$PROMPTS" "$ws" 0 "" 0 0 >/dev/null 2>&1
  assert_identical "$ws/CLAUDE.md" "$ws/claude.orig" "manual CLAUDE.md import left byte-identical"
  assert_no_grep "$ws/CLAUDE.md" "<!-- BEGIN devskills:import -->" "no managed import block injected over a manual one"
  rm -rf "$ws"
}

echo "profile.sh behavior tests"
echo
test_create_new
test_idempotent
test_option_stacking
test_preserve_and_backup_once
test_uninstall_preserves_user
test_uninstall_pure_devskills
test_uninstall_dry_run_deletes_nothing
test_dry_run_writes_nothing
test_manual_import_untouched
echo
echo "passed: ${PASS}, failed: ${FAIL}"
[ "$FAIL" -eq 0 ]
