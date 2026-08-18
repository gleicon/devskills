---
name: ds-humanize
description: "Remove the AI tells from prose — the patterns that read as machine-written. Not `/ds-deslop` (code slop) or `/ds-comment-review` (whether a comment earns its place), and not `/ds-clarity-review` (whether a reader understands): this judges only whether the text sounds machine-written."
disable-model-invocation: true
---

When invoked, remove the AI tells from the prose in scope. Preserve meaning and the author's register — this removes patterns, it never injects new style, idiom, or figurative flourish. The test: **a pattern an experienced writer of this document wouldn't produce.**

This objective is distinct from clarity's: `/ds-clarity-review` optimizes for being understood; this skill optimizes for not reading as machine-written. The two overlap and then diverge — run both when you want both; never let one pass serve the other's goal.

## Arguments

Treat any argument as scope (files, directories, globs). With no scope, work on the prose changed on the current branch.

## What to remove

- **Diff-anchored framing** — prose describing its own edit history ("this was added to replace…", "now uses X instead") rather than what the thing is. Rewrite to the current state.
- **Negative parallelism** — "it's not just X, it's Y" as a reflex construction. State Y.
- **Synonym cycling** — rotating names for one thing to avoid repeating a word. One concept, one name.
- **Inline-header bullets** — a bolded label followed by a sentence restating the label.
- **Title Case Headings** where the document's convention is sentence case.
- **Emoji** decorating headings or bullets.
- **Chatbot artifacts** — conversation residue shipped as content: "Certainly!", "Here's the updated version", "as requested", "I hope this helps".
- **Sycophancy** — flattering the reader, the question, or the codebase.
- **Filler phrases** — "in order to", "it's important to note that", "as previously mentioned". Cut or compress.
- **Stacked hedging** — multiple qualifiers on one claim ("could potentially", "may possibly"). One hedge carrying real uncertainty stays; the stack goes.
- **Signposting** — announcing structure instead of having it: "let's dive in", "in this section we will".
- **Rhetorical openers** — self-answered questions and dramatic one-word leads.

## Guardrails

- Meaning-preserving. Never change what the text claims, only how it sounds.
- Remove tells; don't add voice. Injected idiom or vivid phrasing is out of scope — and would fight the clarity pass.
- Fenced code, inline code, URLs, and quoted material stay byte-identical.
- Don't enforce plainness or restructure for comprehension — that's `/ds-clarity-review`'s lens.

## Output

Apply the edits, then give a concise 1–3 sentence summary of what was removed.
