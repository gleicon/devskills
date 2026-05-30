# Recipes & Workflows

Worked examples of the devskills commands doing real work — and, more usefully, working *together*. These are opinionated suggestions, not the only way. For the dry reference (args, flags, behavior) see [commands.md](commands.md).

Everything here is **GSD-free**: it relies only on the commands, `git`, and `gh`. If you use GSD, these still apply — they're just the smaller, faster moves you reach for between (or instead of) the heavy phases.

---

## Modes stack — run several at once

A **mode** (`/tiger-style`, `/ui`, `/test`, `/caveman-lite` / `/caveman-ultra`) doesn't do a job and return — it changes *how* the agent works for the rest of the session. Modes **compose**: turn on as many as fit the work. Building a tested UI to a strict bar is three at once —

```
/tiger-style             # safety + explicitness bar
/ui                      # component/state discipline + design craft
/test                    # keep the core honestly tested as you build
   ...build it; all three stay active until the session ends...
```

To drop one mid-session, say so ("stop UI mode"); `/caveman-lite` and `/caveman-ultra` are the ones with explicit off-switches ("normal mode"). Everything else here — `/spec`, `/bug-review`, `/verify-this`, … — is an **action**: it runs once and returns a result. The recipes below stitch the two together.

---

## Three ways to scope `/code-quality-review`

`/code-quality-review` treats its argument as the review scope. Pick the scope to match the question you're asking.

```
/code-quality-review
```
**Recent changes.** No argument → the changed files on the current branch. This is the default pre-merge pass: "is the work I just did making the codebase worse?"

```
/code-quality-review src/auth/ scripts/lib/
```
**A specific area.** One or more paths → audit just those. Use when you suspect a module is decaying, or before you start a change in code you don't trust.

```
/code-quality-review the whole codebase, not just recent changes
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
2. **Grill the design**, pointing `/grill-me` at the PR and recording decisions:
   ```
   /grill-me --record  the approach in this PR: <paste PR URL or describe the diff>
   ```
   It interviews you one branch at a time — edge cases, alternatives you skipped, invariants you're assuming — and logs each resolved decision to `DECISIONS.md`.
3. **Incorporate** the decisions: make the code changes, and let `DECISIONS.md` become (or seed) the PR description so reviewers see the *why*.
4. **Mark ready:**
   ```bash
   gh pr ready
   ```

Why this order: the cheapest time to discover the design is wrong is *before* a reviewer spends their attention on it. The draft PR gives the conversation an anchor; `DECISIONS.md` gives the reviewer your reasoning for free.

`/grill-me` does far more than PR review — requirements discovery, design stress-testing, refactor planning, domain/terminology sharpening, even non-coding decisions. See [grill-me.md](grill-me.md) for the full menu.

---

## Generate, then clean: `/deslop` before review

Fresh AI-generated code carries slop — narrating comments, defensive overkill in trusted paths, type escape hatches that only dodge the checker, needless nesting. `/deslop` judges against the language's idioms and the surrounding code (so it won't flag idiomatic Go `any`, or look for `try/catch` in a language that has none). Clean it before anyone (human or `/code-quality-review`) looks at it.

```
# after generating a batch of code on a branch
/deslop
# scoped, if you only want to clean part of it
/deslop src/handlers/
```

`/deslop` is **narrow and behavior-preserving** — it tidies style and removes noise. It is *not* a structural audit. The two compose:

```
/deslop                  # remove the noise first
/code-quality-review     # then judge the structure that remains
```

Doing it in that order means the structural review isn't distracted by slop, and its findings are about real design, not formatting.

---

## Find then prove: `/debug` → `/verify-this` for a bugfix

When you don't yet know *why* it fails, start with `/debug` — it reproduces first, narrows to the root cause one hypothesis at a time, applies a minimal fix, then hands the proven fix to `/verify-this`. A passing test isn't proof the user-visible bug is gone; `/verify-this` captures before/after evidence and returns a hard verdict — **no CI needed**.

```
/verify-this  the fix on this branch makes `mytool parse bad.json` exit 0 instead of panicking
```

It will: restate the claim falsifiably, run the repro against the parent commit (baseline) and your branch (treatment) with the same command/env, diff the artifacts, and return `VERIFIED` / `NOT VERIFIED` / `INCONCLUSIVE`. Use it when:

- a bugfix needs a real before/after repro,
- "is it actually faster?" (same-machine baseline vs treatment timings),
- a test is green but you want to confirm the behavior a user sees.

Give it something measurable. It will refuse "the code is cleaner" — that's a `/code-quality-review` question, not a verification one.

---

## A pre-PR quality gate

Stitch the quality commands into one gate you run before marking a PR ready:

```
/deslop                  # 1. remove slop introduced on the branch
/code-quality-review     # 2. structure: is the diff making the codebase worse?
/bug-review              # 3. correctness: real bugs, not style
/security-review         # 4. exploitability — if it touches input, auth, secrets, or I/O
/test-quality-review     # 5. is the risky logic actually covered, with good tests?
/go-review               # 6. language pass (or /ts-review, /rust-review)
/verify-this  <claim>    # 7. prove the headline change actually works
```

Then write the PR description from what you learned and `gh pr ready`. Each step answers a *different* question — slop (noise), structure, correctness, exploitability, test coverage, language idioms, behavior — so they don't overlap. Not every PR needs all seven: reach for `/security-review` when it touches untrusted input, `/test-quality-review` when the logic is non-trivial. Run the questions that apply.

---

## A standalone build loop (what GSD does, without GSD)

For small-to-medium work you don't need the full `.planning/` machinery. This loop covers the same ground — spec, plan, build, verify, ship — using only standalone commands:

```
/spec                    # 1. WHAT: a SPEC.md with acceptance criteria (optional)
/explore                 # 2. at a fork: lay out approaches → EXPLORE.md (--web to research)
/grill-me --record       # 3. decide the open branches → DECISIONS.md
/zoom-out                # 4. in unfamiliar code: map the area before changing it
/tiger-style             # 5. engineering bar on (stack /test to keep the core covered as you build)
   ...build it, driving the design yourself...
/deslop                  # 6. clean the generated code
/code-quality-review     # 7. audit structure before review
/verify-this <claim>     # 8. prove the acceptance criteria hold
```

Ship with plain `git` + `gh`. The artifacts that persist your thinking are `SPEC.md` and `DECISIONS.md` — commit them. This is deliberately lighter than a phase-based engine: fewer moving parts, no background state, faster to start.

To carry plan and state *across* sessions (so `/clear` is always safe), layer the `.project/` commands on top — `/project-map`, `/project-plan`, `/project-checkpoint`, `/project-resume`. See [project-workflow.md](project-workflow.md).

---

## Building a UI feature

`/ui` is a mode — turn it on and it stays active, shaping every component you build that session. It slots into the build loop above:

```
/ui                      # UI mode on: engineering + design craft, framework-agnostic
/explore                 # at a layout/interaction fork: lay out options (--web for references)
   ...build it: components, minimal co-located state, explicit loading/error/empty states, a11y...
/deslop                  # strip any generated slop
/verify-this "the form shows an inline error and keeps focus when the email is invalid"
```

Because `/ui` encodes design constraints (type scale, spacing tokens, visual hierarchy) up front, you escape the generic AI look without re-prompting for "polish" each time. Verify what the user actually sees — a screenshot or a keyboard-navigation transcript is the evidence, not a green unit test.

---

## Surviving long sessions

Two failure modes on long tasks: the context window fills, and prose burns tokens.

- **Continuity** — when you're switching sessions/machines or pausing mid-task, capture state instead of trusting the transcript:
  ```
  /handoff  next: wire the retry logic into the client and add the timeout test
  ```
  It writes a `handoff.md` (goal, done, remaining, decisions, open questions) and returns the path. Start the next session by pointing the agent at that file. Do this *before* the context gets so full the summary degrades.

- **Compression** — for long iterative back-and-forth, drop the prose overhead:
  ```
  /caveman-lite            # ~25–35% less prose, full explanatory value
  /caveman-ultra           # ~75–85% less; fragments + notation, ask for detail when needed
  ```
  Code, commits, PR bodies, and safety warnings stay written in full regardless.

- **Big inputs** — before pasting a long doc or page into context, compress it losslessly-ish:
  ```
  /tldt https://example.com/long-rfc
  /tldt ./DESIGN.md
  ```

---

## Which command, when

| You want to… | Reach for |
|---|---|
| Turn an idea into a verifiable contract | `/spec` |
| Pressure-test a plan or PR approach | `/grill-me` |
| Understand unfamiliar code before changing it | `/zoom-out` |
| Build with real, refactor-proof tests | `/tdd` |
| Keep the core tested as you work (mode) | `/test` |
| Remove AI slop from a fresh branch | `/deslop` |
| Judge structure / find simplifications | `/code-quality-review` |
| Find real bugs (correctness) | `/bug-review` |
| Audit security, language-agnostic | `/security-review` |
| Check whether the right things are tested | `/test-quality-review` |
| Review language idioms + security | `/go-review` · `/ts-review` · `/rust-review` |
| Find why something fails, then fix it | `/debug` |
| Prove a change actually works | `/verify-this` |
| Hold the session to a strict bar | `/tiger-style` |
| Pause / switch sessions cleanly | `/handoff` |
| Compress a long source doc | `/tldt` |
| Save tokens on a long session | `/caveman-lite` · `/caveman-ultra` |
| Turn a repeated manual flow into a command | `/write-a-command` |
