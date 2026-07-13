---
name: ds-git-mode
description: "Activate git mode for this session."
disable-model-invocation: true
---

When active, manage version control the way a senior engineer does: commit each self-contained, working change as it lands, with terse, human-readable messages — never the verbose, LLM-targeted commit bloat. Activating this mode is your standing authorization to commit completed units without asking each time; it still **never pushes** and **never rewrites history**. Work happens on a branch, not the default.

## When to commit

- A commit is **one self-contained, working change** — the smallest unit that stands on its own: it builds/passes, it's reversible, and a reviewer can understand it alone.
- Commit as soon as such a unit is **done** and the tree is in a good state. Don't hoard a session's work into one giant commit; don't commit mid-thought, broken, or WIP.
- **One logical change per commit.** Don't dribble it across a flurry of `wip`/`fixup`/`oops` commits, and don't bundle unrelated changes together — if the tree holds two logical changes, stage and commit them separately (`git add -p` when needed).
- If a build/test/lint check is already in the loop, the unit should pass it before you commit. Never knowingly commit a broken tree.
- After committing, report **one line** — the commit subject — so the human stays in the loop. Don't ask permission each time.

## Message format (Conventional Commits, kept terse)

- `type(scope): subject`. Types: `feat`, `fix`, `refactor`, `perf`, `docs`, `test`, `build`, `ci`, `chore`, `style`. Scope is optional — include it only when it sharpens the subject (`fix(auth): …`).
- **Subject**: imperative mood, lowercase, no trailing period, ~50 chars. State the *outcome* at the right altitude — what changed and why it matters to a reader — not a file-by-file description the diff already shows.
- **Body**: only when there's a non-obvious **WHY** — a constraint, a tradeoff, a non-local consequence, a bug being worked around. Wrap ~72 cols. **Most commits need no body.** Never restate the diff as bullets.
- **Breaking changes**: `type!: subject` plus a `BREAKING CHANGE: …` footer. Only for real breaks.
- Write for a human skimming `git log --oneline`. **No emoji, no "Generated with…", no AI/co-author attribution trailers, no marketing prose.** If the project's existing history follows a trailer convention, match it; otherwise add none.

## Branch & history

- Work on a branch. If you're on the default branch (`main`/`master`), create a concise, type-prefixed branch (`feat/…`, `fix/…`, `refactor/…`) from the work and switch to it **before the first commit**. Match the project's branch-naming convention if it has one.
- **Never rewrite shared history**: no rebase, no squash that rewrites, no force-push, no amend of a pushed commit. Keep history linear and append-only — each commit is real done work. (Amending the *latest, not-yet-pushed* commit to fix its own message or contents before anyone sees it is fine.)
- **Never push unless explicitly asked.** Opening PRs is out of scope — hand off to `gh` when the user asks.

Confirm activation with "Git mode active." Activating a mode only turns on this posture; it is not approval to begin work — continue with whatever the user already asked for, or wait for their next instruction.
