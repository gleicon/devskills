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
#   1. test_command_paths   — names the original bug: the command-install
#                             paths (Claude, OpenCode, Codex) create nothing
#                             under --dry-run, and a real run still copies
#                             every command.
#   2. test_nothing_written — the broader invariant: a near-full --dry-run
#                             (--lang enabled) writes nothing ANYWHERE under
#                             $HOME or PWD. Catches future leaks at the
#                             install.sh orchestration layer, where the mkdir
#                             bug actually lived. --external stays skipped
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
  mkdir -p "${home}/.codex"
  printf '%s' "$home"
}

# Sandbox every harness path under $home, including XDG_CONFIG_HOME (OpenCode's
# skills root) so a real dev-machine XDG setting can't leak the install out of the
# sandbox. --skip-agy on every call: agy installs into its own store via an
# external binary that can't be sandboxed under $HOME.
run_install() {
  local home="$1"; shift
  HOME="$home" CLAUDE_CONFIG_DIR="${home}/.claude" CODEX_HOME="${home}/.codex" \
    XDG_CONFIG_HOME="${home}/.config" \
    bash "${REPO}/install.sh" --skip-agy "$@" >/dev/null 2>&1
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
  echo "test: --dry-run creates no skill dirs anywhere (names the bug)"
  local home; home="$(sandbox_home)"

  run_install "$home" --dry-run --skip-external

  [ ! -d "${home}/.config/opencode/skills" ] \
    && pass "~/.config/opencode/skills not created" \
    || fail "~/.config/opencode/skills created in dry-run"
  [ ! -d "${home}/.claude/skills" ] \
    && pass "~/.claude/skills not created" \
    || fail "~/.claude/skills created in dry-run"
  [ ! -d "${home}/.codex/skills" ] \
    && pass "~/.codex/skills not created" \
    || fail "~/.codex/skills created in dry-run"

  echo "test: a real run installs skills to every harness and purges legacy dirs"
  local exp_skills; exp_skills="$(ls -d "${REPO}/skills"/*/ | wc -l | tr -d ' ')"

  # Seed each harness's legacy command/prompt dir with devskills files (must be
  # purged) plus a user-authored file (must survive).
  mkdir -p "${home}/.claude/commands" "${home}/.opencode/commands" "${home}/.codex/prompts"
  : > "${home}/.claude/commands/ds-code-review.md"        # removed in overhaul
  : > "${home}/.claude/commands/ds-workflow.md"           # renamed -> ds
  : > "${home}/.claude/commands/my-notes.md"              # user-authored, keep
  : > "${home}/.opencode/commands/ds-caveman-lite-mode.md"
  : > "${home}/.opencode/commands/keep-me.md"             # user-authored, keep
  : > "${home}/.codex/prompts/ds-quality-gate-mode.md"    # renamed -> ds-quality-gate
  : > "${home}/.codex/prompts/user-prompt.md"             # user-authored, keep

  run_install "$home" --skip-external

  local oc cc cx
  oc="$(ls -d "${home}/.config/opencode/skills"/*/ 2>/dev/null | wc -l | tr -d ' ')"
  cc="$(ls -d "${home}/.claude/skills"/*/ 2>/dev/null | wc -l | tr -d ' ')"
  cx="$(ls -d "${home}/.codex/skills"/*/ 2>/dev/null | wc -l | tr -d ' ')"
  [ "$cc" = "$exp_skills" ] \
    && pass "~/.claude/skills has all ${exp_skills} skills" \
    || fail "~/.claude/skills has ${cc}, expected ${exp_skills}"
  [ "$oc" = "$exp_skills" ] \
    && pass "~/.config/opencode/skills has all ${exp_skills} skills" \
    || fail "~/.config/opencode/skills has ${oc}, expected ${exp_skills}"
  [ "$cx" = "$exp_skills" ] \
    && pass "~/.codex/skills has all ${exp_skills} skills" \
    || fail "~/.codex/skills has ${cx}, expected ${exp_skills}"

  # Codex sidecar (AC-6): generated, pins user-only invocation.
  local sc="${home}/.codex/skills/ds-git-mode/agents/openai.yaml"
  { [ -f "$sc" ] && grep -q "allow_implicit_invocation: false" "$sc"; } \
    && pass "Codex sidecar sets allow_implicit_invocation: false" \
    || fail "Codex sidecar missing or wrong"

  # Full legacy purge across all three harnesses; user files survive.
  local purged=1
  [ -e "${home}/.claude/commands/ds-code-review.md" ] && purged=0
  [ -e "${home}/.claude/commands/ds-workflow.md" ] && purged=0
  [ -e "${home}/.opencode/commands/ds-caveman-lite-mode.md" ] && purged=0
  [ -e "${home}/.codex/prompts/ds-quality-gate-mode.md" ] && purged=0
  [ "$purged" = 1 ] \
    && pass "legacy devskills commands purged from all harnesses" \
    || fail "legacy devskills commands survived"
  { [ -e "${home}/.claude/commands/my-notes.md" ] \
      && [ -e "${home}/.opencode/commands/keep-me.md" ] \
      && [ -e "${home}/.codex/prompts/user-prompt.md" ]; } \
    && pass "user-authored files left untouched" \
    || fail "a user-authored file was deleted by purge"

  rm -rf "$home"
}

test_nothing_written() {
  echo "test: a near-full --dry-run writes nothing anywhere (HOME or PWD)"
  local home; home="$(sandbox_home)"
  # A project dir OUTSIDE the repo, so install.sh's in-repo auto-skip block
  # does not fire and --lang is honored.
  local proj; proj="$(mktemp -d "${TMPDIR:-/tmp}/dsk-dryrun-proj.XXXXXX")"

  local home_before proj_before
  home_before="$(fingerprint "$home")"
  proj_before="$(fingerprint "$proj")"

  # --lang exercises the AGENTS.md path. XDG_CONFIG_HOME sandboxes OpenCode's
  # skills root; --skip-agy since agy's store can't be sandboxed under $HOME
  # (its dry-run path is a plain no-op regardless).
  HOME="$home" CLAUDE_CONFIG_DIR="${home}/.claude" CODEX_HOME="${home}/.codex" \
    XDG_CONFIG_HOME="${home}/.config" \
    bash -c "cd '$proj' && bash '${REPO}/install.sh' --dry-run --lang=go --skip-external --skip-agy" \
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
