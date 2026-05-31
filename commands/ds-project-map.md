Map the current repository into a durable project description at `.project/PROJECT.md`.

When invoked, read the actual codebase and write (or refresh) `.project/PROJECT.md` — the stable description a fresh session reads to understand the project. Describe what exists; do not invent direction or make technical decisions.

## Process

1. Create `.project/` if it does not exist.
2. Establish facts efficiently — locate before reading. Lean on cheap signals first: the file tree, manifests (`package.json`, `go.mod`, etc.), and entry points. Search to find the major pieces; read scoped regions, not whole files. Read enough to be accurate, no more — the map is identifiers and one-liners, not a digest of every file.
3. Write `.project/PROJECT.md` with these sections:
   - **Overview** — one paragraph: what the project is and who it's for.
   - **Stack** — languages, runtime, key dependencies, build/test commands.
   - **Repo map** — a short table of the top-level directories and what each holds.
   - **Constraints** — hard rules the project must respect (perf, compatibility, forbidden approaches), if any are evident or already recorded.
4. If `.project/PROJECT.md` already exists, refresh the factual sections (Stack, Repo map) to match the current code and preserve any human-written Overview/Constraints prose.

## Rules

- Facts only — derive everything from the code, not from assumptions.
- Keep it short. This is a map, not documentation. Reference real files and dirs by path.
- Do not record current state or tasks here — that lives in `.project/PLAN.md`.

## Output

Display the path and a one-line summary of what changed (created, or which sections were refreshed).
