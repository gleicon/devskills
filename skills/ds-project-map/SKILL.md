---
name: ds-project-map
description: "Map the current repository into `.project/map.md`."
disable-model-invocation: true
---

When invoked, read the actual codebase and write `.project/map.md` — the structural description a session reads to find its way around. Describe what exists; invent no direction and make no technical decisions.

## Process

1. Create `.project/` if it does not exist.
2. Establish facts efficiently — locate before reading. Lean on cheap signals first: the file tree, manifests (`package.json`, `go.mod`), entry points. Search to find the major pieces and read scoped regions, not whole files. The map is identifiers and one-liners, not a digest of every file.
3. Write `.project/map.md` with three sections:
   - **Overview** — one paragraph: what the project is and who it's for.
   - **Stack** — languages, runtime, key dependencies, build/test commands.
   - **Repo map** — a short table of the top-level directories and what each holds.

## Rules

- **Regenerate the whole file every time.** Never diff it, never preserve part of it, never ask what to keep. Rewriting wholesale is what keeps this skill free of maintenance — and it is only safe because of the next rule.
- **Nothing belongs here that you could not re-derive from the source.** A constraint, a project rule, a hard-won gotcha — anything invisible to the code — goes in `.project/state.md` under `# settled` or `# hazards`, which `/ds-project-checkpoint` writes. Put it here and the next run silently deletes it.
- Facts only, derived from the code. Reference real files and dirs by path.
- Keep it short. This is a map, not documentation. No current state and no tasks — those live in `state.md`.

## Output

The path, and one line on what changed.
