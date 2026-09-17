---
name: ds-blast-radius
description: "Trace what a change could break beyond the diff, name the one fact that makes it safe, prove that fact by running code, and emit the minimal test that would catch the break. `/ds-bug-review` hunts defects inside the diff; this hunts the breakage outside it."
disable-model-invocation: true
---

When invoked, find what a change breaks somewhere else before it ships. Listing the callers is not the job — a symbol search does that in a second. The job is the breakage a search will not show: a consumer in another language, a stored format, a timing assumption, code three hops downstream. Report-only; it edits nothing.

`/ds-bug-review` asks whether the diff is correct in itself. `/ds-verify-this` proves a claim you already hold. This asks what the diff changes for everything that was not in it, and ends with a proof, not a writeup.

## Arguments

Positional args are the scope: a diff, a commit range, files. With none, take the code changed on the current branch.

## Process

1. **Read the change for what it now does differently**, including the part the diff does not spell out: a symbol renamed, a default removed, a unit or ordering changed, a contract tightened.
2. **Find the one fact it is safe because of.** Most risky-looking changes are safe because of a single fact — "only already-dead entries reach this call", "no consumer reads that field". Name it before enumerating maybes; if it holds, most risks clear at once.
3. **Look where a search stops.** Read the source of the library you call, at the version the project pins. Check what runs when: teardown, microtasks, startup order. Follow what a symbol search misses — the JSON an API emits, a database column, a wire or file format, another language reading the same bytes, a feature flag, fixtures and golden files, scripts. A search that finds nothing is an answer; record it.
4. **Weigh each risk honestly.** A real chance of happening and a real cost if it does, anchored to a `file:line`. Never invent a caller or an API. Risks you checked and cleared go in their own list.
5. **Prove the fact by running code.** A script or test that imports the same code the project ships, calls the exact function in question, and fails loudly when you are wrong. Run it and paste what happened. If it cannot be proven cheaply, write *unproven* — never settle it in prose.
6. **Write the test that would have caught it.** The cheapest one that goes red on the real break, for the reader to keep.

## Evidence ladder

For every fact the change's safety rests on, say how far down this list you got. A writeup that sounds right is worth nothing on its own.

1. You said so.
2. You pointed at the line — a real `file:line` or the library's own source.
3. You walked the bad case step by step and showed it cannot reach.
4. You ran it — real code, loud failure if wrong.
5. You reproduced it in the running system.

Anything short of step 4 is reported as *unproven*.

## Output

- **What it does** — what changed, including the part that is not obvious from the diff.
- **The one fact it is safe because of** — stated, with its ladder step and the proof. Or *unproven*.
- **Risks** — only the real ones. Each names how it breaks, the `file:line`, likelihood and cost, and how to check. Proof pasted for the ones that matter.
- **Cleared** — what you checked and why it is fine.
- **Before you merge** — the minimal test or repro that catches the break, including the script you wrote.
