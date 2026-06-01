## Language Profile — JavaScript

Target: ES2022+. Cloudflare Workers, vanilla frontend, Wrangler.

Use this profile for JS-only projects. Prefer TypeScript for new projects — use this when TypeScript isn't practical (rapid prototypes, config scripts, legacy maintenance).

### Toolchain

Runtime: Bun or Node 20+. Workers: Wrangler 3+. Test: Vitest. Lint/format: Biome.

### Error Handling

Every `async` function handles or propagates its errors — no unhandled promise rejections.

```js
async function fetchUser(id) {
  const response = await fetch(`/users/${id}`)
  if (!response.ok) throw new Error(`fetch user ${id}: ${response.status}`)
  return response.json()
}
```

### Cloudflare Workers

Same rules as the TypeScript Workers profile: secrets via `env.SECRET_NAME`, KV gets return `null` on miss (handle it), no Node.js-only APIs.

### Module System

ESM everywhere (`import`/`export`). No CommonJS in new code.

### JSDoc for Public APIs

Without TypeScript, document public signatures with JSDoc:

```js
/**
 * @param {string} id
 * @returns {Promise<User>}
 */
async function getUser(id) { ... }
```

### Testing

Vitest, same structure as the TypeScript profile.

### Tiger Style

- No silent error discard in `catch` blocks.
