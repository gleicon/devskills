Configure project preferences in `.project/config.md` — currently the modes auto-applied on resume.

When invoked, create or update `.project/config.md`, the per-project preferences a session reads at start. Today it holds one thing: the **modes** that `/ds-project-resume` (and `/ds-workflow`) apply automatically. This command edits only `config.md` — it never changes another `.project/` file, and never guesses which modes are active in the current session.

## Arguments

- A space-separated list of mode names sets the `## Modes` list directly (e.g. `ds-git-mode ds-tiger-style-mode`). With no arguments, show the current config and the available modes, then ask which to enable.

## Process

1. Resolve the file: `.project/config.md` if `.project/` exists, else `config.md` in the current directory. Do not create `.project/` itself.
2. **Discover available modes** — list `ds-*-mode` files in your assistant's command directory (`~/.claude/commands/`, `$CLAUDE_CONFIG_DIR/commands/`, `~/.codex/prompts/`, `~/.opencode/commands/`). These are the valid choices.
3. Read the current `## Modes` list if the file already exists.
4. **Determine the desired modes:**
   - From the arguments, if given.
   - Otherwise show current-vs-available and ask which to enable — accept a free-form answer (names to set, or "add X" / "remove Y" / "none"), not a fixed picker.
5. **Validate** each chosen mode against the discovered list. Warn on any name that isn't an installed mode (likely a typo) and leave it out unless the user insists.
6. **Write** `config.md`: a `# Project config` title and a `## Modes` section with one bare mode name per bullet. Create the file with this skeleton if absent; otherwise replace only the `## Modes` section, leaving any other content intact.

## Rules

- Bare names only (`ds-git-mode`), matching the installed command's file stem — no slash prefix.
- Manage the `## Modes` section only. Preserve any other section a human added; the file is general-purpose and may grow.
- Never auto-detect "active" modes from the session — the user chooses explicitly.
- Touch only `config.md`. This command does not read or write `PLAN.md`, `PROJECT.md`, or any other `.project/` file.

## Output

Display the written `## Modes` list and the file path, plus a one-line note that these apply on `/ds-project-resume` and `/ds-workflow` (skip with `--no-modes`).
