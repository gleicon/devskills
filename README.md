# devskills

Installable **command** pack for Claude Code, OpenCode, OpenAI Codex, Cursor, and VSCode Copilot. Opinionated defaults, composable language profiles, a full dev workflow from specification to shipped product.

Despite the name, this is **not** a Skills repo — it ships **commands**, not skills. A Claude Code *Skill* is invoked by the model whenever it decides one applies; a devskills **command** is a slash command *you* invoke, when you choose. That difference is the point: the tools stay under your control. devskills doesn't hand the model more autonomy — it sharpens *your* developer skills with tools you drive directly, and you keep total control over when each one runs.

No magic. Files in the right directories. Prompts that encode real constraints.

## Install

```bash
git clone https://github.com/gleicon/devskills.git ~/.devskills
~/.devskills/install.sh
```

Commands copy to `~/.claude/commands/`, `~/.opencode/commands/`, and `~/.codex/prompts/` (each installed only when that tool is detected). External tools (osv-scanner, tldt) install automatically when their prerequisites are present (Homebrew or Go).

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

`setup.sh` writes a universal engineering baseline to `AGENTS.md` and points `CLAUDE.md` at it via `@AGENTS.md`; `--lang` stacks a language profile on top. See [Repository Setup](#repository-setup) below.

Keep devskills up to date:

```bash
~/.devskills/scripts/update.sh              # pull + reinstall commands
~/.devskills/scripts/update.sh --upgrade-deps  # also force-upgrade tldt
```

> **Renamed in this release:** every command is now namespaced with a `ds-`
> prefix (e.g. `/debug` → `/ds-debug`), and modes carry a `-mode` suffix
> (e.g. `/tiger-style` → `/ds-tiger-style-mode`). This avoids collisions with
> Claude Code built-ins like `/debug` and `/security-review`, and makes a
> command's kind readable at a glance. Re-running `install.sh` or `update.sh`
> drops in the new names and removes the old files automatically — no manual
> cleanup. See the [Commands](#commands) table for the current names.

### install.sh flags

```
--lang=<profile>     go | typescript | javascript | rust | python | java | zig
--claude-dir=<path>  Claude config dir (default: $CLAUDE_CONFIG_DIR or ~/.claude)
--skip-external      skip tldt installation
--skip-cursor        skip Cursor rules
--skip-vscode        skip VSCode Copilot instructions
--concise            add a terse-response directive to AGENTS.md (with --lang)
--dry-run            show what would happen, write nothing
```

### setup.sh flags (per-project)

```
--lang=<profile>     optional; stacks a language profile (go|typescript|javascript|rust|python|java|zig)
--concise            add a terse-response directive to AGENTS.md
--cursor             install Cursor rules into current project
--vscode             install VSCode Copilot instructions into current project
--dry-run            show what would happen, write nothing
```

## Commands

Every command is namespaced with a `ds-` prefix (short for devskills) so it never collides with a Claude Code or OpenCode built-in, and its **suffix tells you its kind**: `-mode` (persists for the session), `-review` (a findings-list audit), `-plan` (a graded, costed plan), or no suffix (a one-shot action). The tables below are grouped by kind; full taxonomy and per-command detail are in [docs/commands.md](docs/commands.md#kinds-of-command).

### Modes — persistent session behavior (`-mode`)

| Command | Description |
|---------|-------------|
| `/ds-tiger-style-mode` | TigerBeetle engineering constraints: safety, performance, experience |
| `/ds-ui-mode` | UI mode: component/state discipline, design craft, a11y, Core Web Vitals |
| `/ds-data-mode` | Data-engineering discipline as you build pipelines: idempotency, late/out-of-order data, schema drift, replay/backfill safety, data-quality assertions. Tool-agnostic |
| `/ds-git-mode` | Senior-engineer commit discipline: commit each self-contained working unit, terse Conventional-Commit messages (no LLM bloat), branch-first, never rewrite history |
| `/ds-step-mode` | User-driven, step-gated execution: smallest step → stop → free-form handback (never a forced picker) → repeat. Drive a plan with `/ds-step-mode current plan` |
| `/ds-tdd-mode` | Test-first, one vertical slice at a time |
| `/ds-test-mode` | Pragmatic testing mode — test by risk, not coverage |
| `/ds-quality-gate-mode` | Seven-pass review pipeline, deslop-bookended (deslop → test → security → bug → data → quality → docs → deslop), implement fixes between passes, toggleable mode |
| `/ds-caveman-lite-mode` | Compressed response mode (~25–35% token reduction) |
| `/ds-caveman-ultra-mode` | Compressed response mode (~75–85% token reduction) |

### Reviews — findings-list audits (`-review`)

| Command | Description |
|---------|-------------|
| `/ds-bug-review` | Language-agnostic correctness audit — hunts real bugs |
| `/ds-security-review` | Language-agnostic security audit — each finding names the attack |
| `/ds-osv` | Scan dependency manifests for known CVEs via OSV Scanner; `--fix` bumps direct deps |
| `/ds-data-review` | Store-agnostic data audit — schema/integrity, query-result correctness, transactions, migration safety. Each finding names the path to wrong/lost data (`--pipelines` also audits ETL/pipeline code) |
| `/ds-code-quality-review` | Strict maintainability audit: abstraction, sprawl, spaghetti |
| `/ds-doc-quality-review` | Strict docs audit: accuracy, dead links, bloat (`--comments` audits code comments) |
| `/ds-test-quality-review` | Strict test audit: is critical code well tested? |
| `/ds-ui-quality-review` | Strict UI audit: async-state/fetch correctness, a11y, Core Web Vitals, design craft |
| `/ds-comment-review` | Bring comments to discipline — WHY-not-WHAT, one-liner default, strip restate/obvious/cruft, keep the rare important long one. Reports by default, `--fix` to apply |
| `/ds-go-review` | Go: idiomatic + security + Tiger Style (`--no-tiger` to skip style) |
| `/ds-ts-review` | TypeScript/Workers: strict mode, React, Cloudflare (`--no-tiger` to skip style) |
| `/ds-rust-review` | Rust: geiger/unsafe, clippy, audit, Tiger Style (`--no-tiger` to skip style) |
| `/ds-python-review` | Python: idioms, typing, security, Tiger Style (`--no-tiger` to skip style) |
| `/ds-java-review` | Java: idioms, records/sealed types, security, Tiger Style (`--no-tiger` to skip style) |
| `/ds-zig-review` | Zig: explicit allocators, errors-as-values, safety, Tiger Style (`--no-tiger` to skip style) |
| `/ds-notebook-review` | Jupyter notebooks: execution/hidden-state, committed outputs & secrets, reproducibility, data-science correctness (leakage, chained assignment). Delegates cell-code review to `/ds-python-review` |

Every `-review` reports by default and changes nothing; pass `--fix` to apply the **mechanical, unambiguous** findings in place (correctness and security fixes, and anything resting on judgment, stay reported either way).

### Plans — graded, sequenced moves (`-plan`)

| Command | Description |
|---------|-------------|
| `/ds-perf-plan` | Language-agnostic optimization plan — moves tagged by architectural cost (L1/L2/L3), each with a cost model (`--max-level` to clamp, `--no-tiger` to skip style) |
| `/ds-architecture-plan` | Module/dependency/boundary analysis of an existing codebase → sequenced refactoring plan. Levels L1/L2/L3, `--max-level` to clamp |

### Actions — one-shot, produce a result and return

| Command | Description |
|---------|-------------|
| `/ds-spec` | Convert a description into a structured specification |
| `/ds-roadmap` | Ordered `## Roadmap` task list from a goal/spec/findings (`.project/PLAN.md`, or `PLAN.md` without `.project/`) |
| `/ds-explore` | Lay out candidate approaches with trade-offs (`--web` for research) |
| `/ds-blueprint` | Design a target architecture for a new system — modules, dependency rules, seams, build order. Decisive counterpart to `ds-explore` |
| `/ds-grill-me` | Relentless plan interview (`--record` logs to DECISIONS.md) |
| `/ds-workflow` | Standalone phase-map orchestrator — orient, then route each phase to its command (uses `.project/` state when present, never requires it) |
| `/ds-project-map` | Scan the repo into `.project/PROJECT.md` |
| `/ds-project-checkpoint` | Persist state to `.project/PLAN.md` (`--handoff` for a full handoff) |
| `/ds-project-resume` | Restore context from `.project/PLAN.md` |
| `/ds-deslop` | Strip AI-generated slop from the branch diff |
| `/ds-debug` | Root-cause a failure with the scientific method |
| `/ds-verify-this` | Prove a falsifiable claim with local before/after evidence |
| `/ds-zoom-out` | Step up a layer — map modules, callers, boundaries |
| `/ds-handoff` | Compact the conversation into a handoff doc |
| `/ds-tldt` | Extractive summary of context or a file — no LLM cost |
| `/ds-write-a-command` | Author a new devskills command in repo conventions |

## Build your own workflow

devskills ships no fixed pipeline. Each command does one job and hands control back — you compose them into whatever flow the work needs. Unlike all-in-one agent frameworks that drive the session for you, the power stays with you: nothing here decides your next step. Start anywhere, reorder freely — spec, plan, build under a mode or two, review, persist. The docs lay out worked flows, not one true path:

- **[docs/recipes.md](docs/recipes.md)** — worked, multi-step workflows (pre-PR gate, find-then-prove, driving a multi-PR queue, …)
- **[docs/commands.md](docs/commands.md)** — every command: args, behavior, and when to reach for it
- **[docs/project-workflow.md](docs/project-workflow.md)** · **[docs/project-recipes.md](docs/project-recipes.md)** — the optional `.project/` memory workflow
- **[docs/grill-me.md](docs/grill-me.md)** · **[docs/tiger-style.md](docs/tiger-style.md)** — the grill playbook and the engineering bar

## Repository Setup

`setup.sh` builds your project's `AGENTS.md` from stacked, independently-managed blocks, and points `CLAUDE.md` at it with a single `@AGENTS.md` import — so Claude Code (which reads `CLAUDE.md`) and OpenCode and OpenAI Codex (which read `AGENTS.md` directly) share the same content with no duplication.

| Block | Flag | Contents |
|-------|------|----------|
| `base` | always | Universal engineering principles — think before coding, simplicity first, surgical changes, goal-driven execution, safe at the boundaries |
| `language` | `--lang=<x>` | Stack-specific idioms, toolchain, and review constraints |
| `concise` | `--concise` | Terse-response directive (caveman-lite behavior, baked in) |

Running `setup.sh` with no flags writes just the baseline. Each block lives between `<!-- BEGIN/END devskills:<id> -->` markers, so re-running is idempotent and swapping `--lang` replaces only that block. Existing `AGENTS.md`/`CLAUDE.md` files are backed up (sibling timestamped `.bak`) once, before any change — these are transient; delete them or keep them out of version control once you've confirmed the result.

`update.sh` refreshes the globally-installed commands, but not a project's `AGENTS.md` — the managed blocks are a point-in-time snapshot. To pull baseline changes into a project after an update, re-run `setup.sh` there (idempotent, so it just refreshes the blocks in place).

The baseline blocks target `AGENTS.md` (Claude Code, OpenCode, and OpenAI Codex). Cursor and VSCode Copilot have their own rule mechanisms — `--cursor` installs `.cursor/rules/*.mdc` and `--vscode` writes `copilot-instructions.md`. Both honor `--lang`: they carry Tiger Style plus the notes for the selected language only (no `--lang` writes Tiger Style alone), but not the `base`/`concise` blocks.

To back out, `setup.sh --uninstall` strips the devskills blocks (and removes a file that held *only* devskills content), leaving your own content untouched — a clean install→uninstall round-trip restores the originals exactly.

## Language Profiles

Each profile encodes idioms, toolchain defaults, and review constraints for its stack, and is stacked under the baseline as a `## Language Profile — <x>` section.

| Profile | Stack | Use case |
|---------|-------|---------|
| `go` | Go 1.24+ | Backend services, CLIs, APIs |
| `typescript` | TypeScript 5+, Wrangler | Cloudflare Workers, Next.js, React |
| `javascript` | ES2022+, Wrangler | Cloudflare Workers, vanilla frontend |
| `rust` | Rust stable | Systems programming, large projects |
| `python` | Python 3.13+ | Backend services, APIs, CLIs, data pipelines |
| `java` | Java 25+ (LTS) | Backend services, APIs, CLIs, systems tooling |
| `zig` | Zig 0.16 | Systems programming, CLIs, embedded (Tiger Style native) |

## Scripts

`install.sh` — one-time global install. Copies commands to Claude Code and OpenCode config dirs. Installs external tools. Run from anywhere.

`scripts/setup.sh` — per-project configurator. Builds `AGENTS.md` (engineering baseline + optional language/concise blocks) and points `CLAUDE.md` at it via `@AGENTS.md`, optionally installs Cursor rules and VSCode Copilot instructions into the current directory. Run from inside a project. See [Repository Setup](#repository-setup).

`scripts/update.sh` — pulls the latest devskills repo and reinstalls skills. Use `--upgrade-deps` to also force-upgrade tldt to its latest published version.

`scripts/upgrade-deps.sh` — force-upgrades external tools regardless of current state. Useful after upstream major version bumps.

## External Tools

Installed by `install.sh`. Managed by `upgrade-deps.sh`.

| Tool | Purpose |
|------|---------|
| [osv-scanner](https://github.com/google/osv-scanner) | Supply-chain vulnerability scan against the OSV/CVE database |
| [tldt](https://github.com/gleicon/tldt) | Extractive text summarization; no LLM, no cost |

## References

devskills ships its own prompt commands based on these upstream sources.

| Reference | Used by |
|-----------|---------|
| [Tiger Style](https://tigerstyle.dev/) | `/ds-tiger-style-mode`, all review commands |
| [Caveman](https://github.com/juliusbrussee/caveman) | `/ds-caveman-lite-mode`, `/ds-caveman-ultra-mode` |
| [mattpocock/skills](https://github.com/mattpocock/skills) | `/ds-grill-me`, `/ds-handoff`, `/ds-zoom-out`, `/ds-tdd-mode`, `/ds-write-a-command` |
| [cursor/plugins — cursor-team-kit](https://github.com/cursor/plugins/tree/main/cursor-team-kit/skills) | `/ds-code-quality-review`, `/ds-deslop`, `/ds-verify-this` |
| [Andrej Karpathy](https://x.com/karpathy/status/2015883857489522876) · [multica-ai/andrej-karpathy-skills](https://github.com/multica-ai/andrej-karpathy-skills) | AGENTS.md baseline (`base` block) |

## License

MIT — see [LICENSE](LICENSE).
