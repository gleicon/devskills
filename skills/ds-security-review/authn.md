# authn — the authentication pass for `/ds-security-review`

Read this when the scope touches **login, registration, password reset, password change,
email/phone change, session issuance, MFA enrolment, or an OAuth/OIDC/SAML integration**.
Nothing here replaces the main pass — it adds the checks that only apply to identity code,
where the same generic weaknesses (injection, access control) sit next to failure modes that
have no analogue elsewhere: a response *time* that leaks account existence, a lockout counter
that becomes a denial-of-service tool, a reset flow that authenticates nothing.

Every check below is a condition to look for, the attack it enables, and the fix. As in the
main pass, a match is a lead — confirm it against the code before reporting it.

Thresholds cite NIST SP 800-63B via the OWASP Authentication Cheat Sheet
(<https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html>). Report the
numbers exactly; a policy that is off by a factor is a finding, "weak-ish" is not.

## 1. Identifiers

- **Sequential or guessable user IDs** exposed in URLs, tokens, or API responses. Attacker
  enumerates the whole user table by counting. Fix: random IDs; a sequential primary key stays
  internal and never leaves the boundary.
- **A backend/service/database account that can log in through the user-facing UI**, or one
  identity provider serving both internal and public access. Compromising the public surface
  reaches internal systems directly. Fix: separate the authentication solution by trust level.

## 2. Password policy

- **Minimum length below 8 with MFA enforced, or below 15 without it.** Fix: raise the minimum;
  state which of the two applies to the code in scope.
- **Maximum length below 64**, or a maximum enforced by silent truncation. Truncation means the
  stored secret is shorter than the user believes. Fix: accept at least 64 characters; reject
  over-length input explicitly rather than cutting it.
- **Composition rules** — required uppercase, digits, or symbols; rejected Unicode, spaces, or
  printable characters. These shrink the search space and push users to predictable patterns.
  Fix: length and blocklist checks instead.
- **No breach/common-password blocklist.** Fix: screen new passwords against a breached-password
  set (Pwned Passwords, or a self-hosted list). This is the control that stops credential
  stuffing at registration time.
- **Forced periodic rotation** with no compromise trigger. Fix: rotate on evidence of compromise
  or an authenticator change, not on a calendar.
- **No maximum on the input length reaching the hash function.** A megabyte password is a CPU
  exhaustion vector against a deliberately slow KDF.

## 3. Verification

- **Password comparison that is not constant-time** — `==`, `strcmp`, an early-exit loop over
  the hash. Fix: the framework's verifier, or a constant-time compare.
- **Untyped or loosely-typed comparison** of a hash against user input, where the language
  coerces (PHP magic hashes, JavaScript `==`). Fix: pin both operand types before comparing.

## 4. Enumeration — text, status, and time

Three separate leaks. Check all three; the first is the one everyone fixes and the third is the
one that survives.

- **Message discrepancy.** "Invalid password" vs "no such user"; "we sent you a reset link" vs
  "that email isn't in our database"; "this user ID is already in use" on registration. Fix:
  one generic response per flow — `Sign-in failed. Check your username and password.`, `If an
  account exists for that address, a reset link is on its way.`, `Check your inbox for an
  activation link.`
- **Status-code or redirect discrepancy** behind an identical page — 200 vs 403, a different
  `Location`, a different error code in a JSON envelope. The body being generic does not matter
  if the envelope is not.
- **Timing discrepancy from a quick-exit branch.** This is a control-flow finding, not a string
  finding. When the hash runs only inside the "user exists" branch, the response time answers
  the question the error message refused to: the expensive work happens for real accounts and
  is skipped for the rest. Fix: look up the record first, then run the hash on every request —
  against the stored hash when there is one and a fixed dummy hash when there isn't — so both
  paths cost the same.

Where a generic message is unacceptable for usability, the compensating control is rate
limiting plus CAPTCHA — note that trade-off in the finding rather than reporting nothing.

## 5. Automated-attack resistance

Name which attack the code fails to stop; they need different controls.

| Attack | Shape |
|---|---|
| Brute force | many passwords against one account |
| Credential stuffing | breached username/password pairs against many accounts |
| Password spraying | one weak password against many accounts |

- **No lockout or throttling at all** on the login, reset, or MFA-verification endpoints.
- **Failure counter keyed by source IP rather than by account.** A distributed attacker never
  trips it. Fix: bind the counter to the account. (Keep the IP counter as a second control if
  present — it catches a different case.)
- **Lockout with no exit** — no observation window, no expiry, or permanent lockout on a low
  threshold. This *is* the denial of service: an attacker locks out every account they can name.
  Fix: threshold, observation window, and lockout duration all defined; consider exponential
  backoff starting near one second; leave a recovery path (forgotten-password) usable while
  locked out.
- **CAPTCHA relied on as the primary defense.** It is defense in depth — solvable and
  outsourceable. Fine after N failures, not instead of throttling.
- **Security questions treated as a second factor.** Both factors are something-you-know, so it
  is not MFA; report any code path that grants MFA-equivalent trust from one.

## 6. Sensitive actions and re-authentication

- **A password change that does not verify the current password.** Abuse case: a session left
  open on a shared machine is a permanent account takeover. Fix: require the current password
  even with a valid session.
- **No re-authentication before** changing email, password, or payment details; adding a trusted
  device; or any high-value transaction. Without it, CSRF or XSS is enough to perform the action
  without ever knowing the credential.
- **No re-authentication after a risk event** — account recovery, a password reset, a new device
  or a changed IP, or flagged suspicious activity.
- **Sessions not invalidated and tokens not rotated after re-authentication or a password
  change.** A stolen session survives the remediation that was supposed to end it.

## 7. Email/phone-address change

The full flow, because a partial one is an account-takeover path:

- The change is stored as **pending**, never applied on submission.
- Identity is re-proven first — **MFA when enrolled, current password when not**.
- **Time-limited nonces**, one per message.
- **Two messages**: a confirmation-required message to the new address, and a message to the
  current address — notification-only when MFA proved identity, confirmation-required when only
  a password did. The current-address message always carries a "this wasn't me" link.

Missing current-address notification is the finding that matters most: it makes the takeover
silent.

## 8. Transport and protocol integration

- **The login page itself served over anything but TLS** — not just the POST target. An attacker
  who can rewrite the page rewrites the form action, and the credentials post elsewhere.
- **Any authenticated page over plaintext** after login — the session ID is then observable.
- **OIDC relying party not validating the ID token**: issuer (`iss`), audience (`aud`),
  signature against the provider's JWKS, and expiry (`exp`). A missing `aud` check accepts a
  token minted for a different application; a missing signature check accepts anything.
- **A new OpenID 2.0 implementation** — obsolete, superseded by OIDC.
- **Credentials handed to a third-party application** to store and replay, where a delegated
  protocol (OAuth 2.0/2.1) applies.

## 9. Logging

- **Authentication failures, password failures, and account lockouts not logged.** Without them
  no spraying campaign is detectable. Fix: log all three.
- The mirror image, and the more common finding: **passwords, tokens, reset nonces, or session
  IDs written into those logs**. Log the event and the account, never the secret.

## 10. Password-manager hostility

Low severity, real consequence — users pick weaker passwords when the manager cannot help:
blocked paste on password or MFA fields, a maximum length under 64, non-standard input widgets
instead of `<input type="password">`, or a broken Tab order between username and password.

## Reporting

Fold findings into the main pass's severity ordering rather than a separate list. Typical
placement: silent takeover paths and missing verification on sensitive actions are critical;
enumeration and missing throttling are high; password-manager hostility and logging gaps are
hardening. Every finding still names the attack.
