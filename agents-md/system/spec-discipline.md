## Spec Amendment Discipline

SPEC.md and GRILL.md are decision records. When implementation reality corrects a decision, amend the file inline — never silently rewrite it.

- Mark every change with a dated `Amended YYYY-MM-DD:` line next to the decision it corrects, stating the new decision and the reason.
- Commit amendments as dedicated `docs(spec):` / `docs(grill):` commits — never folded into feature commits.
- No changelog sections in SPEC.md or GRILL.md; the dated inline amendments and git history are the record.
- Never delete or reword a decision without an amendment line — `/ds-retro` audits the release range and reports undisciplined edits as findings.
