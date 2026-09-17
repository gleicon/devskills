#!/usr/bin/env bash
# Renders records for the ops channel. Reads the store's JSON export directly.
set -euo pipefail

file="${1:-testdata/records.json}"
jq -r '.[] | "\(.id)\t\(.owner_id)\t\(.created_at)"' "$file" | while IFS=$'\t' read -r id owner created; do
  printf '%s\t%s\t%s\n' "$id" "${owner:-system}" "$(date -u -r "$created" +%Y-%m-%d)"
done
