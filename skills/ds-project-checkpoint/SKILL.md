---
name: ds-project-checkpoint
description: "Write the session's durable state to .project/state.md so the next session can pick it up."
disable-model-invocation: true
---

When invoked, record what the next session needs and nothing else. Run it at any point and at the end of a session. It is the only writer of `.project/state.md`, and the counterpart to `/ds-project-resume`.

## The file

Create `.project/state.md` if absent. It has these four sections and no others:

```
# now
<one line — what is in flight>

# next
<one line — the next action on the code>

# settled
- <one line — a call that is made, phrased as what not to redo>

# hazards
- <scope>: <one line — what breaks if you edit it>
```

## Process

1. Read `.project/state.md`.
2. Overwrite `# now` and `# next` — one line each.
3. Append new `# settled` and `# hazards` lines. Delete any line this session reversed, and append its replacement.
4. Report every line added or removed, verbatim.

## Rules

- **One line per entry.** No second sentence, no rationale, no prose. The reasoning lives in the conversation and the commit — this file is state an agent loads, not a record a human reads.
- **A `# settled` line must forbid something.** If it doesn't stop a future session from redoing or re-deciding something, it isn't state. Background, and things that are merely true, go nowhere in this file.
- **A `# hazards` line needs both halves**: a scope (path, glob, or `repo-wide`) and what breaks. Missing either, it isn't a hazard.
- **A reversal replaces the line** — delete the old, append the new. No supersession markers: an entry short enough to delete needs no pointer to its replacement.
- **Never write anything derivable or perishable.** A branch, a SHA, a count, "tests pass", "tree is clean" — these are read from the repo when needed. Frozen here they are wrong by the next commit, including one this session makes.
- **`# next` names work on the code**, never a `/ds-*` invocation. Running this toolchain is not an action the next session needs to be told.
- **This skill writes one file.** Not `map.md`, not `roadmap.md`, not `config.md` — each has its own writer, and `config.md` is the user's standing instructions to you, never yours to edit. Work products outside `.project/` (`SPEC.md`, `GRILL.md`, `EXPLORE.md`) are not yours either: distil what now constrains future work into `# settled`, and leave the artifact alone.

## Output

The lines added or removed, verbatim. Nothing else.
