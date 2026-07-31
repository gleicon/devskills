---
name: ds-structural-diff
description: "Use ast-grep to build a structural inventory of a codebase change: added/removed/modified functions, types, public APIs, and imports. A building block for /ds-quality-gate and large-change review."
disable-model-invocation: true
---

`/ds-structural-diff` is a **structural delta**, not a line-by-line diff. When a large change lands, it tells you what changed at the architectural level: new contracts, signature changes, new dependencies, and deleted or duplicated surface area. It is report-only and additive — it does not judge correctness, it just turns the AST into a concise map so the human and the next review skills know where to focus.

Use it as a standalone pass before a big review, or as the first pass inside `/ds-quality-gate` to orient the later, deeper reviews.

## When to use

- After a large agent generation or refactor — before reading every file.
- Before `/ds-quality-gate` or `/ds-security-review` to surface the scope of the change.
- When you suspect a change introduced duplicate logic, signature drift, or hidden new public API.

## Arguments

- No args: compare the current working tree against the merge-base of `main`/`master`.
- `--base <commit>`: compare against an explicit commit (e.g. `HEAD~1`, a tag, or a branch name).
- `--full`: compare against the full base tree, not just files that changed.
- `--lang <language>`: restrict to one language. ast-grep language names are capitalized (`Go`, `TypeScript`, `Python`, `Rust`, `Java`). Default: detect from the changed files.
- `--scope <path>`: restrict to a path or glob. Only changed files under that path are considered.

## Process

1. **Determine the base.** Use `git merge-base HEAD main` (or `master` if no `main`). If `--base` is given, use that.
2. **Find the changed files.** Use `git diff --name-only <base>..HEAD` (or `git diff --name-only <base>` if comparing against the working tree). If `--scope` is given, filter the list.
3. **Check for `ast-grep`.** If missing, stop and print:
   ```
   Install ast-grep:
     macOS:  brew install ast-grep
     npm:    npm install -g @ast-grep/cli
   Or run: devskills doctor --fix
   ```
4. **Extract structural elements from each changed file.** Run the language-specific patterns below with `ast-grep scan --inline-rules '<rules>' --json=stream`. Run them on the **base** version (`git show <base>:<path>`) and the **current** version of the file. Start with the patterns for the detected language; fall back to the generic patterns if the language is not covered.
5. **Normalize.** For each match, produce a key of the form:
   ```
   <kind>:<qualified-name>:<signature>
   ```
   where `signature` is the name + parameters + return type (or just the import path for imports). Ignore function bodies, comments, and formatting.
6. **Compute the delta.** For each file, compare the base set and the current set:
   - `added` — in current, not in base
   - `removed` — in base, not in current
   - `modified` — same qualified name, different signature
   - `unchanged` — same key in both
7. **Detect slop signals.** Flag:
   - Duplicate declarations with the same local name in the same file or package (possible copy-paste slop).
   - New public API (`export`/`public`/`Exported`) with no doc comment or tests.
   - Added dependency/import that is unused in the current version.
   - Removed dependency/import that is still referenced in the current version.
   - Signature changes that break existing call sites (in the diff or in the rest of the codebase).

## Output

Use this shape:

```text
## Structural diff: <base> → working tree
Scope: <paths or "all changed files">
Languages: <Go, TypeScript, ...>

### Added surface
- <file>:<line> <kind> <qualified-name> — <short purpose>
- <file>:<line> import <package> — new dependency

### Removed surface
- <file>:<line> <kind> <qualified-name>

### Modified signatures
- <file>:<line> <kind> <name>
  - base:     <old signature>
  - current:  <new signature>
  - risk:     <breaking | additive | internal>

### Slop / deviations
- <file>:<line> <kind> <name> — <what looks off>

### Unchanged structural anchors
- <file> <kind> <name> (only if useful for orientation)

### Next steps
- </ds-security-review> for new public API, network, or input paths
- </ds-code-quality-review> for duplicate or dead surface
- </ds-verify-this> for signature changes
```

Keep the report scannable. For a large change, group by package/directory and put the most important signals first (signature changes, new public API, slop).

## Rules

- **Do not edit files.** This skill is report-only.
- **Do not reimplement the AST parser.** Use `ast-grep` only.
- **A match is a lead, not a verdict.** If the pattern says a new public API exists, confirm by reading the surrounding code before declaring it dangerous.
- **If the base checkout is hard**, use `git diff --name-only` and then `git show <base>:<file>` for each file. Do not modify the working tree.
- **If a language is not covered**, say so and run the generic patterns that apply (imports, functions), or ask the user for the language.

## Starter ast-grep patterns

Run with `--json=stream` and parse one match per line. `range.start.line` is 0-based.

### Generic — imports

```yaml
id: imports
language: Go
rule:
  pattern: import($$$)
```

For TypeScript:
```yaml
id: ts-imports
language: TypeScript
rule:
  any:
    - pattern: import { $$$ } from $PATH
    - pattern: import $NAME from $PATH
    - pattern: import $PATH
```

### Functions / methods

**Go**
```yaml
id: go-func
language: Go
rule:
  any:
    - pattern: func $NAME($$$) $RET
    - pattern: func ($R) $NAME($$$) $RET
```

**TypeScript**
```yaml
id: ts-function
language: TypeScript
rule:
  any:
    - pattern: function $NAME($$$) $RET
    - pattern: const $NAME = function($$$) $RET
    - pattern: const $NAME = ($$$) => $RET
    - pattern: $NAME($$$) { $$$ }
      inside: { kind: class_body, stopBy: end }
```

**Python**
```yaml
id: py-function
language: Python
rule:
  any:
    - pattern: def $NAME($$$):
    - pattern: async def $NAME($$$):
```

**Rust**
```yaml
id: rust-function
language: Rust
rule:
  any:
    - pattern: fn $NAME($$$) $RET
    - pattern: async fn $NAME($$$) $RET
```

### Types / contracts

**Go**
```yaml
id: go-type
language: Go
rule:
  any:
    - pattern: type $NAME struct { $$$ }
    - pattern: type $NAME interface { $$$ }
    - pattern: type $NAME $BODY
```

**TypeScript**
```yaml
id: ts-type
language: TypeScript
rule:
  any:
    - pattern: interface $NAME { $$$ }
    - pattern: type $NAME = $$$
    - pattern: class $NAME { $$$ }
    - pattern: export interface $NAME { $$$ }
    - pattern: export class $NAME { $$$ }
```

### Public API surface

**Go**
```yaml
id: go-exported
language: Go
rule:
  any:
    - pattern: func $NAME($$$) $RET
    - pattern: type $NAME $BODY
    - pattern: var $NAME $TYPE
    - pattern: const $NAME $TYPE = $VAL
  regex: ^[A-Z]
```

**TypeScript**
```yaml
id: ts-exported
language: TypeScript
rule:
  any:
    - pattern: export function $NAME($$$)
    - pattern: export const $NAME
    - pattern: export class $NAME { $$$ }
    - pattern: export interface $NAME { $$$ }
```

## Running a pattern

```bash
ast-grep scan --inline-rules '
id: go-func
language: Go
rule:
  pattern: func $NAME($$$) $RET
' --json=stream <path>
```

For the base version of a file:

```bash
git show <base>:path/to/file.go | ast-grep scan --stdin --lang Go --inline-rules '<rules>' --json=stream
```

Combine with `jq` for compact output:

```bash
jq -c '{file: .file, line: .range.start.line, kind: .ruleId, name: .metaVars.$NAME.text}'
```

When comparing two versions, normalize the key by stripping whitespace and comments, but keep the parameter list and return type. A match without a captured name is still useful as a slop signal — group it under "uncategorized".
