# Project Memory Workflow (`.project/`)

A minimal, file-backed workflow for keeping project memory across sessions: a durable description, a plan, and current state in plain markdown — no heavy orchestration, no background agents, no question-driven hand-holding.

The guiding rule: **these skills are scribes, not pilots.** They read the repo and the conversation and persist structure. They never choose your architecture, never impose a methodology, never interrogate you. You drive; they take notes.

For the standalone skills these compose with, see [skills.md](skills.md). For worked use cases (new project, bug fix, big refactor, day-to-day PR flow, keeping `.project/` clean), see [recipes.md](recipes.md#working-with-project-memory).

---

## The state lives in `.project/`

Plain markdown. No hidden state, no checksums. Commit it as shared project memory, **or** add `.project/` to `.gitignore` for a local-only scratch space — the workflow doesn't depend on git either way (and you can commit `PROJECT.md`/`PLAN.md` while ignoring the scratch `EXPLORE.md`/`handoff.md` if you prefer).

```
.project/
├── PROJECT.md     # stable: what it is, stack, repo map, hard constraints, ## Landmines
├── PLAN.md        # living: ## Roadmap (tasks + status), ## Now (state, next, open Qs), ## Watch (open flags)
├── DECISIONS.md   # append-only why-log (written by /ds-grill-me --record and checkpoint's sweep)
├── config.md      # optional: project preferences — modes auto-applied on resume (/ds-project-config)
├── handoff.md     # full handoff, only when you ask (/ds-project-checkpoint --handoff)
├── SPEC.md        # optional, only if you use /ds-spec in this workflow
└── archive/       # spent decisions + finished roadmap, moved here by /ds-project-compact (never read by resume)
```

`PLAN.md` is the heart. Its `## Now` block always holds *where we are* and *the next step*, which is what makes ending a session — or running `/clear` — safe at any time: a fresh session reads `PLAN.md` and continues.

---

## The skills

Six keep the `.project/` memory — `/ds-project-map`, `/ds-project-config`, `/ds-project-checkpoint`, `/ds-project-resume`, `/ds-project-compact`, `/ds-project-verify`; `/ds-roadmap` seeds the plan and works with or without `.project/`.

### `/ds-project-map` → `PROJECT.md`

Reads the actual code and writes (or refreshes) the stable description: overview, stack, a repo map, and any hard constraints. Facts only — it describes what exists. Run it once at the start; re-run when the shape of the repo drifts.

### `/ds-project-config` → `config.md`

Sets the per-project preferences a session reads at start — today, the **modes** that `/ds-project-resume` applies automatically. It discovers your installed `ds-*-mode` skills so you don't have to remember names, writes a `## Modes` bullet list, and warns on an unknown name. Optional and hand-editable; it only ever touches `config.md`.

### `/ds-roadmap` → `PLAN.md` (`## Roadmap`)

Turns input into an ordered task checklist. The input can be a goal, a `SPEC.md`, or **pasted output from another skill** — drop in `/ds-code-quality-review` findings or a bug list and they become ordered tasks. It sequences and scopes; it does not pick libraries or patterns. Tasks are outcomes (`[ ]` / `[~]` / `[x]`), not implementation instructions. Like `/ds-spec`, it's `.project`-aware: it writes `.project/PLAN.md` here, or `PLAN.md` in the current directory when there's no `.project/`.

### `/ds-project-checkpoint [--handoff]` → routes durable context

Run before `/clear` or at end of session. **Sweeps the conversation** for durable context that only lives in the chat and **routes each piece to its owning file** — resolved decisions append to `DECISIONS.md`, a genuinely new structural fact appends to `PROJECT.md` (additive only, with your approval; broad drift is flagged for `/ds-project-map`), a scope change is recorded as a decision and `SPEC.md` is flagged stale in `## Watch` (never edited), and a code-keyed constraint someone will trip over later becomes a scope-keyed row in `PROJECT.md`'s `## Landmines`. It then ticks roadmap statuses, overwrites `## Now` with State / Next / Open questions, and appends open flags to `## Watch`. It **writes what you said and asks only about what it inferred** — an inferred decision, a supersession call — shown as one batch you strike from, never a prompt per item. If nothing durable turns up, it's a fast no-op — just the `## Now` refresh. `--handoff` additionally writes a richer `.project/handoff.md` (context, what was tried, gotchas) — use it when the next session needs more than the plan, e.g. handing to another person or a long pause.

### `/ds-project-resume [--no-modes]` → reads state, applies modes

Run at session start. First **applies any modes in `config.md`** (read-and-adopt each mode's command file; `--no-modes` skips this but still lists them). Then reads `PLAN.md` (and `PROJECT.md` for the map) and loads `DECISIONS.md` in full **silently** — decisions and `## Landmines` rows are loaded to be honored while you work, not read back at you — before reporting a fixed short list: state, next, open questions, `## Watch`, modes. Nothing outside that list, including anything it happened to find interesting. If `handoff.md` exists it is loaded **only if it is newer than `PLAN.md`** (by file modification time — no git required) — otherwise it's flagged as stale and ignored, so a forgotten handoff never misleads a fresh session. Resume itself doesn't modify `.project/` files; an applied mode then governs the session under its own rules.

### `/ds-project-compact` → archives spent state

Housekeeping for a `.project/` that's grown heavy. Classifies each `DECISIONS.md` entry as active (stays), superseded (archived, its residual truth re-recorded as a fresh decision), or expired (archived), and moves completed `## Roadmap` sections out — all into an append-only `.project/archive/<file>-<date>.md`, so nothing is lost: every pre-compact entry stays findable in live ∪ archive. Each classification and reader-affecting write is approved one at a time. It never touches `## Now` or `## Watch`, and `/ds-project-resume` never reads the archive. Run it when superseded decisions and finished roadmap sections start adding noise to every resume. Note what it is *not*: compact sorts entries by whether they still govern and never opens the code, so a compacted file is tidy, not verified — that's `/ds-project-verify`.

### `/ds-project-verify` → reconciles the files against the code

The only skill here that reads the source. Checkpoint sweeps the *session* and compact sorts by governance; neither notices a claim that was true when written and was quietly falsified by a later change. That drift is silent, so it accumulates fastest in projects being maintained well — every session checkpointed, every write approved. It corrects only what is knowable without you (supersession markers on unmarked reversals, landmine rows whose scope no longer resolves) and **flags** the rest into `## Watch`, because a spec disagreeing with the code can equally mean the code regressed — and rewriting the spec to match would bury the regression it was meant to catch. `PLAN.md` carries a `<!-- checkpoints-since-verify: N -->` counter that checkpoint bumps and verify resets. Resume often surfaces it past ~10 — but that nudge is best-effort and does miss, so treat the counter as something to read yourself rather than a reminder that will arrive.

---

## A session, end to end

```
# first time on a repo
/ds-project-map                  # PROJECT.md: what + where
/ds-project-config               # optional: modes to auto-apply on resume → config.md

# starting a piece of work
/ds-spec                         # optional: WHAT → .project/SPEC.md
/ds-explore                      # optional: lay out approaches → .project/EXPLORE.md (--web to research)
/ds-grill-me --record            # optional: decide gray areas → .project/DECISIONS.md
/ds-roadmap                 # ordered tasks → .project/PLAN.md

   ...you write code, driving the design...

/ds-deslop                       # quality gates (standalone skills)
/ds-code-quality-review
/ds-verify-this <claim>

/ds-project-checkpoint           # persist state, then /clear or stop
# next session:
/ds-project-resume               # pick up exactly where you left off

# every so often — resume may nudge, but don't wait on it
/ds-project-compact              # trim spent decisions + finished roadmap
/ds-project-verify               # reconcile the files against the code
```

Every step is engineer-driven and self-contained. The only persistent artifacts are the handful of files in `.project/` — readable, diffable, and yours to edit by hand at any time.

---

## How it relates to the standalone skills

- `/ds-spec` and `/ds-roadmap` are a pair of `.project`-aware generators (not `.project`-only): spec writes `.project/SPEC.md` (else `SPEC.md`) and defines the WHAT; roadmap writes `.project/PLAN.md` (else `PLAN.md`) and turns that into an ordered roadmap.
- `/ds-grill-me --record` appends to `.project/DECISIONS.md` when `.project/` exists. Grill a design, then plan it.
- `/ds-handoff` stays separate and ephemeral (writes to a temp dir, tool-agnostic). The durable handoff is `/ds-project-checkpoint --handoff`.
- `/ds-step-mode current plan` drives `PLAN.md` one user-gated step at a time — the execution complement to these note-taking skills (they're scribes; it's the pilot you stay in control of). It marks steps done as they complete and offers `/ds-project-checkpoint` at milestones.

Nothing here is required to use the standalone skills — `.project/` is opt-in. Create it with `/ds-project-map` (or just `mkdir .project`) and the workflow switches on.
