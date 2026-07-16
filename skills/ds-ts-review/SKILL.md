---
name: ds-ts-review
description: "Review TypeScript code with Tiger Style constraints and TypeScript/Workers idioms."
disable-model-invocation: true
---

Applies to: TypeScript 5.5+. Cloudflare Workers, Next.js, React, edge runtimes.

## Arguments

Scan the invocation for the `--no-tiger`, `--fix`, and `--full` flags. Treat every other argument as review scope (files or directories); if no scope is given, review the changed files on the current branch.

- `--no-tiger` present → skip the Tiger Style section; run the remaining sections only.
- `--no-tiger` absent → run all sections (default).
- `--fix` → after reporting, apply only the violations whose fix is **mechanical and unambiguous** (a rename to the idiom, a missing error check the review is certain about). Anything that changes logic or rests on an unverified assumption — especially security and correctness findings — **stays report-only**. After applying, re-run any build/test/lint check already in the loop and revert any fix that breaks it — or that touched more than the intended mechanical edit. End with a summary of what was applied and what was left.
- `--full` → review the entire codebase instead of just the branch's changes. Explicit positional scope still wins; `--full` only replaces the no-scope default.

Example: `/ds-ts-review --no-tiger src/ worker/` reviews `src/` and `worker/` without Tiger Style.

## Review Checklist

Use the checklist as a lens, not a scorecard: reason about the actual change, report real violations anchored to `file:line`, and flag issues even when they aren't listed. Don't manufacture findings to fill a category. Report only violations — no praise, no summary.

### Tiger Style

Skip this section entirely if `--no-tiger` was passed. Otherwise it is mandatory.
- [ ] No unbounded loops over user-controlled data
- [ ] All promises awaited or explicitly handled — no floating promises
- [ ] Error values propagated explicitly — no silent catch-and-ignore
- [ ] Functions under 70 lines
- [ ] Variable names include units/qualifiers where applicable (`_ms`, `_bytes`, `_count`)

### TypeScript
- [ ] `strict: true` in tsconfig — no `any` unless explicitly justified with comment
- [ ] No `as Type` casts that bypass runtime checks — validate at boundaries
- [ ] `satisfies` to check a value against a type without widening it, instead of an annotation that erases the literal type or an `as` cast
- [ ] Discriminated unions used for state modeling, not boolean flags
- [ ] No optional chaining (`?.`) used to hide missing error handling
- [ ] Prefer unions of string literals over enums entirely (`const enum` breaks under `isolatedModules`/`verbatimModuleSyntax` and most bundlers)
- [ ] Return types explicit on all exported functions
- [ ] No `namespace` — use ES modules
- [ ] `using`/`await using` for scoped cleanup (TS 5.2) where a resource implements `[Symbol.dispose]`/`[Symbol.asyncDispose]`, instead of hand-written try/finally

### Cloudflare Workers (when applicable)
- [ ] No Node.js-only APIs (`fs`, `path`, `crypto` with Node semantics) — use Web APIs
- [ ] `env` bindings typed via `Env` interface, not `any`
- [ ] Durable Objects: state mutations wrapped in `this.state.storage.transaction()`
- [ ] CPU time budget respected — no long synchronous computation in request handler
- [ ] Secrets accessed via `env.SECRET_NAME`, not hardcoded
- [ ] `wrangler.toml` bindings match runtime `Env` interface

### React / Frontend (when applicable)
- [ ] No state mutation — `setState` or dispatch with a new object
- [ ] `useEffect` dependency arrays complete — no stale-closure bugs — and effects reserved for real external synchronization, not derived state
- [ ] With the React Compiler enabled, hand-written `useMemo`/`useCallback`/`React.memo` are dropped in new code — kept only where a measured need survives the compiler
- [ ] `ref` passed as an ordinary prop (React 19) — no `forwardRef` wrapper
- [ ] Promises/context read with `use()` where it fits; forms built on `useActionState`/`useFormStatus` rather than a bespoke state library
- [ ] Data fetching co-located with its loading/error state
- [ ] Unsafe HTML rendering avoided; content sanitized before DOM insertion
- [ ] Keys on lists are stable identifiers, not array indices

### Performance
_Idiom-level checks only — for a ranked, costed optimization plan, use `/ds-perf-plan`._
- [ ] No database/network calls issued inside loops; N+1 patterns absent
- [ ] Independent async work runs concurrently (`Promise.all`), not awaited one-by-one
- [ ] Large payloads streamed, not fully buffered — Workers have a tight memory budget

### Security
- [ ] No dynamic code execution from untrusted strings
- [ ] Raw HTML insertion avoided unless content is sanitized through a trusted library
- [ ] CORS headers explicitly configured, not wildcard `*` on sensitive routes
- [ ] Auth tokens not stored in localStorage (prefer httpOnly cookies or memory)
- [ ] No secrets in client-side bundles
- [ ] All user input validated and bounded at the entry point

### Testing
- [ ] Unit tests for pure functions
- [ ] Integration tests for Worker routes (use Miniflare or `wrangler dev`)
- [ ] No `any` type in test assertions

## Output Format

```
<file>:<line>: <severity>: <problem>. <fix>.
```

Severity: `critical`, `major`, `minor`.
