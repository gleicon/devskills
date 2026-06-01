Run a strict review of documentation quality — accuracy against the code, broken links and stale counts, coverage gaps, and bloat. Reports a findings list by default; `--fix` applies the mechanical, unambiguous fixes.

When invoked, audit the documentation in scope against one governing principle: **documentation earns its length.** Readers skim; they don't read. Every sentence that doesn't help someone do something is noise that hides the sentences that do. So hunt two failures with equal energy — docs that are *wrong* (drifted from the code) and docs that are *bloated* (true, but nobody will read them). Be ambitious about cutting: the best fix is often to delete three paragraphs, not to polish them.

Like `/ds-code-quality-review`, this produces a prioritized list of what needs fixing. **Do not edit any files unless `--fix` is passed** (see Arguments).

## Arguments

- Treat positional args as scope (files, directories, globs). With no scope, review the documentation changed on the current branch (`README`, `docs/`, `*.md`, and the like).
- `--comments` — also audit inline **code comments** in scope, under the docs-accuracy lens. Off by default; prose docs only otherwise. To impose comment discipline and *apply* the fixes, reach for `/ds-comment-review` instead.
- Freeform scope ("the whole docs/ tree", "the README") is interpreted reasonably.
- `--fix` → after reporting, apply the findings whose fix is **mechanical and unambiguous** — a fixable dead link, a stale count, a clearly-drifted line, a bloat-cut that loses no information; leave anything resting on a judgment call (a rewrite, a restructure, a coverage gap to fill) as report-only. For inline **comment** fixes, defer to `/ds-comment-review --fix`. After applying, re-run any build/test/lint check already in the loop and revert any fix that breaks it — or that touched more than the intended mechanical edit. Close with a summary of what was applied and what was left.

## What to check

Verify mechanically wherever you can — resolve links against the filesystem, count the things a doc claims to count, run example commands when they're safe. Do not eyeball what you can confirm.

**1. Accuracy & drift (highest priority).** Docs that contradict the code are worse than missing docs.
- Commands, flags, paths, signatures, env vars, and outputs that no longer match the code.
- Examples and snippets that wouldn't run as written.
- Stale version numbers, renamed or removed things still referenced, text copied from an older state.

**2. Integrity.**
- Broken links — internal (file + anchor), cross-references, images. Resolve them; don't trust them.
- Wrong counts and numbers — "23 commands", table rows, list lengths — recount against reality.
- Duplicated or contradictory statements across files (the same thing documented two ways that have since diverged).

**3. Coverage gaps.**
- The thing a new reader needs first — how to install, run, and do the one core task — missing or buried below the fold.
- Public surface (entry points, config, API, flags) that ships undocumented.
- Non-obvious behavior, gotchas, and required setup that only the author knows.
- Do not demand docs for the obvious. Flag a gap only where a competent reader would actually get stuck.

**4. Bloat (the headline failure).**
- Walls of text where a list, a table, or three sentences would carry the same payload.
- Padding: throat-clearing intros, restating the obvious, marketing adjectives, "as you can see", hedging.
- Redundancy: the same point made three times, or duplicated across docs that now must be maintained in lockstep.
- Sections that exist for completeness but nobody reads — reference dumps that belong next to the code, changelogs nobody updates.
- Prefer deletion to rewriting. Ask of each section: if this were gone, would a reader miss it?

**5. Clarity & wording.**
- Buried lede — the key point arriving in paragraph three.
- Undefined jargon, vague nouns ("the system handles this"), ambiguous antecedents.
- Where the doc is too *terse*: a step that assumes context the reader lacks, an example that needs one line of "why". This is the rare case where more words help — call it out specifically so it isn't lost among the cut-this findings.

**6. Code comments** *(only with `--comments`)*.
- Comments that restate what the code plainly says (`// increment i`). Delete-worthy.
- Wordy or meandering comments that bury the one useful sentence.
- Comments that have drifted — they describe behavior the code no longer has. Treat as accuracy bugs.
- Missing comments *only* where the "why" is genuinely non-obvious (a workaround, a non-local invariant, a surprising constraint). Do not ask for comments on self-explanatory code — that violates the same brevity principle.

## Output

A prioritized findings list, in this order:

1. Accuracy / drift (doc contradicts code)
2. Broken links and wrong counts
3. Missing documentation a reader genuinely needs
4. Bloat — what to cut
5. Clarity and wording (including the rare "needs more")
6. Code-comment findings (if `--comments`)

For each finding:

- Anchor to `file:line` (or `file#section`).
- State the problem in one line, then the **suggested fix** — and when the fix is "cut it", say so plainly, naming what to cut.
- For drift, name both sides: what the doc says and what the code actually does.

Rules:

- Don't flood the list with cosmetic nits when there are real accuracy or coverage problems. A short high-conviction list beats a long pedantic one.
- Judge length against payload, not a word count. A long doc that's all signal is fine; a short doc that's all padding is not.
- Be direct about bloat. "This section can be deleted" is more useful than "this could be tightened".
- Report-only by default — the output is the list. With `--fix`, apply only the mechanical, unambiguous findings above and leave the judgment-dependent ones reported; then summarize what was applied vs. left.
