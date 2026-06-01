## devskills Tooling

This project was configured with devskills. Reference for the available helpers:

**Slash commands** (invoked in Claude Code / OpenCode):
- Modes: `/ds-tiger-style-mode`, `/ds-ui-mode` — engineering / UI discipline for the session (`/ds-modes` multi-selects several at once)
- Reviews: `/ds-code-quality-review`, `/ds-bug-review`, `/ds-security-review`, `/ds-test-quality-review`, `/ds-doc-quality-review`, `/ds-go-review`, `/ds-ts-review`, `/ds-rust-review` (`/ds-review` picks one language-agnostic review to run)
- Build & verify: `/ds-spec`, `/ds-tdd-mode`, `/ds-test-mode`, `/ds-debug`, `/ds-verify-this`, `/ds-deslop`
- Plan & context: `/ds-explore`, `/ds-grill-me`, `/ds-project-map`, `/ds-roadmap`, `/ds-project-checkpoint`, `/ds-project-resume`, `/ds-zoom-out`, `/ds-handoff`, `/ds-workflow`

**Token-saving tools:**
- `/ds-caveman-lite-mode` (~25–35%) or `/ds-caveman-ultra-mode` (~75–85%) — compress responses
- `/ds-tldt [file|url]` — extractive summarization, no LLM cost; compress long docs before adding to context (uses the `tldt` CLI when installed)
- `rtk` — transparent CLI proxy that cuts token use 60–90% on dev commands

Run `devskills list` to see everything available.
