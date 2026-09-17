#!/usr/bin/env bash
set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
prefix="${PREFIX:-$HOME/.local}"

main() {
  mkdir -p "$prefix/bin" "$prefix/etc/tidy"
  cp "$script_dir/../bin/tidy" "$prefix/bin/tidy"
  cp "$script_dir/../tidy.conf" "$prefix/etc/tidy/tidy.conf"
  printf 'installed tidy to %s\n' "$prefix"
}

main "$@"
