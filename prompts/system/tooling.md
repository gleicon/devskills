## devskills Tooling

This project was configured with devskills. Reference for the available helpers:

**Slash commands** (invoked in Claude Code / OpenCode):
- `/tiger-style` — activate engineering constraints for the session
- `/code-quality-review`, `/go-review`, `/ts-review`, `/rust-review` — targeted code reviews
- `/spec`, `/workflow`, `/tdd`, `/grill-me`, `/handoff`, `/zoom-out` — workflow skills

**Token-saving tools:**
- `/caveman-lite` (~35%) or `/caveman-ultra` (~80%) — compress responses
- `/tldt [file|url]` — extractive summarization, no LLM cost; compress long docs before adding to context (uses the `tldt` CLI when installed)
- `rtk` — transparent CLI proxy that cuts token use 60–90% on dev commands

Run `devskills list` to see everything available.
