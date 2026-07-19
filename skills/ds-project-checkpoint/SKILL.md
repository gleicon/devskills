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
<one line — where the code stands>

# next
<one line — the next action on the code>

# settled
- <one line — a call that is made, phrased as what not to redo>

# hazards
- <scope>: <one line — what breaks if you edit it>
```

## Process

1. Read `.project/state.md`.
2. **List candidates before judging any of them.** Sweep everything durable in front of you — the conversation, the files you read, the diff — and enumerate what a session arriving without it would get wrong. A filter applied while you recall cuts what you never listed.
3. Overwrite `# now` and `# next` — one line each.
4. Append the candidates that survive the rules as `# settled` and `# hazards` lines. Delete any line this session reversed, and append its replacement.
5. Report every line added or removed, verbatim — then every candidate you cut, one line each, naming the rule that cut it.

## Rules

- **One line per entry.** No second sentence, no rationale, no prose. The reasoning lives in the conversation and the commit — this file is state an agent loads, not a record a human reads.
- **A `# settled` line must forbid something.** The test is consequence, not novelty: would a session that didn't know this do the wrong thing? A constraint nobody has discussed in months — no telemetry, no third-party scripts, deps capped at four — is the most load-bearing kind of entry and the easiest to mistake for background. What goes nowhere is what forbids nothing.
- **A `# hazards` line needs both halves**: a scope (path, glob, or `repo-wide`) and what breaks. Missing either, it isn't a hazard.
- **A reversal replaces the line** — delete the old, append the new. No supersession markers: an entry short enough to delete needs no pointer to its replacement.
- **Never write anything derivable or perishable.** A branch, a SHA, a count, "tests pass", "tree is clean" — these are read from the repo when needed. Frozen here they are wrong by the next commit, including one this session makes.
- **`# now` and `# next` describe the code**, never this toolchain. A session spent on tooling — migrating `.project/`, editing skills, reorganising docs — changes neither: `# now` stays whatever was true of the project before it started, and "nothing in flight" is a complete answer. The next session is picking up the project, not your last hour.
- **This skill writes one file.** Not `map.md`, not `roadmap.md`, not `config.md` — each has its own writer, and `config.md` is the user's standing instructions to you, never yours to edit. Work products outside `.project/` (`SPEC.md`, `GRILL.md`, `EXPLORE.md`) are not yours either: distil what now constrains future work into `# settled`, and leave the artifact alone.

- **The cut list is output, never content.** It goes in the report so a wrong cut is catchable, and never into `state.md`. Writing too much is visible and someone complains; writing too little looks exactly like a clean checkpoint.

## Output

The lines added or removed, verbatim, then the candidates cut and the rule that cut each.
