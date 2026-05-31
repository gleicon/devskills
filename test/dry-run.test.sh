#!/usr/bin/env bash
# dry-run.test.sh — install.sh --dry-run must write nothing.
#
# Regression guard for the bug where install_claude/install_opencode created the
# target commands/ dir with an unguarded `mkdir -p`, so --dry-run still mutated
# the filesystem. It hid on Claude (whose ~/.claude/commands usually pre-exists,
# making the mkdir a no-op) but visibly created ~/.opencode/commands. The file
# copies were guarded; only the directory creation leaked.
#
# Two tiers, both black-box against the real install.sh under a sandboxed $HOME:
#   1. test_command_paths   — names the original bug: the two command-install
#                             paths create nothing under --dry-run, and a real
#                             run still copies every command.
#   2. test_nothing_written — the broader invariant: a near-full --dry-run
#                             (--lang + cursor + vscode enabled) writes nothing
#                             ANYWHERE under $HOME or PWD. Catches future leaks
#                             at the install.sh orchestration layer, where the
#                             mkdir bug actually lived. --external stays skipped
#                             (it touches brew/curl/network; its own dry guards
#                             are covered by external-tools.test.sh).
#
# Run via `npm test`; exits non-zero on any failure.
set -uo pipefail

REPO="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

PASS=0
FAIL=0
pass() { PASS=$((PASS + 1)); printf '  ok   %s\n' "$1"; }
fail() { FAIL=$((FAIL + 1)); printf '  FAIL %s\n' "$1"; }

# A sandboxed $HOME where opencode looks "installed" (~/.opencode exists) but its
# commands/ subdir does not — the exact condition that exposed the bug.
sandbox_home() {
  local home; home="$(mktemp -d "${TMPDIR:-/tmp}/dsk-dryrun-home.XXXXXX")"
  mkdir -p "${home}/.opencode"
  mkdir -p "${home}/.claude"
  printf '%s' "$home"
}

run_install() {
  local home="$1"; shift
  HOME="$home" CLAUDE_CONFIG_DIR="${home}/.claude" \
    bash "${REPO}/install.sh" "$@" >/dev/null 2>&1
}

# A stable fingerprint of a directory tree: every path plus a hash of each file's
# contents, sorted. Two identical trees produce identical fingerprints; any
# created/removed/modified file changes it.
fingerprint() {
  local dir="$1"
  ( cd "$dir" && find . -print0 \
      | sort -z \
      | while IFS= read -r -d '' p; do
          if [ -f "$p" ]; then
            printf '%s  %s\n' "$(shasum "$p" | awk '{print $1}')" "$p"
          else
            printf 'DIR       %s\n' "$p"
          fi
        done )
}

test_command_paths() {
  echo "test: --dry-run creates no command files or dirs (names the bug)"
  local home; home="$(sandbox_home)"

  run_install "$home" --dry-run --skip-external --skip-cursor --skip-vscode

  [ ! -d "${home}/.opencode/commands" ] \
    && pass "~/.opencode/commands not created" \
    || fail "~/.opencode/commands created in dry-run"
  [ ! -d "${home}/.claude/commands" ] \
    && pass "~/.claude/commands not created" \
    || fail "~/.claude/commands created in dry-run"

  echo "test: a real run still creates dirs and copies commands"
  local expected; expected="$(ls "${REPO}/commands"/*.md | wc -l | tr -d ' ')"
  run_install "$home" --skip-external --skip-cursor --skip-vscode
  local oc; oc="$(ls "${home}/.opencode/commands"/*.md 2>/dev/null | wc -l | tr -d ' ')"
  local cc; cc="$(ls "${home}/.claude/commands"/*.md 2>/dev/null | wc -l | tr -d ' ')"
  [ "$oc" = "$expected" ] \
    && pass "~/.opencode/commands has all ${expected} commands" \
    || fail "~/.opencode/commands has ${oc}, expected ${expected}"
  [ "$cc" = "$expected" ] \
    && pass "~/.claude/commands has all ${expected} commands" \
    || fail "~/.claude/commands has ${cc}, expected ${expected}"

  rm -rf "$home"
}

test_nothing_written() {
  echo "test: a near-full --dry-run writes nothing anywhere (HOME or PWD)"
  local home; home="$(sandbox_home)"
  # A project dir OUTSIDE the repo, so install.sh's in-repo auto-skip block
  # does not fire and --lang is honored. Seed .cursor/.vscode so the editor
  # install paths actually execute regardless of what's on PATH.
  local proj; proj="$(mktemp -d "${TMPDIR:-/tmp}/dsk-dryrun-proj.XXXXXX")"
  mkdir -p "${proj}/.cursor" "${proj}/.vscode"

  local home_before proj_before
  home_before="$(fingerprint "$home")"
  proj_before="$(fingerprint "$proj")"

  # cursor + vscode enabled (not skipped); --lang exercises the AGENTS.md path.
  HOME="$home" CLAUDE_CONFIG_DIR="${home}/.claude" \
    bash -c "cd '$proj' && bash '${REPO}/install.sh' --dry-run --lang=go --skip-external" \
    >/dev/null 2>&1

  local home_after proj_after
  home_after="$(fingerprint "$home")"
  proj_after="$(fingerprint "$proj")"

  [ "$home_before" = "$home_after" ] \
    && pass "\$HOME tree byte-for-byte unchanged" \
    || fail "\$HOME tree changed under --dry-run:"$'\n'"$(diff <(printf '%s' "$home_before") <(printf '%s' "$home_after"))"
  [ "$proj_before" = "$proj_after" ] \
    && pass "project tree byte-for-byte unchanged" \
    || fail "project tree changed under --dry-run:"$'\n'"$(diff <(printf '%s' "$proj_before") <(printf '%s' "$proj_after"))"

  rm -rf "$home" "$proj"
}

echo "install.sh --dry-run tests"
echo
test_command_paths
test_nothing_written
echo
echo "passed: ${PASS}, failed: ${FAIL}"
[ "$FAIL" -eq 0 ]
