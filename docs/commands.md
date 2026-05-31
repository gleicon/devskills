# Command Reference

Every devskills command is a single prompt file invoked as `/<name>` in Claude Code, OpenCode, or Cursor (and as `/prompts:<name>` in OpenAI Codex). This is the reference: what each one does, its arguments, and when to reach for it. For worked, multi-step workflows see [recipes.md](recipes.md).

Commands come in two shapes:

- **Modes** stay active for the rest of the session until you turn them off (`/ds-tiger-style-mode`, `/ds-ui-mode`, `/caveman-*`). They change *how* the agent works.
- **Actions** run once and finish (`/ds-spec`, `/ds-code-quality-review`, `/ds-handoff`, …). They produce an output and return.

Most commands need no external tooling. The only GSD-coupled command is `/ds-workflow`; everything else stands alone.

---

## Specs & planning

### `/ds-spec` — action

Turn a rough description into a structured specification (the WHAT, not the HOW).

- **Args:** an optional description. With one, it proceeds directly; without, it asks three focused questions (primary user action, what success looks like, hard constraints) then writes the spec.
- **Output:** `SPEC.md` in the current directory (or `.planning/SPEC.md` if that exists), shown inline. Sections: Problem, Scope, Users, Functional/Non-Functional Requirements, Interfaces, Constraints, Acceptance Criteria, Open Questions.
- **Reach for it when:** you have an idea and want a verifiable contract before any code.

### `/ds-explore` — action

Surface candidate approaches to a problem with their trade-offs — suggests, never decides, never implements. Reads `.project/` state and decisions (and the code) so options respect reality; can do bounded, cited web research with `--web` (off by default — and if local context is too thin for good options, it says so and suggests `--web` rather than guessing). Writes a scratch `.project/EXPLORE.md` (or a temp path) and lists the open questions the choice hinges on.

- **Args:** a problem/question; `--web` to opt into web research.
- **Reach for it when:** facing a "how should I build this?" fork and you want options laid out before deciding. It's the upstream of `/ds-grill-me`: explore generates the options, `/ds-grill-me` walks you through choosing.

### `/ds-grill-me` — action

Interview you relentlessly about a plan or design until you share the same understanding. One question at a time, each with a recommended answer; explores the codebase instead of asking when it can.

- **Args:** `--record` appends every resolved decision to `DECISIONS.md` (question, answer, one-line rationale).
- **Output:** a resolved-plan summary once no decision branches remain; the `DECISIONS.md` path if `--record` was used.
- **Reach for it when:** a plan feels under-specified, or you want to pressure-test a design (including the approach in a draft PR) before committing to it.
- **More:** `/ds-grill-me` is unusually versatile — see [grill-me.md](grill-me.md) for a full menu of uses (requirements discovery, design/refactor/architecture, domain terminology, non-coding decisions).

### `/ds-workflow` — action *(uses GSD)*

Spec-to-ship orchestration backed by GSD (Get Shit Done). Requires GSD installed separately. If you are not using GSD, use the `/project-*` family below instead.

---

## Project memory (`.project/`)

A minimal, file-backed alternative to GSD — persistent description, plan, and state in plain markdown under `.project/`, so any session is safe to `/clear` or end. These are *scribes, not pilots*: they record what you decide, never steer architecture. Walkthrough: [project-workflow.md](project-workflow.md). Worked use cases: [project-recipes.md](project-recipes.md).

### `/ds-project-map` — action

Scan the repo and write/refresh `.project/PROJECT.md` (overview, stack, repo map, constraints). Facts only — describes what exists. Run once at start; re-run when the repo drifts.

### `/ds-project-plan` — action

Turn input into an ordered task checklist in `.project/PLAN.md` (`## Roadmap`). Input can be a goal, a `SPEC.md`, or pasted output from another command (e.g. `/ds-code-quality-review` findings). Sequences and scopes; does not choose architecture.

### `/ds-project-checkpoint` — action

Update `.project/PLAN.md`'s `## Now` (state, next, open questions) and roadmap statuses. Run before `/clear` or end of session. `--handoff` also writes a richer `.project/handoff.md`.

### `/ds-project-resume` — action

Read `.project/PLAN.md` (+ `PROJECT.md`) and report where to pick up. Loads `handoff.md` only if it's newer than the plan (by file time — no git dependency, so `.project/` can be git-ignored), else flags it stale. Read-only.

---

## Engineering modes

### `/ds-tiger-style-mode` — mode

Activate [Tiger Style](https://tigerstyle.dev/) engineering constraints for the session: safety, explicitness, bounded everything, assertions, no silent fallbacks. Active during all review commands too.

- **Reach for it when:** starting focused implementation work you want held to a strict bar.

### `/ds-ui-mode` — mode

UI mode, framework-agnostic (React/Svelte/Vue/Solid/vanilla, any runtime). Two halves: **engineering** (one-responsibility components, minimal co-located state, derive-don't-store, explicit loading/error/empty/success, typed boundaries, stale-response cancellation) and **design craft** (type scale + spacing tokens, deliberate visual hierarchy, intentional motion — encoding constraints to escape the generic AI look). Plus accessibility as a requirement and Core Web Vitals targets (LCP/INP/CLS). Adapts to the project's existing stack rather than imposing libraries.

- **Reach for it when:** working on components, UI state, API integration, or styling.

---

## Test & build

### `/ds-tdd-mode` — mode

Drive implementation test-first, one **vertical** slice at a time (one test → minimal implementation → repeat). Refuses horizontal slicing (all tests up front). Tests exercise observable behavior through the public interface, never internals.

- **Output:** per slice, the test then the implementation; reports which behaviors remain untested.
- **Reach for it when:** building a feature you want anchored to real, refactor-survivable behavior.

### `/ds-test-mode` — mode

Pragmatic testing mode: as you build normally, ensure the code that matters gets well tested — without `/ds-tdd-mode`'s test-first ceremony. Tests by **risk, not rule** (core logic, edge cases, a regression test for every bug fixed; skips trivia), insists on behavior-through-the-interface tests that survive refactors, and refuses design-locking tests (mock-the-world, asserting internals, snapshot-everything). Coverage is a side effect, never the goal.

- **Reach for it when:** doing normal implementation work you want kept honestly tested. The alongside-work sibling of `/ds-tdd-mode` (test-first); audit existing tests with `/ds-test-quality-review`.

---

## Quality & cleanup

### `/ds-code-quality-review` — action

Extremely strict maintainability audit: abstraction quality, file sprawl (the 1k-line smell), spaghetti-condition growth. Ambitiously hunts "code judo" — restructurings that delete whole categories of complexity while preserving behavior.

- **Args:** treated as review **scope** (files or directories). With no scope, reviews the changed files on the current branch. Freeform scope ("the whole codebase") is interpreted reasonably.
- **Output:** prioritized findings anchored to `file:line`, with an approval verdict.
- **Reach for it when:** before merging non-trivial work, or auditing an area you suspect is decaying.

### `/ds-doc-quality-review` — action

Strict documentation audit governed by one principle — **docs earn their length** (readers skim, they don't read). Hunts wrong docs (drifted from the code) and bloated docs (true, but nobody reads them) with equal energy. Verifies mechanically: resolves links, recounts claimed counts, runs example commands where safe.

- **Args:** treated as scope (files, directories, globs); defaults to docs changed on the current branch. `--comments` also audits inline code comments (off by default).
- **Output:** prioritized findings anchored to `file:line` with a suggested fix — accuracy/drift first, then dead links and wrong counts, missing docs, bloat-to-cut, clarity. Changes nothing.
- **Reach for it when:** before a docs PR, after the code outgrew its README, or when the docs feel long and unread.

### `/ds-test-quality-review` — action

Strict test-suite audit governed by one principle — **test what matters, and test it well — not coverage.** Hunts critical/core code that's under-tested (the bug waiting to ship) and tests that are bad (green but worthless, or design-locking and worse than nothing) with equal energy. Checks edge/failure-mode coverage on risky logic, behavior-vs-implementation quality, and flags tests coupled to internals for rewrite-or-delete. Explicitly rejects coverage-chasing.

- **Args:** treated as scope (files, directories, globs); defaults to tests covering code changed on the current branch.
- **Output:** prioritized findings anchored to `file:line` — untested critical code first, then missing edge cases, design-locking tests, weak tests, bloat-to-cut. Changes nothing.
- **Reach for it when:** before merging logic-heavy work, or when a suite is green but you don't trust it. The audit counterpart to the `/ds-test-mode` mode and `/ds-tdd-mode`.

### `/ds-ui-quality-review` — action

Strict UI audit governed by one principle — **a UI is judged on both halves: it works and it's crafted.** Framework-agnostic. Hunts engineering correctness (missing/broken async states — especially **empty** — fetch waterfalls, uncancelled stale responses, state that should be derived, index-as-key), accessibility barriers (non-semantic controls, keyboard/focus gaps, contrast, missing labels/live regions), Core Web Vitals (layout shift, INP, oversized critical path), and design craft (generic-AI defaults, flat hierarchy, unsystematized type/spacing).

- **Args:** treated as scope (files, directories, globs); defaults to UI code changed on the current branch.
- **Output:** prioritized findings anchored to `file:line` — engineering correctness first, then a11y, performance, design craft. Each names the concrete failure, not "improve a11y." Changes nothing.
- **Reach for it when:** before merging UI work, or auditing an interface you suspect is broken-on-edges, inaccessible, slow, or generic. The audit counterpart to the `/ds-ui-mode` mode.

### `/ds-deslop` — action

Strip AI-generated slop from the branch and align it with the surrounding code. The test is "code an experienced engineer in this language and codebase wouldn't write" — judged against the language's idioms, not a fixed syntax list. Targets narrating comments, defensive overkill abnormal for a trusted path, type escape hatches used only to dodge the checker, needless nesting, and speculative scaffolding.

- **Args:** treated as scope (files or directories); defaults to the branch diff against main.
- **Output:** the edits applied, plus a 1–3 sentence summary. Behavior preserved.
- **Reach for it when:** right after generating a batch of code, before review. Cheaper and narrower than `/ds-code-quality-review`.

---

## Quality Gate

### `/ds-quality-gate-mode` — mode

Six-pass review pipeline for a feature branch or scoped path. Run in order; implement accepted findings between passes before proceeding.

Pipeline: `/ds-deslop` → `/ds-test-quality-review` → `/ds-security-review` → `/ds-bug-review` → `/ds-code-quality-review` → `/ds-doc-quality-review`

After each pass: shows findings for that pass, asks "accept all / reject all / skip N", implements accepted ones, then moves to the next pass. Reports a one-paragraph summary at the end.

- **Args:** optional path scope (applies to every pass). None → branch diff.
- **Toggle:** `/ds-quality-gate-mode` to activate, `/ds-quality-gate-mode off` or "stop quality gate" to deactivate.
- **Reach for it when:** finishing a feature branch before opening a PR; each pass answers a different question so running all six is not redundant.

---

## Reviews

The review commands are a **layered gate, not competing alternatives** — cheapest and narrowest first, deepest last: `/ds-deslop` (noise) → `/ds-bug-review` (correctness) → `/ds-security-review` (exploitability) → the language review (idioms) → `/ds-code-quality-review` (structure). Each answers a different question, so running several on the same code isn't redundant. The full pre-PR sequence is in [recipes.md](recipes.md#a-pre-pr-quality-gate).

### `/ds-bug-review` — action

Language-agnostic **correctness** audit — the bug-hunting pass. Asks one thing: *will this misbehave at runtime?* Hunts logic errors, null/absent-value derefs, swallowed errors and half-done failure paths, resource leaks, races (TOCTOU, lock ordering), boundary/overflow mistakes, and contract misuse. Distinct from `/ds-code-quality-review` (which is maintainability, not bugs) — the same split the harness draws between cleanup and correctness.

- **Args:** treated as scope (files, directories, globs); defaults to code changed on the current branch.
- **Output:** prioritized findings anchored to `file:line` — critical (data loss / reachable crash) first, then likely-wrong, then edge-case. Each names **the exact condition that triggers it** plus the fix and a confidence note. Real defects only; no theoretical nulls. Changes nothing.
- **Reach for it when:** before merging logic-heavy work, or on any code outside go/ts/rust where there's no language review. Confirmed findings hand off to `/ds-debug` (root-cause) and `/ds-verify-this` (prove the fix).

### `/ds-security-review` — action

Language-agnostic **security** audit — the portable counterpart to the per-language Security sections. Traces untrusted data from entry to dangerous sink: injection (SQL/command/path/SSRF/template), output handling (XSS, unsafe deserialization), broken access control (IDOR, privilege escalation), secrets and weak crypto, sensitive-data exposure, mass assignment / unsafe upload / DoS, and transport/config gaps.

- **Args:** treated as scope (files, directories, globs); defaults to code changed on the current branch.
- **Output:** prioritized findings anchored to `file:line` — critical (code exec / breach / auth bypass) → high → hardening. Each **describes the attack** (input → sink) and the fix. Exploitable over theoretical. Changes nothing.
- **Reach for it when:** any change that touches input handling, auth, secrets, or external I/O — and as a pre-PR gate. The deeper language-specific checks live in `/go·ts·rust·python-review`.

### `/ds-go-review` · `/ds-ts-review` · `/ds-rust-review` · `/ds-python-review` — action

Language-specific review passes.

- **`/ds-go-review`** — idiomatic Go + security + Tiger Style.
- **`/ds-ts-review`** — TypeScript/Workers: strict mode, React, Cloudflare.
- **`/ds-rust-review`** — cargo geiger, `unsafe`/panic counts, clippy, audit, Tiger Style + security.
- **`/ds-python-review`** — Python idioms, `mypy --strict` typing, security, Tiger Style.
- **Args:** `--no-tiger` skips the Tiger Style section (all of them).
- **Reach for it when:** reviewing code in that language, or as a pre-PR gate.

---

## Debugging & verification

### `/ds-debug` — action

Find the root cause of a failure with the scientific method, then prove the fix. Reproduce-first, one hypothesis at a time, evidence over intuition — disciplined against the usual AI failure modes (thrashing, changing five things at once, silencing the symptom). Lightweight and stateless — the agent-agnostic counterpart to GSD's heavier `/gsd:debug`.

- **Args:** the failure to chase — a failing test, error, stack trace, or wrong-behavior description. Refuses a vague "it's broken".
- **Output:** root cause (`file:line` + why), the minimal fix, before/after evidence, and the `/ds-verify-this` claim to lock it in.
- **Reach for it when:** something fails and you don't yet know *why*. Pairs with `/ds-verify-this` — `/ds-debug` finds and fixes, `/ds-verify-this` proves it held.

### `/ds-verify-this` — action

Prove or disprove a **falsifiable** claim with fresh local evidence — not a recap. Restates the claim with a metric and threshold, captures a baseline (old state) and treatment (changed state) with the same command/env, compares raw artifacts, and returns exactly one verdict: `VERIFIED`, `NOT VERIFIED`, or `INCONCLUSIVE`. **No CI required** — all surfaces are local (tests, repro scripts, `tmux`/PTY transcripts, local HTTP, screenshots, profiles).

- **Args:** the claim to verify. Refuses vague claims ("the code is cleaner") — give it something measurable.
- **Reach for it when:** "did this actually fix it?", a bugfix needs a before/after repro, or a perf/memory/UI claim needs measurement.

---

## Understanding

### `/ds-zoom-out` — action

Step up one layer of abstraction and map how an area fits the bigger picture: its responsibility, neighbouring modules, callers, and boundaries. No line-by-line read, no code dumps.

- **Reach for it when:** entering unfamiliar code, or before planning a change in an area you don't hold in your head.

---

## Context & continuity

### `/ds-handoff` — action

Compact the current conversation into a handoff document so a fresh agent can continue without re-reading the transcript.

- **Args:** optional — treated as what the next session should focus on.
- **Output:** writes `handoff.md` to a fresh `mktemp -d` and returns the path. Records goal, done, remaining, key decisions, open questions; references existing artifacts by path rather than duplicating them.
- **Reach for it when:** the context window is filling, you're switching machines/sessions, or pausing mid-task.

### `/ds-tldt` — action

Extractive summarization — selects verbatim sentences, no paraphrasing or generation (so no hallucination). Uses the `tldt` CLI when present, else selects manually. Flags prompt-injection patterns in the source.

- **Args:** `/ds-tldt` (last large block of text), `/ds-tldt <file>`, or `/ds-tldt <url>`.
- **Reach for it when:** compressing a long doc/page before adding it to context.

---

## Response compression

### `/ds-caveman-lite-mode` · `/ds-caveman-ultra-mode` — mode

Compress the agent's prose to save tokens. **Lite** drops articles/filler/hedging (~25–35% reduction, full explanatory value). **Ultra** restructures into fragments and notation (~75–85% reduction). Both keep code, commit messages, PR bodies, numbers, and security/irreversible-action warnings written in full. Deactivate with "stop caveman" or "normal mode".

- **Reach for it when:** long iterative sessions where prose is overhead. Use Ultra only when you'll ask for elaboration if you need it.

---

## Meta

### `/ds-write-a-command` — action

Author a new devskills command in the repo's conventions, written to `commands/` (install.sh copies it to Claude Code, OpenCode, and Codex), registered in the README table and `docs/commands.md`. Knows both archetypes (action vs. mode) and enforces the one-job-per-command rule.

- **Reach for it when:** a workflow you repeat by hand should become a command.
