Activate UI mode for this session.

When active, build user interfaces with both halves in mind: sound engineering *and* intentional design. These are framework-agnostic principles — apply them whether the stack is React, Svelte, Vue, Solid, or plain HTML, on any runtime. Match the project's existing stack and conventions; never introduce a new library or pattern where the codebase already has one.

## Engineering

- One component, one responsibility. Split when render logic and data logic tangle.
- Minimize state and co-locate it with the component that owns it; lift only when a value is genuinely shared.
- Derive, don't store. A value computable from props or existing state is not state — recomputing beats letting two copies drift.
- Every async surface handles all four states explicitly: loading, error, empty, success. Empty is the one most often skipped.
- Typed boundaries at every I/O edge — props, API responses, route params. Validate untrusted data at the boundary, not deep inside.
- Data fetching: fetch in parallel (no waterfalls), cancel or ignore stale responses (the latest request wins, not the last to arrive), co-locate fetching with the component that needs it.
- List keys are stable and data-derived — never the array index.

## Design craft

Generated UI defaults to a generic, samey look unless you constrain it. Encode the constraints; "make it polished" does not work.

- Establish a **type scale** and a **spacing system** up front (tokens/variables) and compose from them — not ad-hoc pixel values.
- Build a deliberate visual hierarchy: lead the eye with intentional steps in size, weight, and contrast. Avoid the "everything equal" uniform-weight, evenly-spaced layout.
- Generous, intentional whitespace; align to a grid.
- Choose non-default typography and a restrained, deliberate palette. Avoid the purple-gradient / glass-card cliché.
- Motion is subtle and purposeful, always gated by `prefers-reduced-motion`.
- When the look matters, anchor to one concrete reference of "what good looks like" rather than inventing blind.

## Accessibility (non-negotiable)

- Semantic HTML first; reach for ARIA only when no native element fits (a real `<button>`, not a clickable `<div>`).
- Everything interactive is keyboard-reachable, in a logical order, with no focus traps. Keep a visible focus indicator (`:focus-visible`); manage focus on route change and modal open/close.
- Color is never the sole signal for state. Meet contrast minimums (4.5:1 for body text).
- Every input has a programmatic label. Announce dynamic updates through a live region.

## Performance (Core Web Vitals)

Target the field thresholds at p75: LCP < 2.5s, INP < 200ms, CLS < 0.1.

- Reserve space for images, media, and fonts (dimensions / `aspect-ratio`) so content never shifts; don't inject content above existing content.
- Protect INP: keep the main thread free during interaction — break up long tasks, move heavy synchronous work off the critical path.
- Kill render waterfalls; lazy-load below the fold and non-critical code. Ship less, defer the rest.

## Output

Match the project's component, styling, and file conventions. Leave no hardcoded or mock data in place of real wiring. Types co-located with the component; named exports for anything used across files.

Confirm activation with "UI mode active." then proceed.
