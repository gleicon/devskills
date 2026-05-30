## Language Profile — TypeScript

Target: TypeScript 5+. Cloudflare Workers, Next.js, React, edge runtimes.

Apply these conventions to all TypeScript/JavaScript code in this session.

### Toolchain

- Runtime: Bun (preferred) or Node 20+
- Workers: Wrangler 3+
- Build: `tsc --noEmit` (type check) + bundler (Bun, Vite, or Wrangler)
- Test: Vitest (unit), Miniflare (Workers integration), Playwright (E2E)
- Lint: Biome or ESLint with typescript-eslint
- Format: Biome or Prettier

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

- Prefer discriminated unions over boolean flags for state:
  ```ts
  type State =
    | { status: "idle" }
    | { status: "loading" }
    | { status: "success"; data: Data }
    | { status: "error"; error: Error }
  ```
- Use `satisfies` for object literals that must conform to a type.
- Runtime validation at system boundaries with Zod or Valibot — not manual checks.
- `type` for unions and aliases. `interface` for object shapes that extend.

### Error Handling

No `throw` in library code except for programmer errors (wrong types at runtime boundaries).

```ts
// Result type for recoverable errors
type Result<T, E = Error> =
  | { ok: true; value: T }
  | { ok: false; error: E }

// Or use neverthrow / ts-results for ergonomics
```

All `async` functions: either return `Result` or propagate typed errors. No unhandled promise rejections.

### Cloudflare Workers

```ts
export default {
  async fetch(request: Request, env: Env, ctx: ExecutionContext): Promise<Response> {
    // route, handle, return
  }
} satisfies ExportedHandler<Env>
```

- `env` typed via `interface Env { ... }` — no `any`.
- `ctx.waitUntil()` for background work that must complete after response.
- KV and R2 operations: always handle `null` on get (key not found).
- Durable Objects: all state mutations in `this.state.storage.transaction()`.

### React

- Functional components only. Named exports.
- `useState` for local UI state. Derive everything derivable from props.
- Effects only for synchronization with external systems (timers, subscriptions, DOM).
- Data fetching: Tanstack Query or SWR — not bare `useEffect` + `fetch`.
- No prop drilling past 2 levels — use context or co-location.

### Module Conventions

- Named exports everywhere. Default exports only for pages (Next.js) and Workers entry.
- Barrel files (`index.ts`) only at package boundaries, not within a feature folder.
- Imports grouped: external → internal → relative. No mixing.

### Testing

```ts
// Vitest
import { describe, it, expect } from "vitest"
describe("feature", () => {
  it("does X when Y", () => {
    expect(result).toEqual(expected)
  })
})
```

- Pure functions: unit test directly.
- Workers: use `unstable_dev` or Miniflare.
- Components: Vitest + Testing Library. Test behavior, not implementation.

### Naming

- Files: `kebab-case.ts`
- Types/Interfaces: `PascalCase`
- Functions/variables: `camelCase`
- Constants: `SCREAMING_SNAKE_CASE` for true constants; `camelCase` for frozen objects
- React components: `PascalCase` file and export

### Tiger Style Integration

- All user input validated at entry (request body, URL params, env vars).
- No optional chaining used to hide missing error handling.
- Promises are not fire-and-forget unless `ctx.waitUntil()` or equivalent tracks them.
- Functions over 70 lines are refactored.
