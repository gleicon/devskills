# Lite Planning Workflow (`.project/`)

A minimal, file-backed alternative to GSD for keeping project memory across sessions. It keeps the part of GSD worth keeping — a durable description, a plan, and current state — and drops the rest (heavy orchestration, background/parallel agents, question-driven hand-holding).

The guiding rule: **these commands are scribes, not pilots.** They read the repo and the conversation and persist structure. They never choose your architecture, never impose a methodology, never interrogate you. You drive; they take notes.

For the GSD-managed workflow instead, see [gsd-workflow.md](gsd-workflow.md). For the standalone commands these compose with, see [commands.md](commands.md). For worked use cases (new project, bug fix, big refactor, day-to-day PR flow, keeping `.project/` clean), see [project-recipes.md](project-recipes.md).

---

## The state lives in `.project/`

Plain markdown. No hidden state, no checksums. Commit it as shared project memory, **or** add `.project/` to `.gitignore` for a local-only scratch space — the workflow doesn't depend on git either way (and you can commit `PROJECT.md`/`PLAN.md` while ignoring the scratch `EXPLORE.md`/`handoff.md` if you prefer).

```
.project/
├── PROJECT.md     # stable: what it is, stack, repo map, hard constraints
├── PLAN.md        # living: ## Roadmap (ordered tasks + status) and ## Now (state, next, open Qs)
├── DECISIONS.md   # append-only why-log (written by /grill-me --record)
├── handoff.md     # full handoff, only when you ask (/project-checkpoint --handoff)
└── SPEC.md        # optional, only if you use /spec in this workflow
```

`PLAN.md` is the heart. Its `## Now` block always holds *where we are* and *the next step*, which is what makes ending a session — or running `/clear` — safe at any time: a fresh session reads `PLAN.md` and continues.

---

## The four commands

### `/project-map` → `PROJECT.md`

Reads the actual code and writes (or refreshes) the stable description: overview, stack, a repo map, and any hard constraints. Facts only — it describes what exists. Run it once at the start; re-run when the shape of the repo drifts.

### `/project-plan` → `PLAN.md` (`## Roadmap`)

Turns input into an ordered task checklist. The input can be a goal, a `SPEC.md`, or **pasted output from another command** — drop in `/code-quality-review` findings or a bug list and they become ordered tasks. It sequences and scopes; it does not pick libraries or patterns. Tasks are outcomes (`[ ]` / `[~]` / `[x]`), not implementation instructions.

### `/project-checkpoint [--handoff]` → `PLAN.md` (`## Now`)

Run before `/clear` or at end of session. Ticks roadmap statuses and overwrites `## Now` with State / Next / Open questions. `--handoff` additionally writes a richer `.project/handoff.md` (context, what was tried, gotchas) — use it when the next session needs more than the plan, e.g. handing to another person or a long pause.

### `/project-resume` → reads state

Run at session start. Reads `PLAN.md` (and `PROJECT.md` for the map), then summarizes where to pick up. If `handoff.md` exists it is loaded **only if it is newer than `PLAN.md`** (by file modification time — no git required) — otherwise it's flagged as stale and ignored, so a forgotten handoff never misleads a fresh session. Read-only.

---

## A session, end to end

```
# first time on a repo
/project-map                  # PROJECT.md: what + where

# starting a piece of work
/spec                         # optional: WHAT → .project/SPEC.md
/explore                      # optional: lay out approaches → .project/EXPLORE.md (--web to research)
/grill-me --record            # optional: decide gray areas → .project/DECISIONS.md
/project-plan                 # ordered tasks → .project/PLAN.md

   ...you write code, driving the design...

/deslop                       # quality gates (standalone commands)
/code-quality-review
/verify-this <claim>

/project-checkpoint           # persist state, then /clear or stop
# next session:
/project-resume               # pick up exactly where you left off
```

Every step is engineer-driven and self-contained. The only persistent artifacts are the four files in `.project/` — readable, diffable, and yours to edit by hand at any time.

---

## How it relates to the standalone commands

- `/spec` writes to `.project/SPEC.md` when `.project/` exists (else its usual location). It defines the WHAT; `/project-plan` turns that into ordered tasks.
- `/grill-me --record` appends to `.project/DECISIONS.md` when `.project/` exists. Grill a design, then plan it.
- `/handoff` stays separate and ephemeral (writes to a temp dir, tool-agnostic). The durable handoff is `/project-checkpoint --handoff`.

Nothing here is required to use the standalone commands — `.project/` is opt-in. Create it with `/project-map` (or just `mkdir .project`) and the workflow switches on.
