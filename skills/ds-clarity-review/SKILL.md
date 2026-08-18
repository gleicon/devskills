---
name: ds-clarity-review
description: "Review prose for clarity against the Federal Plain Language Guidelines — any text: docs, comment prose, commit text, UI and error strings. Not `/ds-deslop` (code slop) or `/ds-comment-review` (whether a comment earns its place): this judges only whether a reader understands the words. Reports before/after findings; `--fix` applies the rewrites."
disable-model-invocation: true
---

When invoked, review the prose in scope against the **Federal Plain Language Guidelines**, applied with **Grice's maxims of manner and quantity** in mind. The guidelines are the rule set — bring no private checklist, and order findings by reader impact (what most misleads first), not by which guideline fired. Technical terms the reader already owns count as plain language.

Boundaries: `/ds-deslop` removes code slop from the branch diff; `/ds-comment-review` judges whether comments earn their place; `/ds-doc-quality-review` audits documents as artifacts — accuracy against the code, links, coverage, bloat. This skill has one lens: **does a reader understand the words?** It optimizes for being understood, not for how the text sounds — reading-as-machine-written is `/ds-humanize`'s lens — and it never judges code.

## Arguments

- Positional args are scope (files, directories, globs). With no scope, review the prose changed on the current branch.
- `--full` → review all prose in the repo instead of the branch's changes. Explicit positional scope still wins.
- `--fix` → apply the reported rewrites in place; exceptions stay untouched. After applying, confirm the diff touches prose only — code, identifiers, URLs, and quoted material must be byte-identical; revert anything else.

## Process

1. **Collect the scope's prose.** Fenced code blocks, inline code, URLs, and quoted material keep their exact form — never a finding, never edited.
2. **Run the drift pass mechanically.** While reading, collect candidate pairs that may be one concept under two names ("calc type" / "calculation type"). `grep` the scope for each variant — the tool produces every count and site list. Then judge each enumerated pair: term drift, or genuinely two concepts.
3. **Review the prose against the guidelines**, Grice's manner and quantity in mind — at sentence and document level both (the minor point given the most words is a quantity finding).
4. **Report.** No edits unless `--fix`.

## Output

For each finding: `file:line`, **before** (quoted), **after** (the rewrite). Where a rewrite would lose real nuance, don't rewrite — report it as an **exception** and say what would be lost. Ambiguity a reader would act on wrongly outranks structure, which outranks word choice. A short high-conviction list beats a long pedantic one.

Rules:

- The guidelines govern; do not layer house rules on top of them.
- Never count by eye — every number in a finding comes from the tooling.
- Prose only. Code, identifiers, URLs, and quoted material stay byte-identical even under `--fix`.
