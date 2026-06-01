# Recipes & Workflows

Worked examples of the devskills commands doing real work — and, more usefully, working *together*. These are opinionated suggestions, not the only way. For the dry reference (args, flags, behavior) see [commands.md](commands.md).

Everything here is **GSD-free**: it relies only on the commands, `git`, and `gh`. If you use GSD, these still apply — they're just the smaller, faster moves you reach for between (or instead of) the heavy phases.

---

## Modes stack — run several at once

A **mode** (`/ds-tiger-style-mode`, `/ds-ui-mode`, `/ds-data-mode`, `/ds-test-mode`, `/ds-git-mode`, `/ds-step-mode`, `/ds-caveman-lite-mode` / `/ds-caveman-ultra-mode`) doesn't do a job and return — it changes *how* the agent works for the rest of the session. Modes **compose**: turn on as many as fit the work. Building a tested UI to a strict bar, committed cleanly as you go, is four at once —

```
/ds-tiger-style-mode             # safety + explicitness bar
/ds-ui-mode                      # component/state discipline + design craft
/ds-test-mode                    # keep the core honestly tested as you build
/ds-git-mode                     # commit each working unit, terse human messages, no history rewrite
   ...build it; all four stay active until the session ends...
```

To drop one mid-session, say so ("stop UI mode"); `/ds-caveman-lite-mode` and `/ds-caveman-ultra-mode` are the ones with explicit off-switches ("normal mode"). The caveman modes deliberately leave commit messages "written normally" — `/ds-git-mode` is what governs them, so the two compose cleanly. Everything else here — `/ds-spec`, `/ds-bug-review`, `/ds-verify-this`, … — is an **action**: it runs once and returns a result. The recipes below stitch the two together.

---

## Three ways to scope `/ds-code-quality-review`

`/ds-code-quality-review` treats its argument as the review scope. Pick the scope to match the question you're asking.

```
/ds-code-quality-review
```
**Recent changes.** No argument → the changed files on the current branch. This is the default pre-merge pass: "is the work I just did making the codebase worse?"

```
/ds-code-quality-review src/auth/ scripts/lib/
```
**A specific area.** One or more paths → audit just those. Use when you suspect a module is decaying, or before you start a change in code you don't trust.

```
/ds-code-quality-review the whole codebase, not just recent changes
```
**Everything.** Freeform "full codebase" scope → a project-wide structural audit. Slower and noisier; reach for it when onboarding to a repo, planning a refactor, or doing a periodic health check. Expect it to find cross-file duplication and sprawl that a branch-scoped pass can't see.

> Rule of thumb: branch-scoped before every PR, area-scoped when something smells, full-codebase occasionally.

---

## The draft-PR grill loop

Stress-test the *approach* of a change before you ask a human to review it. Works whether or not you use GSD.

1. **Open a draft PR** from your branch so there's a stable thing to point at:
   ```bash
   gh pr create --draft --fill
   ```
2. **Grill the design**, pointing `/ds-grill-me` at the PR and recording decisions:
   ```
   /ds-grill-me --record  the approach in this PR: <paste PR URL or describe the diff>
   ```
   It interviews you one branch at a time — edge cases, alternatives you skipped, invariants you're assuming — and logs each resolved decision to `DECISIONS.md`.
3. **Incorporate** the decisions: make the code changes, and let `DECISIONS.md` become (or seed) the PR description so reviewers see the *why*.
4. **Mark ready:**
   ```bash
   gh pr ready
   ```

Why this order: the cheapest time to discover the design is wrong is *before* a reviewer spends their attention on it. The draft PR gives the conversation an anchor; `DECISIONS.md` gives the reviewer your reasoning for free.

`/ds-grill-me` does far more than PR review — requirements discovery, design stress-testing, refactor planning, domain/terminology sharpening, even non-coding decisions. See [grill-me.md](grill-me.md) for the full menu.

---

## Generate, then clean: `/ds-deslop` before review

Fresh AI-generated code carries slop — narrating comments, defensive overkill in trusted paths, type escape hatches that only dodge the checker, needless nesting. `/ds-deslop` judges against the language's idioms and the surrounding code (so it won't flag idiomatic Go `any`, or look for `try/catch` in a language that has none). Clean it before anyone (human or `/ds-code-quality-review`) looks at it.

```
# after generating a batch of code on a branch
/ds-deslop
# scoped, if you only want to clean part of it
/ds-deslop src/handlers/
```

`/ds-deslop` is **narrow and behavior-preserving** — it tidies style and removes noise. It is *not* a structural audit. The two compose:

```
/ds-deslop                  # remove the noise first
/ds-code-quality-review     # then judge the structure that remains
```

Doing it in that order means the structural review isn't distracted by slop, and its findings are about real design, not formatting.

---

## Find then prove: `/ds-debug` → `/ds-verify-this` for a bugfix

When you don't yet know *why* it fails, start with `/ds-debug` — it reproduces first, narrows to the root cause one hypothesis at a time, applies a minimal fix, then hands the proven fix to `/ds-verify-this`. A passing test isn't proof the user-visible bug is gone; `/ds-verify-this` captures before/after evidence and returns a hard verdict — **no CI needed**.

```
/ds-verify-this  the fix on this branch makes `mytool parse bad.json` exit 0 instead of panicking
```

It will: restate the claim falsifiably, run the repro against the parent commit (baseline) and your branch (treatment) with the same command/env, diff the artifacts, and return `VERIFIED` / `NOT VERIFIED` / `INCONCLUSIVE`. Use it when:

- a bugfix needs a real before/after repro,
- "is it actually faster?" (same-machine baseline vs treatment timings),
- a test is green but you want to confirm the behavior a user sees.

Give it something measurable. It will refuse "the code is cleaner" — that's a `/ds-code-quality-review` question, not a verification one.

---

## Find then prove: `/ds-perf-plan` → `/ds-verify-this` for a speedup

Performance work is the most hallucination-prone — "this loop looks slow, add a cache" with no measurement breaks clean code and often *pessimizes*. The pairing closes that gap: `/ds-perf-plan` finds the candidates and refuses any without a cost model (Big-O, alloc/IO/query counts, or a profile), tagging each `measured` / `reasoned` / `speculative` and by architectural cost (L1/L2/L3). Then you apply the move you choose and `/ds-verify-this` proves the win with a same-machine baseline/treatment.

```
/ds-perf-plan src/index/            # rank costed moves; free wins (L1) first
/ds-perf-plan --max-level=1         # or: only the behavior- and structure-preserving wins
   ...apply the move you picked...
/ds-verify-this  `bench query big.idx` runs ≥2× faster on this branch vs parent
```

`/ds-perf-plan` names the *price* of each speedup, so trading architecture for speed (L3) is a deliberate choice; `/ds-verify-this` makes sure the speedup is real before you keep it. Reach for it when a path is hot or a change is perf-sensitive — not as a blanket pass over code that isn't.

---

## A pre-PR quality gate

Stitch the quality commands into one gate you run before marking a PR ready:

```
/ds-deslop                  # 1. remove slop introduced on the branch
/ds-code-quality-review     # 2. structure: is the diff making the codebase worse?
/ds-bug-review              # 3. correctness: real bugs, not style
/ds-security-review         # 4. exploitability — if it touches input, auth, secrets, or I/O
/ds-data-review             # 5. data correctness — if it touches schema, queries, transactions, or migrations
/ds-test-quality-review     # 6. is the risky logic actually covered, with good tests?
/ds-perf-plan               # 7. performance: where is it doing more work than needed? (perf-sensitive changes)
/ds-go-review               # 8. language pass (or /ds-ts-review, /ds-rust-review)
/ds-verify-this  <claim>    # 9. prove the headline change actually works
```

Then write the PR description from what you learned and `gh pr ready`. Each step answers a *different* question — slop (noise), structure, correctness, exploitability, data correctness, test coverage, performance, language idioms, behavior — so they don't overlap. Not every PR needs all nine: reach for `/ds-security-review` when it touches untrusted input, `/ds-data-review` when it touches schema/queries/migrations, `/ds-test-quality-review` when the logic is non-trivial, `/ds-perf-plan` when a path is hot or the change is perf-sensitive. Run the questions that apply.

---

## A standalone build loop (what GSD does, without GSD)

For small-to-medium work you don't need the full `.planning/` machinery. This loop covers the same ground — spec, plan, build, verify, ship — using only standalone commands:

```
/ds-spec                    # 1. WHAT: a SPEC.md with acceptance criteria (optional)
/ds-explore                 # 2. at a fork: lay out approaches → EXPLORE.md (--web to research)
/ds-grill-me --record       # 3. decide the open branches → DECISIONS.md
/ds-zoom-out                # 4. in unfamiliar code: map the area before changing it
/ds-tiger-style-mode             # 5. engineering bar on (stack /ds-test-mode to keep the core covered, /ds-git-mode to land each working unit as a clean commit)
   ...build it, driving the design yourself...
/ds-deslop                  # 6. clean the generated code
/ds-code-quality-review     # 7. audit structure before review
/ds-verify-this <claim>     # 8. prove the acceptance criteria hold
```

Ship with plain `git` + `gh`. The artifacts that persist your thinking are `SPEC.md` and `DECISIONS.md` — commit them. This is deliberately lighter than a phase-based engine: fewer moving parts, no background state, faster to start.

To carry plan and state *across* sessions (so `/clear` is always safe), layer the `.project/` commands on top — `/ds-project-map`, `/ds-project-plan`, `/ds-project-checkpoint`, `/ds-project-resume`. See [project-workflow.md](project-workflow.md).

---

## Drive a plan with full control

When you want to *drive* — approve, steer, or replan at every step instead of letting a long autonomous run unfold — turn an existing plan over to `/ds-step-mode`:

```
/ds-project-plan                 # (or any plan: a path, or pasted text)
/ds-step-mode current plan       # work it one step at a time, gated
   ...for each step: it proposes → you approve/amend/combine/redirect → it does one step → stops...
/ds-project-checkpoint           # at a milestone, persist state so /clear is safe
```

The discipline that makes this work: it proposes the next step and **waits** (a free veto before anything changes), does the smallest reviewable unit, then hands control back **in prose** — options are suggestions you can accept *and add to*, never a forced single-select picker. Say "bigger steps" / "smaller steps" to tune granularity live. Stacks with `/ds-git-mode` (each approved step ≈ a commit) for a driven, cleanly-committed build.

---

## Greenfield: design the architecture before you build

`/ds-spec` and `/ds-explore` get you to *what* and *which options* — but neither commits to a structure. `/ds-blueprint` is the decisive step that does: it takes the requirements and recommends one architecture — modules, dependency rules, seams, build order — then you build the walking skeleton first.

```
/ds-spec                    # WHAT: requirements + acceptance criteria → SPEC.md
/ds-explore                 # options at the big forks (--web to research references)
/ds-blueprint  SPEC.md      # commit to one architecture: modules, deps, seams, build order
/ds-project-plan            # turn the build order into an ordered task list
   ...build the walking skeleton, then the increments...
```

`/ds-blueprint`'s spine is YAGNI — every layer/boundary/queue must trace to a stated requirement, so you get the simplest structure that meets the actual scale, not a cargo-culted one. It states what it deliberately deferred and what would justify adding it later.

---

## Onboarding a codebase inherited in a bad state

When you adopt a running project whose architecture is already wrong, `/ds-code-quality-review` won't help — it works *within* the architecture. `/ds-architecture-plan` works *on* it: it questions whether the structure itself is sound and lays out a sequenced, risk-tagged refactoring roadmap.

```
/ds-zoom-out                     # 1. map the system first — modules, callers, boundaries
/ds-architecture-plan            # 2. critique + sequenced roadmap (L1/L2/L3 by blast radius)
/ds-architecture-plan --max-level=1   # or: safe, in-place wins only to start
/ds-project-plan                 # 3. turn the roadmap into ordered tasks
   ...build each step; add characterization tests at the seam before risky moves...
/ds-verify-this <claim>          # 4. prove a risky move preserved behavior
```

Every step in the plan is anchored to a concrete symptom in *your* codebase — a cycle path, files that co-change, logic in the wrong layer at `file:line` — never generic "go DDD" advice. Mind the altitude split: `/ds-architecture-plan` owns the architecture itself; `/ds-code-quality-review` owns file/function/abstraction cleanup within it.

---

## Building a UI feature

`/ds-ui-mode` is a mode — turn it on and it stays active, shaping every component you build that session. It slots into the build loop above:

```
/ds-ui-mode                      # UI mode on: engineering + design craft, framework-agnostic
/ds-explore                 # at a layout/interaction fork: lay out options (--web for references)
   ...build it: components, minimal co-located state, explicit loading/error/empty states, a11y...
/ds-deslop                  # strip any generated slop
/ds-verify-this "the form shows an inline error and keeps focus when the email is invalid"
```

Because `/ds-ui-mode` encodes design constraints (type scale, spacing tokens, visual hierarchy) up front, you escape the generic AI look without re-prompting for "polish" each time. Verify what the user actually sees — a screenshot or a keyboard-navigation transcript is the evidence, not a green unit test.

---

## Building a data pipeline

`/ds-data-mode` is the data analogue of `/ds-ui-mode`: turn it on and every transform you build that session is shaped against the naive ETL defaults (read-all → overwrite, assume data arrives once and in order, crash on a bad record, no replay). Stack it with the test mode and verify the property that actually matters — a backfill reprocesses cleanly:

```
/ds-data-mode                    # discipline on: idempotency, late/out-of-order data, schema drift, replay-safety
/ds-test-mode                    # cover the transforms as you build (stacks)
   ...build it: pure transforms, upsert-on-key writes, event-time windowing, boundary assertions...
/ds-verify-this "re-running yesterday's window produces identical row counts and totals — no double-counting"
/ds-data-review --pipelines      # audit the pipeline code + the store it writes to
```

The mode shapes how the pipeline gets *built*; `/ds-data-review --pipelines` audits the built pipeline code (idempotency, replay-safety, late-data handling, schema-drift contracts), and plain `/ds-data-review` audits the *store* it writes to (schema, constraints, transactions, migrations). They're complements, not duplicates — run the mode while building, the review before merging.

---

## Surviving long sessions

Two failure modes on long tasks: the context window fills, and prose burns tokens.

- **Continuity** — when you're switching sessions/machines or pausing mid-task, capture state instead of trusting the transcript:
  ```
  /ds-handoff  next: wire the retry logic into the client and add the timeout test
  ```
  It writes a `handoff.md` (goal, done, remaining, decisions, open questions) and returns the path. Start the next session by pointing the agent at that file. Do this *before* the context gets so full the summary degrades.

- **Compression** — for long iterative back-and-forth, drop the prose overhead:
  ```
  /ds-caveman-lite-mode            # ~25–35% less prose, full explanatory value
  /ds-caveman-ultra-mode           # ~75–85% less; fragments + notation, ask for detail when needed
  ```
  Code, commits, PR bodies, and safety warnings stay written in full regardless.

- **Big inputs** — before pasting a long doc or page into context, compress it losslessly-ish:
  ```
  /ds-tldt https://example.com/long-rfc
  /ds-tldt ./DESIGN.md
  ```

---

## Which command, when

| You want to… | Reach for |
|---|---|
| Turn an idea into a verifiable contract | `/ds-spec` |
| Pressure-test a plan or PR approach | `/ds-grill-me` |
| Understand unfamiliar code before changing it | `/ds-zoom-out` |
| Build with real, refactor-proof tests | `/ds-tdd-mode` |
| Keep the core tested as you work (mode) | `/ds-test-mode` |
| Build a data pipeline correctly as you go (mode) | `/ds-data-mode` |
| Commit clean, human-readable history as you build (mode) | `/ds-git-mode` |
| Execute step-by-step, keeping control at every break (mode) | `/ds-step-mode` |
| Remove AI slop from a fresh branch | `/ds-deslop` |
| Bring a codebase's comments to discipline | `/ds-comment-review` |
| Judge structure / find simplifications | `/ds-code-quality-review` |
| Find real bugs (correctness) | `/ds-bug-review` |
| Audit security, language-agnostic | `/ds-security-review` |
| Check the data is correct, consistent, and well-modeled | `/ds-data-review` |
| Check whether the right things are tested | `/ds-test-quality-review` |
| Plan a performance optimization (costed) | `/ds-perf-plan` |
| Plan a refactor of an existing architecture | `/ds-architecture-plan` |
| Commit to an architecture for a new system | `/ds-blueprint` |
| Review language idioms + security | `/ds-go-review` · `/ds-ts-review` · `/ds-rust-review` |
| Find why something fails, then fix it | `/ds-debug` |
| Prove a change actually works | `/ds-verify-this` |
| Hold the session to a strict bar | `/ds-tiger-style-mode` |
| Pause / switch sessions cleanly | `/ds-handoff` |
| Compress a long source doc | `/ds-tldt` |
| Save tokens on a long session | `/ds-caveman-lite-mode` · `/ds-caveman-ultra-mode` |
| Run the full pre-PR review pipeline with fixes between passes | `/ds-quality-gate-mode` |
| Turn a repeated manual flow into a command | `/ds-write-a-command` |
