## Phase-Aware Suggestions

Emit an Insight block when the conversation reaches a clear phase transition and a specific command would help — not on every turn. Use this format exactly:

```
`Insight ─────────────────────────────────────`
[one or two concrete suggestions]
`─────────────────────────────────────────────`
```

Trigger only on the signals that genuinely apply right now:

**Starting a new task** — user describes a feature or bug from scratch, no spec yet.
Suggest: `/ds-spec` to lock the WHAT before the HOW.

**At an architectural fork** — user is choosing between approaches or asks "should I use X or Y?"
Suggest: `/ds-explore` to surface trade-offs; `/ds-grill-me` to pressure-test the choice one branch at a time.

**Code just generated** — a significant block of code was just written.
Suggest: `/ds-deslop` to strip narrating comments and defensive overkill before anyone reviews it.

**Done / opening a PR** — user says "I'm done", "ready to review", or "opening a PR".
Suggest: `/ds-code-quality-review` then `/ds-bug-review`; `/ds-security-review` if it touches input/auth/secrets; `/ds-verify-this <claim>` to prove the headline change works.

## Rules

- One Insight block per transition; never repeat a suggestion already made this session, and skip it when the user already knows (they just ran the command or said so).
- Name the exact command and what it gives, one line each.
