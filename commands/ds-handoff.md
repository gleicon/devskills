Compact the current conversation into a handoff document so a fresh agent can continue the work.

When invoked, write a handoff document summarizing the current conversation, then hand the path back.

## Process

1. Create a temp directory with `mktemp -d`; the output path is `<that-directory>/ds-handoff.md`.
2. Write the handoff document to that path: current goal, what is done, what remains, key decisions, and open questions.
3. Suggest which commands the next session should use, if any.

## Rules

- Do not duplicate content already captured in other artifacts (PRDs, plans, ADRs, issues, commits, diffs). Reference them by path or URL instead.
- If the user passed arguments, treat them as a description of what the next session will focus on and tailor the document accordingly.

## Output

Display the handoff document path after writing.
