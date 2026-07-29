---
name: ds-onboarding
description: "Get a new developer oriented in a codebase: map the entry points, conventions, and risks they need to know before making changes."
disable-model-invocation: true
---

Onboarding is not a tour of every file. It is a focused brief that lets a new developer make a safe first change.

## When To Use

- A new teammate joins the project.
- The user asks "how does this repo work?" or "what should I know before touching X?".
- You are returning to a codebase after a long break.

## Workflow

1. Identify the user's role and goal: backend, frontend, infra, reviewer, or full-stack?
2. Map the entry points: build/test commands, main package, key config files, and how to run locally.
3. Summarize the codebase shape: major directories, module boundaries, and the data flow for one representative path.
4. Surface the guardrails: test conventions, lint/format gates, required env vars, auth patterns, and any sensitive areas.
5. Flag the common pitfalls: coupling traps, implicit state, magic files, or places where copy-paste fails.
6. Suggest a safe first change and the tests to run.

## Output

Use this shape:

```text
## Entry points
- build: <command>
- test: <command>
- run locally: <command>

## Codebase shape
- <dir/file>: <what it owns>
- <dir/file>: <what it owns>

## Guardrails
- <convention/tool>: <what it enforces>
- <env/auth>: <how to set it up>

## Common pitfalls
- <area>: <what to watch for>

## Suggested first change
- <small, low-risk task> — run <tests> after.
```

Keep it short enough to read in one sitting. Link to longer docs only when asked.
