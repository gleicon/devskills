#!/usr/bin/env bash
# gemini.test.sh — tests for the Gemini CLI command conversion in editors.sh.
#
# Unlike the Claude/OpenCode/Codex targets (plain .md copies), the Gemini target
# converts each commands/*.md into a TOML custom-command file. That conversion is
# real logic — a bad quote or a dropped key yields a file Gemini silently refuses
# to load — so pin it here directly, plus the .toml purge of stale renamed names.
#
# Plain bash (no bats, no TOML parser): TOML shape is checked with grep/sed.
# Sources the lib and converts into mktemp dirs; the purge test is black-box
# against the real install.sh. No network. Run via `npm test`; non-zero on fail.
set -uo pipefail

REPO="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
export DEVSKILLS_DIR="$REPO"
# shellcheck source=../scripts/lib/editors.sh
source "${REPO}/scripts/lib/editors.sh"

PASS=0
FAIL=0
pass() { PASS=$((PASS + 1)); printf '  ok   %s\n' "$1"; }
fail() { FAIL=$((FAIL + 1)); printf '  FAIL %s\n' "$1"; }

workspace() { mktemp -d "${TMPDIR:-/tmp}/dsk-gemini.XXXXXX"; }

# A well-formed Gemini command .toml: line 1 a non-empty `description = "…"`,
# line 2 exactly `prompt = '''`, last line the closing `'''`, and a non-blank
# body in between. Returns 0 if all hold.
valid_toml() {
  local f="$1"
  head -1 "$f" | grep -qE '^description = ".+"$' || return 1
  [ "$(sed -n 2p "$f")" = "prompt = '''" ]       || return 1
  [ "$(tail -1 "$f")" = "'''" ]                  || return 1
  local body; body="$(sed -n '3,$p' "$f" | sed '$d')"   # strip closing '''
  [ -n "$(printf '%s' "$body" | tr -d '[:space:]')" ]   || return 1
  return 0
}

test_converts_every_command() {
  echo "test: every commands/*.md becomes a well-formed .toml"
  local out; out="$(workspace)"
  DRY_RUN=0 devskills_install_gemini "$out" >/dev/null

  local expected; expected="$(ls "${REPO}/commands"/*.md | wc -l | tr -d ' ')"
  local got; got="$(ls "${out}"/*.toml 2>/dev/null | wc -l | tr -d ' ')"
  [ "$got" = "$expected" ] \
    && pass "one .toml per command (${got})" \
    || fail "expected ${expected} .toml, got ${got}"

  # Filenames map straight across (ds-explore.md -> ds-explore.toml).
  [ -f "${out}/ds-explore.toml" ] \
    && pass "ds-explore.md -> ds-explore.toml" \
    || fail "ds-explore.toml missing (name mapping broke)"

  local bad=0 f
  for f in "${out}"/*.toml; do
    valid_toml "$f" || { bad=$((bad + 1)); [ "$bad" -le 3 ] && printf '       malformed: %s\n' "$(basename "$f")"; }
  done
  [ "$bad" -eq 0 ] \
    && pass "all ${got} files have a non-empty prompt and description" \
    || fail "${bad} malformed .toml file(s)"

  rm -rf "$out"
}

test_description_matches_first_line() {
  echo "test: the description is the command's first line"
  local out; out="$(workspace)"
  DRY_RUN=0 devskills_install_gemini "$out" >/dev/null

  local line1; line1="$(head -1 "${REPO}/commands/ds-explore.md")"
  local desc; desc="$(head -1 "${out}/ds-explore.toml")"
  desc="${desc#description = \"}"; desc="${desc%\"}"
  [ "$desc" = "$line1" ] \
    && pass "ds-explore description equals its first line" \
    || fail "description [$desc] != first line [$line1]"

  rm -rf "$out"
}

test_description_escaping() {
  echo "test: a first line with quotes and a backslash is escaped in the description"
  local fix; fix="$(workspace)"; mkdir -p "${fix}/commands"
  # File line 1 (after printf): He said "hi" and a back\slash
  printf 'He said "hi" and a back\\slash\n\nBody.\n' > "${fix}/commands/ds-quote.md"

  local out; out="$(workspace)"
  # Subshell so the DEVSKILLS_DIR override can't leak into later tests; the lib
  # function is inherited by the subshell.
  ( export DEVSKILLS_DIR="$fix"; DRY_RUN=0 devskills_install_gemini "$out" >/dev/null )

  local got; got="$(head -1 "${out}/ds-quote.toml")"
  local want='description = "He said \"hi\" and a back\\slash"'
  [ "$got" = "$want" ] \
    && pass "quotes and backslash escaped in description" \
    || fail "escaping wrong:"$'\n'"       got:  $got"$'\n'"       want: $want"

  # The embedded quote must not have leaked into the prompt structure.
  valid_toml "${out}/ds-quote.toml" \
    && pass "escaped file is still structurally valid" \
    || fail "escaped file is malformed"

  rm -rf "$fix" "$out"
}

# The TOML body is a ''' literal with no escaping, which is only safe while no
# command contains that delimiter (a Python '''docstring''' or a TOML example
# would break it — Gemini then silently refuses to load the file). This is the
# author-time guard: it fails CI the moment a command introduces ''', long
# before any user hits the silent load failure.
test_no_triple_quote_in_commands() {
  echo "test: no command body contains the ''' TOML delimiter"
  local offenders; offenders="$(grep -lF "'''" "${REPO}/commands"/*.md 2>/dev/null || true)"
  if [ -z "$offenders" ]; then
    pass "no commands/*.md contains '''"
  else
    fail "commands contain ''' (breaks the Gemini TOML literal):"
    printf '       %s\n' $offenders
  fi
}

test_purges_stale_toml() {
  echo "test: a real install purges a stale renamed .toml from the Gemini dir"
  local home; home="$(workspace)"
  mkdir -p "${home}/.gemini/commands" "${home}/.claude" "${home}/.codex"
  : > "${home}/.gemini/commands/ds-modes.toml"   # a retired name (RENAMED_COMMANDS)

  # Isolate the Gemini target; ~/.gemini present so detection fires.
  HOME="$home" CLAUDE_CONFIG_DIR="${home}/.claude" CODEX_HOME="${home}/.codex" \
    bash "${REPO}/install.sh" --skip-external --skip-cursor --skip-vscode \
      --skip-claude --skip-opencode --skip-codex >/dev/null 2>&1

  [ ! -f "${home}/.gemini/commands/ds-modes.toml" ] \
    && pass "stale ds-modes.toml purged" \
    || fail "stale ds-modes.toml survived install"
  local n; n="$(ls "${home}/.gemini/commands"/*.toml 2>/dev/null | wc -l | tr -d ' ')"
  [ "$n" -gt 0 ] \
    && pass "real commands still installed (${n} .toml)" \
    || fail "no real .toml installed"

  rm -rf "$home"
}

echo "editors.sh Gemini conversion tests"
echo
test_no_triple_quote_in_commands
test_converts_every_command
test_description_matches_first_line
test_description_escaping
test_purges_stale_toml
echo
echo "passed: ${PASS}, failed: ${FAIL}"
[ "$FAIL" -eq 0 ]
