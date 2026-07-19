# Recipes & Workflows

Worked examples of the devskills skills doing real work — and, more usefully, working *together*. These are opinionated suggestions, not the only way. For the dry reference (args, flags, behavior) see [skills.md](skills.md).

Everything here relies only on the skills, `git`, and `gh` — no external orchestration.

---

## Modes stack — run several at once

A **mode** (`/ds-tiger-style-mode`, `/ds-ui-mode`, `/ds-data-mode`, `/ds-test-mode`, `/ds-tdd-mode`, `/ds-git-mode`, `/ds-step-mode`) doesn't do a job and return — it changes *how* the agent works for the rest of the session. Modes **compose**: turn on as many as fit the work. Building a tested UI to a strict bar, committed cleanly as you go, is four at once —

```
/ds-tiger-style-mode             # safety + explicitness bar
/ds-ui-mode                      # component/state discipline + design craft
/ds-test-mode                    # keep the core honestly tested as you build
/ds-git-mode                     # commit each working unit, terse human messages, no history rewrite
   ...build it; all four stay active until the session ends...
```

To drop one mid-session, say so ("stop UI mode"). Everything else here — `/ds-spec`, `/ds-bug-review`, `/ds-verify-this`, … — is an **action**: it runs once and returns a result. The recipes below stitch the two together.

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
/ds-code-quality-review --full
```
**Everything.** `--full` → a project-wide structural audit. Slower and noisier; reach for it when onboarding to a repo, planning a refactor, or doing a periodic health check. Expect it to find cross-file duplication and sprawl that a branch-scoped pass can't see. (`--full` works on every review this way.)

> Rule of thumb: branch-scoped before every PR, area-scoped when something smells, `--full` occasionally.

---

## The draft-PR grill loop

Stress-test the *approach* of a change before you ask a human to review it.

1. **Open a draft PR** from your branch so there's a stable thing to point at:
   ```bash
   gh pr create --draft --fill
   ```
2. **Grill the design**, pointing `/ds-grill-me` at the PR and recording decisions:
   ```
   /ds-grill-me --record  the approach in this PR: <paste PR URL or describe the diff>
   ```
   It interviews you one branch at a time — edge cases, alternatives you skipped, invariants you're assuming — and logs each resolved decision to `GRILL.md`.
3. **Incorporate** the decisions: make the code changes, and let `GRILL.md` become (or seed) the PR description so reviewers see the *why*.
4. **Mark ready:**
   ```bash
   gh pr ready
   ```

Why this order: the cheapest time to discover the design is wrong is *before* a reviewer spends their attention on it. The draft PR gives the conversation an anchor; `GRILL.md` gives the reviewer your reasoning for free.

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
/ds-deslop                  # 1. remove noise first
/ds-code-quality-review     # 2. then judge the structure — including single source of truth
```

`/ds-code-quality-review` catches a failure mode alongside structure: whether the branch introduced *another* version of something that already exists — a second HTTP client, a duplicate constant, a helper that competes with one an earlier agent added. Most common in AI-assisted codebases where multiple agents independently solve the same problem without seeing each other's work. Run it after slop removal so it isn't auditing code that should have been deleted outright.

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

Stitch the quality skills into one gate you run before marking a PR ready:

```
/ds-deslop                  # 1. remove slop introduced on the branch
/ds-code-quality-review     # 2. structure + single source of truth: is the diff making the codebase worse?
/ds-bug-review              # 3. correctness: real bugs, not style
/ds-security-review         # 4. exploitability — if it touches input, auth, secrets, or I/O
/ds-data-review             # 5. data correctness — if it touches schema, queries, transactions, or migrations
/ds-test-quality-review     # 6. is the risky logic actually covered, with good tests?
/ds-perf-plan               # 7. performance: where is it doing more work than needed? (perf-sensitive changes)
/ds-go-review               # 8. language pass (or /ds-ts-review, /ds-rust-review)
/ds-verify-this  <claim>    # 9. prove the headline change actually works
```

Then write the PR description from what you learned and `gh pr ready`. Each step answers a *different* question — noise, structure + single-source-of-truth, correctness, exploitability, data correctness, test coverage, performance, language idioms, behavior — so they don't overlap. Not every PR needs all nine: `/ds-code-quality-review` earns its place on any branch (especially when multiple agents touched it — it catches competing implementations); reach for `/ds-security-review` when it touches untrusted input; `/ds-data-review` when it touches schema/queries/migrations; `/ds-test-quality-review` when the logic is non-trivial; `/ds-perf-plan` when a path is hot. Run the questions that apply. To run the core of this as a guided pipeline with fixes applied between passes, use `/ds-quality-gate` — a tighter, deslop-bookended **six-pass** (this manual list adds `/ds-perf-plan`, the language review, and `/ds-verify-this` on top).

---

## A standalone build loop

This loop covers the full spec-to-ship ground — spec, plan, build, verify, ship — using only standalone skills:

```
/ds-spec                    # 1. WHAT: a SPEC.md with acceptance criteria (optional)
/ds-explore                 # 2. at a fork: lay out approaches → EXPLORE.md (--web to research)
/ds-grill-me --record       # 3. decide the open branches → GRILL.md
/ds-zoom-out                # 4. in unfamiliar code: map the area before changing it
/ds-tiger-style-mode             # 5. engineering bar on (stack /ds-test-mode to keep the core covered, /ds-git-mode to land each working unit as a clean commit)
   ...build it, driving the design yourself...
/ds-deslop                  # 6. clean the generated code
/ds-code-quality-review     # 7. audit structure before review
/ds-verify-this <claim>     # 8. prove the acceptance criteria hold
```

Ship with plain `git` + `gh`. The artifacts that persist your thinking are `SPEC.md` and `GRILL.md` — commit them. This is deliberately lighter than a phase-based engine: fewer moving parts, no background state, faster to start.

To carry plan and state *across* sessions (so `/clear` is always safe), layer the `.project/` memory skills on top — `/ds-project-map`, `/ds-project-checkpoint`, `/ds-project-resume` — and seed the plan with `/ds-roadmap`. The full `.project/` playbook is [below](#working-with-project-memory).

---

## Drive a multi-PR contribution queue (one PR per issue, `/clear`-safe)

When you have a backlog of independent changes to land as *separate* PRs — a queue of issues, or a big refactor split into reviewable chunks — the `.project/` files turn it into a loop that survives `/clear` and hands cleanly between sessions. The plan *is* the protocol.

1. **Map once, plan the queue.** `/ds-project-map` writes `.project/map.md` (the repo facts every issue shares). Put the issue order in `.project/roadmap.md`, and let `state.md` carry the pointer: `# now` is the issue and step you're on, `# next` is the single action that moves it. That split is what makes the session reconstructible after a `/clear` — the queue is durable, the cursor is one line.
2. **One branch + one draft PR per issue.** Branch off fresh `main`, implement surgically from the issue's spec, run the tests, open a **draft** PR. Record the PR URL in `# now` — it identifies the work in flight.
3. **Grill the draft, then mark ready.** Point `/ds-grill-me` at the open PR to stress-test the *approach* one decision at a time; apply what it surfaces, then `gh pr ready`. (This is [the draft-PR grill loop](#the-draft-pr-grill-loop) used as one step of the larger loop.)
4. **Confirm-then-advance at the gates.** Treat *mark-ready* and *merged* as human checkpoints — never assume a merge. Once merged: `git fetch upstream && git checkout main && git merge --ff-only upstream/main`, delete the branch, tick the issue off, and move the Current pointer to the next one.
5. **Checkpoint before `/clear`.** Because `# now` always holds the issue, the step, and the PR, you can `/clear` between issues (or hand to a fresh agent via `/ds-project-resume`) and resume exactly where you left off — no transcript needed.

Why it holds up: each issue is an isolated branch+PR (small, reviewable, revertible), the queue's state lives in a file instead of the conversation, and the ready/merged gates stay human. `.project/` is the shared memory; `git` + `gh` do the rest.

---

## Drive a plan with full control

When you want to *drive* — approve, steer, or replan at every step instead of letting a long autonomous run unfold — turn an existing plan over to `/ds-step-mode`:

```
/ds-roadmap                 # (or any plan: a path, or pasted text)
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
/ds-roadmap            # turn the build order into an ordered task list
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
/ds-roadmap                 # 3. turn the roadmap into ordered tasks
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

Because `/ds-ui-mode` encodes design constraints (type scale, spacing tokens, visual hierarchy) up front, you escape the generic AI look without re-prompting for "polish" each time. Verify what the user actually sees — a screenshot or a keyboard-navigation transcript is the evidence, not a green unit test. Audit finished UI with `/ds-ui-quality-review`.

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

Long tasks hit two walls: the context window fills, and a big pasted input blows it out. Handle both by capturing state and compressing inputs.

- **Continuity** — when you're switching sessions/machines or pausing mid-task, capture state instead of trusting the transcript:
  ```
  /ds-handoff  next: wire the retry logic into the client and add the timeout test
  ```
  It writes a `handoff.md` (goal, done, remaining, decisions, open questions) and returns the path. Start the next session by pointing the agent at that file. Do this *before* the context gets so full the summary degrades. (For an ongoing project, `/ds-project-checkpoint` persists the same state into `.project/` instead.)

- **Big inputs** — before pasting a long doc or page into context, compress it losslessly-ish:
  ```
  /ds-tldt https://example.com/long-rfc
  /ds-tldt ./DESIGN.md
  ```

---

## Context recycling with recall (experimental)

Requires [recall](https://github.com/gleicon/recall). The idea: every session with an AI assistant burns tokens building context from scratch. recall indexes your project locally, accumulates cross-project recipes and insights, and routes questions through a local model before hitting the cloud API — so you reuse what you already paid for.

### First-time setup

```
/ds-recall-setup
```

This indexes the current project, seeds default framework recipes (Go, Next.js, Python, Rust, others), and installs recall's session integration into your AI assistant (via `recall install-skill`). Run once per project, re-run after major structural changes.

### Starting a session with recall context

```
/ds-recall
```

Always run this at the start of a session on a known project. recall maps the project (idempotent) and injects a context-rich brief — prior patterns, matched recipes, relevant cross-project knowledge. The main LLM receives a pre-enriched context and needs fewer tokens to orient itself.

```
/ds-recall query "how do I add middleware to this router?"
```

For a specific question, route it through recall first. If a local model can answer from the indexed recipes, you skip the cloud API entirely. If not, recall returns an enriched brief that reduces the token cost of the cloud call.

```
/ds-recall brain
```

Pull the accumulated cross-project knowledge base — patterns and recipes from everything recall has indexed across your projects. Useful when starting work in a new project that shares patterns with ones you've worked on before.

### Capturing a session's outcome

When you resolve something worth keeping — a bug class, a design decision, a framework-specific pattern — store it before clearing the context:

```
/ds-recall-capture
```

recall extracts signal only: **goal** (one line), **result** (one line), **insight** (one to three lines). No reasoning chain. No failed attempts. The compact recipe is stored in recall's local knowledge base and becomes available in future sessions across all your projects.

Run this **before** `/clear` — the context is gone after.

### A full session with recall

```
# session start — orient recall and inject context
/ds-recall

# ... work, debug, build ...

# before ending
/ds-recall-capture   # store the outcome
/ds-project-checkpoint  # update the .project/ plan (if in use)
/clear
```

### Token budget framing

recall is a **staged composition**: local context → local model → cloud API. Each stage is only reached if the previous one can't answer. A session that would have cost 50k tokens without context can drop significantly when recall's brief eliminates the orientation pass and routes common questions locally.

The savings are real only when recall has indexed relevant prior work. Seed it early, capture consistently.

---

## Working with `.project/` memory

The `.project/` skills keep a durable description, plan, session state, and preferences in plain markdown, so any session is safe to `/clear` or end and a fresh agent can pick up exactly where you left off. Everything above works *without* `.project/`; this section is the persistence layer you add for work that spans sittings. For the file layout and each skill's behavior, see [project-workflow.md](project-workflow.md).

> **There is no execute skill — and that's deliberate.** In the sequences below, a `you →` line is *you typing a normal instruction to the agent*. Implementing is its default behavior; you don't invoke a skill to make it write code. The skills bookend the work — decide, structure, check, persist — and the building in the middle is plain conversation.
>
> ```
> /ds-project-resume                # a skill: orient from .project/
> you → "implement task 2: ..."     # you: plain prose, no skill needed
> ```

### Project from scratch

No code yet, so map comes last — there's nothing to map until something exists.

```
you → "I want a CLI that watches a dir and uploads new files to S3"
/ds-spec                       # WHAT + acceptance criteria → SPEC.md
/ds-explore --web              # research stack/approach options → EXPLORE.md
/ds-grill-me --record          # decide the open branches → GRILL.md
/ds-roadmap               # turn the decisions into an ordered roadmap → .project/roadmap.md
/ds-tiger-style-mode                # engineering bar on for the session
you → "implement task 1: project scaffold + the dir-watch loop"
/ds-project-map                # now there's code — capture map.md (description + repo map)
/ds-deslop                     # clean the generated code
/ds-verify-this "watcher emits an event within 1s of a new file appearing"
/ds-project-checkpoint         # persist state, then /clear is safe
```

Next session: `/ds-project-resume` and keep going.

### Adopting it in a live project

The code already exists, so **map first** — establish ground truth before planning.

```
/ds-project-map                # scan the existing repo → .project/map.md
/ds-zoom-out                   # map the area you're about to touch (responsibility, callers, boundaries)
/ds-roadmap               # seed the roadmap from your current goals / backlog
you → "implement the first task"
...
/ds-project-checkpoint
```

If you inherit a long design doc, run `/ds-tldt ./DESIGN.md` first to compress it before it goes into context.

If the project already has a `.project/` from before the rebuild — `PROJECT.md`, `DECISIONS.md`, `PLAN.md` — port it with [migration.md](migration.md) instead of mapping over it.

### Periodic quality pass

No feature — just paying down entropy. The trick is turning findings *into tasks* instead of fixing ad hoc.

```
/ds-deslop                                 # strip slop from recent work
/ds-code-quality-review --full             # project-wide structural audit (or a path to scope it)
/ds-bug-review --full                      # correctness pass: real bugs (logic, null, error paths, races, leaks)
/ds-test-quality-review --full             # is the critical code actually tested — and are those tests any good?
/ds-doc-quality-review --full              # docs entropy too: drift vs. code, dead links, bloat
you → paste the findings into:
/ds-roadmap                           # findings become ordered tasks in roadmap.md
/ds-go-review        (or /ds-ts-review, /ds-rust-review)   # language idioms + security
you → "fix roadmap tasks 1–3"
/ds-verify-this "the auth refactor preserves the existing token behavior"
/ds-project-checkpoint
```

Run it on a cadence (end of a sprint, before a release). Branch-scope the reviews weekly; `--full` occasionally.

### Implementing a new feature

```
/ds-project-resume             # orient: where we are, what's next
/ds-spec                       # if the feature is non-trivial → SPEC.md  (optional)
/ds-explore                    # at a design fork: lay out approaches (add --web to research)
/ds-grill-me --record          # decide → GRILL.md
/ds-roadmap               # add the feature's tasks to the roadmap
/ds-zoom-out                   # if it touches unfamiliar code
/ds-tiger-style-mode                # engineering bar on
/ds-test-mode                       # + keep the core tested as you build (stacks with /ds-tiger-style-mode)
you → "implement task 4: the retry policy with capped backoff"
/ds-deslop
/ds-code-quality-review        # branch-scoped, before review
/ds-bug-review                 # correctness pass on the branch — real bugs, not style
/ds-security-review            # if it touches input, auth, secrets, or external I/O
/ds-test-quality-review        # did the risky logic get good tests, or just happy-path ones?
/ds-doc-quality-review         # if the feature touched README/docs — did they keep up?
/ds-verify-this "requests retry 3× with backoff, then surface the error"
/ds-project-checkpoint
```

### Making a small change

Not everything earns the full ceremony. Match the weight to the work.

```
you → "change the default timeout to 30s and update the test that asserts it"
/ds-deslop                     # optional, if the diff has any slop
/ds-verify-this "client times out at 30s, not 10s"   # if behavior changed
git commit
```

Skip `/ds-roadmap`/`/ds-project-checkpoint` for a one-liner you commit immediately — there's no state worth persisting. Reach for the workflow when work spans more than one sitting.

### Fixing a bug

The find-then-prove loop shown above (`/ds-debug` → `/ds-verify-this`), with a checkpoint if it was more than trivial:

```
/ds-debug "mytool parse empty.json panics"   # reproduce → root cause → minimal fix
/ds-deslop                     # clean the fix if it sprawled
/ds-verify-this "mytool parse empty.json exits 0 with an error message, no panic"
/ds-project-checkpoint         # if it was more than a trivial fix
```

### Shoring up a risky area

Inherited or critical code you don't trust — get it covered *before* you change it, so a later refactor has a net under it.

```
/ds-zoom-out                   # understand the area: responsibility, callers, boundaries
/ds-test-quality-review src/billing/   # where's the critical logic untested, and which tests lie?
/ds-bug-review src/billing/    # while you're in here: any latent defects in that code?
you → paste the gaps into:
/ds-roadmap               # the coverage gaps and bugs become ordered tasks
/ds-test-mode                       # pragmatic testing mode on — cover by risk, not for a number
you → "cover the proration edge cases the review flagged"
/ds-verify-this "a mid-cycle plan change prorates to the day, not the month"
/ds-project-checkpoint
```

Tests first, *then* the change. `/ds-test-quality-review` finds what's unprotected; `/ds-test-mode` keeps you honest writing it; `/ds-bug-review` catches what was already broken.

### Day-to-day: branch → draft PR → ship

The everyday loop, end to end.

```
git checkout -b feat/upload-retries
/ds-project-resume                         # orient
/ds-roadmap                           # if this branch needs its own task list
you → "implement tasks 1 and 2"
/ds-deslop                                 # clean before anyone looks
/ds-code-quality-review                    # structural pass on the branch
git commit && gh pr create --draft --fill
/ds-grill-me --record                      # stress-test the approach against the draft PR
you → "incorporate the decisions we just made"
/ds-verify-this "uploads survive a transient 503 and succeed on retry"
/ds-project-checkpoint                     # persist state
gh pr ready
```

The draft-PR → `/ds-grill-me` → ready loop is [documented in detail above](#the-draft-pr-grill-loop); the `.project/` skills just add persistent state around it.

### A big change (architecture / large refactor)

Big changes span sessions, so lean hard on checkpoint/resume and incremental phases. Understand and decide *before* touching anything.

```
/ds-zoom-out                   # map the current architecture broadly
/ds-explore --web              # research target patterns / approaches → EXPLORE.md
/ds-grill-me --record          # decide; the hard-to-reverse choices land in GRILL.md
/ds-roadmap               # break it into ordered, individually-shippable phases
/ds-tiger-style-mode
you → "implement phase 1: introduce the new interface behind the old one"
/ds-code-quality-review        # audit each phase
/ds-verify-this "behavior is unchanged after phase 1"   # the refactor invariant
/ds-project-checkpoint         # persist before you stop — this will span sessions
# ...next session...
/ds-project-resume             # picks up # now and # next
/ds-project-map                # regenerate map.md once the shape has changed
/ds-doc-quality-review         # the shape changed — hunt docs the refactor silently rotted (renames, moved files, dead links)
```

Checkpoint between *every* phase so you can `/clear` and resume with a clean context window — that's the whole point of the state files for work this size.

### Resuming after time away

```
/ds-project-resume             # reads state.md: # now and # next
/ds-project-map                # re-run if the code drifted while you were away
/ds-zoom-out                   # re-familiarize with the area you'll touch
```

### Handing the project to someone else

```
/ds-handoff                    # a rich, tool-agnostic handoff in a temp dir
```

That one is for a *person*: context, what was tried, the gotchas. A fresh agent needs none of it — it starts with `/ds-project-resume` and gets `# now`, `# next`, and silently, everything `# settled` and `# hazards` record.

### Keeping `.project/` clean

There is no housekeeping skill, because there is nothing to keep house over. `.project/` holds four files, each with one writer, and each bounded by its own shape:

- **`state.md`** — one line per entry. `# now` and `# next` are overwritten every checkpoint; `# settled` and `# hazards` only accept a line that forbids something or names what breaks. It doesn't grow into prose, so it never needs compacting.
- **`map.md`** — regenerated wholesale by `/ds-project-map`, and holds nothing that isn't re-derivable from the source. Re-run it when the repo's shape drifts; there's no incremental state to reconcile.
- **`roadmap.md`** — `/ds-roadmap` appends and `[x]`s tasks. Prune shipped ones by hand when the list stops showing you what's left; git holds the history.
- **`config.md`** — yours, small, stable. Change it when your preferred modes change.

The scratch files aren't in there at all. `SPEC.md`, `GRILL.md`, and `EXPLORE.md` are work products in your working directory: read them, feed them into the next skill, delete them when they've served their purpose. Nothing reads them at session start, so a stale one can't mislead a fresh session.

The one thing worth doing by hand: if a `# settled` line stops being true, delete it. A reversed call that stays on the list governs work it shouldn't.

**Git hygiene.** Commit `.project/` as shared project memory, or git-ignore the whole directory for a purely local workflow — nothing here depends on git either way.

---

## Which skill, when

Indexed by *what you want to do*, not by kind — for the suffix taxonomy (`-mode` / `-review` / `-plan`), see [skills.md](skills.md#kinds-of-skill).

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
| Scan dependencies for known CVEs | `/ds-osv` |
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
| Land a backlog as separate PRs, `/clear`-safe | `/ds-roadmap` + `/ds-project-resume` |
| Compress a long source doc | `/ds-tldt` |
| Run the full pre-PR review pipeline with fixes between passes | `/ds-quality-gate` |
| Inject cross-project context before asking | `/ds-recall` |
| Store this session's outcome for future sessions | `/ds-recall-capture` |
| Set up recall and install its session integration | `/ds-recall-setup` |
