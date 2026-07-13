---
name: ds-project-resume
description: "Restore working context from `.project/PLAN.md` (and a fresh handoff, if any), and apply any configured modes."
disable-model-invocation: true
---

When invoked, read the project's persisted state and report where to pick up — the counterpart to `/ds-project-checkpoint`. Safe to run at the start of any session.

## Arguments

- `--no-modes` — skip applying the modes in `.project/config.md`. Resume still lists what the project configures, so opting out never hides them.

## Process

1. **Apply configured modes** (unless `--no-modes`). If `.project/config.md` exists, read its `## Modes` list and, for each mode, read that mode's installed command file from your assistant's command directory (`~/.claude/commands/`, `$CLAUDE_CONFIG_DIR/commands/`, `~/.codex/prompts/`, `~/.opencode/commands/`) and adopt its rules for the rest of the session (read-and-adopt). Echo which modes you applied; for a mode that grants an authorization or destructive-ish behavior — e.g. `ds-git-mode`'s standing authorization to commit without asking — spell out that consequence in the echo so it is never a surprise. If a listed mode can't be found, say so and tell the user to apply it manually — never silently skip. With `--no-modes`, don't apply them but still list what `config.md` configures. If `.project/config.md` is absent, drop a one-line hint that the user can create one with `/ds-project-config` to auto-apply modes.
2. If `.project/PLAN.md` does not exist, say so and suggest `/ds-project-map` then `/ds-roadmap`. Stop.
3. Read `.project/PLAN.md` — focus on `## Now` (state, next, open questions) and the `## Roadmap` status.
4. Read `.project/PROJECT.md` if present, for the repo map and constraints.
5. Read `.project/DECISIONS.md` if present — note how many decisions it records and surface the most recent few if they bear on the next action. Never dump the whole file; it grows over time.
6. If `.project/handoff.md` exists, check whether it is still current — by **file modification time, not git** (the workflow must work when `.project/` is git-ignored or the repo has no git):
   - If `handoff.md` is newer than `.project/PLAN.md`, load it — it's the freshest context.
   - If it is older than `PLAN.md` (a checkpoint happened after it), treat it as **stale**: mention it exists and its date, but do not rely on it.
   - If the repo uses git, you may *optionally* also flag the handoff as stale when commits have landed since it was written — but never require git; the file-time comparison is the source of truth.
7. Summarize: current state, the next action, open questions, recorded decisions worth recalling, the modes applied, and anything stale worth noting.

## Rules

- Resume does not modify `.project/` files — it reads them. Applying a configured mode is part of resume; once applied, that mode governs the session under its own rules.
- Apply modes via read-and-adopt only — never assume a mode is active without reading its command file.
- Trust `## Now` over a stale `handoff.md`.

## Output

A short orientation — where we are, the next step, open questions, and which modes were applied (or skipped) — enough to start working immediately.
