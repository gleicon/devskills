# Command Reference

Every devskills command is a single prompt file invoked as `/<name>` in Claude Code, OpenCode, or Cursor. This is the reference: what each one does, its arguments, and when to reach for it. For worked, multi-step workflows see [recipes.md](recipes.md).

Commands come in two shapes:

- **Modes** stay active for the rest of the session until you turn them off (`/tiger-style`, `/frontend`, `/caveman-*`). They change *how* the agent works.
- **Actions** run once and finish (`/spec`, `/code-quality-review`, `/handoff`, …). They produce an output and return.

Most commands need no external tooling. The only GSD-coupled command is `/workflow`; everything else stands alone.

---

## Specs & planning

### `/spec` — action

Turn a rough description into a structured specification (the WHAT, not the HOW).

- **Args:** an optional description. With one, it proceeds directly; without, it asks three focused questions (primary user action, what success looks like, hard constraints) then writes the spec.
- **Output:** `SPEC.md` in the current directory (or `.planning/SPEC.md` if that exists), shown inline. Sections: Problem, Scope, Users, Functional/Non-Functional Requirements, Interfaces, Constraints, Acceptance Criteria, Open Questions.
- **Reach for it when:** you have an idea and want a verifiable contract before any code.

### `/explore` — action

Surface candidate approaches to a problem with their trade-offs — suggests, never decides, never implements. Reads `.project/` state and decisions (and the code) so options respect reality; can do bounded, cited web research with `--web` (off by default — and if local context is too thin for good options, it says so and suggests `--web` rather than guessing). Writes a scratch `.project/EXPLORE.md` (or a temp path) and lists the open questions the choice hinges on.

- **Args:** a problem/question; `--web` to opt into web research.
- **Reach for it when:** facing a "how should I build this?" fork and you want options laid out before deciding. It's the upstream of `/grill-me`: explore generates the options, `/grill-me` walks you through choosing.

### `/grill-me` — action

Interview you relentlessly about a plan or design until you share the same understanding. One question at a time, each with a recommended answer; explores the codebase instead of asking when it can.

- **Args:** `--record` appends every resolved decision to `DECISIONS.md` (question, answer, one-line rationale).
- **Output:** a resolved-plan summary once no decision branches remain; the `DECISIONS.md` path if `--record` was used.
- **Reach for it when:** a plan feels under-specified, or you want to pressure-test a design (including the approach in a draft PR) before committing to it.
- **More:** `/grill-me` is unusually versatile — see [grill-me.md](grill-me.md) for a full menu of uses (requirements discovery, design/refactor/architecture, domain terminology, non-coding decisions).

### `/workflow` — action *(uses GSD)*

Spec-to-ship orchestration backed by GSD (Get Shit Done). Requires GSD installed separately. If you are not using GSD, use the `/project-*` family below instead.

---

## Project memory (`.project/`)

A minimal, file-backed alternative to GSD — persistent description, plan, and state in plain markdown under `.project/`, so any session is safe to `/clear` or end. These are *scribes, not pilots*: they record what you decide, never steer architecture. Full walkthrough: [project-workflow.md](project-workflow.md).

### `/project-map` — action

Scan the repo and write/refresh `.project/PROJECT.md` (overview, stack, repo map, constraints). Facts only — describes what exists. Run once at start; re-run when the repo drifts.

### `/project-plan` — action

Turn input into an ordered task checklist in `.project/PLAN.md` (`## Roadmap`). Input can be a goal, a `SPEC.md`, or pasted output from another command (e.g. `/code-quality-review` findings). Sequences and scopes; does not choose architecture.

### `/project-checkpoint` — action

Update `.project/PLAN.md`'s `## Now` (state, next, open questions) and roadmap statuses. Run before `/clear` or end of session. `--handoff` also writes a richer `.project/handoff.md`.

### `/project-resume` — action

Read `.project/PLAN.md` (+ `PROJECT.md`) and report where to pick up. Loads `handoff.md` only if it's newer than the plan (by file time — no git dependency, so `.project/` can be git-ignored), else flags it stale. Read-only.

---

## Engineering modes

### `/tiger-style` — mode

Activate [Tiger Style](https://tigerstyle.dev/) engineering constraints for the session: safety, explicitness, bounded everything, assertions, no silent fallbacks. Active during all review commands too.

- **Reach for it when:** starting focused implementation work you want held to a strict bar.

### `/frontend` — mode

Frontend task mode: composable pure components, minimal co-located state, typed fetch wrappers, accessibility as a requirement. Covers React/Svelte/Vue/vanilla and Cloudflare Workers / edge patterns.

- **Reach for it when:** working on components, UI state, API integration, or styling.

---

## Test & build

### `/tdd` — action/mode

Drive implementation test-first, one **vertical** slice at a time (one test → minimal implementation → repeat). Refuses horizontal slicing (all tests up front). Tests exercise observable behavior through the public interface, never internals.

- **Output:** per slice, the test then the implementation; reports which behaviors remain untested.
- **Reach for it when:** building a feature you want anchored to real, refactor-survivable behavior.

---

## Quality & cleanup

### `/code-quality-review` — action

Extremely strict maintainability audit: abstraction quality, file sprawl (the 1k-line smell), spaghetti-condition growth. Ambitiously hunts "code judo" — restructurings that delete whole categories of complexity while preserving behavior.

- **Args:** treated as review **scope** (files or directories). With no scope, reviews the changed files on the current branch. Freeform scope ("the whole codebase") is interpreted reasonably.
- **Output:** prioritized findings anchored to `file:line`, with an approval verdict.
- **Reach for it when:** before merging non-trivial work, or auditing an area you suspect is decaying.

### `/deslop` — action

Strip AI-generated slop from the branch and align it with the surrounding code. Targets stray/inconsistent comments, defensive `try/catch` in trusted paths, `any`-casts that only dodge types, and needless nesting.

- **Args:** treated as scope (files or directories); defaults to the branch diff against main.
- **Output:** the edits applied, plus a 1–3 sentence summary. Behavior preserved.
- **Reach for it when:** right after generating a batch of code, before review. Cheaper and narrower than `/code-quality-review`.

---

## Reviews

### `/go-review` · `/ts-review` · `/rust-review` — action

Language-specific review passes.

- **`/go-review`** — idiomatic Go + security + Tiger Style.
- **`/ts-review`** — TypeScript/Workers: strict mode, React, Cloudflare.
- **`/rust-review`** — cargo geiger, `unsafe` counts, clippy, audit.
- **Args:** `--no-tiger` skips the Tiger Style section (Go/TS).
- **Reach for it when:** reviewing code in that language, or as a pre-PR gate.

---

## Verification

### `/verify-this` — action

Prove or disprove a **falsifiable** claim with fresh local evidence — not a recap. Restates the claim with a metric and threshold, captures a baseline (old state) and treatment (changed state) with the same command/env, compares raw artifacts, and returns exactly one verdict: `VERIFIED`, `NOT VERIFIED`, or `INCONCLUSIVE`. **No CI required** — all surfaces are local (tests, repro scripts, `tmux`/PTY transcripts, local HTTP, screenshots, profiles).

- **Args:** the claim to verify. Refuses vague claims ("the code is cleaner") — give it something measurable.
- **Reach for it when:** "did this actually fix it?", a bugfix needs a before/after repro, or a perf/memory/UI claim needs measurement.

---

## Understanding

### `/zoom-out` — action

Step up one layer of abstraction and map how an area fits the bigger picture: its responsibility, neighbouring modules, callers, and boundaries. No line-by-line read, no code dumps.

- **Reach for it when:** entering unfamiliar code, or before planning a change in an area you don't hold in your head.

---

## Context & continuity

### `/handoff` — action

Compact the current conversation into a handoff document so a fresh agent can continue without re-reading the transcript.

- **Args:** optional — treated as what the next session should focus on.
- **Output:** writes `handoff.md` to a fresh `mktemp -d` and returns the path. Records goal, done, remaining, key decisions, open questions; references existing artifacts by path rather than duplicating them.
- **Reach for it when:** the context window is filling, you're switching machines/sessions, or pausing mid-task.

### `/tldt` — action

Extractive summarization — selects verbatim sentences, no paraphrasing or generation (so no hallucination). Uses the `tldt` CLI when present, else selects manually. Flags prompt-injection patterns in the source.

- **Args:** `/tldt` (last large block of text), `/tldt <file>`, or `/tldt <url>`.
- **Reach for it when:** compressing a long doc/page before adding it to context.

---

## Response compression

### `/caveman-lite` · `/caveman-ultra` — mode

Compress the agent's prose to save tokens. **Lite** drops articles/filler/hedging (~25–35% reduction, full explanatory value). **Ultra** restructures into fragments and notation (~75–85% reduction). Both keep code, commit messages, PR bodies, numbers, and security/irreversible-action warnings written in full. Deactivate with "stop caveman" or "normal mode".

- **Reach for it when:** long iterative sessions where prose is overhead. Use Ultra only when you'll ask for elaboration if you need it.

---

## Meta

### `/write-a-skill` — action

Author a new devskills command in the repo's conventions, written byte-identical to both `claude/commands/` and `opencode/commands/`, with the README table updated. Enforces the one-job-per-command rule.

- **Reach for it when:** a workflow you repeat by hand should become a command.
