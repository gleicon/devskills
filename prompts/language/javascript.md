## Language Profile — JavaScript

Target: ES2022+. Cloudflare Workers, vanilla frontend, Wrangler.

Use this profile for JS-only projects. Prefer TypeScript for new projects — use this profile when TypeScript is not practical (rapid prototypes, configuration scripts, legacy maintenance).

### Toolchain

- Runtime: Bun or Node 20+
- Workers: Wrangler 3+
- Test: Vitest
- Lint: Biome or ESLint
- Format: Biome or Prettier

### Code Style

- `const` by default. `let` when reassignment is required. Never `var`.
- Arrow functions for callbacks. Named function declarations for top-level functions.
- Template literals over string concatenation.
- Destructuring at function entry for clarity.
- `async/await` over raw Promise chains.

### Error Handling

Every `async` function handles errors explicitly. No unhandled rejections.

```js
// Explicit handling
async function fetchUser(id) {
  const response = await fetch(`/users/${id}`)
  if (!response.ok) {
    throw new Error(`fetch user ${id}: ${response.status}`)
  }
  return response.json()
}

// Caller wraps in try/catch or uses Result pattern
```

### Cloudflare Workers

```js
export default {
  async fetch(request, env, ctx) {
    const url = new URL(request.url)
    // route by url.pathname
  }
}
```

Same rules as TypeScript Workers profile apply:
- Secrets via `env.SECRET_NAME`
- KV gets return `null` on miss — handle explicitly
- No Node.js-only APIs

### Module System

ESM everywhere (`import`/`export`). No CommonJS (`require`, `module.exports`) in new code.

### JSDoc for Public APIs

When TypeScript is not in use, document public function signatures with JSDoc:

```js
/**
 * @param {string} id
 * @returns {Promise<User>}
 */
async function getUser(id) { ... }
```

### Testing

Vitest. Same structure as TypeScript profile.

### Tiger Style Integration

- Input validation at every external boundary (request, env, file read).
- No silent error discard in catch blocks.
- Loops over user-controlled arrays have explicit length checks.
- Functions over 70 lines are refactored.
