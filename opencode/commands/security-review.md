Run a strict security review of code changes — find exploitable weaknesses. Language-agnostic. Reports a findings list; changes nothing.

When invoked, audit the code in scope against one question: **how would an attacker abuse this?** Look for the weaknesses that lead to real compromise — injection, broken access control, leaked secrets, untrusted input trusted too far. This is the portable, cross-language pass; for deeper language specifics, `/go-review`, `/ts-review`, and `/rust-review` carry their own Security sections. Every finding names the attack that exploits it. **Do not edit any files.**

## Arguments

- Treat positional args as scope (files, directories, globs). With no scope, review the code changed on the current branch.
- Freeform scope ("the auth handler", "the upload path") is interpreted reasonably.

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
- Change nothing. The output is the list.
