## devskills Tooling

This project was configured with devskills. Reference for the available helpers:

**Slash commands** (invoked in Claude Code / OpenCode):
- Modes: `/tiger-style`, `/ui` — engineering / UI discipline for the session
- Reviews: `/code-quality-review`, `/bug-review`, `/security-review`, `/test-quality-review`, `/doc-quality-review`, `/go-review`, `/ts-review`, `/rust-review`
- Build & verify: `/spec`, `/tdd`, `/test`, `/debug`, `/verify-this`, `/deslop`
- Plan & context: `/explore`, `/grill-me`, `/project-map`, `/project-plan`, `/project-checkpoint`, `/project-resume`, `/zoom-out`, `/handoff`, `/workflow`

**Token-saving tools:**
- `/caveman-lite` (~25–35%) or `/caveman-ultra` (~75–85%) — compress responses
- `/tldt [file|url]` — extractive summarization, no LLM cost; compress long docs before adding to context (uses the `tldt` CLI when installed)
- `rtk` — transparent CLI proxy that cuts token use 60–90% on dev commands

Run `devskills list` to see everything available.
