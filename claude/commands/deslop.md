Remove AI-generated code slop introduced on the current branch and clean it up to match the surrounding codebase.

When invoked, diff the current branch against the main branch and remove AI-generated slop the branch introduced. Preserve behavior.

## Arguments

Treat any argument as scope (files or directories). With no scope, review the branch's diff against the main branch.

## Focus Areas

- Extra comments that are unnecessary or inconsistent with local style
- Defensive checks or try/catch blocks that are abnormal for trusted code paths
- Casts to `any` (or the language's equivalent escape hatch) used only to bypass type issues
- Deeply nested code that should be simplified with early returns
- Other patterns inconsistent with the file and surrounding codebase

## Guardrails

- Keep behavior unchanged unless fixing a clear bug.
- Prefer minimal, focused edits over broad rewrites.
- Match the conventions already present in each file — do not impose a new style.

## Output

Apply the edits, then give a concise 1-3 sentence summary of what was removed or simplified.
