Run a strict data review of code changes — find data-correctness, integrity, and migration-safety defects. Store-agnostic (relational or NoSQL). Reports a findings list by default; `--fix` applies the mechanical, unambiguous fixes (logic-changing or uncertain ones stay reported).

When invoked, audit the code in scope against one question: **is the data correct, consistent, and well-modeled?** Not "can it be injected" (that's `/ds-security-review` — assume queries are parameterized here), not "is the query fast" (that's `/ds-perf-plan`; the line is *consequence* — a slow query is perf, a query that returns *wrong or duplicate data* is here), not general code logic (that's `/ds-bug-review`). Every finding names the concrete condition that produces wrong, lost, or inconsistent data; a "best practice" with no demonstrated hazard is noise. Adapt to the store — don't demand relational constructs from a document database. **Do not edit any files unless `--fix` is passed** (see Arguments). When a finding is confirmed, `/ds-verify-this` proves the fix against real before/after data.

## Arguments

- Treat positional args as scope (files, directories, globs — including schema and migration files). With no scope, review the code changed on the current branch.
- Freeform scope ("the orders schema", "the reporting queries") is interpreted reasonably.
- State the store/engine if you know it (Postgres, MySQL, SQLite, Mongo, DynamoDB…) — isolation defaults and SQL dialects differ. Otherwise infer from config/driver and **state the assumption**.
- `--pipelines` extends the review with a sixth area — data-pipeline / ETL correctness (idempotency, replay/backfill safety, late & out-of-order data, dedup, schema drift). **Off by default**: the default review covers the operational store. With the flag on, it's the after-the-fact audit counterpart to the `/ds-data-mode` build mode (mode shapes the pipeline build; this audits the built pipeline code).
- `--fix` → after reporting, apply only the findings whose fix is **mechanical and unambiguous** — a single obvious edit, no design judgment (e.g. adding a missing `NOT NULL`/`UNIQUE` the review is certain about, correcting a column type). A wrong fix to a data finding is worse than none, so anything that changes behavior, alters a migration's effect, or rests on an assumption you couldn't verify **stays report-only**. After applying, re-run any build/test/lint check already in the loop and revert any fix that breaks it — or that touched more than the intended mechanical edit. Close with a summary of what was applied and what was left.

## What to check

**1. Schema & data modeling.** Missing constraints (`NOT NULL`, `FK`, `UNIQUE`, `CHECK`) that let invalid state exist; wrong types (money as float, naive timestamps without timezone, enums as free text, ids as nullable); normalization/denormalization that doesn't fit the access pattern; referential integrity gaps (deletes that orphan rows). NoSQL: embedding vs referencing for the read pattern, hot/low-cardinality partition keys, documents that grow without bound, multiple sources of truth that can diverge.

**2. Integrity & invariants.** Invariants the application *assumes* but only enforces in app code (so a concurrent path or a second writer breaks them) — uniqueness, sums that must reconcile, state machines. Prefer the store enforcing them (constraint, unique index) over a check-then-write that races.

**3. Query result correctness.** Wrong JOIN type silently dropping or duplicating rows; `NULL` semantics in `WHERE`/comparisons and aggregates (`NULL != NULL`, `COUNT(col)` skipping NULLs, `AVG`/`SUM` over NULLs); implicit type coercion; timezone/precision loss in aggregates; `LIMIT` without a total `ORDER BY` giving nondeterministic rows; pagination drift (OFFSET over a changing set).

**4. Transactions & consistency.** Multi-statement invariants not wrapped in a transaction (partial writes on failure); wrong or assumed isolation level (lost updates, non-repeatable reads, phantoms under the engine's *actual* default); read-modify-write races; cross-aggregate writes assuming atomicity a NoSQL store doesn't provide; eventual-consistency reads treated as strongly consistent.

**5. Migration safety.** Backward-incompatible schema changes deployed against running code (drop/rename a column the old version still reads); locking DDL on large tables (add-column-with-volatile-default, non-concurrent index build, type change that rewrites the table); non-idempotent migrations; backfills that are wrong, unbatched, or racy against live writes; missing or untested rollback path.

**6. Pipeline & ETL correctness (`--pipelines` only).** The after-the-fact audit of pipeline code — the mirror of `/ds-data-mode`'s build-time constraints. Non-idempotent writes (blind append where re-running double-writes, instead of upsert/merge on a key); transform logic that depends on wall-clock, randomness, or arrival order. Processing-time windowing where event time is needed (late/out-of-order data silently dropped or miscounted; no watermark/cutoff); no dedup under at-least-once delivery; bad records that crash the batch instead of being quarantined; schema drift silently coerced or dropped instead of failing loudly at a pinned contract. Backfills that double-count (not replayable); no time-partitioning to bound reprocessing; no checkpoint to resume from; destructive overwrite with no recovery path; a delivery guarantee (at-least-once vs exactly-once) the sink doesn't actually satisfy. Missing boundary assertions (counts, ranges, referential, reconciliation) that let silently-wrong data publish.

## Output

A prioritized findings list, ordered by severity (likelihood × impact):

1. **Critical** — silent data loss or corruption, or a migration that can lock/break production.
2. **Wrong results** — queries that return incorrect, duplicated, or missing data under normal use.
3. **Integrity gap** — an invariant the app assumes but the store doesn't enforce (breakable by a race or a second writer).
4. **Hardening** — a constraint or type that would prevent a class of future data bug, not yet causing one.

With `--pipelines`, pipeline/ETL findings join the same ladder — a backfill that double-counts or an overwrite with no recovery path is *Critical*; a missing boundary assertion that lets silently-wrong data publish is *wrong-results* or *integrity-gap*.

For each finding:

- Anchor to `file:line` (or the schema/migration file).
- State the hazard in one line, **name the exact condition or data that triggers wrong/lost/inconsistent data**, then the fix — prefer a store-enforced constraint over an app-side check when it removes a race.
- Give your confidence and state the store/engine assumption it rests on (isolation level, dialect).

Rules:

- Real data hazards only. Name the path to wrong, lost, or inconsistent data; a "weakness" with no such path is hardening at most.
- Store-agnostic: don't flag the absence of relational constructs in a document store — judge against *that* store's correctness model.
- A short, high-confidence list beats a long speculative one.
- Report-only by default — the output is the list. With `--fix`, apply only the mechanical, unambiguous findings above and leave every judgment- or assumption-dependent one reported; then summarize what was applied vs. left.
