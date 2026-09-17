#!/bin/bash
set -euo pipefail

script_dir="$(dirname "$(readlink -f "$0")")"
prefix="${PREFIX:-$HOME/.local}"

usage() {
  echo "usage: install.sh [--uninstall]" >&2
  exit 2
}

install() {
  mkdir -p "$prefix/bin" "$prefix/etc/tidy"
  cp "$script_dir/../bin/tidy" "$prefix/bin/tidy"
  cp "$script_dir/../tidy.conf" "$prefix/etc/tidy/tidy.conf"
  sed -i "s|^cache_dir=.*|cache_dir=$prefix/var/tidy|" "$prefix/etc/tidy/tidy.conf"
  printf 'installed tidy to %s\n' "$prefix"
}

uninstall() {
  rm -rf $prefix/bin/tidy $prefix/etc/tidy
  printf 'removed tidy from %s\n' "$prefix"
}

main() {
  case "${1:-}" in
    "") install ;;
    --uninstall) uninstall ;;
    *) usage ;;
  esac
}

main "$@"
