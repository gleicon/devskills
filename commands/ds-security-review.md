Run a strict security review of code changes — find exploitable weaknesses. Language-agnostic. Reports a findings list by default; `--fix` applies the mechanical, unambiguous fixes (logic-changing or uncertain ones stay reported).

When invoked, audit the code in scope against one question: **how would an attacker abuse this?** Look for the weaknesses that lead to real compromise — injection, broken access control, leaked secrets, untrusted input trusted too far. This is the portable, cross-language pass; for deeper language specifics, `/ds-go-review`, `/ds-ts-review`, and `/ds-rust-review` carry their own Security sections. Every finding names the attack that exploits it. **Do not edit any files unless `--fix` is passed** (see Arguments).

## Arguments

- Treat positional args as scope (files, directories, globs). With no scope, review the code changed on the current branch.
- Freeform scope ("the auth handler", "the upload path") is interpreted reasonably.
- `--fix` → after reporting, apply only the findings whose fix is **mechanical and unambiguous** — a single obvious edit, no design judgment (e.g. removing a secret committed to source, tightening over-permissive file modes). A wrong fix to a security finding is worse than none, so anything that changes behavior or rests on an assumption you couldn't verify **stays report-only**. After applying, re-run any build/test/lint check already in the loop and revert any fix that breaks it — or that touched more than the intended mechanical edit. Close with a summary of what was applied and what was left.

## ast-grep pre-filter (when available)

If `ast-grep` is installed (`command -v ast-grep`), run it first as a structural pre-filter before reading full files. This extracts only the nodes that match dangerous patterns — Claude reviews those nodes instead of entire files, reducing noise and token cost.

Run `ast-grep scan --json --inline-rules` with patterns matched to the languages in scope. Example patterns to emit as inline YAML rules:

**Injection sinks** (adapt `language:` to the file's language):
```yaml
id: eval-call
language: JavaScript
rule:
  pattern: eval($INPUT)
---
id: exec-call
language: Python
rule:
  any:
    - pattern: os.system($CMD)
    - pattern: subprocess.call($CMD, shell=True)
    - pattern: subprocess.run($CMD, shell=True)
---
id: sql-concat
language: Go
rule:
  pattern: fmt.Sprintf($QUERY, $$$ARGS)
  inside:
    kind: call_expression
```

Generate and run the appropriate `--inline-rules` block for the detected languages. Feed the JSON output (`.[].text`, `.[].file`, `.[].range`) into your review — inspect only the matched nodes for exploitability.

**If `ast-grep` is not installed:** proceed with full file read. Note at the start of your output: `ast-grep not found — running full read. For faster, more precise results: brew install ast-grep or npm i -g @ast-grep/cli`

## What to check

Trace untrusted data from where it enters to where it's used. Most vulnerabilities are an input that reaches a dangerous sink without validation in between.

**1. Injection.** Untrusted input reaching an interpreter — SQL/NoSQL, OS commands, file paths (traversal), URLs (SSRF), templates, `eval`-like calls, LDAP. Look for string-built queries or commands instead of parameterized/escaped APIs.

**2. Output handling.** Untrusted data rendered without context-correct encoding (XSS), unsafe deserialization of attacker-controlled data, content-type confusion.

**3. Access control.** Missing or wrong authorization on an action; object-level checks absent (IDOR — can user A reach user B's record?); privilege escalation; trusting a client-supplied role or id; an auth check bypassable by ordering or a missing branch.

**4. Secrets & crypto.** Hardcoded credentials, keys, or tokens; secrets in logs, errors, or responses; rolled-own or weak crypto; predictable randomness used for security (tokens, IDs, salts); missing encryption for sensitive data in transit or at rest.

**5. Sensitive-data exposure.** PII or secrets in logs, stack traces, or verbose errors returned to the caller; over-broad API responses; debug endpoints or stack traces reachable in production paths.

**6. Untrusted input trusted too far.** Mass assignment / binding attacker-controlled fields; unvalidated redirects; unsafe file upload (type, size, path); unbounded input enabling resource exhaustion (DoS) — allocation, recursion, regex backtracking.

**7. Configuration & transport.** Missing TLS or certificate validation, permissive CORS, missing security headers, default credentials, overly broad permissions or IAM.

**8. Dependencies.** Known-vulnerable or untrusted dependencies introduced by the change. (The language reviews run the deeper audit tooling — flag the obvious here.)

## Output

A prioritized findings list, ordered by exploitability × impact:

1. Critical — directly exploitable for code execution, data breach, or auth bypass
2. High — exploitable under realistic conditions, or a clear data-exposure path
3. Hardening — defense-in-depth gap, not directly exploitable on its own

For each finding:

- Anchor to `file:line`.
- State the weakness in one line, **describe the attack that exploits it** (the input and the sink), then the fix — prefer the standard safe API over hand-rolled escaping.
- Note your confidence and any assumption about what's trusted.

Rules:

- Exploitable findings over theoretical ones. Name a path from attacker-controlled input to impact; a "weakness" on a fully-trusted internal path is hardening at most.
- A short, high-confidence list beats a long speculative one.
- Report-only by default — the output is the list. With `--fix`, apply only the mechanical, unambiguous findings above and leave every judgment- or assumption-dependent one reported; then summarize what was applied vs. left.
