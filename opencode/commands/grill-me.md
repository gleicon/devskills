Interview the user relentlessly about a plan or design until reaching shared understanding.

When invoked, stress-test every aspect of the plan. Walk down each branch of the design tree, resolving dependencies between decisions one at a time.

## Arguments

If invoked with `--record`, append every resolved decision to `.project/DECISIONS.md` if `.project/` exists, else `DECISIONS.md` in the current directory, as the interview proceeds — one entry per decision: the question, the chosen answer, and a one-line rationale. Plain Markdown, no fixed schema. Without the flag, keep decisions in the conversation only.

## Process

- Ask questions one at a time. Wait for the answer before the next question.
- For every question, provide your recommended answer.
- If a question can be answered by exploring the codebase, explore the codebase instead of asking.
- Continue until every branch of the decision tree is resolved and you and the user share the same understanding.

## Output

End when no unresolved decision branches remain. Summarize the resolved plan. If `--record` was passed, also report the path to the updated `DECISIONS.md`.
