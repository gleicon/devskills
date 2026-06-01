Pick one language-agnostic review to run from a quick menu.

A launcher, not a review itself: it shows the menu, you choose one, and it runs that review on the current branch diff (or a scope you pass). For the per-language reviews (`/ds-go-review`, `/ds-ts-review`, …) call those directly.

## On activation

1. Present the menu below as a **single-select** (pick one). If the host supports an interactive picker, use it; otherwise list the reviews and ask the user to pick one.

   - **bug** — correctness audit: real runtime bugs (logic, null/absent, races, leaks, boundaries)
   - **security** — exploitability: untrusted input → dangerous sink, each finding names the attack
   - **data** — data correctness: schema/integrity, query results, transactions, migration safety (`--pipelines` adds ETL)
   - **code-quality** — maintainability: abstraction, file sprawl, spaghetti conditions
   - **test-quality** — is the code that matters tested, and tested well (not coverage)
   - **doc-quality** — docs accuracy, dead links, stale counts, bloat (`--comments` adds code comments)
   - **ui-quality** — UI engineering correctness, accessibility, Core Web Vitals, design craft
   - **comment** — comment discipline: WHY-not-WHAT, strip restate/obvious/cruft

2. Run the selected review by activating its command `/ds-<name>-review` — read that command file (e.g. `ds-bug-review.md`) from your commands directory if its exact rules aren't already in context. Load **only the selected** review; never the whole set.
3. Forward any scope or flags passed to `/ds-review` (a path, `--fix`, `--pipelines`, `--comments`, `--no-tiger`) to the chosen review. With no scope, it reviews the changed files on the current branch.

## Notes

- Single-select — one review per run; rerun to run another. To run several in order as a pre-PR gate, use `/ds-quality-gate-mode`.
- The menu is hardcoded for speed (showing it reads nothing); only the picked review's file is loaded.
- Keep the menu in sync when a language-agnostic review is added or removed.
- The per-language reviews are intentionally excluded — invoke `/ds-go-review` · `/ds-ts-review` · `/ds-rust-review` · `/ds-python-review` · `/ds-java-review` · `/ds-zig-review` directly.
