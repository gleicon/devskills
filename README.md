# devskills

Installable skill package for Claude Code, OpenCode, OpenAI Codex, Cursor, and VSCode Copilot. Opinionated defaults, composable language profiles, full dev workflow from specification to shipped product.

No magic. Files in the right directories. Prompts that encode real constraints.

## Install

```bash
git clone https://github.com/gleicon/devskills.git ~/.devskills
~/.devskills/install.sh
```

Skills copy to `~/.claude/commands/`, `~/.opencode/commands/`, and `~/.codex/prompts/` (each installed only when that tool is detected). External tools (GSD, RTK, tldt) install automatically if prerequisites are present.

In Codex, devskills commands are invoked under the `prompts:` namespace — `/ds-debug` becomes `/prompts:ds-debug`. Codex reads a project's `AGENTS.md` natively, so `setup.sh` covers its persistent surface with no extra step.

Skip external tools:

```bash
~/.devskills/install.sh --skip-external
```

Custom Claude config dir:

```bash
~/.devskills/install.sh --claude-dir=~/.config/claude
```

Per-project setup (run from inside a project):

```bash
~/.devskills/scripts/setup.sh                        # baseline AGENTS.md only
~/.devskills/scripts/setup.sh --lang=go              # baseline + Go profile
~/.devskills/scripts/setup.sh --lang=typescript --cursor --vscode
```

`setup.sh` writes a universal engineering baseline to `AGENTS.md` and points `CLAUDE.md` at it via `@AGENTS.md`; `--lang` stacks a language profile on top. See [Project Setup](#project-setup) below.

Keep devskills up to date:

```bash
~/.devskills/scripts/update.sh              # pull + reinstall skills
~/.devskills/scripts/update.sh --upgrade-deps  # also force-upgrade GSD, RTK, tldt
```

> **Renamed in this release:** every command is now namespaced with a `ds-`
> prefix (e.g. `/debug` → `/ds-debug`), and modes carry a `-mode` suffix
> (e.g. `/tiger-style` → `/ds-tiger-style-mode`). This avoids collisions with
> Claude Code built-ins like `/debug` and `/security-review`, and makes a
> command's kind readable at a glance. Re-running `install.sh` or `update.sh`
> drops in the new names and removes the old files automatically — no manual
> cleanup. See the [Skills](#skills) table for the current names.

### install.sh flags

```
--lang=<profile>     go | typescript | javascript | rust | python
--claude-dir=<path>  Claude config dir (default: $CLAUDE_CONFIG_DIR or ~/.claude)
--skip-external      skip GSD, RTK, tldt installation
--skip-cursor        skip Cursor rules
--skip-vscode        skip VSCode Copilot instructions
--concise            add a terse-response directive to AGENTS.md (with --lang)
--hints              add a devskills tooling reference to AGENTS.md (with --lang)
--dry-run            show what would happen, write nothing
```

### setup.sh flags (per-project)

```
--lang=<profile>     optional; stacks a language profile (go|typescript|javascript|rust|python)
--concise            add a terse-response directive to AGENTS.md
--hints              add a devskills tooling reference to AGENTS.md
--cursor             install Cursor rules into current project
--vscode             install VSCode Copilot instructions into current project
--dry-run            show what would happen, write nothing
```

## Workflow

The intended flow moves from a rough idea to a shipped, reviewed product. Each stage has a skill or tool.

**1. Write a spec**

Use any external AI (ChatGPT, Claude.ai, Gemini) to draft a product description. Bring it into your project directory.

**2. Interrogate the spec**

```
/ds-grill-me            # interview the spec; surface every unresolved decision branch
```

Use `--record` to log decisions to `DECISIONS.md`. Feed long reference docs through `/ds-tldt` first to compress them before adding to context.

**3. Build a project with GSD**

GSD manages context rot across long builds by storing state in `.planning/` and using focused sub-agents per phase.

```
/ds-spec  (devskills)   → produce SPEC.md with acceptance criteria
/gsd-new-project     → initialize .planning/, build ROADMAP.md from SPEC.md
/gsd-discuss-phase   → capture decisions and constraints before planning
/gsd-plan-phase      → produce PLAN.md with numbered tasks
/gsd-execute-phase   → implement; sub-agents stay context-lean
/gsd-verify-work     → validate against acceptance criteria
/gsd-ship            → create PR from verified work
```

Note: `/gsd-*` commands are provided by GSD Redux, not devskills. Install GSD separately (`npx @opengsd/get-shit-done-redux@latest`) or let `install.sh` handle it.

**4. Keep quality high between phases**

```
/ds-tiger-style-mode         # activate engineering constraints for the session
/ds-go-review           # Go: idiomatic + security + Tiger Style
/ds-go-review --no-tiger  # skip Tiger Style section
/ds-ts-review           # TypeScript/Workers: strict, React, Cloudflare
/ds-rust-review         # Rust: cargo geiger, unsafe counts, clippy, audit
/ds-zoom-out            # map modules, callers, boundaries — useful before planning
/ds-caveman-lite-mode        # compress responses during iterative work (~25–35% reduction)
```

Full walkthrough: [docs/gsd-workflow.md](docs/gsd-workflow.md)

**Lite planning, without GSD**

If GSD is more machinery than you want, a smaller `.project/` workflow keeps just the parts worth keeping — a project description, a plan, and current state in plain files, so any session is safe to `/clear` or end:

```
/ds-project-map         # scan the repo → .project/PROJECT.md (description + map)
/ds-project-plan        # ordered tasks → .project/PLAN.md (feed it a goal, SPEC.md, or command output)
   ...you drive the work...
/ds-project-checkpoint  # persist state → .project/PLAN.md (--handoff for a full .project/handoff.md)
/ds-project-resume      # restore context from .project/PLAN.md (loads handoff.md only if fresh)
```

These are scribes, not pilots: they record what you decide, never steer architecture. Commit `.project/` as shared memory or add it to `.gitignore` for a local-only scratch space — the workflow doesn't rely on git. Walkthrough: [docs/project-workflow.md](docs/project-workflow.md) · use cases: [docs/project-recipes.md](docs/project-recipes.md)

## Skills

Every command is namespaced with a `ds-` prefix (short for devskills) so it
never collides with a Claude Code or OpenCode built-in. **Modes** — which turn
on persistent session behavior — also carry a `-mode` suffix, so a name tells
you both its origin and its kind: `/ds-tiger-style-mode` is a mode you toggle,
`/ds-bug-review` is an action that runs once and finishes.

| Skill | Command | Description |
|-------|---------|-------------|
| Tiger Style | `/ds-tiger-style-mode` | TigerBeetle engineering constraints: safety, performance, experience |
| Caveman Lite | `/ds-caveman-lite-mode` | Compressed response mode (~25–35% token reduction) |
| Caveman Ultra | `/ds-caveman-ultra-mode` | Compressed response mode (~75–85% token reduction) |
| TLDT | `/ds-tldt` | Extractive summary of context or a file — no LLM cost |
| Workflow | `/ds-workflow` | Spec-to-ship orchestration using GSD |
| Project Map | `/ds-project-map` | Scan the repo into `.project/PROJECT.md` |
| Project Plan | `/ds-project-plan` | Ordered task roadmap in `.project/PLAN.md` |
| Project Checkpoint | `/ds-project-checkpoint` | Persist state to `.project/PLAN.md` (`--handoff` for a full handoff) |
| Project Resume | `/ds-project-resume` | Restore context from `.project/PLAN.md` |
| Spec | `/ds-spec` | Convert a description into a structured specification |
| Code Quality Review | `/ds-code-quality-review` | Strict maintainability audit: abstraction, sprawl, spaghetti |
| Doc Quality Review | `/ds-doc-quality-review` | Strict docs audit: accuracy, dead links, bloat (`--comments` audits code comments) |
| Deslop | `/ds-deslop` | Strip AI-generated slop from the branch diff |
| Go Review | `/ds-go-review` | Go: idiomatic + security + Tiger Style (`--no-tiger` to skip style) |
| TS Review | `/ds-ts-review` | TypeScript/Workers: strict mode, React, Cloudflare (`--no-tiger` to skip style) |
| Rust Review | `/ds-rust-review` | Rust: geiger/unsafe, clippy, audit, Tiger Style (`--no-tiger` to skip style) |
| Python Review | `/ds-python-review` | Python: idioms, typing, security, Tiger Style (`--no-tiger` to skip style) |
| Bug Review | `/ds-bug-review` | Language-agnostic correctness audit — hunts real bugs |
| Security Review | `/ds-security-review` | Language-agnostic security audit — each finding names the attack |
| UI | `/ds-ui-mode` | UI mode: component/state discipline, design craft, a11y, Core Web Vitals |
| UI Quality Review | `/ds-ui-quality-review` | Strict UI audit: async-state/fetch correctness, a11y, Core Web Vitals, design craft |
| Explore | `/ds-explore` | Lay out candidate approaches with trade-offs (`--web` for research) |
| Grill Me | `/ds-grill-me` | Relentless plan interview (`--record` logs to DECISIONS.md) |
| Handoff | `/ds-handoff` | Compact the conversation into a handoff doc |
| Zoom Out | `/ds-zoom-out` | Step up a layer — map modules, callers, boundaries |
| TDD | `/ds-tdd-mode` | Test-first, one vertical slice at a time |
| Test | `/ds-test-mode` | Pragmatic testing mode — test by risk, not coverage |
| Test Quality Review | `/ds-test-quality-review` | Strict test audit: is critical code well tested? |
| Debug | `/ds-debug` | Root-cause a failure with the scientific method |
| Verify This | `/ds-verify-this` | Prove a falsifiable claim with local before/after evidence |
| Quality Gate | `/ds-quality-gate-mode` | Six-pass review pipeline (deslop → test → security → bug → quality → docs), implement fixes between passes, toggleable mode |
| Write a Command | `/ds-write-a-command` | Author a new devskills command in repo conventions |

Full per-command reference: [docs/commands.md](docs/commands.md). Worked, GSD-free workflows and examples: [docs/recipes.md](docs/recipes.md). Extended `/ds-grill-me` playbook: [docs/grill-me.md](docs/grill-me.md). Tiger Style principles: [docs/tiger-style.md](docs/tiger-style.md).

## Project Setup

`setup.sh` builds your project's `AGENTS.md` from stacked, independently-managed blocks, and points `CLAUDE.md` at it with a single `@AGENTS.md` import — so Claude Code (which reads `CLAUDE.md`) and OpenCode and OpenAI Codex (which read `AGENTS.md` directly) share the same content with no duplication.

| Block | Flag | Contents |
|-------|------|----------|
| `base` | always | Universal engineering principles — think before coding, simplicity first, surgical changes, goal-driven execution, safe at the boundaries |
| `language` | `--lang=<x>` | Stack-specific idioms, toolchain, and review constraints |
| `concise` | `--concise` | Terse-response directive (caveman-lite behavior, baked in) |
| `tooling` | `--hints` | Reference list of devskills commands, tldt, and RTK |

Running `setup.sh` with no flags writes just the baseline. Each block lives between `<!-- BEGIN/END devskills:<id> -->` markers, so re-running is idempotent and swapping `--lang` replaces only that block. Existing `AGENTS.md`/`CLAUDE.md` files are backed up (sibling timestamped `.bak`) once, before any change — these are transient; delete them or keep them out of version control once you've confirmed the result.

`update.sh` refreshes the globally-installed skills, but not a project's `AGENTS.md` — the managed blocks are a point-in-time snapshot. To pull baseline or tooling changes into a project after an update, re-run `setup.sh` there (idempotent, so it just refreshes the blocks in place).

The baseline blocks target `AGENTS.md` (Claude Code, OpenCode, and OpenAI Codex). Cursor and VSCode Copilot have their own rule mechanisms — `--cursor` installs `.cursor/rules/*.mdc` and `--vscode` writes `copilot-instructions.md`. Both honor `--lang`: they carry Tiger Style plus the notes for the selected language only (no `--lang` writes Tiger Style alone), but not the `base`/`concise`/`tooling` blocks.

To back out, `setup.sh --uninstall` strips the devskills blocks (and removes a file that held *only* devskills content), leaving your own content untouched — a clean install→uninstall round-trip restores the originals exactly.

## Language Profiles

Each profile encodes idioms, toolchain defaults, and review constraints for its stack, and is stacked under the baseline as a `## Language Profile — <x>` section.

| Profile | Stack | Use case |
|---------|-------|---------|
| `go` | Go 1.22+ | Backend services, CLIs, APIs |
| `typescript` | TypeScript 5+, Wrangler | Cloudflare Workers, Next.js, React |
| `javascript` | ES2022+, Wrangler | Cloudflare Workers, vanilla frontend |
| `rust` | Rust stable | Systems programming, large projects |
| `python` | Python 3.11+ | Backend services, APIs, CLIs, data pipelines |

## Scripts

`install.sh` — one-time global install. Copies skills to Claude Code and OpenCode config dirs. Installs external tools. Run from anywhere.

`scripts/setup.sh` — per-project configurator. Builds `AGENTS.md` (engineering baseline + optional language/concise/tooling blocks) and points `CLAUDE.md` at it via `@AGENTS.md`, optionally installs Cursor rules and VSCode Copilot instructions into the current directory. Run from inside a project. See [Project Setup](#project-setup).

`scripts/update.sh` — pulls the latest devskills repo and reinstalls skills. Use `--upgrade-deps` to also force-upgrade GSD, RTK, and tldt to their latest published versions.

`scripts/upgrade-deps.sh` — force-upgrades external tools regardless of current state. Useful after upstream major version bumps.

## External Tools

Installed by `install.sh`. Managed by `upgrade-deps.sh`.

| Tool | Purpose |
|------|---------|
| [GSD Redux](https://github.com/open-gsd/get-shit-done-redux) | Full dev lifecycle: discuss, plan, execute, verify, ship |
| [RTK](https://github.com/rtk-ai/rtk) | CLI proxy; reduces AI context token use 60-90% |
| [tldt](https://github.com/gleicon/tldt) | Extractive text summarization; no LLM, no cost |

## References

devskills ships its own prompt commands based on these upstream sources.

| Reference | Used by |
|-----------|---------|
| [Tiger Style](https://tigerstyle.dev/) | `/ds-tiger-style-mode`, all review skills |
| [Caveman](https://github.com/juliusbrussee/caveman) | `/ds-caveman-lite-mode`, `/ds-caveman-ultra-mode` |
| [mattpocock/skills](https://github.com/mattpocock/skills) | `/ds-grill-me`, `/ds-handoff`, `/ds-zoom-out`, `/ds-tdd-mode`, `/ds-write-a-command` |
| [cursor/plugins — cursor-team-kit](https://github.com/cursor/plugins/tree/main/cursor-team-kit/skills) | `/ds-code-quality-review`, `/ds-deslop`, `/ds-verify-this` |
| [Andrej Karpathy](https://x.com/karpathy/status/2015883857489522876) · [multica-ai/andrej-karpathy-skills](https://github.com/multica-ai/andrej-karpathy-skills) | AGENTS.md baseline (`base` block) |

## Adding Skills

Drop a `.md` file into `commands/`. The filename becomes the command name. The file content is the system prompt injected when the command runs. `install.sh` copies it to `~/.claude/commands/`, `~/.opencode/commands/`, and `~/.codex/prompts/` (the latter invoked as `/prompts:<name>`).

For Cursor, drop a `.mdc` file into `cursor/rules/`. Use YAML frontmatter:
- `alwaysApply: true` — inject regardless of open file
- `globs: ["**/*.go"]` — inject only when matched file is open

System-level prompts (session preamble, not slash commands) go in `prompts/system/`. Reference them in `CLAUDE.md`.

## License

MIT — see [LICENSE](LICENSE).
