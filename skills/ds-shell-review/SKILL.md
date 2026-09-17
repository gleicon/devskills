---
name: ds-shell-review
description: "Review shell scripts with Tiger Style constraints, bash idioms, and macOS/Linux portability."
disable-model-invocation: true
---

Applies to: bash 3.2+ on macOS and Linux. Build glue, CI steps, installers, operational tooling.

Shell has no compiler and no type system, so the review carries the weight a toolchain carries elsewhere. Two failure classes dominate: a script that works on the author's machine and breaks on the other platform, and an unquoted expansion that turns a filename or an argument into code. Weight those over style.

## Arguments

Scan the invocation for the `--no-tiger`, `--fix`, and `--full` flags. Treat every other argument as review scope (files or directories); if no scope is given, review the changed files on the current branch.

- `--no-tiger` present → skip the Tiger Style section; run Preamble, Portability, Quoting & Expansion, Error Handling, Security, and Testing only.
- `--no-tiger` absent → run all sections (default).
- `--fix` → after reporting, apply only the violations whose fix is **mechanical and unambiguous** (quoting an expansion, `$(...)` for backticks, `printf` for `echo -e`, `[[ ]]` for `[ ]`). Anything that changes logic or rests on an unverified assumption — especially security and portability findings that need a platform branch — **stays report-only**. After applying, re-run any build/test/lint check already in the loop and revert any fix that breaks it — or that touched more than the intended mechanical edit. End with a summary of what was applied and what was left.
- `--full` → review the entire codebase instead of just the branch's changes. Explicit positional scope still wins; `--full` only replaces the no-scope default.

Example: `/ds-shell-review --no-tiger scripts/` reviews `scripts/` without Tiger Style.

Scope is every file with a `.sh`/`.bash` extension or a bash shebang. A `#!/bin/sh` script is reviewed as POSIX sh: bash-only constructs (`[[ ]]`, arrays, `local -n`, `pipefail`) are findings there, not idioms.

## Automated Checks (run first if tools are available)

```bash
shellcheck -x -S style <files>       # static analysis; -x follows source'd files
shfmt -d -i 2 -ci -bn <files>        # formatting drift
bats tests/                          # if the project has bats tests
```

Run these, report what they surface, then do the manual review below. They are baseline context — anchor findings to the code in scope; don't report pre-existing failures outside the change as if it introduced them. `shellcheck` cannot see platform divergence (a GNU-only flag passes it), so the Portability section is always manual.

## Review Checklist

Use the checklist as a lens, not a scorecard: reason about the actual change, report real violations anchored to `file:line`, and flag issues even when they aren't listed. Don't manufacture findings to fill a category. Report only violations — no praise, no summary.

### Tiger Style

Skip this section entirely if `--no-tiger` was passed. Otherwise it is mandatory.
- [ ] The script asserts its preconditions at the top: required commands via `command -v`, required variables via `${VAR:?}`, required argument count — before any side effect
- [ ] Every loop over external input has an explicit bound; `while true` has a visible exit
- [ ] Named constants (`readonly`) for limits, paths, and retry counts — no unexplained magic values
- [ ] Functions under 70 lines; a script past ~200 lines is flagged as a candidate for a real language
- [ ] `rm -rf` on a variable path uses `"${var:?}"` so an empty value fails loudly instead of expanding to `/`

### Preamble
- [ ] `#!/usr/bin/env bash` — `#!/bin/bash` pins macOS to stock 3.2 even when a newer bash is installed; `#!/bin/sh` is dash on Debian
- [ ] `set -euo pipefail` present, and the places where `-e` does not fire (inside `if`, `while`, `&&`/`||` chains, `$(...)` in assignments) are handled explicitly
- [ ] A pipeline whose consumer exits early (`| head`) under `pipefail` is guarded deliberately, not by dropping `pipefail`
- [ ] Temp files are created after the `trap ... EXIT` that removes them, never before
- [ ] `main "$@"` at the bottom so the file can be sourced for tests without running

### Portability — macOS and Linux
Both platforms are targets unless the script says otherwise. macOS ships bash 3.2 and BSD userland; Linux ships bash 5 and GNU coreutils.
- [ ] No bash 4+ construct on a 3.2 floor: `declare -A`, `mapfile`/`readarray`, `${var,,}`/`${var^^}`, `${var@Q}`, `|&`, negative array indices, `$EPOCHSECONDS`. A script that needs bash 4+ says so and checks `((BASH_VERSINFO[0] >= 4))` before doing anything else
- [ ] No `sed -i` in either form — BSD requires `-i ''`, GNU rejects it; write to a temp file and `mv`, or `perl -pi -e`
- [ ] No GNU-only flags: `date -d`, `stat -c`, `grep -P`, `find -printf`, `xargs -d`, `sort -V`, `cp --parents`, `readlink -f`, `mktemp` without a template, `sed` with `\+`/`\|`/`\t` in basic regex
- [ ] No BSD-only flags either: `date -v`, `stat -f`, `sed -i ''`
- [ ] No tool absent from one platform used bare: `timeout`, `flock`, `nproc`, `tac`, `realpath`, `sha256sum` (macOS has `shasum -a 256`) — detected once and wrapped in a function, never branched inline
- [ ] `echo -e`/`echo -n` replaced by `printf '%s\n'`
- [ ] Script directory resolved with `"$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"`, not `readlink -f` or `realpath`
- [ ] Platform branching lives in one place (`case "$(uname -s)"`), not scattered through the script
- [ ] No dependence on Homebrew's `g`-prefixed GNU tools (`gsed`, `gdate`, `gstat`)

### Quoting & Expansion
- [ ] Every expansion quoted: `"$var"`, `"${arr[@]}"`, `"$(cmd)"` — an unquoted one is a finding unless the split is deliberate and commented
- [ ] Argument lists built in arrays, not whitespace-joined strings
- [ ] `[[ ]]` over `[ ]`, `(( ))` for arithmetic, `$(...)` over backticks
- [ ] No parsing of `ls` output; globs with `shopt -s nullglob` for file iteration
- [ ] `find`/`xargs` pairs use `-print0`/`-0`; `read` loops use `-r` and `IFS=`
- [ ] `${var:-default}`/`${var:?msg}` over manual emptiness checks

### Error Handling
- [ ] Every command whose failure matters is checked — `cmd || die "…"`; a `die()` writes to stderr and exits non-zero
- [ ] Diagnostics go to stderr; stdout carries only output another command can consume
- [ ] Exit codes are meaningful: 0 success, 1 failure, 2 usage error
- [ ] No `|| true` that swallows a failure the script should surface
- [ ] `cd` failures are checked (`cd dir || exit`) — under `set -e` inside a function they are, elsewhere they are not

### Security
- [ ] No `eval` on any value derived from input; no command built by string interpolation and then run
- [ ] `curl … | sh` and `curl … | bash` are absent, or download to a file and verify a checksum first
- [ ] Secrets never on the command line (`ps` shows them) or in `set -x` output; passed via env or a file with restricted mode
- [ ] Temp files created with `mktemp`, never a predictable path in `/tmp`
- [ ] Paths from input are not joined into `rm`, `cp`, or redirection without validation; `--` separates options from operands where a filename can start with `-`
- [ ] No `sudo` inside a script that did not announce it; no `chmod 777`

### Testing
- [ ] Scripts with branches have `bats` tests; run on both `ubuntu-latest` and `macos-latest` in CI
- [ ] Tests source the script and call functions; spawning and grepping output is reserved for the CLI surface
- [ ] External commands stubbed by a directory first on `PATH`, not by network or real side effects

## Output Format

```
<file>:<line>: <severity>: <problem>. <fix>.
```

Severity levels: `critical` (security / destructive on the wrong path), `major` (breaks on one platform / silent failure), `minor` (idiom/style).

Skip formatting nits unless they affect correctness or readability significantly.
