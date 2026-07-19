---
name: ds-project-map
description: "Map the current repository into a durable project description at `.project/PROJECT.md`."
disable-model-invocation: true
---

When invoked, read the actual codebase and write (or refresh) `.project/PROJECT.md` — the stable description a fresh session reads to understand the project. Describe what exists; do not invent direction or make technical decisions.

## Process

1. Create `.project/` if it does not exist.
2. Establish facts efficiently — locate before reading. Lean on cheap signals first: the file tree, manifests (`package.json`, `go.mod`, etc.), and entry points. Search to find the major pieces; read scoped regions, not whole files. Read enough to be accurate, no more — the map is identifiers and one-liners, not a digest of every file.
3. Write `.project/PROJECT.md` with these sections:
   - **Overview** — one paragraph: what the project is and who it's for.
   - **Stack** — languages, runtime, key dependencies, build/test commands.
   - **Repo map** — a short table of the top-level directories and what each holds.
   - **Constraints** — hard rules the project must respect (perf, compatibility, forbidden approaches), if any are evident or already recorded.
4. **If `.project/PROJECT.md` already exists, treat it as partly human-authored.** Refresh the code-derived sections (Stack, Repo map) to match the current source — but before overwriting, compare what you would generate against what's already there. Anything present that you would **not** re-derive from the code — a manually-added project truth, a constraint or rule not visible in the source, a custom section, hand-written Overview prose — is human-added knowledge. **List those items and ask whether to keep, update, or drop each one before you write.** Never silently discard them; default to keeping.

## Rules

- Facts only — derive everything from the code, not from assumptions.
- Keep it short. This is a map, not documentation. Reference real files and dirs by path.
- Do not record current state or tasks here — that lives in `.project/PLAN.md`.
- **On a rerun, human edits win by default.** Code-derived sections are refreshed freely, but anything you can't re-derive from the code is preserved unless the user agrees to change it — and when something would be lost, ask rather than assume.
- **`## Landmines` is never yours to touch.** `/ds-project-checkpoint` appends its rows and `/ds-project-verify` proposes removals; you preserve the section verbatim and keep it out of step 4's ask-list. A landmine is invisible to the code by definition — that is why someone wrote it down — so it would fail every re-derivation test and be offered for deletion on every rerun. Nobody notices a missing warning until they hit the thing it warned about.

## Output

Display the path and a one-line summary of what changed (created, or which sections were refreshed).
