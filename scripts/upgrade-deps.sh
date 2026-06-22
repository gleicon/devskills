#!/usr/bin/env bash
# upgrade-deps.sh: force-upgrade all external tools to their latest versions
#
# Unlike install.sh (idempotent, skips already-installed binaries), this
# script forces reinstall of every tool regardless of current state.
# Run it when tools feel stale or after a major version bump upstream.
set -euo pipefail

log()  { printf '[devskills:upgrade] %s\n' "$1"; }
warn() { printf '[devskills:upgrade] WARN: %s\n' "$1" >&2; }

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

DRY_RUN=0
for arg in "$@"; do
  case "$arg" in
    --dry-run) DRY_RUN=1 ;;
    --help|-h)
      echo "Usage: upgrade-deps.sh [--dry-run]"
      echo "Force-reinstalls tldt to its latest published version."
      exit 0
      ;;
  esac
done

# Shared tldt logic (depends on log/warn/DRY_RUN above).
# shellcheck source=lib/external-tools.sh
source "${SCRIPT_DIR}/lib/external-tools.sh"

log "Force-upgrading all external tools..."
devskills_osv upgrade
devskills_tldt upgrade
devskills_astgrep upgrade
log "Caveman: bundled in devskills prompt files. Run scripts/update.sh to update."
log "Done."
