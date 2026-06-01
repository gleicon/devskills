Activate data mode for this session.

When active, build data pipelines and transforms with correctness designed in, not bolted on. These are tool-agnostic principles — apply them whether the work is batch or streaming, on any orchestrator or engine (Spark, Airflow, dbt, Flink, or plain scripts). Match the project's existing stack and conventions; never introduce a new framework where the codebase already has one. This mode is the build-time complement of `/ds-data-review`, which audits the operational store (schema, queries, transactions, migrations) after the fact — and, run as `/ds-data-review --pipelines`, audits the pipeline code itself. This mode shapes how the pipeline gets built; that review audits it once it's built.

## Correctness & determinism

- Transforms are **idempotent**: re-running a step on the same input yields the same output and the same end state. Write by upsert/merge on a key, not blind append.
- No hidden nondeterminism — no dependence on wall-clock, randomness, or arrival order inside transform logic. If a step needs "now" or a seed, inject it so a run is reproducible.
- Separate pure transformation from IO: pure functions you can test in isolation, IO pushed to the edges.

## Real-world data

- Assume data arrives **late and out of order**. Key and window by *event* time, not processing time; state explicitly how late is too late (watermark / cutoff).
- Assume **duplicates**. Dedup on a stable business key; design every write to be safe under at-least-once delivery.
- Don't trust the shape. Validate records at the boundary; **quarantine** malformed rows (dead-letter) instead of crashing the batch or silently skewing it.
- **Schema drift is expected.** Pin an explicit schema/contract at each boundary and evolve it deliberately (additive, versioned). Fail loudly on an incompatible change rather than silently dropping or coercing columns.

## Reprocessing & recovery

- Every pipeline is **replayable**: backfilling a past window produces correct results with no double-counting (idempotent writes + time partitioning).
- Partition by time so you can reprocess a bounded window, not the whole history.
- **Checkpoint** progress; on failure, resume without reprocessing committed work or dropping in-flight work.
- No destructive overwrite without a recovery path — write a new partition/table and swap, or retain the prior version.
- State the delivery guarantee you target (at-least-once vs exactly-once) and design the sink for it (idempotent or transactional writes).

## Data quality & observability

- Assert invariants at boundaries — row counts in an expected range, no nulls in keys, referential checks, totals reconcile. **Fail the run** on violation; never publish silently-wrong data.
- Emit freshness, volume, and basic distribution signals so a stalled or skewed pipeline is visible here, not discovered downstream.
- Keep failures loud and localized — a bad partition fails its own partition, not the whole job.

## Structure

- Separate extract / transform / load; keep transforms free of source- and sink-specifics so they stay testable and portable.
- Prefer incremental processing over full-refresh where the data model allows; full-refresh only when it's genuinely cheaper or safer.
- Align partitioning to both the read access pattern and the reprocessing unit.

Confirm activation with "Data mode active." then proceed.
