<div align="center">

# devskills

**Opinionated engineering skills for AI coding assistants.**
You invoke them. You stay in control.

`Claude Code` · `OpenCode` · `Codex`

</div>

---

devskills are Agent Skills with one rule inverted: **you** decide when each one runs, never the model. Every skill ships with `disable-model-invocation` set — it surfaces only when you type `/ds-<name>`. Nothing fires on its own, watches your session, or picks your next step.

Each skill encodes an opinionated engineering default — Tiger Style, a spec→ship workflow, strict review passes, per-language idioms — as a tool you reach for, not autonomy you hand over. A single Go binary installs the whole catalog and scaffolds your project's `AGENTS.md`.

*No magic. Files in the right directories. Prompts that encode real constraints.*

## Install

Install the CLI — pick one:

```bash
# quick install — macOS / Linux
curl -fsSL https://raw.githubusercontent.com/gleicon/devskills/main/install.sh | sh

# or, with Go
go install github.com/gleicon/devskills@latest
```

…or grab a prebuilt binary from the [latest release](https://github.com/gleicon/devskills/releases).

Then sync the skills into your assistants:

```bash
devskills install
```

`install` detects Claude Code, OpenCode, and Codex, lets you pick which to target, and copies the catalog into each one's skills directory. Re-run it any time to update — it prunes skills that were renamed or dropped and never touches anything it didn't ship. Use `--local` to install into the current repo instead of globally, `--dry-run` to preview the plan, `--uninstall` to remove.

*Enforcement is per-assistant: Claude Code and Codex honor the invoke-only flag (Codex via a generated policy sidecar); OpenCode [doesn't yet](https://github.com/anomalyco/opencode/issues/34498), so a skill there can still be model-invoked until upstream lands it.*

## Use

Type `/ds` in your assistant to browse the catalog, then `/ds-<name>` to run a skill:

```
/ds-spec         turn a rough idea into a structured spec
/ds-tiger-style-mode   set the engineering bar for the session
/ds-security-review    audit the diff for exploitable weaknesses
```

To give a project a persistent engineering baseline, scaffold its `AGENTS.md` (and a `CLAUDE.md` import) from your stack:

```bash
devskills init --lang go
```

## The catalog

The **phase spine** follows the arc of a change. The groups below it — modes, language reviews, project memory, utilities — cut across phases; reach for them whenever they apply. Everything works standalone: no skill requires `.project/` or any other.

| Phase | Skill | What it does |
|-------|-------|--------------|
| **Orient** | `/ds-zoom-out` | step up a level — map how the code fits before you change it |
| **Spec** | `/ds-spec` | turn a description into a structured spec with acceptance criteria |
| | `/ds-explore` | at a fork, lay out candidate approaches without deciding |
| | `/ds-grill-me` | get interviewed about a plan until the gaps are exposed |
| **Plan** | `/ds-blueprint` | design a **new** system: module boundaries, seams, build order |
| | `/ds-architecture-plan` | assess an **existing** codebase → sequenced refactoring plan |
| | `/ds-roadmap` | turn a goal or another skill's output into an ordered roadmap |
| | `/ds-perf-plan` | a graded, ranked optimization plan with a cost model |
| **Build** | `/ds-debug` | root-cause a failure: reproduce, isolate, fix, prove |
| **Clean** | `/ds-deslop` | strip AI slop from the diff before any review |
| **Review** | `/ds-code-quality-review` | maintainability + single source of truth |
| | `/ds-bug-review` | correctness — real bugs, not style |
| | `/ds-security-review` | exploitable weaknesses; each finding names the attack |
| | `/ds-data-review` | data correctness, integrity, migration safety |
| | `/ds-test-quality-review` | is the risky logic actually covered, and are the tests real? |
| | `/ds-doc-quality-review` | docs accuracy against the code, dead links, staleness |
| | `/ds-ui-quality-review` | UI soundness, craft, and accessibility |
| | `/ds-comment-review` | do the comments earn their place? |
| | `/ds-notebook-review` | notebook state, output hygiene, reproducibility |
| | `/ds-quality-gate` | run the review pipeline as a gate over the whole branch |
| | `/ds-osv` | scan dependencies for known CVEs (OSV) |
| **Verify** | `/ds-verify-this` | a before/after repro with a hard verdict |
| **Ship** | `/ds-handoff` | compact the session into a handoff doc |

Every review reports by default and changes nothing; pass `--fix` to apply the mechanical, unambiguous findings, or `--full` to widen scope from the branch diff to the whole codebase.

### Modes — standing postures, compose anytime

Turn one on and it governs the rest of the session; several can be active at once.

| Mode | Posture |
|------|---------|
| `/ds-senior-mode` | write like a super-senior engineer everywhere — terse, step-gated, no slop (folds in git + test + deslop + step) |
| `/ds-tiger-style-mode` | the safety + explicitness engineering bar |
| `/ds-git-mode` | commit each working unit as it lands; terse messages; never pushes |
| `/ds-step-mode` | smallest reviewable step, then hand back |
| `/ds-test-mode` | keep the core tested by risk as you build |
| `/ds-tdd-mode` | drive implementation test-first, one vertical slice at a time |
| `/ds-ui-mode` | component/state discipline + design craft for UI work |
| `/ds-data-mode` | idempotency, late/out-of-order data, schema drift, replay safety |

### Language reviews

`/ds-go-review` · `/ds-python-review` · `/ds-rust-review` · `/ds-java-review` · `/ds-ts-review` · `/ds-zig-review`

Each folds Tiger Style, that language's idioms, and security into one pass, and detects the project's target version to layer on version-specific checks. Prefer these over the general reviews when the diff is single-language.

### Project memory — optional `.project/` persistence

Durable context across `/clear` and session ends. The workflow runs fine without any of these.

| Skill | What it does |
|-------|--------------|
| `/ds-project-map` | map the repo into `.project/PROJECT.md` |
| `/ds-project-config` | set preferences (e.g. modes auto-applied on resume) |
| `/ds-project-resume` | restore where you left off and apply configured modes |
| `/ds-project-checkpoint` | sweep the session, route durable context to its owning file |
| `/ds-project-compact` | housekeeping over the persisted `.project/` state |
| `/ds-project-verify` | reconcile the `.project/` files against the code |

### Utilities

| Skill | What it does |
|-------|--------------|
| `/ds-tldt` | extractive summary of a doc before it enters context — no LLM cost |
| `/ds-recall` | inject prior local context from [recall](https://github.com/gleicon/recall) into the session |
| `/ds-recall-capture` | store this session's outcome in recall's knowledge base |
| `/ds-recall-setup` | initialize recall and its session integration |

> The `recall` skills are experimental and need the external [recall](https://github.com/gleicon/recall) engine installed.

## The CLI

One binary, four commands:

| Command | Does |
|---------|------|
| `devskills install` | sync the skills into Claude Code / OpenCode / Codex (`--local`, `--harness`, `--dry-run`, `--uninstall`) |
| `devskills init` | scaffold a project's `AGENTS.md` + a `CLAUDE.md` import (`--lang`, `--concise`, `--phases`) |
| `devskills doctor` | check — or, with `--fix`, install — the external tools some skills need |
| `devskills version` | print version and build info |

`init` builds `AGENTS.md` from stacked, independently-managed blocks marked `<!-- BEGIN/END devskills:<id> -->`, so re-running is idempotent and swapping a language replaces only that block. It's harness-agnostic — Claude Code reads the `CLAUDE.md` import; OpenCode and Codex read `AGENTS.md` directly.

### Language profiles

`init --lang` stacks a stack-specific profile — idioms, toolchain defaults, and review constraints — under the universal baseline. Pass several (`--lang go,typescript`) for a polyglot repo.

| Profile | Target |
|---------|--------|
| `go` | Go 1.24+ — backend services, CLIs, APIs |
| `typescript` | TypeScript 5.5+ — Workers, Next.js, React, edge |
| `javascript` | ES2022+ — Workers, vanilla frontend |
| `rust` | Rust stable — systems, performance-critical services |
| `python` | Python 3.13+ — backend, APIs, CLIs, data pipelines |
| `java` | Java 25+ (LTS) — backend, APIs, systems tooling |
| `zig` | Zig 0.16 — systems, CLIs, embedded |

### External tools

`doctor` provisions the binaries a few skills shell out to — each detects and instructs at use time, so these are optional until you run the skill.

| Tool | Skill | Purpose |
|------|-------|---------|
| [osv-scanner](https://github.com/google/osv-scanner) | `/ds-osv` | supply-chain vulnerability scan against the OSV/CVE database |
| [ast-grep](https://github.com/ast-grep/ast-grep) | `/ds-security-review` | structural pattern search that widens the security pass ([cookbook](docs/ast-grep.md)) |
| [tldt](https://github.com/gleicon/tldt) | `/ds-tldt` | extractive text summarization — no LLM, no cost |

## Docs

devskills ships no fixed pipeline. Each skill does one job and hands control back — you compose them into whatever the work needs. The docs lay out worked flows, not one true path.

- **[docs/skills.md](docs/skills.md)** — every skill: args, behavior, and when to reach for it
- **[docs/recipes.md](docs/recipes.md)** — worked, multi-step workflows (pre-PR gate, find-then-prove, driving a multi-PR queue, the optional `.project/` memory loop)
- **[docs/project-workflow.md](docs/project-workflow.md)** — the `.project/` memory model
- **[docs/grill-me.md](docs/grill-me.md)** · **[docs/tiger-style.md](docs/tiger-style.md)** — the grill playbook and the engineering bar
- **[docs/ast-grep.md](docs/ast-grep.md)** — the optional structural pass for `/ds-security-review`

## References

devskills builds on these upstream sources.

| Reference | Used by |
|-----------|---------|
| [Tiger Style](https://tigerstyle.dev/) | `/ds-tiger-style-mode`, all review skills |
| [mattpocock/skills](https://github.com/mattpocock/skills) | `/ds-grill-me`, `/ds-handoff`, `/ds-zoom-out`, `/ds-tdd-mode` |
| [cursor-team-kit](https://github.com/cursor/plugins/tree/main/cursor-team-kit/skills) | `/ds-code-quality-review`, `/ds-deslop`, `/ds-verify-this` |
| [Andrej Karpathy](https://x.com/karpathy/status/2015883857489522876) · [andrej-karpathy-skills](https://github.com/multica-ai/andrej-karpathy-skills) | the `AGENTS.md` baseline |
| [recall](https://github.com/gleicon/recall) | `/ds-recall`, `/ds-recall-capture`, `/ds-recall-setup` |

## License

MIT — see [LICENSE](LICENSE).

## Development

No Makefile — it's a plain Go module (needs **Go 1.26+**). The common tasks:

```bash
go build .                                          # build the ./devskills binary
go run . --help                                     # run the CLI without building first
go test -race ./...                                 # unit tests (embed integrity, sync, scaffold, cli, …)
go test -tags integration ./internal/acceptance/    # end-to-end acceptance — builds the binary, drives it in a sandbox
golangci-lint run                                   # lint gate (golangci-lint v2; config in .golangci.yml)
gofmt -w .                                          # format
```

Skills and profiles live in `skills/` and `agents-md/` and are embedded into the binary at build time — edit the source there, not an installed copy. Releases are cut by pushing a `v*` tag (goreleaser, via `.github/workflows/release.yml`); `goreleaser build --snapshot --clean` dry-runs the cross-build locally.
