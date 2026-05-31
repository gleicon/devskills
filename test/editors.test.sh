#!/usr/bin/env bash
# editors.test.sh — tests for the Cursor install path in scripts/lib/editors.sh.
#
# Cursor rules are always-apply engineering guardrails (tiger-style + context);
# if devskills_install_cursor silently stops shipping one, every project that
# installs loses that rule with no error. The lib is otherwise only exercised by
# install.sh's dry-run integration, so pin the real install here.
#
# Plain bash (no bats). Sources the lib and installs into mktemp dirs; no network.
# Run via `npm test`; exits non-zero on any failure.
set -uo pipefail

REPO="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
export DEVSKILLS_DIR="$REPO"
# shellcheck source=../scripts/lib/editors.sh
source "${REPO}/scripts/lib/editors.sh"

PASS=0
FAIL=0
pass() { PASS=$((PASS + 1)); printf '  ok   %s\n' "$1"; }
fail() { FAIL=$((FAIL + 1)); printf '  FAIL %s\n' "$1"; }

workspace() { mktemp -d "${TMPDIR:-/tmp}/dsk-editors.XXXXXX"; }

test_cursor_installs_always_apply_rules() {
  echo "test: Cursor install ships the always-apply rules plus the language rule"
  local ws; ws="$(workspace)"

  DRY_RUN=0 devskills_install_cursor "$ws" go >/dev/null

  local r="${ws}/.cursor/rules"
  [ -f "${r}/tiger-style.mdc" ] && pass "tiger-style.mdc installed" || fail "tiger-style.mdc missing"
  [ -f "${r}/context.mdc" ]     && pass "context.mdc installed"     || fail "context.mdc missing (always-apply rule dropped)"
  [ -f "${r}/go.mdc" ]          && pass "go.mdc installed for --lang=go" || fail "go.mdc missing"

  # context.mdc must actually be always-apply, or Cursor won't load it passively.
  grep -q '^alwaysApply: true' "${r}/context.mdc" \
    && pass "context.mdc is alwaysApply: true" \
    || fail "context.mdc is not marked alwaysApply: true"

  rm -rf "$ws"
}

test_cursor_no_lang_installs_always_apply_only() {
  echo "test: an empty lang installs the always-apply rules and no language rule"
  local ws; ws="$(workspace)"

  DRY_RUN=0 devskills_install_cursor "$ws" "" >/dev/null

  local r="${ws}/.cursor/rules"
  [ -f "${r}/tiger-style.mdc" ] && pass "tiger-style.mdc installed" || fail "tiger-style.mdc missing"
  [ -f "${r}/context.mdc" ]     && pass "context.mdc installed"     || fail "context.mdc missing"
  local count; count="$(ls "${r}"/*.mdc | wc -l | tr -d ' ')"
  [ "$count" = "2" ] && pass "exactly the 2 always-apply rules, no language rule" \
    || fail "expected 2 rules, got ${count}"

  rm -rf "$ws"
}

echo "editors.sh Cursor install tests"
echo
test_cursor_installs_always_apply_rules
test_cursor_no_lang_installs_always_apply_only
echo
echo "passed: ${PASS}, failed: ${FAIL}"
[ "$FAIL" -eq 0 ]
