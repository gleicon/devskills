#!/usr/bin/env bash
# external-tools.test.sh — tests for the RTK archive-validation helpers in
# scripts/lib/external-tools.sh. They guard the only path that writes an
# executable into the user's $HOME, so a regression that weakens them is an
# arbitrary-file-write risk. Pin them.
#
# Plain bash (no bats). The helpers are pure (status only), so the tests need no
# network: _rtk_archive_has_unsafe_paths takes a tar member list, and
# _rtk_staged_binary_ok inspects a staged dir built under mktemp. Run via
# `npm test`; exits non-zero on any failure.
set -uo pipefail

REPO="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

# The lib expects callers to define log/warn before sourcing; honor that even
# though the pure helpers under test never call them.
log()  { :; }
warn() { :; }
# shellcheck source=../scripts/lib/external-tools.sh
source "${REPO}/scripts/lib/external-tools.sh"

PASS=0
FAIL=0
pass() { PASS=$((PASS + 1)); printf '  ok   %s\n' "$1"; }
fail() { FAIL=$((FAIL + 1)); printf '  FAIL %s\n' "$1"; }

assert_unsafe()     { _rtk_archive_has_unsafe_paths "$1" && pass "$2" || fail "$2 (allowed unsafe list)"; }
assert_safe()       { _rtk_archive_has_unsafe_paths "$1" && fail "$2 (rejected safe list)" || pass "$2"; }
assert_binary_ok()  { _rtk_staged_binary_ok "$1" && pass "$2" || fail "$2 (rejected good staging)"; }
assert_binary_bad() { _rtk_staged_binary_ok "$1" && fail "$2 (accepted bad staging)" || pass "$2"; }
assert_eq()         { [ "$1" = "$2" ] && pass "$3" || fail "$3 (expected '$2', got '$1')"; }

workspace() { mktemp -d "${TMPDIR:-/tmp}/dsk-rtk.XXXXXX"; }

test_unsafe_paths_rejected() {
  echo "test: archive members that escape the extract dir are refused"
  assert_safe   "rtk"                   "plain top-level binary is safe"
  assert_safe   "./rtk"                 "leading ./ is safe"
  assert_safe   "rtkbin/rtk"            "nested non-traversal path is safe"
  assert_safe   $'README\nrtk'          "multi-member safe list is safe"
  assert_unsafe $'rtk\n../escaped'      "a '..' member anywhere makes the list unsafe"
  assert_unsafe "foo/../../etc/passwd"  "an embedded '..' component is unsafe"
  assert_unsafe "/etc/passwd"           "an absolute path is unsafe"
  assert_unsafe ".."                    "a bare '..' is unsafe"
}

test_staged_binary_check() {
  echo "test: only a real, top-level rtk binary is accepted"
  local ws; ws="$(workspace)"

  mkdir -p "$ws/good"; printf '#!/bin/sh\n' > "$ws/good/rtk"
  assert_binary_ok "$ws/good" "a real rtk file is accepted"

  mkdir -p "$ws/symlink"; ln -s /etc/passwd "$ws/symlink/rtk"
  assert_binary_bad "$ws/symlink" "a symlinked rtk is refused"

  mkdir -p "$ws/empty"
  assert_binary_bad "$ws/empty" "a missing rtk is refused"

  mkdir -p "$ws/nested/rtkdir"; printf '#!/bin/sh\n' > "$ws/nested/rtkdir/rtk"
  assert_binary_bad "$ws/nested" "rtk only inside a subdir is refused"

  mkdir -p "$ws/dir/rtk"
  assert_binary_bad "$ws/dir" "an rtk that is a directory is refused"

  rm -rf "$ws"
}

test_expected_sha256_lookup() {
  echo "test: the expected sha256 is read from a checksums.txt body by exact asset name"
  # coreutils format: "<hash>  <name>" — mirrors a real rtk checksums.txt.
  local body
  body=$'aaa111  rtk-x86_64-unknown-linux-musl.tar.gz\nbbb222  rtk-aarch64-unknown-linux-gnu.tar.gz\nccc333  rtk-x86_64-apple-darwin.tar.gz'
  assert_eq "$(_rtk_expected_sha256 "$body" "rtk-x86_64-unknown-linux-musl.tar.gz")" "aaa111" "first asset's hash"
  assert_eq "$(_rtk_expected_sha256 "$body" "rtk-aarch64-unknown-linux-gnu.tar.gz")" "bbb222" "middle asset's hash"
  assert_eq "$(_rtk_expected_sha256 "$body" "rtk-not-a-real-asset.tar.gz")"          ""       "unlisted asset yields empty (no false match; caller refuses to install)"
  assert_eq "$(_rtk_expected_sha256 ""     "rtk-x86_64-unknown-linux-musl.tar.gz")"  ""       "empty checksums body yields empty"
  # Exactness: a name that is a substring of a listed entry must not match.
  assert_eq "$(_rtk_expected_sha256 "$body" "linux-musl.tar.gz")"                    ""       "substring of a listed name does not match"
}

echo "external-tools.sh RTK validation tests"
echo
test_unsafe_paths_rejected
test_staged_binary_check
test_expected_sha256_lookup
echo
echo "passed: ${PASS}, failed: ${FAIL}"
[ "$FAIL" -eq 0 ]
