## Language Profile — TypeScript

Target: TypeScript 5+. Cloudflare Workers, Next.js, React, edge runtimes.

Apply these conventions to all TypeScript/JavaScript code in this session.

### Toolchain

Runtime: Bun (preferred) or Node 20+. Workers: Wrangler 3+. Test: Vitest (unit), Miniflare (Workers), Playwright (E2E). Lint/format: Biome (or ESLint + Prettier).

### tsconfig Baseline

```json
{
  "compilerOptions": {
    "strict": true,
    "noUncheckedIndexedAccess": true,
    "exactOptionalPropertyTypes": true,
    "noImplicitReturns": true,
    "noFallthroughCasesInSwitch": true
  }
}
```

No `any`. Justify every `unknown` cast with a runtime check.

### Type Design

- Discriminated unions over boolean flags for state:
  ```ts
  type State =
    | { status: "idle" }
    | { status: "loading" }
    | { status: "success"; data: Data }
    | { status: "error"; error: Error }
  ```
- `satisfies` for object literals that must conform to a type.
- Runtime validation at system boundaries with Zod or Valibot — not manual checks.
- `type` for unions and aliases; `interface` for object shapes that extend.

### Error Handling

No `throw` in library code except programmer errors. Recoverable errors return a `Result`:

```ts
type Result<T, E = Error> =
  | { ok: true; value: T }
  | { ok: false; error: E }
```

`async` functions return `Result` or propagate typed errors. No unhandled rejections.

### Cloudflare Workers

- `env` typed via `interface Env` — no `any`.
- `ctx.waitUntil()` for background work that must outlive the response.
- KV/R2 gets return `null` on miss — handle it.
- Durable Objects: state mutations in `this.state.storage.transaction()`.

### React

- Functional components, named exports. Derive everything derivable from props.
- Effects only for synchronization with external systems (timers, subscriptions, DOM) — not data fetching. Use Tanstack Query or SWR.
- No prop drilling past 2 levels — context or co-location.

### Module Conventions

Named exports everywhere; default exports only for Next.js pages and the Workers entry. Barrel files (`index.ts`) only at package boundaries.

### Testing

Vitest. Workers via `unstable_dev`/Miniflare. Components with Testing Library — test behavior, not implementation. No `any` in assertions.

### Tiger Style

- Optional chaining never used to hide missing error handling.
- Promises tracked, not fire-and-forget (unless `ctx.waitUntil()` owns them).
