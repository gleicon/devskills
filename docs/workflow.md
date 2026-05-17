# Development Workflow Reference

This document describes the standard specification-to-product workflow used in devskills.

The execution engine is GSD (Get Shit Done): https://github.com/gsd-build/get-shit-done

---

## The Problem: Context Rot

AI coding sessions degrade over time as the context window fills. Decisions made early in a session become unavailable later. State is lost between sessions. This workflow solves that with persistent artifacts and scoped sub-agents.

---

## Phases

### 1. Specification (`/spec`)

Convert a description or vague requirement into a structured SPEC.md with:
- Problem statement
- Functional requirements (numbered, verifiable)
- Non-functional requirements (latency, scale, availability)
- Acceptance criteria (pass/fail testable)
- Open questions that block implementation

Output: `SPEC.md` (or `.planning/SPEC.md`)

### 2. Discussion (`/gsd-discuss-phase`)

Before planning: capture implementation decisions, constraints, and assumptions. This prevents re-litigating settled decisions during execution.

Key questions to settle:
- What is the language and runtime?
- What external services are required?
- What are the performance targets?
- What must not be done (scope boundaries)?

Output: decisions recorded in `.planning/` artifacts.

### 3. Planning (`/gsd-plan-phase`)

Research and create an execution plan (PLAN.md) with:
- Task breakdown
- File-level implementation targets
- Dependency order
- Risk identification

Output: `.planning/PLAN.md`

### 4. Execution (`/gsd-execute-phase`)

Implement the plan. GSD executes in parallel sub-agents with fresh context windows, keeping the main context at 30-40% utilization. Each sub-agent works on a scoped task from PLAN.md.

Tiger Style is active during execution.

### 5. Verification (`/gsd-verify-work`)

Test and validate against the specification:
- Does every acceptance criterion pass?
- Are there regressions in adjacent functionality?
- Does the code pass the Tiger Style review checklist?

If verification fails: generate a fix plan and loop back to execution.

### 6. Ship (`/gsd-ship`)

Create a pull request from verified work. Includes:
- PR description with spec summary and acceptance criteria
- Review readiness check
- Merge guidance

---

## Context Management

GSD maintains shared state in `.planning/`:

```
.planning/
├── ROADMAP.md          # phases and milestones
├── SPEC.md             # current specification
├── PLAN.md             # current execution plan
├── VERIFICATION.md     # latest verification report
└── state/              # session state, checksums
```

This directory should be committed to version control. It is the memory of the project that persists across sessions and team members.

---

## Language Profile Integration

When a language profile is active (set via `./scripts/setup.sh --lang=<lang>`):

1. The `/spec` skill includes a Technical Profile section.
2. The `/workflow` skill applies language-specific idioms during planning.
3. Language review skills (`/go-review`, `/ts-review`) are applied during verification.

---

## External Tools Integration

| Tool | Phase | Benefit |
|------|-------|---------|
| RTK | All | 60-90% token reduction on CLI operations |
| tldt | Discuss, Verify | Compress long documents before feeding to context |
| GSD | Plan → Ship | Persistent state and parallel execution |

---

## Shortcuts

```bash
# In Claude Code or OpenCode:
/spec                     # start a new specification
/workflow                 # enter GSD workflow from current state
/workflow status          # report phase and blockers
/tiger-style              # activate style enforcement
/go-review                # review Go code
/ts-review                # review TypeScript code
/tldt                     # summarize current context
/caveman-lite             # compress future responses
```
