---
name: ds-onboarding
description: "Onboard a new team member to a project: team context, ownership, rituals, and the first safe contribution path — not a generic code tour."
disable-model-invocation: true
---

`/ds-onboarding` is for team growth, not personal navigation. It answers: *what does a new teammate need to know to participate without stepping on the team?* Use `/ds-zoom-out` when you need to understand code before a change; use `/ds-handoff` when compacting a session for someone else to continue. This skill is the arrival brief for a new contributor.

## When To Use

- A new teammate joins the project.
- A contractor or reviewer needs to become productive quickly.
- The user asks "what should a new team member know before their first PR?".

Do not use this for your own "what does this file do?" questions — reach for `/ds-zoom-out` instead.

## Workflow

1. Identify the new contributor's role: backend, frontend, infra, data, reviewer, or full-stack?
2. Surface team context: owner names, review rituals, Slack/Discord channels, on-call rotation, and where decisions are recorded.
3. Map project entry points: build, test, local run, and CI checks.
4. Summarize the codebase shape: major modules, data flow for one representative path, and which parts are owned by which team.
5. List guardrails: conventions, lint/format gates, required env/auth, and sensitive areas.
6. Flag common pitfalls: coupling traps, implicit state, magic files, tribal knowledge, or areas where copy-paste fails.
7. Propose a safe first contribution: a small, low-risk task, the reviewer to ping, and the tests to run.

## Output

Use this shape:

```text
## Team context
- owners: <names/teams>
- rituals: <standups, reviews, incident process>
- decisions: <ADR, RFC, or docs location>

## Entry points
- build: <command>
- test: <command>
- run locally: <command>
- CI: <command / link>

## Codebase shape
- <dir/module>: <what it owns>
- <dir/module>: <what it owns>

## Guardrails
- <convention/tool>: <what it enforces>
- <env/auth>: <how to set it up>

## Common pitfalls
- <area>: <what to watch for>

## Suggested first contribution
- <small, low-risk task> — review with <owner/reviewer> — run <tests>.
```

Keep it short enough to read in one sitting. Link to longer docs only when asked.
