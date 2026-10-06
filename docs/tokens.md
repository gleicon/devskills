# Token tools

devskills bundles no output compressor, code-graph index, or API proxy, and recommends none as a way to cut your bill. Measured end to end, these tools rarely lower cost, sometimes raise it, and most of them change how your assistant works in ways you can't see.

This page gives the evidence, the bar a tool would have to clear, and the two tools we do recommend: a language server plugin for catching errors, and `ccusage` for seeing where your spend goes.

## Where the money goes

A coding agent re-sends its whole conversation on every turn. Assistants cache that prefix, so most of what you pay for is cache traffic, not the text a tool just printed.

In the Claude Code benchmark runs of [Weinberger & Hozez](#sources):

- Cache writes and reads were **about 80% of the bill**.
- Tool output — what compressors shrink — was **about 3.3%**.
- Everything a user-side tool can touch at all (tool output, tool-call arguments, file reads, injected context, history) was **about 6%**.

So a compressor that halves tool output can move the bill by a couple of percent at best. If the model then needs one extra turn to recover what was cut, that turn re-sends the whole cached context and the saving is gone.

The same paper estimates a larger share in long interactive sessions — roughly 30% of cost within reach, and a ceiling of about 9% savings — but that setting was not measured.

## What the measurements show

| Tool | Kind | Measured result |
|---|---|---|
| rtk | Rewrites shell output through a hook | **+7.6% cost** at low effort, flat at high effort, quality unchanged ([JetBrains](#sources)). **−2.7%** pooled in [Weinberger & Hozez](#sources), but their held-out tasks alone show no clear change. |
| The authors' own aggressive compressor | Hook plus a stack of filtering gates | Cut delivered tool output **38.4%** and **raised cost 6.8%** ([Weinberger & Hozez](#sources)). |
| Headroom | Proxy in front of the API | **+48.4% cost** on Claude Code; **12–14% cheaper** on Codex (estimated from token counts), with three times the wall time. Its compressor reported nothing compressed on either ([Weinberger & Hozez](#sources)). |
| Codebase-Memory | Code graph served over MCP | **10× fewer tokens**, but **83% answer quality against 92%** for plain file reading — from the authors' evaluation of their own tool ([Vogel et al.](#sources)). |

Three patterns run through these results:

- **Tokens cut is not cost cut.** Across 100 Haiku tasks, how much tool output a compressor removed barely predicted how much cost changed (r = 0.15).
- **A tool's own counter measures the wrong thing.** rtk reported 96.2 million tokens saved while the bill for the same runs went up. It counts the full raw output as the alternative, but the assistant would have truncated that output anyway, and most input is billed at the cached rate.
- **The result depends on the combination.** The same proxy cost 48% more on one assistant and 12% less on another. A claim measured on someone else's setup doesn't transfer to yours.

## The bar a tool must clear

We would adopt, bundle, or recommend a token tool only if it passes all five:

1. **Its premise still holds** once the assistant's own caching, output truncation, and context compaction are counted.
2. **It stays out of the loop** between the assistant and its tools or API: no hook that rewrites or blocks the built-in read, search, and shell tools, and no proxy that sees your API traffic.
3. **It is a single binary** with no daemon, database, or telemetry.
4. **Any developer would be glad to have it installed,** the way they are with osv-scanner or gitleaks.
5. **Its gain shows on our bench** as lower cost per successful task, cache included.

Every tool we surveyed failed at least one of the first four, so none reached the bench. The usual failures were hooks that deny or rewrite the built-in tools, proxies in front of the API, background indexers and file watchers, telemetry switched on by default, and licenses that rule out commercial use.

## What we recommend instead

### A language server plugin, for correctness

Claude Code's [code intelligence plugins](https://code.claude.com/docs/en/plugins/code-intelligence) connect a language server, so Claude sees the type errors and missing imports its own edits introduce before you run a build. For Go:

```text
go install golang.org/x/tools/gopls@latest
/plugin install gopls-lsp@claude-plugins-official
```

Other languages have their own plugin; the page above lists them. The plugins work in Claude Code terminal sessions only.

Install one for the diagnostics, not to save tokens. Anthropic's cost guide suggests the plugins reduce file reads, but the one study we found on the question found a language server raised tokens for capable models on most task types it tried ([Xu](#sources)). That study ran its own agent loop, not these plugins, so neither claim is settled.

### `ccusage`, to see your spend

[ccusage](https://github.com/ccusage/ccusage) reads the usage logs your assistants already keep on disk — Claude Code, Codex, OpenCode, and others — and reports tokens and cost by day, week, month, or session:

```bash
npx ccusage@20.0.26 session --offline
```

- `--offline` prices usage from its bundled price table instead of fetching prices over the network.
- Pin a version you have looked at rather than running `@latest`.
- Costs are list-price estimates. A subscription bills differently.

It doesn't change how your assistant works; it only reads what's already there.

### `devskills bench`, to measure a change

To see whether a change to a skill or an `AGENTS.md` block moves cost, [`devskills bench`](bench.md) runs the old and new versions side by side and reports each one's **cost per successful run**, failures and cache included, next to its median cost — the measure these studies argue for.

## Sources

Each study was read in full. Checked 2026-09-29.

- **Weinberger & Hozez**, [*Token Reduction Is Not Cost Reduction: An Empirical Study of End-to-End Efficiency in API-Based Coding Agents*](https://arxiv.org/abs/2607.12161), arXiv 2607.12161v5, August 2026. Claude Code 2.1.201; Haiku 4.5, Sonnet 5, Opus 4.8; 2,848 paired billed runs on 103 tasks. The authors work at PointFive and built and configured every arm themselves; they are not Headroom's developers.
- **Shiryaev**, [*Does "rtk" skill really cut agent tokens by 60–90%? We tested it*](https://blog.jetbrains.com/ai/2026/07/rtk-claude-code-token-savings/), JetBrains AI blog, July 2026. rtk 0.43.0; Claude Code 2.1.201; claude-sonnet-5 at low and high effort; 86 SkillsBench tasks, 425 billed trials.
- **Vogel et al.**, [*Codebase-Memory: Tree-Sitter-Based Knowledge Graphs for LLM Code Exploration via MCP*](https://arxiv.org/abs/2603.27277), arXiv 2603.27277, March 2026. The authors evaluate their own tool on one repository per language with Opus 4.6; the first author graded the answers; it counts tokens, not billed cost.
- **Xu**, [*Does a Language Server Save Tokens for Coding Agents? A Measurement Methodology and Preliminary Study*](https://arxiv.org/abs/2608.13568), arXiv 2608.13568v1, June 2026. A single-author preliminary study with small task sets; its own agent loop with pylsp and pyright; counts context tokens, not billed cost.
- **Anthropic**, [Code intelligence plugins](https://code.claude.com/docs/en/plugins/code-intelligence) and [Manage costs effectively](https://code.claude.com/docs/en/costs).
