Initialize recall and install its session integration into your AI assistant.

Sets up [recall](https://github.com/gleicon/recall) for the current project and installs its session integration by delegating to recall's own `install-skill` — no hand-written host config.

## Check

Verify `recall` is on PATH:

```bash
command -v recall
```

If missing:

```
recall is not installed.

Install:
  Go:     go install github.com/gleicon/recall@latest
  Source: https://github.com/gleicon/recall

Re-run /ds-recall-setup after install.
```

Stop. Do not proceed.

## Process

1. **Index project** — run `recall map` to build the project map.

2. **Seed recipes** — run `recall recipes seed` to load default framework patterns (Go, Next.js, Python, Rust, and others). This is the cross-project knowledge base kickstart.

3. **Install recall's session integration** — delegate to recall's own installer; do not hand-write host config:
   - `recall install-skill --target claude` — always.
   - `recall install-skill --target opencode` — only if `~/.config/opencode/` exists.
   - `recall install-skill --target cursor` — only if `~/.cursor/` exists.
   - `recall install-skill --target codex` — only if `~/.codex/` exists.

   `install-skill` installs recall's hook and merges the assistant's `settings.json`, backing it up first.

## Rules

- Let recall own host config — never hand-edit `settings.json` or write plugins; `recall install-skill` does this and backs up `settings.json` first.
- Only run a `--target` whose assistant config dir exists (`~/.claude`, `~/.config/opencode`, `~/.cursor`, `~/.codex`).
- Setup is idempotent: re-running refreshes the index and seeds, and `install-skill` is safe to re-run.

## Output

```
recall setup complete

  project indexed:    recall map ✓
  recipes seeded:     recall recipes seed ✓
  recall integration: install-skill ✓ (claude; opencode/cursor/codex when present)

Run /ds-recall to inject context into any session.
Run /ds-recall-capture before /clear to store what you built.
```
