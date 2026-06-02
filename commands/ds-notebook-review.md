Run a strict review of Jupyter notebooks — find notebook-state, output-hygiene, reproducibility, and data-science correctness defects a line-by-line code review can't see. Reports a findings list by default; `--fix` applies the mechanical, unambiguous fixes (logic-changing or uncertain ones stay reported).

When invoked, audit the `.ipynb` files in scope against one question: **will this notebook reproduce its stated results cleanly, on another machine, without leaking state or secrets?** Code-level Python quality is **not** this command's job — idioms, typing, and generic security in the cell code go to `/ds-python-review`; a ranked, costed optimization plan is `/ds-perf-plan`; pipeline/ETL correctness is `/ds-data-review --pipelines`. Every finding names a concrete failure — a hidden-state break, a committed secret, a leakage path — anchored to a cell, not a vague best-practice. **Do not edit any files unless `--fix` is passed** (see Arguments).

## Arguments

- Treat positional args as scope (notebook files, directories, globs). With no scope, review the `.ipynb` files changed on the current branch.
- Freeform scope ("the training notebook", "the EDA folder") is interpreted reasonably.
- `--fix` → after reporting, apply only the findings whose fix is **mechanical and unambiguous**: strip committed output cells, reset `execution_count` to `null`, remove a scratch cell the review is certain is dead. Anything that changes execution semantics or rests on an assumption — setting a seed, reordering cells, rewriting leakage-prone preprocessing, deleting a cell whose effect you can't confirm — **stays report-only**. Use `NotebookEdit` (or `jupyter nbconvert --clear-output`) for the mechanical edits, confirm the notebook JSON still parses, then close with a summary of what was applied and what was left.

## What to check

**1. Execution & hidden state.** Non-monotonic or gapped `execution_count` (cells run out of order); outputs present while `execution_count` is `null` (ran, then the code was edited — outputs are stale); a cell that reads a name defined only in a *later* cell, or in a cell since deleted; reliance on manual run order so a clean *Restart & Run All* would fail; results that only hold if a cell is run more than once (accumulating mutable state).

**2. Committed outputs & secrets.** Output cells committed at all — stale against the code, repo bloat, large/binary blobs (base64 images, full DataFrame dumps) inflating diffs. Critically: **secrets, API tokens, credentials, connection strings, or PII visible in code *or* output cells** — a printed `df.head()`, an env dump, a stack trace exposing a key. Absolute local paths or machine-specific data baked into outputs.

**3. Reproducibility.** No seed set for stochastic work (`numpy`, `random`, `torch`, `tensorflow`, sklearn `random_state`) when the notebook reports a result; unpinned/implicit dependencies and `%pip install` run mid-notebook instead of declared; reliance on ambient kernel or global state rather than explicit setup; hardcoded absolute paths (`/Users/...`, `C:\\...`); results that depend on nondeterministic ordering (dict/set iteration, unsorted glob).

**4. Data-science correctness.** Train/test **leakage** — fitting a scaler, encoder, imputer, or feature selector on the full dataset *before* the split; target leakage from a feature derived after the outcome is known; evaluating on training data. Pandas chained assignment / `SettingWithCopyWarning`; a surprising `inplace=True`; silent dtype coercion (ids to float, dates to object); `df` re-bound across cells so its state at any point is ambiguous; a merge/join that silently duplicates or drops rows (wrong key cardinality, missing `validate=`).

**5. Notebook hygiene & structure.** Exploration tangled with production logic that should be extracted to an importable module; mega-cells doing many unrelated things; dead scratch cells; top-level mutable globals where a function would isolate state; exceptions silently swallowed in a `try/except` to keep a cell "green".

**6. Delegations — state, don't duplicate.** When the *cell code* itself warrants a deeper pass — language idioms, ranked perf, ETL correctness — name the gap and route it to the owning command (named above) rather than reviewing it here.

## Output

A prioritized findings list, ordered by severity (likelihood × impact):

1. **Critical** — a committed secret or credential, a data-leakage path that invalidates the reported results, or irrecoverable data loss.
2. **Major** — a hidden-state break that defeats *Restart & Run All*, committed outputs stale against the code, or a missing seed under a result the notebook reports.
3. **Minor** — hygiene and structure: scratch cells, mega-cells, extractable logic.

For each finding:

- Anchor to `<file> cell <N>` (add the line within the cell when it helps).
- State the hazard in one line, **name the exact condition that triggers it** (the stale output, the pre-split `fit`, the forward reference), then the fix.

Rules:

- Real defects only. A "best practice" with no demonstrated hazard is noise.
- A short, high-confidence list beats a long speculative one.
- Report-only by default — the output is the list. With `--fix`, apply only the mechanical, unambiguous fixes above and leave every semantic or assumption-dependent finding reported; then summarize what was applied vs. left.
