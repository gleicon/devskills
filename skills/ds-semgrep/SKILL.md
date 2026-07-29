---
name: ds-semgrep
description: "Run a local static application security test (SAST) with Semgrep to find exploitable patterns in source code."
disable-model-invocation: true
---

`/ds-semgrep` is the local SAST counterpart to `/ds-osv` (dependency scanning). It runs Semgrep against the repo to catch dangerous patterns — injection sinks, secret leaks, auth bypass shapes, and other code-level weaknesses — without leaving the machine. Complements `/ds-security-review` by providing fast, mechanical findings that the review then triages and verifies in context.

## When To Use

- Before `/ds-security-review` to surface mechanical SAST leads.
- After a change to confirm no new dangerous patterns were introduced.
- As a standalone quick scan when `semgrep` is already installed.

## Arguments

- No args: scan the current directory with the default Semgrep ruleset.
- `<path>`: scan a specific file or directory.
- `--config <id>`: use a specific ruleset (e.g. `p/default`, `p/security-audit`, `p/owasp-top-ten`). Defaults to `p/default`.
- `--json`: emit machine-readable JSON instead of the human summary.
- `--fix`: do **not** auto-fix; Semgrep autofix is experimental and unsafe for security findings. Report only.

## Process

1. Check for `semgrep` binary. If missing:
   ```
   Install semgrep:
     macOS:  brew install semgrep
     pip:    pip install semgrep
     Other:  https://semgrep.dev/docs/getting-started/
   ```
   Stop here. Do not proceed without the binary.

2. Run the scan with the default security ruleset:
   ```bash
   semgrep --config=p/default --error --json <path-or-.>
   ```

3. Parse the JSON output. Classify findings:
   - **ERROR** (blocking) — high-confidence security matches
   - **WARN** — medium-confidence patterns that need review
   - **INFO** — low-confidence or style-oriented findings

4. For each finding report:
   - File path and line range
   - Rule ID and message
   - Confidence/experimental flag
   - A one-line description of why the pattern matters
   - Whether it is a known-exploit class (e.g. SQL injection, SSRF, hardcoded secret) or a defense-in-depth gap

5. Do not auto-fix. Summarize actionable next steps and hand off to `/ds-security-review` for context and verification.

## Output

```text
Semgrep SAST: <N> findings (<B> blocking, <M> warnings, <I> info)

BLOCKING
  <file>:<line>-<line>  <rule-id>
  <short description>
  Why it matters: <one-line attack vector or risk>

WARNINGS
  <file>:<line>  <rule-id>
  ...

INFO: <N> findings — run with --config=p/security-audit to expand.

Next steps:
  Review blocking findings with /ds-security-review.
  Re-run /ds-semgrep after fixes to confirm the pattern is gone.
```

If zero findings: report clean with the ruleset and scope scanned.

## Rules

- Never apply Semgrep's autofix for security rules without human review — experimental fixes can change behavior or miss context.
- Treat Semgrep matches as leads, not verdicts. Every finding must be traced through `/ds-security-review` to confirm exploitability.
- Use `p/default` for speed; use `p/security-audit` or `p/owasp-top-ten` only when explicitly requested.
