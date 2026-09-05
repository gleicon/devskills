---
name: ds-agent-review
description: "Security review of an AI agent system — tool permissions, prompt-injection exposure, memory isolation, approval gates, agent-to-agent trust, and cost limits. Reviews configuration and code you point it at, including a deployed system's config; static and read-only, it never probes a live target."
disable-model-invocation: true
---

When invoked, audit an **agent system** against one question: **what can an attacker make this agent do?** The subject is not ordinary application code — it is the configuration that decides what an autonomous system is permitted to do: tool manifests, MCP server definitions, system prompts, retrieval sources, memory stores, approval logic, and the wiring between agents. `/ds-security-review` traces untrusted data to a dangerous sink; here the sink is a *tool call*, and the interpreter is a model whose output is never trustworthy.

Report-only by default — no edits unless `--fix` is passed.

## Scope

**This skill reviews whatever you point it at. It is not about the repository it happens to be installed in.**

- Positional args are scope: files, directories, globs, a config path, an exported manifest. With no scope, review the agent-related files changed on the current branch.
- `--full` → review every agent surface in the current codebase instead of just the branch's changes.
- Freeform scope ("our support agent", "the MCP config", "the tool definitions") is interpreted reasonably.
- Scope may be **a production system's configuration** — an exported tool manifest, a deployed prompt, an `mcp.json` pulled off a server, an agent definition from a platform console. Review it exactly as you would source in a repo. You are reading artifacts, not touching a system.

**Boundaries, because the subject is often live:**

- This pass is **static and read-only**. Never invoke the agent under review, never send it crafted input, never call its tools, never connect to an MCP server to enumerate it, never authenticate to a production console. Read what you were given.
- Findings describe what an attacker *could* do. Do not demonstrate it against a running system. Live adversarial testing is a separate, authorized activity — see the abuse-case matrix below for what such a suite should cover, and hand it off rather than performing it here.
- If a config in scope contains real credentials, API keys, or customer data, that is itself a critical finding — report the location, never the value.

## What to find

**1. Tool permission and least privilege.** A tool granting more than its stated job needs: shell execution with a wildcard command list, a filesystem tool with no path allowlist, database access that is not read-scoped when the agent only reads, credentials shared across tools of different sensitivity. Check that a path or resource allowlist exists *and* that it cannot be escaped — `../` traversal, symlinks, glob expansion beyond the intended root, no block for `.env`/`*.key`/`*.pem`/secret-bearing names. Check that agents at different trust levels (internal vs. user-facing) do not share one tool set.

**2. Prompt injection exposure.** Trace every path by which content the operator does not control reaches the model's context: user messages, retrieved documents, web pages, emails, API responses, file contents, another agent's output, a tool's error string. Each is untrusted. Look for missing sanitization, no delimiting between instruction and data, and — the finding that matters most — untrusted content reaching a context that has a high-impact tool available. Indirect injection is the case teams miss: the attacker never talks to the agent, they plant text in a document the agent will read.

**3. Memory and context poisoning.** Data written to persistent memory without validation; memory shared across users or sessions with no isolation key; no expiry or size limit; no integrity check on long-term storage; sensitive data persisted unredacted. The attack is delayed and cross-tenant: content planted in one session steers a different user's session later.

**4. Autonomy and approval gates.** Actions classified by risk (read → write → external/financial → irreversible), and a gate that matches. Findings: no human approval on irreversible, financial, administrative, or externally visible actions; approval that is a bare yes/no not bound to the exact action; approval records missing actor, tool name, target resource, normalized parameters, timestamp, or expiry — all replayable or swappable; the agent both deciding and executing with no independent policy check between; no step-up authentication on critical actions; non-idempotent high-impact operations with no duplicate confirmation; no interrupt or rollback path; and a risk classifier, policy lookup, or audit write that **fails open**.

**5. Output handling.** Model output consumed without validation — parsed as a command, rendered as HTML, passed to a shell, used as a file path, or used to make an authorization decision. That last one is its own finding: authorization derived from model output is not authorization. Check for schema validation on structured output and for sensitive-data filtering before display.

**6. Cost and loop control.** No cap on tool-chain depth, recursion, retries, tokens, or spend per session/user. Unbounded loops are a denial-of-wallet attack — the damage is a bill, and it is reachable without any other compromise. Check circuit breakers exist and trip.

**7. Multi-agent trust.** Messages between agents accepted without validation, sanitization, or signature verification; a low-privilege agent able to reach a high-privilege one's tools through a chain; shared execution environment or shared memory across trust boundaries; no circuit breaker, so one compromised agent cascades. Treat another agent's output as untrusted input — it is.

**8. Data protection.** More sensitive data in context than the task needs; no classification or handling rules; unencrypted at rest or in transit; no retention or deletion policy; PII or credentials in plain-text logs; regulated data (GDPR/CCPA scope) with no stated basis.

**9. Supply chain.** Third-party MCP servers, plugins, tools, or retrieval sources pulled in unpinned or unvetted; a tool description that itself carries instructions to the model; an agent console or template that can be reconfigured by data it consumes.

**10. Observability.** Tool calls, decisions, and outcomes not logged; no structured decision metadata on high-risk actions (classification, authorization outcome, approval identifier, execution result, policy version); no anomaly detection on tool-invocation frequency, privilege use, or approval-bypass attempts; no per-session cost tracking. Absent logging is what makes every other finding undetectable in production.

## Abuse-case coverage

Check whether the system has a repeatable adversarial suite, and report the gaps as a table. The cases that must exist:

| Abuse case | What it must prove |
|---|---|
| Prompt override | system and developer instructions survive hostile user or retrieved content |
| Tool misuse | an unauthorized tool is denied even when the model requests it confidently |
| Privilege escalation | a low-trust session cannot reach privileged tools, credentials, or admin actions |
| Memory poisoning | malicious content is sanitized, scoped, expired, or rejected before it persists |
| Data exfiltration | sensitive context does not leak through tool calls, citations, logs, or output |
| Recursive tool abuse | depth, retry, token, and cost limits stop a runaway loop |
| Approval bypass | a high-impact action cannot run without a valid, unexpired, parameter-bound approval |
| Multi-agent chaining | one compromised agent cannot push another past its trust boundary |

Also flag: no adversarial run in CI on prompt, tool-policy, memory, retrieval, or model-provider changes; no regression test for a previously observed failure; secrets or live customer data in test fixtures; and **security tests weakened or deleted in the same change that alters agent behavior** — review that pattern as hostile until proven otherwise.

## Output

A prioritized findings list, ordered by exploitability × impact:

1. Critical — a path from attacker-controlled content to an irreversible, financial, administrative, or data-exfiltrating action
2. High — exploitable under realistic conditions, or a clear cross-user/data-exposure path
3. Hardening — defense-in-depth gap, not directly exploitable on its own

For each finding:

- Anchor to `file:line`, or to the config key and its source when the artifact has no line numbers.
- State the weakness in one line, **describe the attack** — where the attacker's content enters, which tool it reaches, what it achieves — then the fix.
- Note your confidence, and state what you assumed to be trusted. Say plainly when a control might exist outside the scope you were given; a missing gate in one file may be enforced in a layer you were not shown.

Rules:

- Exploitable over theoretical. Name the path from untrusted content to impact. A permissive tool an attacker cannot reach is hardening at most.
- Do not report the absence of a control the architecture makes unnecessary — say why it is unnecessary instead.
- A short, high-confidence list beats a long speculative one.
- With `--fix`: apply only findings whose fix is **mechanical and unambiguous** — narrowing a wildcard to a stated allowlist, adding a blocked-pattern entry, pinning a dependency version, removing a logged secret. Anything that changes what the agent is able to do, alters approval logic, or rests on an assumption you could not verify **stays report-only**: a wrong fix to an agent's permissions is worse than none. Never edit a live system's configuration. After applying, re-run any check already in the loop and revert anything that breaks or that touched more than intended. Close with what was applied and what was left.
