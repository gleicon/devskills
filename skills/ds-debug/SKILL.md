---
name: ds-debug
description: "Find the root cause of a failure with the scientific method — reproduce, isolate, fix, then prove it."
disable-model-invocation: true
---

When invoked, debug a specific failure by investigation, not guesswork. The goal is the *cause*, not a symptom that happens to go quiet. Change one thing at a time and let evidence — not intuition — drive each step. End by proving the fix with `/ds-verify-this`.

This is the lightweight, stateless counterpart to confirmation: `/ds-debug` finds and fixes; `/ds-verify-this` proves the fix held.

## Arguments

Treat the argument as the failure to chase: a failing test, an error message, a stack trace, or a description of wrong behavior. If no concrete symptom is given, ask for one — don't debug a vague "it's broken".

## Process

1. **Build the feedback loop first.** Before forming any theory, get a fast, repeatable pass/fail signal that goes red on the bug — this is the whole skill; everything after is mechanical. Reach for the cheapest loop that reproduces it: a failing test at the right seam, a `curl`/CLI + fixture diff, a replayed trace, a throwaway harness, a bisection or old-vs-new differential run, a property/fuzz loop, a headless-browser script. **No red-capable loop, no hypothesis.**
   - *Flaky or intermittent?* Aim for a higher reproduction **rate**, not a clean repro — loop the trigger 100×, parallelize, add load, narrow the timing window. A 50%-flake is debuggable; 1% isn't — keep raising the rate.
   - *Can't build a loop at all?* Stop and ask for what's missing: environment access, a captured artifact (HAR, log dump, core dump, timestamped recording), or permission for temporary instrumentation. Don't hypothesize without a loop.
2. **Capture the failing baseline** — the exact observable: error text, wrong output, transcript. This is what "fixed" will be measured against.
3. **Form hypotheses, ranked** by likelihood × cheapness to test. Each must carry a falsifiable prediction — *if this is the cause, changing X makes the failure vanish (or Y makes it worse)*; a hypothesis with no prediction is a vibe — drop it. Write them down; then, **once**, you may show the ranked list to the user for a domain-knowledge re-rank ("we just deployed #3") — proceed without blocking if they're away.
4. **Test one hypothesis at a time** with the cheapest probe that can *disprove* it — a breakpoint or REPL beats ten log lines; a bisect; a narrowed input. Instrument to learn, not to change code "to see if it helps" — and tag any debug logging with a unique prefix (`[DEBUG-a4f2]`) so cleanup is a single grep.
5. **Narrow to the root cause** — the specific line or condition that produces the failure, *and why*. State it explicitly. "The error stopped" is not a root cause.
6. **Apply the minimal fix** that addresses the cause. Surgical only — no drive-by refactors or unrelated cleanup (see `/ds-deslop`).
7. **Prove it.** Re-run the repro: the failure is gone. Then hand to `/ds-verify-this "<the repro now passes / behavior X now holds>"` for a baseline-vs-treatment record.
8. **Close the loop.** Delete the instrumentation and throwaway harnesses (grep your prefix). Record the winning hypothesis in the commit/PR so the next debugger inherits it, and ask "what would have prevented this?" — route an architectural answer to `/ds-architecture-plan`.

## Discipline

- One change at a time. Five changes that "work" teach you nothing and leave landmines.
- Reproduce before fixing; confirm by reproducing after. A fix you can't demonstrate isn't a fix.
- Find the cause, not a symptom. Silencing the error is not the same as resolving it.
- Evidence over intuition. When they conflict, instrument until one wins.
- Perf regression? Logs mislead — establish a baseline measurement (timing, profiler, query plan) first, then bisect. Measure, then fix.
- Do your own narrowing — don't guess-and-check by polling the user step by step. The one-shot hypothesis re-rank (step 3) is the single allowed checkpoint, not a running dialogue.
- On a dead end, widen the repro or question an assumption you skipped; don't keep poking the same spot.
- Keep a short trail of what you've ruled out, so you (or a fresh agent) don't re-test it.

## Output

- **Root cause** — stated plainly, anchored to `file:line`, with *why* it fails.
- **Fix** — what changed and how it addresses the cause, not the symptom.
- **Evidence** — the before/after, and the `/ds-verify-this` claim that locks it in.
- **If unresolved** — the ranked hypotheses tested, what each ruled out, and the next probe to run.
