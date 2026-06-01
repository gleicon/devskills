Enter the development workflow — orient, spec, build, verify, ship.

A standalone phase-map orchestrator: it orients you, then routes each phase to its primary command. It requires nothing else — every phase works on its own. When `.project/` is present it *also* reads that state to orient faster, but the `.project/` commands are optional persistence, never required: you get the same workflow with or without them.

## On activation

1. If `.project/` is present, read `.project/PROJECT.md` (context) and `.project/PLAN.md` (where to pick up) when they exist, and report where to resume. If no plan is found — or `.project/` isn't in use — treat it as a fresh start and point at `/ds-spec` (lock the WHAT) or `/ds-explore` (lay out approaches) to begin.
2. Report current state: what exists, what phase you appear to be in, what's next.
3. Activate `/ds-tiger-style-mode` implicitly — engineering bar on for the session.
4. If a language profile is set in `AGENTS.md` (look for `<!-- profile: <lang> -->`), apply its conventions.

## Phase map

Each phase has a primary command and the question it answers. Emit an Insight block when entering each phase.

### Orient
*"Where am I? What does this code do?"*
```
`Insight ─────────────────────────────────────`
/ds-zoom-out — map the area before changing it
/ds-project-resume — restore where you left off (only if .project/ is in use)
`─────────────────────────────────────────────`
```

### Spec
*"What exactly are we building?"*
```
`Insight ─────────────────────────────────────`
/ds-spec — lock the WHAT + acceptance criteria → SPEC.md
/ds-explore — lay out approaches at a fork (--web to research)
/ds-grill-me --record — decide open branches → DECISIONS.md
`─────────────────────────────────────────────`
```

### Plan
*"In what order do we build it?"*
```
`Insight ─────────────────────────────────────`
/ds-blueprint — for a new system, commit to a structure first (modules, deps, build order)
/ds-project-plan — sequence spec/decisions into ordered tasks (only if you persist state under .project/)
`─────────────────────────────────────────────`
```

### Build
*"Build it."*
```
`Insight ─────────────────────────────────────`
/ds-tiger-style-mode — safety + explicitness bar (already active)
/ds-tdd-mode — drive implementation with tests
/ds-test-mode — keep the core tested as you build (mode, stays on)
/ds-ui-mode — component/state discipline + design craft (UI work)
`─────────────────────────────────────────────`
```
Modes compose: `/ds-tiger-style-mode /ds-test-mode /ds-ui-mode` can all be active at once.

### Clean
*"Strip the AI slop before anyone looks at it."*
```
`Insight ─────────────────────────────────────`
/ds-deslop — remove narrating comments, defensive overkill, type escape hatches
Run this before any review pass.
`─────────────────────────────────────────────`
```

### Review
*"Is it correct, safe, and idiomatic?"*
```
`Insight ─────────────────────────────────────`
/ds-code-quality-review — structure: is the diff making the codebase worse?
/ds-bug-review — correctness: real bugs, not style
/ds-security-review — if it touches input, auth, secrets, or I/O
/ds-test-quality-review — is the risky logic actually covered?
/ds-go-review · /ds-ts-review · /ds-rust-review — language idioms + security
`─────────────────────────────────────────────`
```

### Verify
*"Does it actually do what I claimed?"*
```
`Insight ─────────────────────────────────────`
/ds-verify-this <claim> — before/after repro, hard verdict: VERIFIED / NOT VERIFIED
Give it something measurable. It refuses vague claims.
`─────────────────────────────────────────────`
```

### Persist / ship
*"Save state, open the PR."*
```
`Insight ─────────────────────────────────────`
/ds-handoff — capture goal/done/remaining into a handoff doc for a long pause or another person (works anywhere)
/ds-project-checkpoint — persist state before /clear or end of session (only if .project/ is in use)
Then: git push + gh pr create
`─────────────────────────────────────────────`
```

## Shortcuts

```
/ds-workflow orient   → zoom out, then report state
/ds-workflow spec     → jump to specification phase
/ds-workflow build    → activate tiger-style + suggest build modes
/ds-workflow review   → run the full pre-PR gate (deslop → reviews → verify)
/ds-workflow status   → report current phase, what exists, what's next
/ds-workflow ship     → persist state (handoff, or checkpoint if .project/ is in use) + gh pr create guidance
```

## Surviving long sessions

```
`Insight ─────────────────────────────────────`
/ds-caveman-lite-mode (~30% savings) or /ds-caveman-ultra-mode (~80%) — compress responses
/ds-tldt <file|url> — compress a long doc before adding it to context
`─────────────────────────────────────────────`
```

Respond with: current state (what files exist, what phase), the recommended next command, and the Insight block for that phase.
