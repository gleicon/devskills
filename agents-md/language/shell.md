## Language Profile — Shell

Target: bash 3.2+ on macOS and Linux. Build glue, CI steps, installers, operational tooling.

Apply these conventions to all shell scripts in this session.

### Toolchain

Lint with `shellcheck` (`shellcheck -x -S style`) and treat every warning as a finding. Format with `shfmt -i 2 -ci -bn`. Test with `bats-core`. Run both `shellcheck` and `shfmt -d` in CI; they are cheap and catch most of what follows.

### Preamble

- `#!/usr/bin/env bash` — never `#!/bin/bash`, which pins macOS to the stock 3.2 even when a newer bash is installed, and never `#!/bin/sh`, which is dash on Debian and POSIX-only there.
- `set -euo pipefail` on the first line after the shebang. Know the gaps: `-e` is disabled inside `if`, `while`, `&&`/`||` chains, and command substitutions in assignments; `pipefail` fails a pipeline whose consumer exits early (`| head`), so guard those with `|| true` deliberately.
- Resolve the script's own directory with `"$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"`, not `readlink -f` (BSD `readlink` has no `-f`).
- Clean up temp files with `trap 'rm -rf "$tmp"' EXIT`, set before the file is created.

### Portability — macOS and Linux

Scripts must run on both. macOS ships bash 3.2 and BSD userland; Linux ships bash 5 and GNU coreutils. Every difference below is a real breakage, not a style point.

- **bash 3.2 is the floor.** No associative arrays (`declare -A`), `mapfile`/`readarray`, `${var,,}`/`${var^^}`, `${var@Q}`, `|&`, or negative array indices. A script that needs bash 4+ says so in a comment and checks `((BASH_VERSINFO[0] >= 4))` up front, before doing anything else.
- `sed -i` takes a mandatory argument on BSD (`-i ''`) and an optional one on GNU (`-i`); the two forms are mutually incompatible. Write to a temp file and `mv`, or use `perl -pi -e`.
- BSD `sed` has no `\+`, `\|`, or `\t` in basic regex; use `-E` and `[[:space:]]`, or `awk`.
- `date -d` is GNU only; BSD uses `date -v` and `-j -f`. For epoch math use `date +%s` and arithmetic, or `$EPOCHSECONDS` (bash 5+, so only where the floor is raised).
- `stat` differs entirely: GNU `stat -c '%s'` versus BSD `stat -f '%z'`. Use `wc -c < file` for size and `ls -l`/`find -newer` for mtime comparisons.
- `grep -P`, `find -printf`, `xargs -d`, `sort -V`, `cp --parents`, `mktemp` without a template on Linux, and `realpath` on older macOS are GNU-isms. `mktemp -d "${TMPDIR:-/tmp}/name.XXXXXX"` works on both.
- `echo -e` and `echo -n` are not portable; use `printf '%s\n'` always.
- `xargs` and `find -exec` behave the same everywhere only with `-print0`/`-0`; never split on whitespace.
- `ls`, `ps`, `mktemp`, `tar`, `base64` flag sets diverge; pin the intersection or branch on `uname -s` in one place, never inline.
- `sha256sum` is GNU; macOS has `shasum -a 256`. Detect once and assign to a variable.
- Never assume GNU tools are installed under their plain names on macOS: Homebrew ships them as `gsed`, `gdate`, `gstat`. Do not depend on those either; write for the intersection.
- `timeout`, `flock`, `nproc`, `tac` are absent from stock macOS. Use `sysctl -n hw.ncpu` / `nproc` behind a function, and `tail -r` for `tac`.

### Quoting and Expansion

- Quote every expansion: `"$var"`, `"${arr[@]}"`, `"$(cmd)"`. The exceptions are deliberate word-splitting of a variable that holds flags, and then `read -ra` into an array is the better tool.
- `[[ ]]` over `[ ]` for tests; `(( ))` for arithmetic. `[ ]` only inside a `#!/bin/sh` script.
- `$(...)` over backticks; `${var:-default}` and `${var:?message}` over manual checks.
- Build argument lists in arrays, not strings: `args=(-x "$path"); cmd "${args[@]}"`.
- Never `eval` on input. Never build a command from user-supplied text; pass it as an argument.
- Loop over files with a glob and `nullglob` (`shopt -s nullglob`), never by parsing `ls`.

### Error Handling

- Every command whose failure matters is checked. `cmd || die "message"` with a `die()` that prints to stderr and exits non-zero.
- Diagnostics go to stderr (`>&2`); stdout is for output another command can consume.
- Exit codes carry meaning: 0 success, 1 general failure, 2 usage error. Reserve 64–78 (`sysexits`) only if the project already uses them.
- Functions return status, not text, unless the text is the product; capture with `$(...)` and check `$?` in the same statement.

### Structure

- `main "$@"` at the bottom, after all functions, so the file can be sourced for tests without running.
- `local` on every variable inside a function; `readonly` for constants.
- Prefer a function over a repeated pipeline; a script past ~200 lines is a candidate for a real language.
- Parse options with a `while`/`case` loop over `"$@"`; `getopts` for short flags only. No `getopt` — the BSD one does not support long options.

### Testing

- `bats-core` tests for every script that has branches; run them on both platforms in CI (`ubuntu-latest` and `macos-latest`).
- Source the script under test and call its functions; do not spawn it and grep output unless the CLI surface is the thing under test.
- No network in tests; fake external commands by putting a stub directory first on `PATH`.

### Tiger Style

- Assert preconditions at the top: required commands (`command -v jq >/dev/null || die "jq required"`), required variables (`: "${TOKEN:?TOKEN is required}"`), required arguments (`[[ $# -eq 2 ]] || usage`).
- Bound every loop over external input; `while read` over a command substitution has an implicit bound, `while true` does not.
- No `rm -rf "$dir/"` where `$dir` can be empty: `"${dir:?}"` makes the expansion fail loudly instead of deleting `/`.
