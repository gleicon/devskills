Inject local context from recall into the current session.

Requires [recall](https://github.com/gleicon/recall) — a local-first context engine that indexes your project and accumulates cross-project knowledge. If recall is not installed, print install instructions and stop. Do not attempt to emulate its behavior.

## Check

Before anything else, verify `recall` is on PATH:

```bash
command -v recall
```

If the binary is missing:

```
recall is not installed.

Install:
  Go:     go install github.com/gleicon/recall@latest
  Source: https://github.com/gleicon/recall

After install, run /ds-recall-setup to initialize.
```

Stop. Do not proceed.

## Arguments

- No args — index the project then generate a context-rich brief and inject it into the session.
- `query "<question>"` — route a question through recall's smart local-first routing (local model if available, else an enriched brief).
- `brain` — pull cross-project patterns, recipes, and accumulated insights.

## Process

### No args (default)

1. Run `recall map` to index the current project. Always run this — no detection needed; map is idempotent.
2. Run `recall brief` to generate a context-rich brief.
3. Inject the brief into the working context. Summarize what recall surfaced: project shape, relevant prior patterns, any matching recipes.

### `query "<question>"`

1. Check binary (see above).
2. Run `recall query "<question>"`. Report what the router returned: local answer, enriched brief, or a flag that no local context was relevant.
3. Use the returned context to inform your response.

### `brain`

1. Check binary (see above).
2. Read the current session to decide which facets of the global brain are relevant, then dispatch only the matching content subcommands:
   - `recall brain conversations` — when prior sessions likely solved something similar.
   - `recall brain lessons` — when the work is debugging, gotchas, or hard-won constraints.
   - `recall brain snippets` — when reusable code patterns would help.
3. Surface an overview from the stats-like subcommands: `recall brain stats` and `recall brain frameworks`.
4. Report the cross-project patterns returned and highlight anything directly relevant to the current session.

## Rules

- Never emulate recall's output if the binary is absent — that defeats the purpose.
- Always run `recall map` before `recall brief` — brief errors without a prior map.
- Do not pass user input to `recall query` without quoting it.

## Output

```
recall: project indexed
recall: brief injected — <N> recipes matched, <M> prior patterns surfaced

[brief content here]
```

If no recipes match: report that cleanly. Do not fabricate patterns.
