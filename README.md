# devskills

Installable skill package for Claude Code, OpenCode, Cursor, and VSCode Copilot. Opinionated defaults, composable language profiles, full dev workflow from specification to shipped product.

No magic. Files in the right directories. Prompts that encode real constraints.

## Install

```bash
git clone https://github.com/gleicon/devskills.git ~/.devskills
~/.devskills/install.sh
```

Skills copy to `~/.claude/commands/` and `~/.opencode/commands/`. External tools (GSD, RTK, tldt) install automatically if prerequisites are present.

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

### install.sh flags

```
--lang=<profile>     go | typescript | javascript | rust
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
--lang=<profile>     optional; stacks a language profile (go|typescript|javascript|rust)
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
/grill-me            # interview the spec; surface every unresolved decision branch
```

Use `--record` to log decisions to `DECISIONS.md`. Feed long reference docs through `/tldt` first to compress them before adding to context.

**3. Build a project with GSD**

GSD manages context rot across long builds by storing state in `.planning/` and using focused sub-agents per phase.

```
/spec  (devskills)   → produce SPEC.md with acceptance criteria
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
/tiger-style         # activate engineering constraints for the session
/go-review           # Go: idiomatic + security + Tiger Style
/go-review --no-tiger  # skip Tiger Style section
/ts-review           # TypeScript/Workers: strict, React, Cloudflare
/rust-review         # Rust: cargo geiger, unsafe counts, clippy, audit
/zoom-out            # map modules, callers, boundaries — useful before planning
/caveman-lite        # compress responses during iterative work (~25–35% reduction)
```

Full walkthrough: [docs/gsd-workflow.md](docs/gsd-workflow.md)

**Lite planning, without GSD**

If GSD is more machinery than you want, a smaller `.project/` workflow keeps just the parts worth keeping — a project description, a plan, and current state in plain files, so any session is safe to `/clear` or end:

```
/project-map         # scan the repo → .project/PROJECT.md (description + map)
/project-plan        # ordered tasks → .project/PLAN.md (feed it a goal, SPEC.md, or command output)
   ...you drive the work...
/project-checkpoint  # persist state → .project/PLAN.md (--handoff for a full .project/handoff.md)
/project-resume      # restore context from .project/PLAN.md (loads handoff.md only if fresh)
```

These are scribes, not pilots: they record what you decide, never steer architecture. Commit `.project/` as shared memory or add it to `.gitignore` for a local-only scratch space — the workflow doesn't rely on git. Walkthrough: [docs/project-workflow.md](docs/project-workflow.md) · use cases: [docs/project-recipes.md](docs/project-recipes.md)

## Skills

| Skill | Command | Description |
|-------|---------|-------------|
| Tiger Style | `/tiger-style` | TigerBeetle engineering constraints: safety, performance, experience |
| Caveman Lite | `/caveman-lite` | Compressed response mode (~25–35% token reduction) |
| Caveman Ultra | `/caveman-ultra` | Compressed response mode (~75–85% token reduction) |
| TLDT | `/tldt` | Extractive summary of context or a file — no LLM cost |
| Workflow | `/workflow` | Spec-to-ship orchestration using GSD |
| Project Map | `/project-map` | Scan the repo into `.project/PROJECT.md` |
| Project Plan | `/project-plan` | Ordered task roadmap in `.project/PLAN.md` |
| Project Checkpoint | `/project-checkpoint` | Persist state to `.project/PLAN.md` (`--handoff` for a full handoff) |
| Project Resume | `/project-resume` | Restore context from `.project/PLAN.md` |
| Spec | `/spec` | Convert a description into a structured specification |
| Code Quality Review | `/code-quality-review` | Strict maintainability audit: abstraction, sprawl, spaghetti |
| Doc Quality Review | `/doc-quality-review` | Strict docs audit: accuracy, dead links, bloat (`--comments` audits code comments) |
| Deslop | `/deslop` | Strip AI-generated slop from the branch diff |
| Go Review | `/go-review` | Go: idiomatic + security + Tiger Style (`--no-tiger` to skip style) |
| TS Review | `/ts-review` | TypeScript/Workers: strict mode, React, Cloudflare (`--no-tiger` to skip style) |
| Rust Review | `/rust-review` | Rust: geiger/unsafe, clippy, audit, Tiger Style (`--no-tiger` to skip style) |
| Bug Review | `/bug-review` | Language-agnostic correctness audit — hunts real bugs |
| Security Review | `/security-review` | Language-agnostic security audit — each finding names the attack |
| UI | `/ui` | UI mode: component/state discipline, design craft, a11y, Core Web Vitals |
| Explore | `/explore` | Lay out candidate approaches with trade-offs (`--web` for research) |
| Grill Me | `/grill-me` | Relentless plan interview (`--record` logs to DECISIONS.md) |
| Handoff | `/handoff` | Compact the conversation into a handoff doc |
| Zoom Out | `/zoom-out` | Step up a layer — map modules, callers, boundaries |
| TDD | `/tdd` | Test-first, one vertical slice at a time |
| Test | `/test` | Pragmatic testing mode — test by risk, not coverage |
| Test Quality Review | `/test-quality-review` | Strict test audit: is critical code well tested? |
| Debug | `/debug` | Root-cause a failure with the scientific method |
| Verify This | `/verify-this` | Prove a falsifiable claim with local before/after evidence |
| Write a Command | `/write-a-command` | Author a new devskills command in repo conventions |

Full per-command reference: [docs/commands.md](docs/commands.md). Worked, GSD-free workflows and examples: [docs/recipes.md](docs/recipes.md). Extended `/grill-me` playbook: [docs/grill-me.md](docs/grill-me.md). Tiger Style principles: [docs/tiger-style.md](docs/tiger-style.md).

## Project Setup

`setup.sh` builds your project's `AGENTS.md` from stacked, independently-managed blocks, and points `CLAUDE.md` at it with a single `@AGENTS.md` import — so Claude Code (which reads `CLAUDE.md`) and OpenCode (which reads `AGENTS.md`) share the same content with no duplication.

| Block | Flag | Contents |
|-------|------|----------|
| `base` | always | Universal engineering principles — think before coding, simplicity first, surgical changes, goal-driven execution, safe at the boundaries |
| `language` | `--lang=<x>` | Stack-specific idioms, toolchain, and review constraints |
| `concise` | `--concise` | Terse-response directive (caveman-lite behavior, baked in) |
| `tooling` | `--hints` | Reference list of devskills commands, tldt, and RTK |

Running `setup.sh` with no flags writes just the baseline. Each block lives between `<!-- BEGIN/END devskills:<id> -->` markers, so re-running is idempotent and swapping `--lang` replaces only that block. Existing `AGENTS.md`/`CLAUDE.md` files are backed up (sibling timestamped `.bak`) once, before any change — these are transient; delete them or keep them out of version control once you've confirmed the result.

`update.sh` refreshes the globally-installed skills, but not a project's `AGENTS.md` — the managed blocks are a point-in-time snapshot. To pull baseline or tooling changes into a project after an update, re-run `setup.sh` there (idempotent, so it just refreshes the blocks in place).

The baseline blocks target `AGENTS.md` (Claude Code and OpenCode). Cursor and VSCode Copilot have their own rule mechanisms — `--cursor` installs `.cursor/rules/*.mdc` and `--vscode` writes `copilot-instructions.md`; those paths carry Tiger Style and the language rules but not the `base`/`concise`/`tooling` blocks.

To back out, `setup.sh --uninstall` strips the devskills blocks (and removes a file that held *only* devskills content), leaving your own content untouched — a clean install→uninstall round-trip restores the originals exactly.

## Language Profiles

Each profile encodes idioms, toolchain defaults, and review constraints for its stack, and is stacked under the baseline as a `## Language Profile — <x>` section.

| Profile | Stack | Use case |
|---------|-------|---------|
| `go` | Go 1.22+ | Backend services, CLIs, APIs |
| `typescript` | TypeScript 5+, Wrangler | Cloudflare Workers, Next.js, React |
| `javascript` | ES2022+, Wrangler | Cloudflare Workers, vanilla frontend |
| `rust` | Rust stable | Systems programming, large projects |

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
| [Tiger Style](https://tigerstyle.dev/) | `/tiger-style`, all review skills |
| [Caveman](https://github.com/juliusbrussee/caveman) | `/caveman-lite`, `/caveman-ultra` |
| [mattpocock/skills](https://github.com/mattpocock/skills) | `/grill-me`, `/handoff`, `/zoom-out`, `/tdd`, `/write-a-command` |
| [cursor/plugins — cursor-team-kit](https://github.com/cursor/plugins/tree/main/cursor-team-kit/skills) | `/code-quality-review`, `/deslop`, `/verify-this` |
| [Andrej Karpathy](https://x.com/karpathy/status/2015883857489522876) · [multica-ai/andrej-karpathy-skills](https://github.com/multica-ai/andrej-karpathy-skills) | AGENTS.md baseline (`base` block) |

## Adding Skills

Drop a `.md` file into `claude/commands/`. The filename becomes the command name. The file content is the system prompt injected when the command runs. Copy to `opencode/commands/` for OpenCode parity.

For Cursor, drop a `.mdc` file into `cursor/rules/`. Use YAML frontmatter:
- `alwaysApply: true` — inject regardless of open file
- `globs: ["**/*.go"]` — inject only when matched file is open

System-level prompts (session preamble, not slash commands) go in `prompts/system/`. Reference them in `CLAUDE.md`.

## License

MIT
