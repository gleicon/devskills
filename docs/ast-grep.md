# ast-grep — the structural pass for `/ds-security-review`

[ast-grep](https://ast-grep.github.io/) is a structural search tool: it matches code by its
syntax tree, not by text, so a rule like `eval($INPUT)` matches every call regardless of
spacing, variable names, or line breaks. devskills installs it (via `install.sh`) as an
**optional, additive** aid to `/ds-security-review`.

> **Experimental.** This integration is new and meant to be exercised in the field.

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

## The operational reference

How `/ds-security-review` actually runs a pass — the inline-rule invocation, the rule
language (atomic / metavariable / relational / composite), and a starter cookbook of injection,
output, and randomness rules — lives in the skill's own companion,
[`skills/ds-security-review/ast-grep.md`](../skills/ds-security-review/ast-grep.md). It installs
alongside the skill, so the model reads it at runtime whenever ast-grep is present.

**Harvest proven rules there, not here:** a rule that surfaces a real finding belongs in the
seed library the review inherits, which is the companion.

## References

- [Rule object reference](https://ast-grep.github.io/reference/rule.html)
- [`scan` CLI](https://ast-grep.github.io/reference/cli/scan.html) · [JSON mode](https://ast-grep.github.io/guide/tools/json.html)
- [Pattern syntax](https://ast-grep.github.io/guide/pattern-syntax.html) · [Playground](https://ast-grep.github.io/playground.html)
