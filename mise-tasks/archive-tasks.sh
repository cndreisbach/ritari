#!/usr/bin/env bash
#MISE description="Archive task notes whose state is done (pass --dry-run to preview)"
#MISE depends=["build"]
# Archive every task note with `state: done`. Archived notes stay linkable but
# drop out of search and move under `# Archived` in tasks/index.md (§5.6).
set -euo pipefail

dry_run=false
if [[ "${1:-}" == "--dry-run" ]]; then
  dry_run=true
elif [[ $# -gt 0 ]]; then
  echo "usage: mise run archive-tasks [--dry-run]" >&2
  exit 2
fi

# Prefer the checkout's binary when it has been built; fall back to PATH.
root=$(git rev-parse --show-toplevel 2>/dev/null || echo .)
rit=rit
if [[ -x "$root/bin/rit" ]]; then
  rit="$root/bin/rit"
fi

ids=$("$rit" search --type task --meta state=done --limit 0 | sed -n 's/^id: //p')
if [[ -z "$ids" ]]; then
  echo "no done tasks to archive"
  exit 0
fi

while IFS= read -r id; do
  [[ -z "$id" ]] && continue
  if [[ "$dry_run" == true ]]; then
    echo "would archive $id"
  else
    "$rit" note archive "$id" >/dev/null
    echo "archived $id"
  fi
done <<<"$ids"
