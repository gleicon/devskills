Author a new devskills command.

When invoked, help the user create a command file that fits the devskills conventions, then place it in `commands/`.

## devskills command format

A command is a single Markdown file — no YAML frontmatter. Structure:

- **Line 1** — one imperative sentence stating what the command does. This is what the agent sees when deciding to load it, so make it specific.
- **Blank line, then a framing paragraph** — for an *action* command, what it does when invoked; for a *mode* command, what behavior it switches on for the session.
- **Concrete `##` sections** — ordered instructions. Keep prose tight; fragments are fine. Two archetypes, two natural shapes:
  - **Action commands** (do a job, produce a result — `/ds-spec`, `/ds-deslop`, the reviews): `## Process` / `## Rules` / `## Output`, or domain-fit variants like `## Review Checklist` + `## Output Format`. Always end with the result the user gets.
  - **Mode commands** (switch on a session behavior — `/ds-tiger-style-mode`, `/ds-caveman-lite-mode`, `/ds-ui-mode`): describe the behavior the mode enforces. No `## Output` — the effect is ongoing, not a single deliverable.

No companion files. If a command needs more than one file, it is doing too much — split it into separate commands.

## Process

1. Gather the requirement: what task the command automates, and when the user would invoke it.
2. Draft the command file following the format above. Keep it focused — one job per command.
3. Write it to `commands/ds-<name>.md` (mode commands: `commands/ds-<name>-mode.md`) — install.sh copies it to Claude Code, OpenCode, and Codex (the last invoked as `/prompts:ds-<name>`).
4. Register it: add a row to the README "Commands" table and an entry to `docs/commands.md`.
5. Show the draft to the user for review before finalizing.

## Rules

- The filename is the command name: `/ds-<name>` invokes `ds-<name>.md`. Every devskills command carries the `ds-` prefix (avoids collisions with Claude Code / OpenCode built-ins); modes additionally end in `-mode` (e.g. `ds-tiger-style-mode.md`) so a name shows its kind at a glance.
- Match the voice of the existing commands — imperative, no filler.
- If adapting from an external source, add the upstream repo to the README "References" section instead of a `Source:` line in the command.

## Output

The new command file and the registration edits (README + `docs/commands.md`), shown for review.
