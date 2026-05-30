Find the root cause of a failure with the scientific method — reproduce, isolate, fix, then prove it.

When invoked, debug a specific failure by investigation, not guesswork. The goal is the *cause*, not a symptom that happens to go quiet. Change one thing at a time and let evidence — not intuition — drive each step. End by proving the fix with `/verify-this`.

This is the lightweight, stateless counterpart to confirmation: `/debug` finds and fixes; `/verify-this` proves the fix held.

## Arguments

Treat the argument as the failure to chase: a failing test, an error message, a stack trace, or a description of wrong behavior. If no concrete symptom is given, ask for one — don't debug a vague "it's broken".

## Process

1. **Reproduce first.** Get a deterministic, minimal repro before touching any code. If you can't reproduce it, that *is* the first problem — gather logs, inputs, and conditions until you can. No repro, no fix.
2. **Capture the failing baseline** — the exact observable: error text, wrong output, transcript. This is what "fixed" will be measured against.
3. **Form hypotheses, ranked** by likelihood × cheapness to test. Write them down.
4. **Test one hypothesis at a time** with the cheapest probe that can *disprove* it — a log line, a breakpoint, a bisect, a narrowed input. Instrument to learn; don't change code "to see if it helps".
5. **Narrow to the root cause** — the specific line or condition that produces the failure, *and why*. State it explicitly. "The error stopped" is not a root cause.
6. **Apply the minimal fix** that addresses the cause. Surgical only — no drive-by refactors or unrelated cleanup (see `/deslop`).
7. **Prove it.** Re-run the repro: the failure is gone. Then hand to `/verify-this "<the repro now passes / behavior X now holds>"` for a baseline-vs-treatment record.

## Discipline

- One change at a time. Five changes that "work" teach you nothing and leave landmines.
- Reproduce before fixing; confirm by reproducing after. A fix you can't demonstrate isn't a fix.
- Find the cause, not a symptom. Silencing the error is not the same as resolving it.
- Evidence over intuition. When they conflict, instrument until one wins.
- Don't guess-and-check against the user — narrow it yourself before asking.
- On a dead end, widen the repro or question an assumption you skipped; don't keep poking the same spot.
- Keep a short trail of what you've ruled out, so you (or a fresh agent) don't re-test it.

## Output

- **Root cause** — stated plainly, anchored to `file:line`, with *why* it fails.
- **Fix** — what changed and how it addresses the cause, not the symptom.
- **Evidence** — the before/after, and the `/verify-this` claim that locks it in.
- **If unresolved** — the ranked hypotheses tested, what each ruled out, and the next probe to run.
