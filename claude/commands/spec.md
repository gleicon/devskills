Convert a description into a structured specification.

When invoked, gather information and produce a spec document that defines WHAT the system does, not HOW it is implemented. The spec becomes the input to `/workflow` planning phases.

## Process

If the user has provided a description: proceed directly.
If not: ask three focused questions, then produce the spec without further prompting.

Questions (ask only what is missing):
1. What is the primary user action this system enables?
2. What does success look like — how do you know it works?
3. What are the hard constraints: scale, latency, cost, compliance?

## Spec Structure

Produce a document with these sections:

```
# Specification: <name>

## Problem
One paragraph. What problem does this solve and for whom?

## Scope
What is in scope. What is explicitly out of scope.

## Users
Who uses this. What they need to accomplish. No personas — just roles and goals.

## Functional Requirements
Numbered list. Each requirement is a verifiable statement.
FR-1: The system shall <verb> <object> when <condition>.

## Non-Functional Requirements
NFR-1: Latency: <P99 target> under <load description>
NFR-2: Scale: <concurrent users or requests/sec>
NFR-3: Availability: <uptime target>
NFR-4: Data retention: <policy>

## Interfaces
- API endpoints or CLI commands (names and purpose, not full schema)
- External systems this integrates with
- Data formats at boundaries

## Constraints
- Language/runtime (from language profile if set)
- Infrastructure (from project context if set)
- Forbidden approaches

## Acceptance Criteria
Numbered list. Each item is a pass/fail test that can be run.
AC-1: Given <state>, when <action>, then <outcome>.

Coverage — map every FR to the AC(s) that verify it. An uncovered FR is a spec defect:
FR-1 → AC-1, AC-2
FR-2 → AC-3

## Open Questions
Questions that must be answered before implementation begins.
```

## Requirement Quality Rules

- Every requirement is verifiable. If it can't be tested pass/fail, it's a wish, not a requirement — rewrite it or move it to Constraints.
- SHALL = mandatory, SHOULD = recommended, MAY = optional. Choose deliberately.
- No ambiguous quantifiers: not "fast" or "reasonable" — state the number.
- One requirement per statement. Split compound requirements; list exceptions as sub-items (`FR-N.a`, `FR-N.b`).

## Language Profile Integration

If a language profile is active, add a "Technical Profile" section:
- Primary language and version
- Runtime target
- Build toolchain
- Testing framework

## Output

Write the spec to `.project/SPEC.md` if `.project/` exists, else `.planning/SPEC.md` if `.planning/` exists, else `SPEC.md` in the current directory; then display it inline for review.

Ask: "Should I proceed to planning with `/workflow`?" after displaying the spec.
