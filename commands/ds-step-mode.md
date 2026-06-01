Activate step mode for this session.

When active, execute work in small, user-gated steps instead of long autonomous runs. Do the smallest meaningful, reviewable step, then stop and hand control back so the user can approve, steer, change the plan, or checkpoint before the next step. The user drives; you take one step at a time. Invoked with a plan — `/ds-step-mode current plan`, a path like `.project/PLAN.md`, or pasted text — read it and work through it step by step with these breakpoints.

## The loop

- Work in the **smallest meaningful, reviewable step** — one logical change or action. Default to pausing *more*, not less; never silently chain multiple steps.
- **Propose before you act.** State the single next step and *wait* — do not make the change until the user approves. This gives them a veto before anything happens.
- After completing a step, **stop and report concisely**: what you did, what changed (files/commands), and the next step you propose. A few lines, not a wall of text.
- Then **yield**: end your turn and wait. Do not start the next step until the user responds.
- The user may approve ("go" / "next"), amend the next step, change the plan, ask for a checkpoint, ask a question, or redirect entirely — honor whatever they say, including "do the next three steps" (then pause again).
- **Granularity is tunable live**: "bigger steps" / "smaller steps" adjusts how much you do before stopping. Start at one logical unit.

## Handing control back (always steerable)

- At any decision or handback point, **return control in prose** — end with a short, open question the user can answer freely. Never hand back via the multiple-choice picker: the user must always be able to accept an option *and* add instructions, or combine and redirect.
- When you present options, list them as **prose recommendations the user can accept, amend, or combine** — never force a single selection.
- The multiple-choice picker is acceptable **only** for a trivial either/or disambiguation (e.g. "did you mean file A or file B?") — never for "what should I do next" or steering a step.

## Driving a plan

- Restate the next unstarted step from the plan, then run the loop above on it.
- Keep the plan honest: mark steps done as they complete; when the user changes course, update the plan to match. Don't rewrite the whole plan unprompted.
- Offer a checkpoint (`/ds-project-checkpoint`) at natural milestones or on request, so the session stays safe to pause or `/clear`.

Confirm activation with "Step mode active." then proceed.
