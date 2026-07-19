# Migrating an old `.project/`

Repos set up before the rebuild carry `PROJECT.md`, `DECISIONS.md`, `PLAN.md`,
sometimes `handoff.md` and an `archive/`. The new `.project/` is four files with
one writer each — `config.md`, `state.md`, `map.md`, `roadmap.md`.

**The old files are claims, not truth.** They were written by the tools this
release replaces, appended to every session for the life of the project, and never
verified against anything. Some entries were wrong when written; more went stale as
the code moved. Migrating means auditing them against the code, not copying them
forward.

That's also why this can't be one sweep. `/ds-project-checkpoint` distils a
*session* — an hour of context with a handful of decisions in it. Pointed at a
project's entire accumulated history it silently under-reads, because a filter
applied while recalling can only reject what it happened to surface. So the old
files get enumerated and checked first, and checkpoint runs against a list.

`DECISIONS.md` is also a different document from `state.md`. It logs what was
decided, including everything now plainly visible in the code. `state.md` holds
only what a future session would otherwise get wrong. Most entries won't survive,
and that's the migration working.

---

1. `devskills install`

2. `devskills config`

3. **Audit the old files.** One pass, one row per entry. Paste this:

   ```
   Read these, top level of .project/ only: PROJECT.md, DECISIONS.md, PLAN.md,
   SPEC.md, handoff.md — whichever exist. Skip .project/archive/ (retired content
   that reads exactly like live content) and config.md (not project state).

   From SPEC.md take constraints, forbidden lists, out-of-scope and NFRs. Skip its
   FR/AC bodies — those describe features, and features live in the code.

   Emit one row for every discrete entry in those files: every bullet, every table
   row, every numbered constraint. Work through one file at a time and finish it
   before starting the next. Do not group, merge, or skip entries that look minor
   or obvious — the row count should match what is in the files. This is the only
   pass that reads them.

   Each row:

     ENTRY  the claim in the source's own words, one line
     KIND   code     — a claim about how this repo is built or must not change
            working  — a standing instruction about how to work with the user
            desc     — a statement of what was built that forbids nothing
     CHECK  KIND=code: read the code and answer CONFIRMED <file:line>, or STALE
            if the code contradicts it or the file or symbol is gone
            KIND=working: carry — the repo cannot confirm these and their
            staleness is the user's call, not the code's
            KIND=desc: skip

   These files were never verified and have been drifting for the life of the
   project. Every entry is a claim to check. STALE is a normal outcome, not a
   failure to explain away.

   Judge truth only. Do not decide what deserves to be in state.md — the next step
   owns that, and answering both questions at once means a true constraint gets
   cut for looking unimportant while a loud one survives.

   Write no files.
   ```

4. **`/ds-project-checkpoint`**, with this alongside it:

   ```
   Sweep only the CONFIRMED and carry rows from that table. Ignore STALE and desc.
   ```

   Its rules now decide which of them earn a line. Expect it to cut heavily — the
   audit answered *is this true*, and checkpoint answers *does a future session
   need it*. Both questions have to be asked, and neither answers the other.

5. **`/ds-roadmap`** — only if `PLAN.md` had unchecked `## Roadmap` items; paste
   them as the argument. Nothing open means no `roadmap.md`; don't create an empty
   one.

6. **`/ds-project-map`** — writes `map.md` from the code. Nothing carries over
   into it; it regenerates wholesale by design.

7. **Delete the leftovers.** `PROJECT.md`, `DECISIONS.md`, `PLAN.md` and
   `handoff.md` are done. `SPEC.md` / `GRILL.md` / `EXPLORE.md` are work products —
   move them to the repo root if you still want them. `.project/` holds only its
   four files.

   `.project/` is usually gitignored, so there is no `git checkout` to fall back
   on. Copy it aside before this step.

8. Start a fresh session and run **`/ds-project-resume`**.
