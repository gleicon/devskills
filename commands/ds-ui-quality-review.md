Run a strict review of UI quality — whether the interface is soundly engineered, crafted, accessible, and fast. Reports a findings list by default; `--fix` applies the mechanical, unambiguous fixes (design-judgment ones stay reported).

When invoked, audit the UI in scope against one governing principle: **a UI is judged on both halves — it works and it's crafted.** Correct rendering is the floor, not the bar. Hunt the failures that ship broken or generic interfaces with equal energy: components that mishandle state and async edges (the bug a user hits), and interfaces that are inaccessible, slow, or generic-by-default (the experience nobody chose). Framework-agnostic — apply it whether the stack is React, Svelte, Vue, Solid, or plain HTML, on any runtime.

Like `/ds-code-quality-review`, `/ds-doc-quality-review`, and `/ds-test-quality-review`, this produces a prioritized list. It's the after-the-fact audit; to build UI to this standard from the start, use `/ds-ui-mode` (mode). **Do not edit any files unless `--fix` is passed** (see Arguments).

## Arguments

- Treat positional args as scope (files, directories, globs). With no scope, review the UI code changed on the current branch.
- Freeform scope ("the checkout flow", "the settings page") is interpreted reasonably.
- `--fix` → after reporting, apply only the findings whose fix is **mechanical and unambiguous** — a missing `alt`/`aria-label`, a non-semantic `<div>` that should be a `<button>`, an index-as-key, a missing image dimension that causes layout shift. Anything touching design craft, visual hierarchy, or async/state logic rests on judgment and **stays report-only**. After applying, re-run any build/test/lint check already in the loop and revert any fix that breaks it — or that touched more than the intended mechanical edit. Close with a summary of what was applied and what was left.

## What to check

Match the project's existing stack and conventions — a finding is "wrong for this codebase," never "doesn't use my preferred library." Judge against the four dimensions below, in priority order.

**1. Engineering correctness (highest priority — this is where it breaks).**
- **Async states.** Every async surface handles all four explicitly: loading, error, **empty**, success. Empty is the one most often skipped — flag it specifically.
- **Fetch discipline.** Render waterfalls (sequential awaits that could be parallel), stale responses not cancelled or ignored (the last-to-arrive wins instead of the latest request), data fetched far from the component that needs it.
- **State.** State that should be derived from props/existing state but is stored instead (two copies that drift); state lifted higher than its only owner; a component doing two jobs because render logic and data logic are tangled.
- **Boundaries.** Untrusted I/O (API responses, route params, props from outside) used without validation or types at the edge.
- **List keys** that are the array index instead of a stable, data-derived id — flag wherever the list can reorder, insert, or delete.

**2. Accessibility (non-negotiable — a real barrier, not a nit).**
- Clickable non-semantic elements (`<div onClick>` where a real `<button>`/`<a>` belongs); ARIA used where a native element exists.
- Interactive elements not keyboard-reachable, illogical tab order, focus traps, missing visible focus (`:focus-visible`), focus not managed on route change / modal open-close.
- Color as the **sole** signal for state; body text below 4.5:1 contrast.
- Inputs without a programmatic label; dynamic updates not announced via a live region.

**3. Performance (Core Web Vitals — field thresholds at p75: LCP < 2.5s, INP < 200ms, CLS < 0.1).**
- Layout shift: images/media/fonts without reserved space (dimensions / `aspect-ratio`); content injected above existing content.
- INP risk: long synchronous work on the main thread during interaction.
- Wasted bytes: no lazy-loading below the fold, no code-splitting of non-critical paths, render waterfalls on the critical path.

**4. Design craft (the difference between generic and intentional).**
- Ad-hoc pixel values instead of a type scale and spacing system (tokens/variables).
- Flat, uniform-weight hierarchy — everything the same size/weight/spacing, nothing leading the eye.
- The generic-AI defaults: purple-gradient / glass-card cliché, default system typography where the look matters, cramped or gridless spacing.
- Motion that isn't gated by `prefers-reduced-motion`.

Also flag the tells of unfinished work: hardcoded or mock data left in place of real wiring, and inline/default-export types where the project's convention is co-located, named exports.

## Output

A prioritized findings list, in this order:

1. Engineering correctness — broken/missing async states, fetch and stale-response bugs, drifting state (the failure a user hits)
2. Accessibility barriers — keyboard, focus, labels, contrast
3. Performance — layout shift, INP, oversized critical path
4. Design craft — generic-by-default, broken hierarchy, unsystematized spacing/type

For each finding:

- Anchor to `file:line`.
- State the problem in one line, then the fix. Name the concrete failure — *which* of the four async states is missing, *which* element isn't keyboard-reachable, *what* shifts the layout — not "improve a11y" or "needs polish."

Rules:

- A short high-conviction list beats a long pedantic one. Don't pad it with subjective styling nits — design findings must point to a concrete, defensible problem (broken hierarchy, missing token system), not personal taste.
- Respect the project's conventions and stack. Never recommend a new library or pattern where the codebase already has one.
- Report-only by default — the output is the list. With `--fix`, apply only the mechanical, unambiguous findings above and leave the judgment-dependent ones reported; then summarize what was applied vs. left.
