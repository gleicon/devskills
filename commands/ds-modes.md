Open a quick multi-select of devskills session modes and activate the ones you pick.

A launcher, not a mode itself: it shows the mode menu, you choose, and it turns the selected modes on for the session. The chosen set *is* the active set — an unchecked mode is off. Modes compose, so several can run at once.

## On activation

1. Present the menu below as a **multi-select**. If the host supports an interactive multi-select (e.g. Claude Code's question UI), use it; otherwise list the modes and ask the user to reply with the ones to enable. Pre-check any mode already active this session.

   - **tiger-style** — safety + explicitness engineering bar: assertions, bounded loops, no silent fallbacks
   - **test** — keep the code that matters tested by risk as you build (not coverage-chasing)
   - **tdd** — test-first, one vertical slice at a time
   - **ui** — component/state discipline + design craft, accessibility, Core Web Vitals
   - **data** — data-pipeline discipline: idempotency, late/out-of-order data, replay/backfill safety
   - **git** — commit each working unit, terse Conventional-Commit messages, branch-first, no history rewrite
   - **step** — user-gated execution: smallest step → stop → free-form handback → repeat
   - **quality-gate** — the seven-pass, deslop-bookended pre-PR review pipeline
   - **caveman-lite** — compress responses ~25–35%
   - **caveman-ultra** — compress responses ~75–85% (mutually exclusive with caveman-lite)

2. For each **selected** mode, adopt its full behavior by activating its command `/ds-<name>-mode` — read that command file (e.g. `ds-tiger-style-mode.md`) from your commands directory if its exact rules aren't already in context. Read **only the selected** modes' files; never load the whole set.
3. If both caveman-lite and caveman-ultra are selected, keep ultra and note the conflict. If the selection removes a currently-active mode, turn it off.
4. Confirm in one line which modes are now active (and which, if any, were turned off).

## Notes

- This is the fast path: the menu above is hardcoded, so showing it costs no file reads — only the modes you pick get loaded.
- Keep the menu in sync when a mode is added or removed (one line per mode).
- The interactive picker is host-dependent; where it isn't available the prose fallback works the same. In rules-only editors (Cursor) that can't read other command files, modes are applied from the summaries above plus the always-on Tiger Style rule.
