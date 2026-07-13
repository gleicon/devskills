#!/usr/bin/env bash
# sync.test.sh — skills/ is the single source. Every skill is a directory with a
# non-empty SKILL.md carrying the required frontmatter (name matching the dir,
# a description, and disable-model-invocation: true — the user-invoked contract).
set -uo pipefail

REPO="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SKILLS="${REPO}/skills"

echo "skills/ source integrity"

if [ ! -d "$SKILLS" ]; then
  echo "  FAIL skills/ directory missing"
  exit 1
fi

fail=0
count=0
for d in "${SKILLS}"/*/; do
  name="$(basename "$d")"
  f="${d}SKILL.md"
  count=$((count + 1))
  [ -s "$f" ] || { echo "  FAIL missing/empty SKILL.md: ${name}"; fail=1; continue; }
  grep -q "^name: ${name}$" "$f" || { echo "  FAIL ${name}: frontmatter name mismatch"; fail=1; }
  grep -qE "^description: .+" "$f" || { echo "  FAIL ${name}: missing description"; fail=1; }
  grep -q "^disable-model-invocation: true$" "$f" || { echo "  FAIL ${name}: not marked user-invoked"; fail=1; }
done

if [ "$fail" -eq 0 ]; then
  echo "  ok   ${count} skills found, all valid"
  exit 0
fi
exit 1
