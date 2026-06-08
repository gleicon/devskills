#!/usr/bin/env bash
# dry-run.test.sh — install.sh --dry-run must write nothing.
#
# Regression guard for the bug where install_claude/install_opencode created the
# target commands/ dir with an unguarded `mkdir -p`, so --dry-run still mutated
# the filesystem. It hid on Claude (whose ~/.claude/commands usually pre-exists,
# making the mkdir a no-op) but visibly created ~/.opencode/commands. The file
# copies were guarded; only the directory creation leaked.
#
# Three tiers, all black-box against the real install.sh under a sandboxed $HOME:
#   1. test_command_paths   — names the original bug: the command-install
#                             paths (Claude, OpenCode, Codex, Gemini) create
#                             nothing under --dry-run, and a real run still
#                             installs every command (Gemini as converted .toml).
#   2. test_skip_flags      — each --skip-<target> suppresses only its own
#                             command install and leaves the other three intact.
#   3. test_nothing_written — the broader invariant: a near-full --dry-run
#                             (--lang + cursor + vscode enabled) writes nothing
#                             ANYWHERE under $HOME or PWD — including the project
#                             GEMINI.md the --lang path would write. Catches
#                             future leaks at the install.sh orchestration layer,
#                             where the mkdir bug actually lived. --external stays
#                             skipped (it touches brew/curl/network; its own dry
#                             guards are covered by external-tools.test.sh).
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
  mkdir -p "${home}/.codex"
  mkdir -p "${home}/.gemini"
  printf '%s' "$home"
}

run_install() {
  local home="$1"; shift
  HOME="$home" CLAUDE_CONFIG_DIR="${home}/.claude" CODEX_HOME="${home}/.codex" \
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
  [ ! -d "${home}/.codex/prompts" ] \
    && pass "~/.codex/prompts not created" \
    || fail "~/.codex/prompts created in dry-run"
  [ ! -d "${home}/.gemini/commands" ] \
    && pass "~/.gemini/commands not created" \
    || fail "~/.gemini/commands created in dry-run"

  echo "test: a real run still creates dirs and installs commands"
  local expected; expected="$(ls "${REPO}/commands"/*.md | wc -l | tr -d ' ')"
  run_install "$home" --skip-external --skip-cursor --skip-vscode
  local oc; oc="$(ls "${home}/.opencode/commands"/*.md 2>/dev/null | wc -l | tr -d ' ')"
  local cc; cc="$(ls "${home}/.claude/commands"/*.md 2>/dev/null | wc -l | tr -d ' ')"
  local cx; cx="$(ls "${home}/.codex/prompts"/*.md 2>/dev/null | wc -l | tr -d ' ')"
  # Gemini ships TOML, not markdown — count .toml.
  local gm; gm="$(ls "${home}/.gemini/commands"/*.toml 2>/dev/null | wc -l | tr -d ' ')"
  [ "$oc" = "$expected" ] \
    && pass "~/.opencode/commands has all ${expected} commands" \
    || fail "~/.opencode/commands has ${oc}, expected ${expected}"
  [ "$cc" = "$expected" ] \
    && pass "~/.claude/commands has all ${expected} commands" \
    || fail "~/.claude/commands has ${cc}, expected ${expected}"
  [ "$cx" = "$expected" ] \
    && pass "~/.codex/prompts has all ${expected} commands" \
    || fail "~/.codex/prompts has ${cx}, expected ${expected}"
  [ "$gm" = "$expected" ] \
    && pass "~/.gemini/commands has all ${expected} .toml commands" \
    || fail "~/.gemini/commands has ${gm}, expected ${expected}"

  rm -rf "$home"
}

test_skip_flags() {
  echo "test: each --skip-<target> suppresses only its own command install"
  local expected; expected="$(ls "${REPO}/commands"/*.md | wc -l | tr -d ' ')"

  # Install with one skip flag into a fresh sandbox, count all four targets, and
  # assert the named one is empty while the other three install in full.
  check_skip() {
    local flag="$1" skipped="$2"
    local h; h="$(sandbox_home)"
    run_install "$h" --skip-external --skip-cursor --skip-vscode "$flag"
    local c o x g
    c="$(ls "${h}/.claude/commands"/*.md 2>/dev/null | wc -l | tr -d ' ')"
    o="$(ls "${h}/.opencode/commands"/*.md 2>/dev/null | wc -l | tr -d ' ')"
    x="$(ls "${h}/.codex/prompts"/*.md 2>/dev/null | wc -l | tr -d ' ')"
    g="$(ls "${h}/.gemini/commands"/*.toml 2>/dev/null | wc -l | tr -d ' ')"
    local want_c=$expected want_o=$expected want_x=$expected want_g=$expected
    case "$skipped" in
      claude) want_c=0 ;; opencode) want_o=0 ;; codex) want_x=0 ;; gemini) want_g=0 ;;
    esac
    if [ "$c" = "$want_c" ] && [ "$o" = "$want_o" ] && [ "$x" = "$want_x" ] && [ "$g" = "$want_g" ]; then
      pass "${flag}: claude=$c opencode=$o codex=$x gemini=$g (only ${skipped} skipped)"
    else
      fail "${flag}: claude=$c opencode=$o codex=$x gemini=$g; wanted ${skipped}=0, others=${expected}"
    fi
    rm -rf "$h"
  }

  check_skip --skip-claude   claude
  check_skip --skip-opencode opencode
  check_skip --skip-codex    codex
  check_skip --skip-gemini   gemini
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
  HOME="$home" CLAUDE_CONFIG_DIR="${home}/.claude" CODEX_HOME="${home}/.codex" \
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
test_skip_flags
test_nothing_written
echo
echo "passed: ${PASS}, failed: ${FAIL}"
[ "$FAIL" -eq 0 ]
