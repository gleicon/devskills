# Command Reference

Every devskills command is a single prompt file invoked as `/<name>` in Claude Code, OpenCode, or Cursor (and as `/prompts:<name>` in OpenAI Codex). This is the reference: what each one does, its arguments, and when to reach for it. For worked, multi-step workflows see [recipes.md](recipes.md).

## Kinds of command

A command's **suffix tells you its kind**:

- **`-mode`** — persistent, toggleable session behavior; changes *how* the agent works until you turn it off. *tiger-style, ui, test, tdd, data, git, step, quality-gate, caveman-lite/ultra.*
- **`-review`** — a findings-list audit. Report-only by default (several take `--fix`); findings are independent and fixable in any order. *bug, security, data, code-quality, doc-quality, test-quality, ui-quality, comment, and the six language reviews.*
- **`-plan`** — graded, sequenced moves that each carry a trade-off or dependency, so the output is a *plan*, not a verdict. *perf-plan, architecture-plan.*
- **no suffix** — a one-shot action that produces a result and returns. *spec, roadmap, explore, blueprint, grill-me, modes, review, handoff, zoom-out, tldt, verify-this, debug, deslop, write-a-command, and the project-\* family.*
- **language profiles** — configured per project via `--lang=<x>`, not invoked as slash commands (see the [README](../README.md#language-profiles)).

Everything except `-mode` runs once and finishes; a `-mode` stays on. The per-command headings below tag each one with its kind. No command needs external tooling — every one stands alone.

---

## Specs & planning

### `/ds-spec` — action

Turn a rough description into a structured specification (the WHAT, not the HOW).

- **Args:** an optional description. With one, it proceeds directly; without, it asks three focused questions (primary user action, what success looks like, hard constraints) then writes the spec.
- **Output:** `.project/SPEC.md` if `.project/` exists, else `SPEC.md` in the current directory, shown inline. Sections: Problem, Scope, Users, Functional/Non-Functional Requirements, Interfaces, Constraints, Acceptance Criteria, Open Questions.
- **Reach for it when:** you have an idea and want a verifiable contract before any code.

### `/ds-roadmap` — action

Turn a goal, a `SPEC.md`, or another command's output (e.g. `/ds-code-quality-review` findings, a bug list) into an ordered `## Roadmap` task checklist. The companion to `/ds-spec` — spec defines the WHAT, this orders the work to get there. Sequences and scopes; does not choose architecture.

- **Output:** `.project/PLAN.md` if `.project/` exists, else `PLAN.md` in the current directory (`.project/` is never created for you). Appends to existing tasks; preserves a `## Now` section (that belongs to `/ds-project-checkpoint`).
- **Reach for it when:** you have a spec, a goal, or a pile of findings and want them turned into an ordered, shippable task list.

### `/ds-explore` — action

Surface candidate approaches to a problem with their trade-offs — suggests, never decides, never implements. Reads `.project/` state and decisions (and the code) so options respect reality; can do bounded, cited web research with `--web` (off by default — and if local context is too thin for good options, it says so and suggests `--web` rather than guessing). Writes a scratch `.project/EXPLORE.md` (or a temp path) and lists the open questions the choice hinges on.

- **Args:** a problem/question; `--web` to opt into web research.
- **Reach for it when:** facing a "how should I build this?" fork and you want options laid out before deciding. It's the upstream of `/ds-grill-me`: explore generates the options, `/ds-grill-me` walks you through choosing.

### `/ds-blueprint` — action

Design a target architecture for a **new** system from its requirements — module/boundary decomposition, dependency-direction rules, seams, and build order (walking skeleton first). The **decisive** counterpart to `/ds-explore`: where explore surveys options and abstains, blueprint commits to one structure and notes the strongest alternative briefly. Describes the structural *how*, not the behavioral *what* (`/ds-spec`); to critique an architecture that *already exists*, use `/ds-architecture-plan`. The spine is YAGNI: **no structural element without a requirement that demands it** — generic patterns (hexagonal, DDD, microservices) appear only when a stated need forces them.

- **Args:** a requirements source (a `SPEC.md` path, a freeform description, or a chosen approach from `/ds-explore`); with none, reads `SPEC.md` if present or asks. `--no-tiger` skips the Tiger Style section.
- **Output:** a blueprint — shape (+ the tier it's pitched at), modules/boundaries, dependency rules (acyclic), seams, build order, what's deferred (and what would justify adding it), and the alternative considered. Changes nothing.
- **Reach for it when:** starting a new system or subsystem and you want the structural HOW committed before building. Hand the build order to `/ds-roadmap`.

### `/ds-grill-me` — action

Interview you relentlessly about a plan or design until you share the same understanding. One question at a time, each with a recommended answer; explores the codebase instead of asking when it can.

- **Args:** `--record` appends every resolved decision to `DECISIONS.md` (question, answer, one-line rationale).
- **Output:** a resolved-plan summary once no decision branches remain; the `DECISIONS.md` path if `--record` was used.
- **Reach for it when:** a plan feels under-specified, or you want to pressure-test a design (including the approach in a draft PR) before committing to it.
- **More:** `/ds-grill-me` is unusually versatile — see [grill-me.md](grill-me.md) for a full menu of uses (requirements discovery, design/refactor/architecture, domain terminology, non-coding decisions).

### `/ds-workflow` — action

A standalone phase-map orchestrator — orients you, then routes each phase (orient → spec → plan → build → clean → review → verify → ship) to its primary command. Works fully on its own. When `.project/` is in use it *also* reads that state to orient faster (`PROJECT.md` for context, `PLAN.md` for where to resume) and surfaces the `/ds-project-*` commands as options — but never requires them; with no plan or no `.project/` it points at `/ds-spec` or `/ds-explore` to begin. The `/ds-project-*` family is optional persistence layered on top, not a dependency.

---

## Project memory (`.project/`)

A minimal, file-backed project memory — three commands that keep a durable description, plan, and session state in plain markdown under `.project/`, so any session is safe to `/clear` or end. (The plan's `## Roadmap` is seeded by `/ds-roadmap` above; these maintain and restore it.) These are *scribes, not pilots*: they record what you decide, never steer architecture. Walkthrough: [project-workflow.md](project-workflow.md). Worked use cases: [project-recipes.md](project-recipes.md).

### `/ds-project-map` — action

Scan the repo and write/refresh `.project/PROJECT.md` (overview, stack, repo map, constraints). Facts only — describes what exists. Run once at start; re-run when the repo drifts.

### `/ds-project-checkpoint` — action

Update `.project/PLAN.md`'s `## Now` (state, next, open questions) and roadmap statuses. Run before `/clear` or end of session. `--handoff` also writes a richer `.project/handoff.md`.

### `/ds-project-resume` — action

Read `.project/PLAN.md` (+ `PROJECT.md`) and report where to pick up. Loads `handoff.md` only if it's newer than the plan (by file time — no git dependency, so `.project/` can be git-ignored), else flags it stale. Read-only.

---

## Engineering modes

### `/ds-tiger-style-mode` — mode

Activate [Tiger Style](https://tigerstyle.dev/) engineering constraints for the session: safety, explicitness, bounded everything, assertions, no silent fallbacks. Active during all review commands too.

- **Reach for it when:** starting focused implementation work you want held to a strict bar.

### `/ds-ui-mode` — mode

UI mode, framework-agnostic (React/Svelte/Vue/Solid/vanilla, any runtime). Two halves: **engineering** (one-responsibility components, minimal co-located state, derive-don't-store, explicit loading/error/empty/success, typed boundaries, stale-response cancellation) and **design craft** (type scale + spacing tokens, deliberate visual hierarchy, intentional motion — encoding constraints to escape the generic AI look). Plus accessibility as a requirement and Core Web Vitals targets (LCP/INP/CLS). Adapts to the project's existing stack rather than imposing libraries.

- **Reach for it when:** working on components, UI state, API integration, or styling.

### `/ds-data-mode` — mode

Data-engineering discipline, tool-agnostic (Spark/Airflow/dbt/Flink/plain scripts, batch **and** streaming). Counters the naive ETL shape LLMs default to (read-all → transform → overwrite, assume data arrives once and in order, crash on a bad record, no replay) by encoding constraints up front: **idempotent** transforms (upsert on a key) with no wall-clock/random/order dependence, real-world data handling (event-time windowing with explicit watermarks, dedup on a business key safe under at-least-once, quarantine malformed rows, explicit schema-drift contracts), reprocessing & recovery (replayable/backfillable with no double-counting, time-partitioned, checkpointed, no destructive overwrite without a recovery path), boundary data-quality assertions that **fail the run**, and E/T/L separation with testable pure transforms. Matches the project's existing stack rather than imposing a framework.

- **Reach for it when:** building or extending a data pipeline or transform. Stacks with `/ds-tiger-style-mode` + `/ds-test-mode`. The after-the-fact audit counterpart is `/ds-data-review` (the store) and `/ds-data-review --pipelines` (the pipeline code itself) — mode shapes the build, review audits it.

### `/ds-git-mode` — mode

Senior-engineer commit discipline — commits written **for humans, not LLMs**. Counters the two LLM failure modes: commit-message bloat (diff re-narrated as bullet lists, emoji, "Generated with…" trailers) and mis-timed commits (a flurry of WIP commits, or one giant end-of-session dump). When active: commit each **self-contained, working** unit as it's done (builds/passes, reversible, reviewable — one logical change, not bundled, not dribbled), with terse Conventional-Commit messages (`type(scope): subject`, imperative, ~50 chars, body only for non-obvious WHY). Activating the mode is standing authorization to commit completed units without asking; it reports a one-line subject per commit. Branch-first (auto-creates a type-prefixed branch off the default branch before the first commit). **Never pushes, never rewrites shared history** (no rebase/squash/force-push) — the irreversible actions stay user-driven.

- **Reach for it when:** any build session where you want clean, honest, human-readable history as you go. Stacks with `/ds-tiger-style-mode` + `/ds-test-mode` (disciplined code *and* disciplined history). It also governs the commit messages the caveman modes deliberately leave "written normally."

### `/ds-step-mode` — mode

User-driven, step-gated execution — deliberately **inverts** the default autonomy, pausing *more*, not less. When active: do the smallest meaningful, reviewable step, **propose before acting and wait** (a free veto before any change), then after the step stop and report concisely (did / changed / next) and yield. Never silently chains steps; granularity is tunable live ("bigger/smaller steps"). The handback rule is the sharp part: **always return control in prose, never via the multiple-choice picker** — options are prose suggestions the user can accept, amend, *or combine*, never a forced single-select. The picker is allowed only for trivial either/or disambiguation ("did you mean file A or B?"), never for "what next." Invoke with a plan — `/ds-step-mode current plan`, a path, or pasted text — to drive an existing plan one step at a time, keeping it honest and offering `/ds-project-checkpoint` at milestones.

- **Reach for it when:** you want to drive the work yourself, keeping control at every decision point. Composes with `/ds-git-mode` (a "step" ≈ a commit unit), `/ds-tiger-style-mode`, `/ds-test-mode`. Unlike `/ds-tdd-mode` (steps for design reasons) or `/ds-grill-me` (interrogates a plan, doesn't execute), this gates *execution* under your control.

---

## Test & build

### `/ds-tdd-mode` — mode

Drive implementation test-first, one **vertical** slice at a time (one test → minimal implementation → repeat). Refuses horizontal slicing (all tests up front). Tests exercise observable behavior through the public interface, never internals.

- **Output:** per slice, the test then the implementation; reports which behaviors remain untested.
- **Reach for it when:** building a feature you want anchored to real, refactor-survivable behavior.

### `/ds-test-mode` — mode

Pragmatic testing mode: as you build normally, ensure the code that matters gets well tested — without `/ds-tdd-mode`'s test-first ceremony. Tests by **risk, not rule** (core logic, edge cases, a regression test for every bug fixed; skips trivia), insists on behavior-through-the-interface tests that survive refactors, and refuses design-locking tests (mock-the-world, asserting internals, snapshot-everything). Coverage is a side effect, never the goal.

- **Reach for it when:** doing normal implementation work you want kept honestly tested. The alongside-work sibling of `/ds-tdd-mode` (test-first); audit existing tests with `/ds-test-quality-review`.

---

## Quality & cleanup

### `/ds-code-quality-review` — review

Extremely strict maintainability audit: abstraction quality, file sprawl (the 1k-line smell), spaghetti-condition growth. Ambitiously hunts "code judo" — restructurings that delete whole categories of complexity while preserving behavior.

- **Args:** treated as review **scope** (files or directories). With no scope, reviews the changed files on the current branch. Freeform scope ("the whole codebase") is interpreted reasonably.
- **Output:** prioritized findings anchored to `file:line`, with an approval verdict. Changes nothing by default; `--fix` applies the mechanical, behavior-preserving findings — structural/code-judo restructurings stay reported.
- **Reach for it when:** before merging non-trivial work, or auditing an area you suspect is decaying.

### `/ds-doc-quality-review` — review

Strict documentation audit governed by one principle — **docs earn their length** (readers skim, they don't read). Hunts wrong docs (drifted from the code) and bloated docs (true, but nobody reads them) with equal energy. Verifies mechanically: resolves links, recounts claimed counts, runs example commands where safe.

- **Args:** treated as scope (files, directories, globs); defaults to docs changed on the current branch. `--comments` also audits inline code comments (off by default).
- **Output:** prioritized findings anchored to `file:line` with a suggested fix — accuracy/drift first, then dead links and wrong counts, missing docs, bloat-to-cut, clarity. Changes nothing by default; `--fix` applies the mechanical, unambiguous findings (dead links, stale counts, lossless bloat-cuts).
- **Reach for it when:** before a docs PR, after the code outgrew its README, or when the docs feel long and unread.

### `/ds-test-quality-review` — review

Strict test-suite audit governed by one principle — **test what matters, and test it well — not coverage.** Hunts critical/core code that's under-tested (the bug waiting to ship) and tests that are bad (green but worthless, or design-locking and worse than nothing) with equal energy. Checks edge/failure-mode coverage on risky logic, behavior-vs-implementation quality, and flags tests coupled to internals for rewrite-or-delete. Explicitly rejects coverage-chasing.

- **Args:** treated as scope (files, directories, globs); defaults to tests covering code changed on the current branch.
- **Output:** prioritized findings anchored to `file:line` — untested critical code first, then missing edge cases, design-locking tests, weak tests, bloat-to-cut. Changes nothing by default; `--fix` applies the mechanical, unambiguous findings (deleting worthless or duplicate tests) — writing or redesigning tests stays reported.
- **Reach for it when:** before merging logic-heavy work, or when a suite is green but you don't trust it. The audit counterpart to the `/ds-test-mode` mode and `/ds-tdd-mode`.

### `/ds-ui-quality-review` — review

Strict UI audit governed by one principle — **a UI is judged on both halves: it works and it's crafted.** Framework-agnostic. Hunts engineering correctness (missing/broken async states — especially **empty** — fetch waterfalls, uncancelled stale responses, state that should be derived, index-as-key), accessibility barriers (non-semantic controls, keyboard/focus gaps, contrast, missing labels/live regions), Core Web Vitals (layout shift, INP, oversized critical path), and design craft (generic-AI defaults, flat hierarchy, unsystematized type/spacing).

- **Args:** treated as scope (files, directories, globs); defaults to UI code changed on the current branch.
- **Output:** prioritized findings anchored to `file:line` — engineering correctness first, then a11y, performance, design craft. Each names the concrete failure, not "improve a11y." Changes nothing by default; `--fix` applies the mechanical, unambiguous findings (a missing label, a non-semantic control, index-as-key) — design-craft and async-state fixes stay reported.
- **Reach for it when:** before merging UI work, or auditing an interface you suspect is broken-on-edges, inaccessible, slow, or generic. The audit counterpart to the `/ds-ui-mode` mode.

### `/ds-deslop` — action

Strip AI-generated slop from the branch and align it with the surrounding code. The test is "code an experienced engineer in this language and codebase wouldn't write" — judged against the language's idioms, not a fixed syntax list. Targets narrating comments, defensive overkill abnormal for a trusted path, type escape hatches used only to dodge the checker, needless nesting, and speculative scaffolding.

- **Args:** treated as scope (files or directories); defaults to the branch diff against main.
- **Output:** the edits applied, plus a 1–3 sentence summary. Behavior preserved.
- **Reach for it when:** right after generating a batch of code, before review. Cheaper and narrower than `/ds-code-quality-review`.

### `/ds-comment-review` — review

Strict review of code comments under one lens — **does each comment earn its place, and is it as short as it can be?** Comments are for humans and explain **WHY, not WHAT** — one line by default, only where the reason isn't obvious, never restating code or citing plan/ticket IDs; a long comment is rare and signals importance. Unlike `/ds-doc-quality-review --comments` (reports under a docs-accuracy lens) and `/ds-deslop` (branch-diff, matches existing style), this **imposes** the discipline regardless of the codebase's existing habits, works on any scope, and can apply the fix. Comment-only and behavior-preserving — never changes code logic.

- **Args:** treated as scope (files, directories, globs, or the whole codebase); defaults to comments in the code changed on the current branch. `--fix` applies the edits in place instead of only reporting.
- **Output:** by default a prioritized findings list anchored to `file:line`, grouped delete / tighten / drifted (correctness) / kept-as-important, each with the fix. With `--fix`, the edits applied plus a 1–3 sentence summary.
- **Reach for it when:** a codebase has accumulated comment bloat, or you want a fresh branch's comments brought to standard. Flags restate-the-code, obvious/ceremonial, planning cruft, and buried WHY; keeps and respects the rare legitimately-long comment.

---

## Quality Gate

### `/ds-quality-gate-mode` — mode

Seven-pass review pipeline for a feature branch or scoped path, **bookended by `/ds-deslop`**. Run in order; implement accepted findings between passes before proceeding.

Pipeline: `/ds-deslop` → `/ds-test-quality-review` → `/ds-security-review` → `/ds-bug-review` → `/ds-data-review` → `/ds-code-quality-review` → `/ds-doc-quality-review` → `/ds-deslop`

Deslop runs first to clean the incoming diff, then again last because the gate implements fixes between passes — that freshly-generated fix code can carry its own slop. `/ds-data-review` runs only when the change touches schema, queries, transactions, or migrations; otherwise it's skipped with a note.

After each pass: shows findings for that pass, asks "accept all / reject all / skip N", implements accepted ones, then moves to the next pass. Reports a one-paragraph summary at the end.

- **Args:** optional path scope (applies to every pass). None → branch diff.
- **Toggle:** `/ds-quality-gate-mode` to activate, `/ds-quality-gate-mode off` or "stop quality gate" to deactivate.
- **Reach for it when:** finishing a feature branch before opening a PR; each pass answers a different question so running all six is not redundant.

---

## Reviews

The review commands are a **layered gate, not competing alternatives** — cheapest and narrowest first, deepest last: `/ds-deslop` (noise) → `/ds-bug-review` (correctness) → `/ds-security-review` (exploitability) → `/ds-data-review` (data correctness, when the change touches schema/queries/transactions/migrations) → the language review (idioms) → `/ds-code-quality-review` (structure). Each answers a different question, so running several on the same code isn't redundant. The full pre-PR sequence is in [recipes.md](recipes.md#a-pre-pr-quality-gate).

### `/ds-bug-review` — review

Language-agnostic **correctness** audit — the bug-hunting pass. Asks one thing: *will this misbehave at runtime?* Hunts logic errors, null/absent-value derefs, swallowed errors and half-done failure paths, resource leaks, races (TOCTOU, lock ordering), boundary/overflow mistakes, and contract misuse. Distinct from `/ds-code-quality-review` (which is maintainability, not bugs) — the same split the harness draws between cleanup and correctness.

- **Args:** treated as scope (files, directories, globs); defaults to code changed on the current branch.
- **Output:** prioritized findings anchored to `file:line` — critical (data loss / reachable crash) first, then likely-wrong, then edge-case. Each names **the exact condition that triggers it** plus the fix and a confidence note. Real defects only; no theoretical nulls. Changes nothing by default; `--fix` applies only mechanical, unambiguous fixes — logic-changing or uncertain ones stay reported.
- **Reach for it when:** before merging logic-heavy work, or on any code outside go/ts/rust where there's no language review. Confirmed findings hand off to `/ds-debug` (root-cause) and `/ds-verify-this` (prove the fix).

### `/ds-security-review` — review

Language-agnostic **security** audit — the portable counterpart to the per-language Security sections. Traces untrusted data from entry to dangerous sink: injection (SQL/command/path/SSRF/template), output handling (XSS, unsafe deserialization), broken access control (IDOR, privilege escalation), secrets and weak crypto, sensitive-data exposure, mass assignment / unsafe upload / DoS, and transport/config gaps.

- **Args:** treated as scope (files, directories, globs); defaults to code changed on the current branch.
- **Output:** prioritized findings anchored to `file:line` — critical (code exec / breach / auth bypass) → high → hardening. Each **describes the attack** (input → sink) and the fix. Exploitable over theoretical. Changes nothing by default; `--fix` applies only mechanical, unambiguous fixes — anything that changes behavior or rests on an assumption stays reported.
- **Reach for it when:** any change that touches input handling, auth, secrets, or external I/O — and as a pre-PR gate. The deeper language-specific checks live in `/go·ts·rust·python·java·zig-review`.

### `/ds-data-review` — review

Store-agnostic **data correctness** audit — the question no other review owns: *is the data correct, consistent, and well-modeled?* Works on relational **and** NoSQL, adapting to the store (won't demand FKs from a document database). Checks schema & integrity (missing constraints, wrong types, referential gaps, partition-key hotspots, unbounded documents), query-result correctness (JOINs that drop/duplicate rows, NULL/aggregate semantics, `LIMIT` without `ORDER BY`, pagination drift), transactions & consistency (missing boundaries, wrong isolation level, lost updates, eventual-consistency-read-as-strong), and migration safety (backward-incompatible DDL against running code, locking DDL on large tables, racy backfills, missing rollback). The line vs neighbors is drawn by *consequence*: a query that's slow → `/ds-perf-plan`; one that returns wrong/duplicate data → here. Injection stays with `/ds-security-review` (assumes parameterized queries); general code-logic races stay with `/ds-bug-review`.

- **Args:** scope (files, directories, globs — including schema and migration files); defaults to code changed on the current branch. State the store/engine (Postgres, Mongo, DynamoDB…) when known — isolation defaults and dialects differ; otherwise it infers and **states the assumption**. `--pipelines` adds a sixth area — data-pipeline / ETL correctness (idempotency, replay/backfill safety, late & out-of-order data, dedup, schema drift); off by default, and the after-the-fact audit counterpart to the `/ds-data-mode` build mode.
- **Output:** prioritized findings anchored to `file:line` — critical (silent data loss/corruption, or a migration that can lock production) → wrong-results → integrity-gap → hardening. Each names **the exact condition that triggers wrong/lost/inconsistent data**, the fix (prefer a store-enforced constraint over an app-side check that races), and the store/engine assumption it rests on. Changes nothing by default; `--fix` applies only mechanical, unambiguous fixes — migration-altering or uncertain ones stay reported.
- **Reach for it when:** a change touches schema, queries, transactions, or migrations (add `--pipelines` for ETL/pipeline code). Confirmed findings hand off to `/ds-verify-this` (prove the fix against real before/after data). The build-time complement is the `/ds-data-mode` mode.

### `/ds-go-review` · `/ds-ts-review` · `/ds-rust-review` · `/ds-python-review` · `/ds-java-review` · `/ds-zig-review` — review

Language-specific review passes.

- **`/ds-go-review`** — idiomatic Go + security + Tiger Style.
- **`/ds-ts-review`** — TypeScript/Workers: strict mode, React, Cloudflare.
- **`/ds-rust-review`** — cargo geiger, `unsafe`/panic counts, clippy, audit, Tiger Style + security.
- **`/ds-python-review`** — Python idioms, `mypy --strict` typing, security, Tiger Style.
- **`/ds-java-review`** — Java idioms (records, sealed types, pattern matching), security, Tiger Style.
- **`/ds-zig-review`** — explicit allocators, errors-as-values, no hidden control flow, safety, Tiger Style (its native context).
- **Args:** `--no-tiger` skips the Tiger Style section (all of them). `--fix` (all of them) applies the mechanical, unambiguous violations; security and correctness findings stay reported.
- **Reach for it when:** reviewing code in that language, or as a pre-PR gate.

### `/ds-notebook-review` — review

Jupyter-notebook review for the defects a line-by-line code review structurally can't see — the ones that live in notebook *state* and *structure*, not in the cell source. Audits four areas: **execution & hidden state** (out-of-order/gapped `execution_count`, stale outputs with `null` count, forward references, reliance on manual run order so *Restart & Run All* would fail), **committed outputs & secrets** (outputs committed at all, repo-bloating blobs, and credentials/tokens/PII leaked into code or output cells), **reproducibility** (no seed on a reported result, mid-notebook `%pip install`, ambient-kernel reliance, hardcoded absolute paths), and **data-science correctness** (train/test leakage from pre-split `fit`, target leakage, pandas chained assignment, ambiguous `df` re-binding, merges that silently dup/drop rows). The boundary that keeps it in its lane: cell-code idioms/typing/generic security → `/ds-python-review`; ranked perf → `/ds-perf-plan`; pipeline/ETL correctness → `/ds-data-review --pipelines`. It states those delegations rather than duplicating them.

- **Args:** scope (notebook files, directories, globs); defaults to the `.ipynb` files changed on the current branch. `--fix` applies only mechanical, unambiguous fixes — strip committed outputs, reset `execution_count`, drop a confirmed-dead scratch cell — by editing the notebook JSON directly; anything that changes execution semantics (setting a seed, reordering, rewriting leakage-prone preprocessing) stays report-only.
- **Output:** prioritized findings anchored to `<file> cell <N>` — critical (committed secret, leakage that invalidates results, irrecoverable loss) → major (hidden-state break that defeats *Restart & Run All*, stale committed outputs, missing seed) → minor (hygiene/structure). Each names the exact condition that triggers it, then the fix.
- **Reach for it when:** reviewing a `.ipynb` change, or before sharing/committing a notebook. The reproducibility bar is the *Restart & Run All* test. Pair it with `/ds-python-review` on the cell code for the language-level pass.

---

## Plans

A *plan* is not a findings list. Where the reviews above report independent defects you fix in any order, a plan produces **graded, costed moves** — each tagged by the architectural cost it incurs (L1/L2/L3) and ranked so the cheap, high-impact wins come first. The output is an actionable, trade-off-aware plan; it still changes nothing.

### `/ds-perf-plan` — plan

Language-agnostic performance pass governed by one question — **where is this doing more work than it needs to, and what would each speedup cost?** A *plan*, not a verdict: every candidate move is tagged by the architectural cost of applying it (**L1** free win, **L2** localized restructuring, **L3** architectural/boundary-breaking), and ranked by impact ÷ cost so free wins float up. The spine is the anti-hallucination guardrail: **no finding without a cost model** (Big-O, alloc/IO/query counts, or a measured profile), each labeled `measured` / `reasoned` / `speculative`. Distinct from the language reviews' idiom-level `### Performance` checklist, from `/ds-code-quality-review` (which disclaims micro-opts), and from `/ds-ui-quality-review` (frontend rendering).

- **Args:** scope (files, directories, globs); defaults to code changed on the current branch. `--max-level=<1|2|3>` (or freeform "free wins only") clamps higher tiers; `--no-tiger` skips the Tiger Style section.
- **Output:** ranked moves grouped by level, each anchored to `file:line` with its cost model, level tag (and the architecture/clarity cost for L2/L3), evidence label, and the `/ds-verify-this` claim that would prove the win. Changes nothing.
- **Reach for it when:** a path is hot or a change is perf-sensitive. Pairs with `/ds-verify-this` to prove the speedup with a same-machine baseline/treatment.

### `/ds-architecture-plan` — plan

Assess an **existing** codebase's architecture and produce a sequenced refactoring plan, governed by one question — **is the architecture itself sound, and if not, what's the highest-leverage way to fix it, in what order?** Operates at the **module / dependency / boundary** altitude: god packages, import cycles, dependency-direction violations, logic in the wrong layer, shotgun-surgery coupling, duplicated subsystems. Distinct from `/ds-code-quality-review` (file/function altitude, *within* the architecture) and `/ds-zoom-out` (maps, renders no judgment). The spine against cargo-culting: **no recommendation without a concrete symptom in this codebase** — a cycle path, files that co-change, logic at `file:line` — generic "adopt hexagonal/DDD" with no local evidence is banned.

- **Args:** scope (directories, packages, the repo); defaults to the whole project. `--max-level=<1|2|3>` clamps (`--max-level=1` = safe, in-place wins only); `--no-tiger` skips the Tiger Style section.
- **Output:** a 3–5 line assessment, then ordered steps ranked by leverage — each with its level tag (L1 in-place / L2 restructure-within-style / L3 architecture-style change), the symptom it fixes, why-now/what-it-unblocks, blast radius & risk, and whether characterization tests are needed at the seam first. Changes nothing.
- **Reach for it when:** onboarding a codebase inherited in a bad state, or before a structural refactor. Map first with `/ds-zoom-out`; turn the roadmap into tasks with `/ds-roadmap`. (To design a *new* architecture, use `/ds-blueprint`.)

---

## Debugging & verification

### `/ds-debug` — action

Find the root cause of a failure with the scientific method, then prove the fix. Reproduce-first, one hypothesis at a time, evidence over intuition — disciplined against the usual AI failure modes (thrashing, changing five things at once, silencing the symptom). Lightweight and stateless.

- **Args:** the failure to chase — a failing test, error, stack trace, or wrong-behavior description. Refuses a vague "it's broken".
- **Output:** root cause (`file:line` + why), the minimal fix, before/after evidence, and the `/ds-verify-this` claim to lock it in.
- **Reach for it when:** something fails and you don't yet know *why*. Pairs with `/ds-verify-this` — `/ds-debug` finds and fixes, `/ds-verify-this` proves it held.

### `/ds-verify-this` — action

Prove or disprove a **falsifiable** claim with fresh local evidence — not a recap. Restates the claim with a metric and threshold, captures a baseline (old state) and treatment (changed state) with the same command/env, compares raw artifacts, and returns exactly one verdict: `VERIFIED`, `NOT VERIFIED`, or `INCONCLUSIVE`. **No CI required** — all surfaces are local (tests, repro scripts, `tmux`/PTY transcripts, local HTTP, screenshots, profiles).

- **Args:** the claim to verify. Refuses vague claims ("the code is cleaner") — give it something measurable.
- **Reach for it when:** "did this actually fix it?", a bugfix needs a before/after repro, or a perf/memory/UI claim needs measurement.

---

## Understanding

### `/ds-zoom-out` — action

Step up one layer of abstraction and map how an area fits the bigger picture: its responsibility, neighbouring modules, callers, and boundaries. No line-by-line read, no code dumps.

- **Reach for it when:** entering unfamiliar code, or before planning a change in an area you don't hold in your head.

---

## Context & continuity

### `/ds-handoff` — action

Compact the current conversation into a handoff document so a fresh agent can continue without re-reading the transcript.

- **Args:** optional — treated as what the next session should focus on.
- **Output:** writes `handoff.md` to a fresh `mktemp -d` and returns the path. Records goal, done, remaining, key decisions, open questions; references existing artifacts by path rather than duplicating them.
- **Reach for it when:** the context window is filling, you're switching machines/sessions, or pausing mid-task.

### `/ds-tldt` — action

Extractive summarization — selects verbatim sentences, no paraphrasing or generation (so no hallucination). Uses the `tldt` CLI when present, else selects manually. Flags prompt-injection patterns in the source.

- **Args:** `/ds-tldt` (last large block of text), `/ds-tldt <file>`, or `/ds-tldt <url>`.
- **Reach for it when:** compressing a long doc/page before adding it to context.

---

## Response compression

### `/ds-caveman-lite-mode` · `/ds-caveman-ultra-mode` — mode

Compress the agent's prose to save tokens. **Lite** drops articles/filler/hedging (~25–35% reduction, full explanatory value). **Ultra** restructures into fragments and notation (~75–85% reduction). Both keep code, commit messages, PR bodies, numbers, and security/irreversible-action warnings written in full. Deactivate with "stop caveman" or "normal mode".

- **Reach for it when:** long iterative sessions where prose is overhead. Use Ultra only when you'll ask for elaboration if you need it.

---

## Meta

### `/ds-write-a-command` — action

Author a new devskills command in the repo's conventions, written to `commands/` (install.sh copies it to Claude Code, OpenCode, and Codex), registered in the README table and `docs/commands.md`. Knows both archetypes (action vs. mode) and enforces the one-job-per-command rule.

- **Reach for it when:** a workflow you repeat by hand should become a command.
