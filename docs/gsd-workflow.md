# Get Shit Done: Spec to Production

Source: https://github.com/open-gsd/get-shit-done-redux  
Install: `npx @opengsd/get-shit-done-redux@latest`

GSD is a meta-prompting framework for AI coding tools. It manages context rot (the quality degradation that happens as your session fills up) by persisting state in a `.planning/` directory and delegating heavy work to sub-agents with fresh context windows.

This document covers how to move from a `/spec` output into a GSD-managed project with full phase tracking.

---

## Installation

```bash
# Install GSD skill set into Claude Code, OpenCode, or other supported tools
npx @opengsd/get-shit-done-redux@latest

# Verify
claude /gsd-help    # in Claude Code
```

GSD supports 15+ AI coding runtimes. For non-Claude tools, see the GSD README for tool-specific setup.

---

## From Spec to Managed Project

### Step 1: Write the spec

Use devskills `/spec` skill to produce a structured `SPEC.md`:

```
/spec
```

This prompts you for the problem, constraints, and acceptance criteria, then writes `SPEC.md`.

### Step 2: Initialize GSD

```
/gsd-new-project
```

GSD reads your SPEC.md (if present) and creates:

```
.planning/
├── ROADMAP.md          # milestone and phase breakdown
├── PROJECT.md          # vision, users, constraints
└── state/              # session checksums and progress
```

It will ask a series of questions to flesh out the roadmap. Answer them — this is the "discuss" work that prevents mid-execution pivots.

### Step 3: Discuss each phase before planning

Before planning any individual phase, run:

```
/gsd-discuss-phase
```

This captures:
- Implementation decisions (which library, which pattern)
- Scope boundaries (what this phase does NOT do)
- Dependencies on other phases
- Risk factors

Answers are written to `.planning/` so they persist across sessions.

### Step 4: Plan the phase

```
/gsd-plan-phase
```

GSD researches the implementation and produces `.planning/PLAN.md` with:
- Numbered tasks in dependency order
- File-level targets (which files to create or modify)
- Risk notes

Review PLAN.md before executing. Change anything you disagree with.

### Step 5: Execute

```
/gsd-execute-phase
```

GSD runs PLAN.md tasks using parallel sub-agents. Each sub-agent gets a fresh context window scoped to its task. The main session stays at 30-40% context utilization.

Tiger Style is applied during execution. Language profile conventions (from the `<!-- profile: <lang> -->` block in your `AGENTS.md`) are applied automatically if set.

### Step 6: Verify

```
/gsd-verify-work
```

GSD checks the execution output against the acceptance criteria in SPEC.md and the phase goals in ROADMAP.md. Failures generate a fix plan — re-run `/gsd-execute-phase` to apply.

### Step 7: Ship

```
/gsd-ship
```

Creates a pull request with:
- Summary from SPEC.md
- Phase description from ROADMAP.md
- Acceptance criteria results from VERIFICATION.md

---

## Phase Structure

GSD organizes work into phases within a milestone. A phase is a coherent unit of work with a single goal.

```
ROADMAP.md:

## Milestone 1: MVP

### Phase 1: Data layer
Goal: working database schema and repository layer
Status: complete

### Phase 2: API
Goal: REST endpoints over the data layer
Status: in progress

### Phase 3: Frontend
Goal: React UI consuming the API
Status: planned
```

Each phase runs through discuss → plan → execute → verify → ship independently. You do not need to plan all phases upfront — plan each phase just before executing it.

---

## Multi-Session Work

GSD is designed for sessions that span days or weeks. The `.planning/` directory is the memory:

```bash
# Start a new session — GSD reads state from .planning/
/gsd-resume-work

# Check status without entering a session
cat .planning/ROADMAP.md
```

Commit `.planning/` to git. It is the canonical record of decisions and progress.

---

## Combining devskills with GSD

| Scenario | Command sequence |
|----------|-----------------|
| New project | `/spec` → `/gsd-new-project` → `/gsd-discuss-phase` → `/gsd-plan-phase` → `/gsd-execute-phase` |
| Resume work | `/gsd-resume-work` → `/workflow status` |
| Review after execute | `/go-review` or `/rust-review` → `/gsd-verify-work` |
| Compress long docs | `/tldt` before feeding to GSD discuss |
| Reduce response noise | `/caveman-lite` or `/caveman-ultra` during iterative execution |
| Style enforcement | `/tiger-style` at session start |
| Token savings on CLI ops | `rtk` wraps git/build/test commands automatically (60–90% reduction) |
| Language conventions | set a profile via `setup.sh --lang=<lang>`; `/spec` adds a Technical Profile section, execution applies the idioms, and the matching review skill (`/go-review`, `/ts-review`, `/rust-review`) runs at verify |

---

## Context Budget

GSD keeps context lean. For best results:

- Use `/tldt` to summarize long files before pasting them into context
- Use `/caveman-lite` during planning and discuss phases
- Let GSD sub-agents handle heavy research — do not do the research in the main session
- Run `/gsd-execute-phase` for implementation — do not implement inline in the planning session

---

## Troubleshooting

**GSD commands not found:**
```bash
npx @opengsd/get-shit-done-redux@latest   # reinstall
```

**Context too full mid-session:**
```
/gsd-pause-work    # saves state and creates handoff notes
```
Start a new session and `/gsd-resume-work`.

**Execution output doesn't match spec:**
```
/gsd-verify-work   # generates fix plan
/gsd-execute-phase # applies fix plan
```

**Lost track of phase state:**
```bash
cat .planning/ROADMAP.md
cat .planning/PLAN.md
```
