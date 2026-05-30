# Project Workflow — Recipes

Use-case walkthroughs for working with devskills — the `.project/` workflow plus the modes, reviews, and checks that bookend it. For the commands and file layout, see [project-workflow.md](project-workflow.md); for the full command reference, see [commands.md](commands.md); for non-`.project/` workflows, see [recipes.md](recipes.md).

> **There is no execute command — and that's deliberate.** In the sequences below, a `you →` line is *you typing a normal instruction to the agent*. Implementing is its default behavior; you don't invoke a command to make it write code. The slash-commands bookend the work — decide, structure, check, persist — and the building in the middle is plain conversation.
>
> ```
> /project-resume                   # a command: orient from .project/
> you → "implement task 2: ..."     # you: plain prose, no command needed
> ```

---

## Project from scratch

No code yet, so map comes last — there's nothing to map until something exists.

```
you → "I want a CLI that watches a dir and uploads new files to S3"
/spec                       # WHAT + acceptance criteria → .project/SPEC.md
/explore --web              # research stack/approach options → .project/EXPLORE.md
/grill-me --record          # decide the open branches → .project/DECISIONS.md
/project-plan               # turn the decisions into an ordered roadmap → .project/PLAN.md
/tiger-style                # engineering bar on for the session
you → "implement task 1: project scaffold + the dir-watch loop"
/project-map                # now there's code — capture PROJECT.md (description + repo map)
/deslop                     # clean the generated code
/verify-this "watcher emits an event within 1s of a new file appearing"
/project-checkpoint         # persist state, then /clear is safe
```

Next session: `/project-resume` and keep going.

## Adopting it in a live project

The code already exists, so **map first** — establish ground truth before planning.

```
/project-map                # scan the existing repo → .project/PROJECT.md
/zoom-out                   # map the area you're about to touch (responsibility, callers, boundaries)
/project-plan               # seed the roadmap from your current goals / backlog
you → "implement the first task"
...
/project-checkpoint
```

If you inherit a long design doc, run `/tldt ./DESIGN.md` first to compress it before it goes into context.

## Periodic quality pass

No feature — just paying down entropy. The trick is turning findings *into tasks* instead of fixing ad hoc.

```
/deslop                                 # strip slop from recent work
/code-quality-review .                  # full-codebase structural audit (or a path to scope it)
/bug-review .                           # correctness pass: real bugs (logic, null, error paths, races, leaks)
/test-quality-review .                  # is the critical code actually tested — and are those tests any good?
/doc-quality-review .                   # docs entropy too: drift vs. code, dead links, bloat (--comments for code comments)
you → paste the findings into:
/project-plan                           # findings become ordered tasks in PLAN.md
/go-review        (or /ts-review, /rust-review)   # language idioms + security
you → "fix roadmap tasks 1–3"
/verify-this "the auth refactor preserves the existing token behavior"
/project-checkpoint
```

Run it on a cadence (end of a sprint, before a release). Branch-scope `/code-quality-review` weekly; full-codebase occasionally.

## Implementing a new feature

```
/project-resume             # orient: where we are, what's next
/spec                       # if the feature is non-trivial → .project/SPEC.md  (optional)
/explore                    # at a design fork: lay out approaches (add --web to research)
/grill-me --record          # decide → DECISIONS.md
/project-plan               # add the feature's tasks to the roadmap
/zoom-out                   # if it touches unfamiliar code
/tiger-style                # engineering bar on
/test                       # + keep the core tested as you build (stacks with /tiger-style)
you → "implement task 4: the retry policy with capped backoff"
/deslop
/code-quality-review        # branch-scoped, before review
/bug-review                 # correctness pass on the branch — real bugs, not style
/security-review            # if it touches input, auth, secrets, or external I/O
/test-quality-review        # did the risky logic get good tests, or just happy-path ones?
/doc-quality-review         # if the feature touched README/docs — did they keep up?
/verify-this "requests retry 3× with backoff, then surface the error"
/project-checkpoint
```

## Making a small change

Not everything earns the full ceremony. Match the weight to the work.

```
you → "change the default timeout to 30s and update the test that asserts it"
/deslop                     # optional, if the diff has any slop
/verify-this "client times out at 30s, not 10s"   # if behavior changed
git commit
```

Skip `/project-plan`/`/project-checkpoint` for a one-liner you commit immediately — there's no state worth persisting. Reach for the workflow when work spans more than one sitting.

## Fixing a bug

Root-cause it, fix it, prove it — `/debug` runs that loop and hands the proven fix to `/verify-this`.

```
/debug "mytool parse empty.json panics"   # reproduce → root cause → minimal fix
/deslop                     # clean the fix if it sprawled
/verify-this "mytool parse empty.json exits 0 with an error message, no panic"  # /debug hands off here to prove it
/project-checkpoint         # if it was more than a trivial fix
```

## Shoring up a risky area

Inherited or critical code you don't trust — get it covered *before* you change it, so a later refactor has a net under it.

```
/zoom-out                   # understand the area: responsibility, callers, boundaries
/test-quality-review src/billing/   # where's the critical logic untested, and which tests lie?
/bug-review src/billing/    # while you're in here: any latent defects in that code?
you → paste the gaps into:
/project-plan               # the coverage gaps and bugs become ordered tasks
/test                       # pragmatic testing mode on — cover by risk, not for a number
you → "cover the proration edge cases the review flagged"
/verify-this "a mid-cycle plan change prorates to the day, not the month"
/project-checkpoint
```

Tests first, *then* the change. `/test-quality-review` finds what's unprotected; `/test` keeps you honest writing it; `/bug-review` catches what was already broken.

## Day-to-day: branch → draft PR → ship

The everyday loop, end to end.

```
git checkout -b feat/upload-retries
/project-resume                         # orient
/project-plan                           # if this branch needs its own task list
you → "implement tasks 1 and 2"
/deslop                                 # clean before anyone looks
/code-quality-review                    # structural pass on the branch
git commit && gh pr create --draft --fill
/grill-me --record                      # stress-test the approach against the draft PR
you → "incorporate the decisions we just made"
/verify-this "uploads survive a transient 503 and succeed on retry"
/project-checkpoint                     # persist state
gh pr ready
```

The draft-PR → `/grill-me` → ready loop is documented in detail in [recipes.md](recipes.md#the-draft-pr-grill-loop); the `/project-*` commands just add persistent state around it.

## A big change (architecture / large refactor)

Big changes span sessions, so lean hard on checkpoint/resume and incremental phases. Understand and decide *before* touching anything.

```
/zoom-out                   # map the current architecture broadly
/explore --web              # research target patterns / approaches → EXPLORE.md
/grill-me --record          # decide; the hard-to-reverse choices go in DECISIONS.md (ADR-worthy)
/project-plan               # break it into ordered, individually-shippable phases
/tiger-style
you → "implement phase 1: introduce the new interface behind the old one"
/code-quality-review        # audit each phase
/verify-this "behavior is unchanged after phase 1"   # the refactor invariant
/project-checkpoint --handoff   # rich handoff before you stop — this will span sessions
# ...next session...
/project-resume             # picks up PLAN.md + the fresh handoff
/project-map                # refresh PROJECT.md once the shape has changed
/doc-quality-review         # the shape changed — hunt docs the refactor silently rotted (renames, moved files, dead links)
```

Checkpoint between *every* phase so you can `/clear` and resume with a clean context window — that's the whole point of the state files for work this size.

## Resuming after time away

```
/project-resume             # reads PLAN.md ## Now; flags handoff.md if it's stale
/project-map                # re-run if the code drifted while you were away (refreshes the map)
/zoom-out                   # re-familiarize with the area you'll touch
```

If `/project-resume` reports a stale `handoff.md`, trust `## Now` over it (and delete the stale file — see below).

## Handing the project to someone else

```
/project-checkpoint --handoff   # writes a rich .project/handoff.md (context, what was tried, gotchas)
```

They (or a fresh agent) start with `/project-resume`, which loads `PROJECT.md` (the map), `PLAN.md` (`## Now` + roadmap), and the fresh `handoff.md`. `DECISIONS.md` answers "why is it like this?" without a meeting.

---

## Keeping `.project/` clean

The files have different lifetimes. Knowing which are durable and which are scratch keeps the directory from rotting.

**`PLAN.md` — living, prune it.**
- `/project-checkpoint` marks tasks `[x]` and overwrites `## Now`; `## Now` never needs manual cleanup.
- The `## Roadmap` accumulates `[x]` tasks. When a feature or milestone ships, **prune the completed tasks** so the roadmap shows what's *left*, not a growing history. If you want a record, move them to a short `## Shipped` list at the bottom — but git and `DECISIONS.md` already hold the history, so deleting is fine.
- Rule of thumb: if you can't see the next three things to do without scrolling, prune.

**`EXPLORE.md` — scratch, disposable.**
- Overwritten on every `/explore`. Once you've decided (`/grill-me` → `DECISIONS.md`), its content is captured where it matters. Delete it or just let the next `/explore` overwrite it. Never cite it as a durable record.

**`handoff.md` — point-in-time, expires.**
- Written only by `/project-checkpoint --handoff`. `/project-resume` ignores it once it's older than `PLAN.md`. After a successful resume, `rm .project/handoff.md` so it doesn't linger and confuse — `## Now` is the source of truth, not a past handoff.

**`PROJECT.md` / `DECISIONS.md` — durable, keep.**
- `PROJECT.md`: refresh with `/project-map` when the repo's shape drifts; otherwise leave it.
- `DECISIONS.md`: append-only "why" log. Don't prune it — a long decision history is a feature, not clutter.

**Git hygiene.** Commit the durable trio (`PROJECT.md`, `PLAN.md`, `DECISIONS.md`) as shared memory; the scratch files don't belong in history. If you commit `.project/`, ignore the scratch:

```gitignore
.project/EXPLORE.md
.project/handoff.md
```

Or git-ignore the whole `.project/` directory for a purely local workflow — nothing here depends on git.
