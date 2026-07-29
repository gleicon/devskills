# ast-grep — additive structural pass for `/ds-security-review`

Read this when `ast-grep` is installed (`command -v ast-grep`). It's an **additive** aid, never a filter: it mechanically enumerates known-dangerous structural patterns across the whole scope — injection sinks, every call site of a risky API, where an untrusted type flows — so no candidate is skipped by skimming a large diff. It never narrows what you read, never gates on scope size, never saves tokens. Each match is **one more branch to investigate**: open the surrounding code and trace the data flow in full context, exactly as without it. A match is a lead, not a verdict — confirm exploitability by reading the code, never from the match alone. (If ast-grep is absent, the full read is the baseline and nothing changes.)

## Running a pass

Author rules for the languages and risky APIs actually in scope, run them inline, and read matches back with `jq`:

```bash
ast-grep scan --inline-rules '
id: py-shell-exec
language: Python
rule:
  any:
    - pattern: os.system($CMD)
    - pattern: subprocess.run($CMD, shell=True)
' --json=stream . | jq -c '{file, line: .range.start.line, text}'
```

- An inline rule needs `id`, `language`, and `rule:`; separate multiple rules with `---`.
- `--json=stream` prints one match per line (good for piping); the `=<style>` is required to pick a style.
- **Ranges are 0-based** (`range.start.line` starts at 0 — the tree-sitter convention). Open each file at that line and trace the flow.

## Rule language

YAML; three families compose freely (full reference: <https://ast-grep.github.io/reference/rule.html>):

- **Atomic** — `pattern` matches by code shape (`eval($INPUT)`); `kind` matches a tree-sitter node kind; `regex` matches node text.
- **Metavariables** inside a pattern — `$VAR` captures one node, `$$$ARGS` a variadic list (e.g. `console.log($MSG, $$$REST)`).
- **Relational** — constrain a match by its surroundings: `inside:` / `has:` / `follows:` / `precedes:` (add `stopBy: end` to search past immediate neighbors) — e.g. a Go `fmt.Sprintf($Q, $$$ARGS)` `inside:` a database-query call.
- **Composite** — boolean logic: `any:`, `all:`, `not:`.

## Starter rules

Seeds, not a closed list — extend per the code in scope. Set `language:` to the file's language; ast-grep language names are capitalized (`JavaScript`, `TypeScript`, `Python`, `Go`, `Rust`, `Java`, …).

```yaml
# Injection sinks
id: js-eval
language: JavaScript
rule:
  any:
    - pattern: eval($INPUT)
    - pattern: new Function($$$ARGS)
---
id: py-shell-exec
language: Python
rule:
  any:
    - pattern: os.system($CMD)
    - pattern: subprocess.run($CMD, shell=True)
    - pattern: subprocess.call($CMD, shell=True)
---
id: go-sql-sprintf
language: Go
rule:
  pattern: fmt.Sprintf($QUERY, $$$ARGS)
  inside: { kind: call_expression, stopBy: end }
---
# Unsafe output / rendering
id: react-dangerous-html
language: TypeScript
rule:
  has: { pattern: dangerouslySetInnerHTML={$_}, stopBy: end }
---
# Weak randomness for security values
id: py-insecure-random
language: Python
rule:
  any:
    - pattern: random.random()
    - pattern: random.randint($$$ARGS)
```

Each rule's matches are leads — trace the source→sink path in the surrounding code to judge exploitability.

**Harvest:** when a rule surfaces a real finding in the field, add it here so the next review inherits it — this file is the seed library, meant to grow.

## Install

`brew install ast-grep` (macOS) or `npm i -g @ast-grep/cli` (anywhere with Node) — or let `devskills doctor --fix` install it for you.
