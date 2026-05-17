# devskills

A curated, installable skill package for AI-powered development environments. Covers Claude Code, OpenCode, Cursor, and VSCode Copilot. Inspired by the LazyVim model: opinionated defaults, composable layers, and a clear path from specification to shipped product.

## Philosophy

This package standardizes the full development loop — specification, planning, implementation, review, and delivery — across language stacks and AI tools. It bundles proven external frameworks with first-class language profiles for Go, Rust, TypeScript, and JavaScript.

No magic. Files in the right directories. Prompts that encode real constraints.

## External Dependencies

The following are installed as separate tools. The install script handles this.

| Tool | Source | Purpose |
|------|--------|---------|
| GSD (Get Shit Done) | https://github.com/gsd-build/get-shit-done | Full dev lifecycle workflow: discuss, plan, execute, verify, ship |
| RTK | https://github.com/rtk-ai/rtk | CLI proxy; reduces AI context token consumption 60-90% |
| Caveman | https://github.com/juliusbrussee/caveman | Response compression; reduces output tokens ~65% |
| tldt | https://github.com/gleicon/tldt | Extractive text summarization; no LLM, no cost |
| Tiger Style | https://tigerstyle.dev/ | Engineering principles — safety, performance, developer experience |

## Included Skills

| Skill | Command | Description |
|-------|---------|-------------|
| Tiger Style | `/tiger-style` | Enforces TigerBeetle coding principles: safety, performance, experience |
| Caveman Lite | `/caveman-lite` | Compressed response mode — lite level (~35% token reduction) |
| Caveman Ultra | `/caveman-ultra` | Compressed response mode — ultra level (~80% token reduction) |
| TLDT | `/tldt` | Summarize context or file with extractive techniques, no LLM cost |
| Workflow | `/workflow` | Spec-to-ship orchestration using GSD |
| Go Review | `/go-review` | Go code review: Tiger Style + idiomatic Go + security |
| TS Review | `/ts-review` | TypeScript/Workers review: strict mode, React, Cloudflare |
| Rust Review | `/rust-review` | Rust review: cargo geiger, unsafe counts, clippy, audit |
| Frontend | `/frontend` | Frontend task mode: components, state, API integration, a11y |
| Spec | `/spec` | Convert a description into a verifiable structured specification |

## Language Profiles

Each profile encodes idioms, toolchain defaults, and review constraints for its stack.

| Profile | Stack | Use case |
|---------|-------|---------|
| `go` | Go 1.22+ | Backend services, CLIs, APIs |
| `typescript` | TypeScript 5+, Wrangler | Cloudflare Workers, Next.js, React |
| `javascript` | ES2022+, Wrangler | Cloudflare Workers, vanilla frontend |
| `rust` | Rust stable | Systems programming, experimental large projects |

## Directory Structure

```
devskills/
├── README.md
├── PUBLISHING.md             # npm publish, GitHub releases, CI
├── package.json              # npm package
├── install.sh                # shell installer (--dry-run, --skip-external, --lang, --claude-dir)
├── claude/commands/          # Claude Code skills (10 .md files)
│   ├── tiger-style.md
│   ├── caveman-lite.md
│   ├── caveman-ultra.md
│   ├── tldt.md
│   ├── workflow.md
│   ├── spec.md
│   ├── go-review.md
│   ├── ts-review.md
│   ├── rust-review.md
│   └── frontend.md
├── opencode/commands/        # OpenCode skills (same files)
├── cursor/rules/             # Cursor rules (auto-activate by file glob)
│   ├── tiger-style.mdc       # alwaysApply: true
│   ├── go.mdc                # *.go
│   ├── typescript.mdc        # *.ts, *.tsx
│   ├── javascript.mdc        # *.js, *.mjs
│   └── rust.mdc              # *.rs
├── vscode/
│   └── copilot-instructions.md
├── prompts/
│   ├── language/             # go, typescript, javascript, rust
│   └── system/               # specification.md
├── scripts/
│   ├── setup.sh              # per-project configurator
│   └── update.sh             # pull + reinstall
└── docs/
    ├── tiger-style.md        # Tiger Style rationale and full rules
    ├── workflow.md           # GSD artifact structure reference
    └── gsd-workflow.md       # spec → GSD phases walkthrough
```

## Installation

### Clone and install

```bash
git clone https://github.com/gleicon/devskills.git ~/.devskills
~/.devskills/install.sh
```

That's it. Skills are copied to `~/.claude/commands/` and `~/.opencode/commands/`. External tools (GSD, RTK, Caveman, tldt) are installed automatically if their prerequisites are present.

Skip external tools:

```bash
~/.devskills/install.sh --skip-external
```

Custom Claude config dir:

```bash
~/.devskills/install.sh --claude-dir=~/.config/claude
CLAUDE_CONFIG_DIR=~/.config/claude ~/.devskills/install.sh
```

### Per-project language profile

Run from inside a project directory:

```bash
~/.devskills/scripts/setup.sh --lang=go
~/.devskills/scripts/setup.sh --lang=typescript --cursor --vscode
~/.devskills/scripts/setup.sh --lang=rust --cursor
```

The installer writes to:
- `~/.claude/commands/` — Claude Code user-level skills (override with `--claude-dir`)
- `~/.opencode/commands/` — OpenCode user-level skills
- `.cursor/rules/` — Cursor rules (project-local)
- `.github/copilot-instructions.md` — VSCode Copilot (project-local)

### Flags

```
--lang=<profile>     go | typescript | javascript | rust
--claude-dir=<path>  Claude config dir (default: $CLAUDE_CONFIG_DIR or ~/.claude)
--skip-external      skip GSD, RTK, caveman, tldt installation
--cursor             install Cursor rules into current project
--vscode             install VSCode Copilot instructions into current project
--dry-run            show what would happen, write nothing
```

### Publishing

See [PUBLISHING.md](PUBLISHING.md) for npm publish, GitHub releases, and CI automation.

## Tiger Style

Tiger Style is TigerBeetle's engineering philosophy: safety first, performance second, developer experience third. It is the opinionated foundation for all code generated or reviewed in this package.

Source: https://tigerstyle.dev/  
Full reference: TigerBeetle TIGER_STYLE.md (https://github.com/tigerbeetle/tigerbeetle/blob/main/docs/TIGER_STYLE.md)

Key constraints enforced:
- Assertions: minimum 2 per function (arguments, return values, invariants)
- No dynamic memory allocation after initialization
- No recursion unless termination is formally proven
- All loops and queues have explicit upper bounds
- Zero dependencies policy
- Functions capped at 70 lines
- Variable names include units and qualifiers
- Static analysis must pass clean before review

## Workflow: Specification to Product

GSD (Get Shit Done) is the execution engine. Install it:

```bash
npx get-shit-done-cc@latest
```

Full walkthrough — including how to take a spec through phases: [docs/gsd-workflow.md](docs/gsd-workflow.md)

Quick path:

```
/spec              → produce SPEC.md with verifiable acceptance criteria
/gsd-new-project   → initialize .planning/ from SPEC.md, build ROADMAP.md
/gsd-discuss-phase → capture decisions and constraints before planning
/gsd-plan-phase    → produce PLAN.md with numbered tasks
/gsd-execute-phase → implement in parallel sub-agents (context stays lean)
/gsd-verify-work   → validate against spec acceptance criteria
/gsd-ship          → create PR from verified work
```

Between phases, use devskills skills to keep quality high:

```
/tiger-style       → activate style enforcement
/go-review         → review Go code before verify
/rust-review       → review Rust code (runs cargo geiger, counts unsafe/unwrap)
/ts-review         → review TypeScript/Workers code
/tldt              → compress long files before feeding them to context
/caveman-lite      → reduce response verbosity during iterative work
```

See `docs/workflow.md` for the GSD artifact structure and multi-session guidance.

## Adding Skills

Drop a `.md` file into `claude/commands/`. The filename becomes the command name. The file content is the system prompt injected when the command runs. Copy to `opencode/commands/` for OpenCode parity.

For Cursor, drop a `.mdc` file into `cursor/rules/`. Use YAML frontmatter:
- `alwaysApply: true` — inject regardless of open file
- `globs: ["**/*.go"]` — inject only when the matched file is open
- `description:` — shown in Cursor's rule list

System-level prompts (for session preamble, not slash commands) go in `prompts/system/`. Reference them in `CLAUDE.md` or inject manually at session start.

## License

MIT
