#!/usr/bin/env bash
# sync.test.sh — the claude/ and opencode/ command trees must stay byte-identical.
# Every command ships to both, and /write-a-command writes the two copies by hand;
# nothing else guards the invariant, so a stray edit to one tree would drift
# silently. This pins it: `npm test` fails the moment they diverge.
set -uo pipefail

REPO="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

echo "claude/ ↔ opencode/ command parity"
if diff -rq "${REPO}/claude/commands" "${REPO}/opencode/commands"; then
  echo "  ok   command trees are byte-identical"
  exit 0
fi
echo "  FAIL claude/commands and opencode/commands diverge (see diff above)"
exit 1
