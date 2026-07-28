---
name: ds-project-resume
description: "Read .project/ state and report where to pick up."
disable-model-invocation: true
---

When invoked, load the project's state and report where to pick up — the counterpart to `/ds-project-checkpoint`. Safe to run at the start of any session. It writes nothing.

## Arguments

- `--no-modes` — read `.project/config.md` but don't apply what it lists. Still name the modes it configures, so opting out never hides them.

## Process

1. **Apply the modes in `.project/config.md`.** For each one, read that mode's `SKILL.md` from your assistant's skills directory (`~/.claude/skills/`, `$CLAUDE_CONFIG_DIR/skills/`, `~/.codex/skills/`, `~/.config/opencode/skills/`) and adopt its rules for the session. Each mode ends with its own confirmation line — those are the echo; don't restate the list as prose. Where a mode grants a standing authorization, name it inline on that line. If a listed mode isn't installed, say so and tell the user to apply it by hand — never skip one silently.
2. Read `.project/state.md`. If it is absent, say so, suggest `/ds-project-map` then `/ds-project-checkpoint`, and stop.
3. Report the mode confirmations, then `# now`, then `# next`.

## Rules

- **Two sections are reported, two are loaded.** `# now` and `# next` are the report. `# settled` and `# hazards` enter context so they govern what you do — never say them back to the person who wrote them. Name one only when it changed what you did: "that probe is load-bearing, so I left it."
- **Report the file, not around it.** No git, no counts, no branch position, no observations about the repo. A fact `state.md` doesn't hold is not part of the report, however useful it looks.
- **Modes come first.** A standing authorization read out after the work context is backwards.

## Output

Mode confirmations, `# now`, `# next`. Usually four lines.
