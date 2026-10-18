#!/usr/bin/env bash
# Prints markdown release notes for commits since the given tag (or all
# commits if none), grouped by Conventional Commit type. Non-breaking docs,
# chore, and ci commits are omitted.
# Usage: release-notes.sh [previous-tag]
set -euo pipefail

range=HEAD
if [ -n "${1:-}" ]; then
  range="$1..HEAD"
fi

breaking=() features=() fixes=() other=()
while IFS= read -r sha; do
  subject=$(git log -1 --format=%s "$sha")
  body=$(git log -1 --format=%b "$sha")
  # "type(scope)!: description" -> "**scope:** description"
  entry=$(sed -E 's/^[a-z]+(\(([^)]+)\))?!?: (.*)$/\2|\3/' <<<"$subject")
  scope=${entry%%|*} desc=${entry#*|}
  [ "$entry" = "$subject" ] && scope="" desc=$subject
  line="- ${scope:+**$scope:** }$desc (${sha:0:7})"
  if [[ "$subject" =~ ^[a-z]+(\([^\)]+\))?!: ]] || grep -qE '^BREAKING[ -]CHANGE:' <<<"$body"; then
    breaking+=("$line")
  elif [[ "$subject" =~ ^feat(\(|:) ]]; then
    features+=("$line")
  elif [[ "$subject" =~ ^(fix|perf)(\(|:) ]]; then
    fixes+=("$line")
  elif [[ "$subject" =~ ^(docs|chore|ci)(\(|:) ]]; then
    # Housekeeping is left out of the notes unless it is a breaking change.
    continue
  else
    other+=("$line")
  fi
done < <(git log --no-merges --format=%H "$range")

section() {
  local title=$1
  shift
  [ $# -eq 0 ] && return
  printf '## %s\n\n' "$title"
  printf '%s\n' "$@"
  printf '\n'
}
section "Breaking changes" ${breaking[@]+"${breaking[@]}"}
section "Features" ${features[@]+"${features[@]}"}
section "Fixes" ${fixes[@]+"${fixes[@]}"}
section "Other changes" ${other[@]+"${other[@]}"}
