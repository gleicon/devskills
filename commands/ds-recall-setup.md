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

1. **Index project** — run `recall map` to build the project index.

2. **Seed recipes** — run `recall recipes seed` to load default framework patterns (Go, Next.js, Python, Rust, and others). This is the cross-project knowledge base kickstart.

3. **Configure Claude Code reminder** — detect if `~/.claude/settings.json` exists. Add a `Stop` hook that prints a reminder when Claude finishes a session:

   ```json
   {
     "hooks": {
       "Stop": [
         {
           "matcher": "",
           "hooks": [
             {
               "type": "command",
               "command": "echo 'recall: run /ds-recall-capture to store this session'"
             }
           ]
         }
       ]
     }
   }
   ```

   Merge into existing hooks if the file already has them — do not overwrite unrelated config. If `~/.claude/settings.json` does not exist, create it with this content only.

4. **Configure OpenCode reminder** — write a JS plugin to `~/.config/opencode/plugins/recall-reminder.js` that fires on `session.idle`:

   ```js
   export const RecallReminderPlugin = async ({ $ }) => {
     return {
       "session.idle": async () => {
         await $`echo "recall: run /ds-recall-capture to store this session"`
       }
     }
   }
   ```

   Create `~/.config/opencode/plugins/` if it does not exist.

5. **Confirm opt-in** — ask once whether to enable automatic capture (writes `.recall/.devskills-capture`). This is the same gate as `/ds-recall-capture`; skip if the file already exists.

## Rules

- Do not overwrite unrelated content in `settings.json` — merge the hook, don't replace.
- Do not install the OpenCode plugin if `~/.config/opencode/` does not exist (OpenCode not installed).
- The hooks are **reminders only** — they print a message, nothing more. Actual capture requires the user to run `/ds-recall-capture` explicitly.
- Setup is idempotent: re-running refreshes the index and seeds, but does not duplicate hooks.

## Output

```
recall setup complete

  project indexed:   recall map ✓
  recipes seeded:    recall recipes seed ✓
  Claude Code hook:  ~/.claude/settings.json (Stop reminder added)
  OpenCode plugin:   ~/.config/opencode/plugins/recall-reminder.js
  capture opt-in:    <enabled|disabled>

Run /ds-recall to inject context into any session.
Run /ds-recall-capture before /clear to store what you built.
```
