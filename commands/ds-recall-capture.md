Store the current session's outcome into recall's local knowledge base.

Extracts the signal from the session — goal, result, key insight — and feeds it to [recall](https://github.com/gleicon/recall) so the work survives the context window and becomes reusable across projects. Strips reasoning chains and failed attempts; keeps only the compact recipe.

## Check

Verify `recall` is on PATH:

```bash
command -v recall
```

If missing, print install instructions (see `/ds-recall`) and stop.

## Process

1. Check binary.
2. Use `/ds-tldt` to compress the session into a compact summary — three fields only:
   - **Task** — what was being implemented or resolved (one line)
   - **Result** — what was produced or fixed (one line)
   - **Insight** — the non-obvious takeaway, constraint, or pattern (one to three lines)

   Exclude: intermediate reasoning, failed approaches, scaffolding discussion.

3. Call recall:
   ```bash
   recall run record --task "<task>: <result>"
   recall learn "<insight>"
   ```

4. Report what was stored.

## Rules

- Never capture the full session transcript — signal only.
- If the session had no clear resolution (still in progress, inconclusive debug), report that and skip capture. Do not store partial or speculative outcomes.
- Run this **before** `/clear` or session end, not after — the context is gone once cleared.

## Output

```
recall: stored
  task:    <task line>
  result:  <result line>
  insight: <insight>
```

Or, if capture is disabled:

```
recall: capture disabled — run /ds-recall-capture --enable to turn it on
```
