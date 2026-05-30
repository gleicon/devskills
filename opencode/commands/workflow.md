Enter the standard development workflow: specification to shipped product.

This workflow uses GSD (Get Shit Done) as its execution engine.

## Workflow Phases

```
1. Spec      → Structure the problem. Define what, not how.
2. Discuss   → Capture decisions, constraints, assumptions.
3. Plan      → Research and create an execution plan.
4. Execute   → Implement the plan in parallel sub-agents.
5. Verify    → Test and validate against the specification.
6. Ship      → Create PR with review and merge readiness.
```

## Phase Commands

| Phase | GSD Command | When to use |
|-------|-------------|-------------|
| Spec | `/gsd-new-project` or `/spec` | Starting a new project or feature |
| Discuss | `/gsd-discuss-phase` | Before planning any phase |
| Plan | `/gsd-plan-phase` | After discussion is complete |
| Execute | `/gsd-execute-phase` | After plan is reviewed and approved |
| Verify | `/gsd-verify-work` | After execution completes |
| Ship | `/gsd-ship` | After verification passes |

## Context Management

GSD keeps main context at 30-40% utilization by delegating heavy work to sub-agents. The `.planning/` directory persists state across sessions.

Key artifacts:
- `.planning/ROADMAP.md` — phases and milestones
- `.planning/PLAN.md` — current phase execution plan
- `.planning/state/` — session state and checksums

## Entering the Workflow

When this skill is activated:

1. Ask: "What are you building?" if no project context exists.
2. If a `.planning/` directory exists, read `ROADMAP.md` and report current phase.
3. Suggest the appropriate next GSD command.
4. Apply Tiger Style constraints to all generated code (activate `/tiger-style` implicitly).

## Language Profile

If a language profile is set (go, typescript, javascript, rust), apply its conventions to all code in this session. Check `AGENTS.md` in the project root for a `<!-- profile: <lang> -->` marker.

## Shortcuts

```
/workflow spec    → jump to specification phase
/workflow plan    → jump to planning (requires existing spec)
/workflow execute → jump to execution (requires existing plan)
/workflow status  → report current phase and blockers
/workflow verify  → run verification on current phase output
```

Respond with current workflow status and the recommended next step.
