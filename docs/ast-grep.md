# ast-grep — the structural pass for `/ds-security-review`

[ast-grep](https://ast-grep.github.io/) is a structural search tool: it matches code by its
syntax tree, not by text, so a rule like `eval($INPUT)` matches every call regardless of
spacing, variable names, or line breaks. devskills installs it (via `install.sh`) as an
**optional, additive** aid to `/ds-security-review`.

> **Experimental.** This integration is new and meant to be exercised in the field. The
> starter rules below are a seed, not a finished library — the intent is to harvest the
> rules that actually surface real findings back into this file over time.

## The frame: additive, never a filter

ast-grep does **not** replace reading code, and it does **not** exist to save tokens. A
security review's job is to trace untrusted data from where it enters to where it's used —
the value is the *source → sink path*, which a bare matched node can't show you. So ast-grep
is used only to **widen** the review:

- It mechanically enumerates structural matches across the whole scope — dangerous sinks,
  every call site of a risky API, where an untrusted type flows — so nothing is skipped by
  skimming a large diff.
- Each match is **one more candidate branch** to investigate. The reviewer still opens the
  surrounding code and traces the flow in full context.

It never narrows what gets read, never gates on scope size, and never judges a match in
isolation. If it's not installed, the review runs exactly as it always has.

## Install

```bash
brew install ast-grep          # macOS
npm i -g @ast-grep/cli         # anywhere with Node
```

`install.sh` does this for you (Homebrew → npm fallback); `scripts/upgrade-deps.sh` upgrades it.

## Running a pass

Author rules for the languages and risky APIs in scope, run them inline, and read the matches
back with `jq`:

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

Notes that matter:

- An inline rule needs `id`, `language`, and `rule:`. Separate multiple rules with `---`.
- `--json=stream` prints one match object per line (good for piping); `--json` alone defaults
  to `pretty`. The `=<style>` syntax is required to pick a style.
- A match object carries `file`, `text`, `lines`, `range`, and `metaVariables`
  (`single` / `multi` / `transformed`).
- **Ranges are 0-based** (`range.start.line` starts at 0 — the LSP/tree-sitter convention).
  Account for it when you jump to a line.

Then, for each match: open the file and trace the data flow in full context. The match tells
you *where to look*, not *whether it's exploitable*.

## The rule language

A rule is YAML. Three families compose freely
([full reference](https://ast-grep.github.io/reference/rule.html)):

**Atomic** — match a node directly:

| Key | Matches |
|-----|---------|
| `pattern` | code by shape, e.g. `eval($INPUT)` |
| `kind` | a tree-sitter node kind, e.g. `call_expression` |
| `regex` | node text against a Rust regex |
| `nthChild` / `range` | position among siblings / a literal source range |

**Metavariables** inside a pattern: `$VAR` captures one node, `$$$ARGS` captures a variadic
list. Example: `pattern: console.log($MSG, $$$REST)`.

**Relational** — match by surroundings (each takes `stopBy: neighbor | end` and an optional
`field`):

| Key | Meaning |
|-----|---------|
| `inside` | node is inside another matching node |
| `has` | node has a descendant matching the sub-rule |
| `follows` / `precedes` | node comes after / before another |

**Composite** — boolean logic: `all:` (every sub-rule), `any:` (any sub-rule), `not:`.

## Starter cookbook

Seeds, not a closed list — extend per the code in scope. Set `language:` to the file's
language; ast-grep language names are capitalized (`JavaScript`, `TypeScript`, `Python`, `Go`,
`Rust`, `Java`, …).

**Injection sinks**

```yaml
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
  inside:
    kind: call_expression
    stopBy: end
```

**Unsafe output / rendering**

```yaml
id: react-dangerous-html
language: TypeScript
rule:
  has:
    pattern: dangerouslySetInnerHTML={$_}
    stopBy: end
```

**Weak randomness for security values**

```yaml
id: py-insecure-random
language: Python
rule:
  any:
    - pattern: random.random()
    - pattern: random.randint($$$ARGS)
```

Each rule's matches are leads — confirm exploitability by reading the surrounding code, never
from the match alone.

## References

- [Rule object reference](https://ast-grep.github.io/reference/rule.html)
- [`scan` CLI](https://ast-grep.github.io/reference/cli/scan.html) · [JSON mode](https://ast-grep.github.io/guide/tools/json.html)
- [Pattern syntax](https://ast-grep.github.io/guide/pattern-syntax.html) · [Playground](https://ast-grep.github.io/playground.html)
