---
name: ds-grill-me
description: "Interview the user relentlessly about a plan or design until reaching shared understanding."
disable-model-invocation: true
---

When invoked, stress-test every aspect of the plan. Walk down each branch of the decision tree, resolving dependencies between decisions one at a time.

## Arguments

If invoked with `--record`, append each resolved decision **worth keeping** to `.project/DECISIONS.md` if `.project/` exists, else `DECISIONS.md` in the current directory, as the interview proceeds. Record the ones that are hard to reverse, surprising without context, or the product of a real trade-off — skip the trivial and the self-evident. One entry per decision: the question, the chosen answer, and a one-line rationale. Plain Markdown, no fixed schema. Without the flag, keep decisions in the conversation only.

## Process

- Ask questions one at a time. Wait for the answer before the next question.
- For every question, provide your recommended answer.
- If a question can be answered by exploring the codebase, explore the codebase instead of asking.
- When the user asserts how something works, check it against the code — if they disagree, surface the conflict: "your code cancels the whole Order, but you said partial cancellation is possible — which is right?"
- Stress-test boundaries with invented edge-case scenarios that force the user to be precise about where one concept ends and another begins.
- Pin down vague or overloaded terms — "you're saying 'account' — the Customer or the User? those are different things."
- Continue until every branch of the decision tree is resolved and you and the user share the same understanding.

## Output

End when no unresolved decision branches remain. Summarize the resolved plan. If `--record` was passed, also report the path to the updated `DECISIONS.md`.
