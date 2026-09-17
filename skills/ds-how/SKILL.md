---
name: ds-how
description: "Explain how a subsystem works as a mental model — overview, key concepts, the runtime flow, where things live, gotchas. `/ds-zoom-out` places an area among its neighbours before a change; this walks the machinery inside it."
disable-model-invocation: true
---

When invoked, answer "how does X work?" at the level a senior engineer needs to start working in a subsystem: enough to hold a working mental model, not annotated source. It explains and judges nothing. For an area's place among its neighbours use `/ds-zoom-out`; for a new teammate's arrival brief use `/ds-onboarding`.

## Arguments

The question: a subsystem, a runtime flow, or a placement question ("where should this live", "which package owns this", "is this the right layer"). If the scope is ambiguous, state your reading and proceed — the user can redirect.

## Process

1. **Size the question.** One module, one utility, or one function: explore and explain in a single pass. A subsystem spanning several files or services, a cross-cutting feature, or a whole-architecture question: fan out first. When in doubt, take the single pass.
2. **Fan out (large questions only).** Split the question into 2 to 4 angles, each a distinct slice of the subsystem, and hand each to a read-only explorer in one message with the brief below. Explorers return facts; you write the prose.
3. **Read the code, not the names.** If `.project/map.md` exists, start from it. Find the entry point — what triggers the behaviour — then follow the call chain and the data that moves through it, read the definitions of the central types and interfaces, and find where the subsystem meets its neighbours. A part you cannot trace is reported as a gap, never filled in by guessing.
4. **Synthesize.** Where explorers overlap or disagree, settle it by checking the code yourself. Write the explanation in the shape below.

## Explorer brief

Give each explorer the question and its angle, then:

> Gather facts, not prose; another agent writes the explanation. Read the implementation — do not infer behaviour from names. Find the entry point, trace the flow and the data between steps, read the central types, find the boundaries with other subsystems, and note anything surprising or easy to misread. Return: components found (name, path, one line each), the flow step by step, boundaries (inputs, outputs, neighbours), non-obvious things, open questions you could not trace, and every file you read.

## Output

Use these headings, in this order. Drop one only when it has nothing to say — never Overview or How it works.

- `## Overview` — one or two paragraphs: what it is, what it does, why it exists. Enough to decide whether to read on.
- `## Key concepts` — the types, services and abstractions needed to follow the rest. Brief definitions, not a catalogue.
- `## How it works` — the longest section. What triggers it, what happens step by step, where data goes, the decision points. Prose that names files and functions so the reader knows where to look; a snippet only when a point needs it. A diagram (mermaid or ASCII) only when components talk to each other or data transforms through stages and prose would blur it.
- `## Where things live` — the files and directories someone needs to start working here.
- `## Gotchas` — surprising behaviour, historical artifacts, what a newcomer would misread. Omit when there is nothing.

Concrete over abstract: "`Server.Serve` calls `router.Dispatch`", not "the server delegates to the router". Explain why something is complex instead of describing the complexity; do not pad the simple. Gaps the exploration could not close are stated, not hidden.
