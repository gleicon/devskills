# Project Memory Workflow (`.project/`)

A minimal, file-backed workflow for keeping project memory across sessions — plain markdown, no hidden state, no background agents, no question-driven hand-holding.

The guiding rule: **these skills are scribes, not pilots.** They read the repo and the conversation and persist structure. They never choose your architecture, never impose a methodology, never interrogate you. You drive; they take notes.

For the standalone skills these compose with, see [skills.md](skills.md). For worked use cases, see [recipes.md](recipes.md#working-with-project-memory).

---

## Four files, one writer each

```
.project/
├── config.md      # yours: the modes to apply at session start (devskills config)
├── state.md       # what is in flight, what is settled, what will bite you
├── map.md         # what the repo is and where things live (/ds-project-map)
└── roadmap.md     # ordered task checklist (/ds-roadmap)
```

Every file has exactly one writer, and that is the design rather than a convention. **`config.md` is written by the `devskills config` command — no skill can edit it**, which is what lets it hold standing instructions the assistant can't quietly drop. Put `ds-step-mode` in there and every session starts gated, whether or not you remember to ask.

Commit the directory as shared project memory, or add it to `.gitignore` for a local scratch space. Nothing here depends on git.

Carrying a `.project/` from before the rebuild — `PROJECT.md`, `DECISIONS.md`, `PLAN.md` — see [migration.md](migration.md).

### `state.md` is the one a session actually loads

```
# now
migrating the auth middleware off the legacy session store

# next
port the refresh-token path, then delete sessionAdapter

# settled
- redis is the only session backend; do not reintroduce the in-process cache
- tokens are opaque, never JWTs

# hazards
- internal/authz/: the deny-by-default branch is load-bearing, do not "simplify"
- repo-wide: never log a raw refresh token
```

Four sections, one line per entry, and the format is the whole discipline. `# now` and `# next` are overwritten each checkpoint; `# settled` and `# hazards` are appended. A `# settled` line has to *forbid* something — if it doesn't stop a future session from redoing a call, it isn't state and it doesn't go in. A `# hazards` line names a scope and what breaks.

That bound is why there is no housekeeping skill here. A file of one-line entries doesn't grow into something that needs compacting, and it doesn't drift into prose that needs reconciling.

### Work products live outside `.project/`

`/ds-spec` writes `SPEC.md`, `/ds-grill-me --record` writes `GRILL.md`, `/ds-explore` writes `EXPLORE.md` — all in your working directory, none of them read at session start. They're things to read, hand to another skill, and retire. `.project/` holds only what a fresh session needs to work.

---

## The three skills

### `/ds-project-map` → `map.md`

Reads the code and writes the repo's structural description: overview, stack, a map of the top-level directories. It regenerates the whole file every run — nothing to preserve, nothing to ask about — which is safe precisely because nothing goes in it that couldn't be re-derived from the source. A constraint or a gotcha belongs in `state.md` instead. Re-run when the repo's shape drifts.

### `/ds-project-checkpoint` → `state.md`

Run it at any point and before `/clear`. It overwrites `# now` and `# next`, appends any new `# settled` and `# hazards` lines, and reports exactly what it added or removed. It writes one file and no others.

Reversing a call deletes its line and appends the replacement — there are no supersession markers, because an entry short enough to delete needs no pointer to its successor.

### `/ds-project-resume [--no-modes]` → reads it back

Run at session start. It applies the modes in `config.md` (read-and-adopt; `--no-modes` skips them but still names them), then reports the modes, `# now`, and `# next`.

`# settled` and `# hazards` load into context and are never spoken. They exist so the constraint is honored six turns later, silently — not recited back at the person who wrote them. Resume writes nothing.

---

## A session, end to end

```
# first time on a repo
devskills config                 # optional: modes to auto-apply → .project/config.md
/ds-project-map                  # map.md: what + where

# starting a piece of work
/ds-spec                         # optional: WHAT → SPEC.md
/ds-explore                      # optional: lay out approaches → EXPLORE.md (--web to research)
/ds-grill-me --record            # optional: decide the gray areas → GRILL.md
/ds-roadmap                      # ordered tasks → .project/roadmap.md

   ...you write code, driving the design...

/ds-deslop                       # quality gates (standalone skills)
/ds-code-quality-review
/ds-verify-this <claim>

/ds-project-checkpoint           # persist state, then /clear or stop
# next session:
/ds-project-resume               # pick up exactly where you left off

# after the release ships:
/ds-retro --record               # decided vs shipped → RETRO.md rules for the next cycle
```

Every step is engineer-driven and self-contained. The only persistent artifacts are four small files — readable, diffable, and yours to edit by hand at any time.

---

## How it relates to the standalone skills

- `/ds-spec` and `/ds-roadmap` are a pair: spec defines the WHAT as a work product, roadmap turns it into `.project/roadmap.md`.
- `/ds-grill-me --record` writes `GRILL.md`. Checkpoint reads the *session*, not the file, and distils whatever now constrains future work into one-line `# settled` entries — so the artifact keeps the full reasoning and `state.md` keeps only what forbids something.
- `/ds-retro` closes the loop after a release: it compares what `SPEC.md`/`GRILL.md` decided against what shipped, and with `--record` distills the lessons into `RETRO.md` — which `/ds-spec` and `/ds-grill-me` read back as priors at their next invocation.
- `/ds-handoff` stays separate and ephemeral: a temp dir, tool-agnostic, for handing work to a person.
- `/ds-step-mode .project/roadmap.md` drives the roadmap one user-gated step at a time — the execution complement to these note-taking skills.

Nothing here is required to use the standalone skills — `.project/` is opt-in. Create it with `/ds-project-map` (or just `mkdir .project`) and the workflow switches on.
