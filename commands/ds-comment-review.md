Run a strict review of code comments — flag the comments that don't earn their place and tighten the rest to the discipline. Language-agnostic. Reports a findings list by default; `--fix` applies the edits in place.

When invoked, audit the comments in scope against one question: **does each comment earn its place, and is it as short as it can be?** The discipline: comments are for humans and explain **WHY, not WHAT** — one line by default, only where the reason isn't obvious from the code, never restating code or citing plan/ticket IDs; a long comment is rare and signals importance. This is **not** `/ds-doc-quality-review --comments` (which reports comments under a docs-accuracy/consistency lens) and **not** `/ds-deslop` (branch-diff, matches existing style) — this command's only lens is comment discipline, it works on any scope, and it **imposes the standard regardless of the codebase's existing comment habits**. **Comment-only and behavior-preserving: never change code logic.**

## Arguments

- Positional args are scope (files, directories, globs, or the whole codebase). With no scope, review comments in the code changed on the current branch.
- `--fix` → apply the edits in place (delete the noise, tighten the verbose, strip cruft, keep the genuine ones) instead of only reporting.

## What to flag

1. **Restating code** — comments that narrate what the code plainly does (`// increment i`, `// loop over users`). Delete.
2. **Obvious / ceremonial** — section banners, `// constructor`, getter/setter narration, docstrings that just repeat the signature. Delete.
3. **Planning cruft** — references to plan/ticket/step IDs, "as per step 3", author-stamped noise, a commit message pasted as a comment. Strip.
4. **Buried WHY** — a genuine reason wrapped in five lines of prose. Tighten to one line.
5. **Drifted** — comments describing behavior the code no longer has. Treat as a correctness issue: fix or delete.
6. **Keep — and respect — the rare legitimately-long comment**: a non-obvious algorithm, a subtle invariant, a hard-won workaround. Don't cut these; they are the signal. Make sure they still read as important.

Doc-comments: where idiomatic (Go doc comments, Python docstrings on public API, Rust `///`), keep them — but short and to the point, not prose. Don't strip required public-API docs; don't demand docs on the obvious.

## Output

**Default (report):** a prioritized findings list. For each: anchor to `file:line`, quote the comment, name why it fails (restates / obvious / cruft / verbose / drifted), and give the fix — `delete`, or `tighten to: "<one-liner>"`. Group by delete / tighten / drifted (correctness) / kept-as-important. A short high-conviction list beats a long pedantic one.

**With `--fix`:** apply the edits — delete the noise, tighten the verbose to one line, strip the cruft, preserve and (if needed) sharpen the genuine ones — then give a concise 1–3 sentence summary of what changed. Never touch code logic. **After applying, confirm the resulting diff touches only comment lines — if any non-comment line changed, that's a bug: revert it.**

Rules:

- Comment-only, behavior-preserving. Never change code.
- **Impose the discipline regardless of the codebase's existing comment style** — do not match bad patterns (this is the point of the command).
- Keep the rare important long comment; don't flatten everything to one line.
- Change nothing else. In report mode, the output is the list.
