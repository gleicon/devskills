---
name: ds-project-compact
description: "Compact the .project/ memory files — archive spent decisions and finished roadmap sections, keep the live files to what still governs. Loss-free and approval-gated."
disable-model-invocation: true
---

Compact the `.project/` memory so the live files stay small and current while nothing is lost.

When invoked, sweep `.project/` for material that has stopped governing and move it to a dated archive, leaving each live file as the current contract. The principle is **live stays, spent moves, essence promotes**: an active decision stays in `DECISIONS.md`; a superseded or expired one moves to the archive (its residual truth, if any, re-recorded as a fresh active decision); a finished roadmap section moves out of `PLAN.md`; `## Now` is never touched. Run it when `DECISIONS.md`/`PLAN.md` have grown enough to cost tokens or to let stale entries keep steering. The counterpart to `/ds-project-checkpoint`, which grows these files; compact is what trims them.

## Process

1. If `.project/` is absent, say so and stop — there is nothing to compact.
2. **Classify each `DECISIONS.md` entry** as one of:
   - **active** — still describes a current contract or constraint. Stays in `DECISIONS.md`.
   - **superseded** — a later decision replaced it. Archive it; if it still carries residual truth the replacement doesn't state, re-record that as a fresh active decision. An entry carrying a supersession marker is superseded by definition and needs no judgment; concluding that an *unmarked* entry is superseded is a judgment call, so present it as one.
   - **expired** — described a one-time choice already carried out, no longer governing anything. Archive it.
3. **Classify `PLAN.md` `## Roadmap` sections.** A section whose tasks are all complete (`[x]`) moves to the archive. `## Now` is never modified. A section with open tasks stays.
4. **Offer stale scratch files for deletion.** If `handoff.md` or `EXPLORE.md` exists and is stale (older than `PLAN.md`, or its work is done), offer to delete it — durable findings should already have been routed to `DECISIONS.md`/`PROJECT.md` by checkpoint, so the file itself is disposable.
5. **Get per-item approval.** Show each classification and each archive move one at a time and wait for the nod before writing — same model as checkpoint. Honor a batch approval ("archive them all"). Nothing is written without approval.
6. **Move spent material to the archive.** Append it to `.project/archive/<file>-<date>.md` (e.g. `DECISIONS-2026-07-12.md`), append-only, plain Markdown. Then remove the archived entries/sections from the live file. Every archived item stays readable in the archive — compact never deletes decision content.
7. **Promote essence only on approval.** If an archived entry distills to a standing constraint that belongs in the map, propose it as a single additive line to `PROJECT.md` (the additive-only rule `/ds-project-checkpoint` uses) — never rewrite the map, and only with approval.
8. **Report** what moved where — and be exact about what that means. You sorted entries by whether they still *govern*; you did not check a single one against the code. "The live files hold only what still governs" is a claim you have not tested: a decision can read as perfectly active and describe behaviour the codebase walked away from months ago. Reconciling against source is `/ds-project-verify`'s job — say plainly that compact didn't do it, rather than handing back something that sounds like a clean bill of health.

## Rules

- **Loss-free.** The union of the live files and the archive contains every pre-compact decision entry. Compact relocates; it never discards.
- **`## Now` and `## Watch` are off-limits.** Compact touches `## Roadmap` sections and the decision log — current state and open flags are checkpoint's. A `## Watch` entry that looks spent may be flagging something you can't see from inside the file; leave it.
- **Approval-gated.** Every classification and every reader-affecting write is approved per item (batch approval honored). Default to keeping an entry live when its status is unclear — err toward not archiving.
- **Archive is write-only memory.** `.project/archive/` is append-only and is never read by `/ds-project-resume` — it exists for a human who wants the history, not for orientation.
- **Owned files stay owned.** `SPEC.md` is never touched (flag it in `## Watch` if a decision moved past it); `PROJECT.md` gets additive constraint promotions only, with approval. Removing a dead `## Landmines` row belongs to `/ds-project-verify` — it is the only skill that checks a row's scope against the source, and a landmine deleted on a guess is a warning nobody gets twice.

## Output

A list of what was archived (decisions by classification, roadmap sections), the archive file paths written, any `PROJECT.md` promotion made, any scratch file deleted, and a one-line confirmation that the live files are compacted and nothing was lost — scoped to what compact actually checked, which is governance, not accuracy.
