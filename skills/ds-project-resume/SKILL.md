---
name: ds-project-resume
description: "Restore working context from `.project/PLAN.md` (and a fresh handoff, if any), and apply any configured modes."
disable-model-invocation: true
---

When invoked, read the project's persisted state and report where to pick up — the counterpart to `/ds-project-checkpoint`. Safe to run at the start of any session.

## Arguments

- `--no-modes` — skip applying the modes in `.project/config.md`. Resume still lists what the project configures, so opting out never hides them.

## Process

1. **Apply configured modes** (unless `--no-modes`). If `.project/config.md` exists, read its `## Modes` list and, for each mode, read that mode's installed skill file — `<name>/SKILL.md` under your assistant's skills directory (`~/.claude/skills/`, `$CLAUDE_CONFIG_DIR/skills/`, `~/.codex/skills/`, `~/.config/opencode/skills/`) — and adopt its rules for the rest of the session (read-and-adopt). Every mode's `SKILL.md` ends with its own confirmation line — **those confirmations are the echo**; don't restate the list a second time as prose. Where a mode grants a standing authorization, name it inline on that same line ("git mode — commits finished units without asking, never pushes") instead of in a paragraph of its own; it is a one-time warning, not something to re-read at every resume. If a listed mode can't be found, say so and tell the user to apply it manually — never silently skip. With `--no-modes`, don't apply them but still list what `config.md` configures. If `.project/config.md` is absent, drop a one-line hint that the user can create one with `/ds-project-config` to auto-apply modes.
2. If `.project/PLAN.md` does not exist, say so and suggest `/ds-project-map` then `/ds-roadmap`. Stop.
3. Read `.project/PLAN.md` — `## Now` (state, next, open questions), the top-level `## Watch` section, and the `## Roadmap` status.
4. Read `.project/PROJECT.md` if present, for the repo map and constraints. Its `## Landmines` rows load **silently** — they exist so you honor them when you touch that scope, not so you can read them back to the person who wrote them.
5. Read `.project/DECISIONS.md` in full if present, and say nothing about it. It loads so settled questions stay settled — a decision you never mention still stops you re-proposing what it killed. No count, no recent-few, no summary. Its size is `/ds-project-compact`'s problem, not a reason to read selectively: old decisions bind as hard as new ones.
6. If `.project/handoff.md` exists, check whether it is still current — by **file modification time, not git** (the workflow must work when `.project/` is git-ignored or the repo has no git):
   - If `handoff.md` is newer than `.project/PLAN.md`, load it — it's the freshest context.
   - If it is older than `PLAN.md` (a checkpoint happened after it), treat it as **stale**: mention it exists and its date, but do not rely on it.
   - If the repo uses git, you may *optionally* also flag the handoff as stale when commits have landed since it was written — but never require git; the file-time comparison is the source of truth.
   - If there is no `handoff.md`, say nothing. An absent optional file is not news.
7. **Report this, and nothing else:**
   - **State** — where things stand, from `## Now`.
   - **Next** — the single next action.
   - **Open questions** — only if `## Now` has any.
   - **Watch** — only if `## Watch` has entries.
   - **Modes** — the confirmation lines from step 1.
   - **Nudge** — at most one, and only when a mechanical threshold fires (below).

   Nothing outside that list appears, *including* something you noticed while reading and judged worth mentioning. No decision counts, no landmines, no shipped-feature inventories, no remarks about files that don't exist, no observations about the repo. A slot with nothing in it is dropped, never written as "none".

## Rules

- Resume does not modify `.project/` files — it reads them. Applying a configured mode is part of resume; once applied, that mode governs the session under its own rules.
- Apply modes via read-and-adopt only — never assume a mode is active without reading its `SKILL.md`.
- Trust `## Now` over a stale `handoff.md`.
- **Load to apply, not to announce.** Decisions and landmines enter context so they shape the work. Narrating one unprompted is the failure, not the service — the person wrote them and needs them honored six turns later, silently. State a constraint only when it changed what you did ("you asked me to simplify `importVideo.ts`; that probe is load-bearing, so I left it"), never as an opening recital.
- **Derivable staleness is self-healing, not a finding.** If `PLAN.md` names a SHA, a count, or a branch position that the repo contradicts, use the repo's value and say nothing. With no git available, ignore the recorded value rather than reporting it — never open a session by telling someone their own file is out of date. Report staleness only where it needs a human: a described feature that no longer exists, a plan step the code contradicts.
- **A nudge needs a measurable trigger.** One line, at most one per resume, and only from something countable — today that is `DECISIONS.md` past roughly 50KB, which earns a one-line `/ds-project-compact` suggestion. Never invent a nudge from judgment; without a threshold it is just an opinion, and opinions are how a four-line report grows into a wall.

## Output

Step 7's list and nothing beyond it — enough to start working immediately. Most resumes are a handful of lines; a clean project with nothing in flight should produce almost none.
