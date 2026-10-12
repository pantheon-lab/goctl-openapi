#!/usr/bin/env bash
# Prints the next semver tag based on Conventional Commits since the last v*
# tag, or nothing if no commit warrants a release.
#   breaking (type!: or BREAKING CHANGE) -> major
#   feat                                 -> minor
#   fix, perf, refactor                  -> patch
#   anything else (chore, ci, docs, ...) -> no release
set -euo pipefail

last=$(git describe --tags --abbrev=0 --match 'v[0-9]*' 2>/dev/null || true)
if [[ -n "$last" ]]; then
  range="$last..HEAD"
else
  last="v0.0.0"
  range="HEAD"
fi

log=$(git log --no-merges --format='%s%n%b' "$range")
[[ -z "$log" ]] && exit 0

if grep -qE '^[a-z]+(\([^)]*\))?!:|^BREAKING[ -]CHANGE:' <<<"$log"; then
  bump=major
elif grep -qE '^feat(\([^)]*\))?:' <<<"$log"; then
  bump=minor
elif grep -qE '^(fix|perf|refactor)(\([^)]*\))?:' <<<"$log"; then
  bump=patch
else
  exit 0
fi

IFS=. read -r major minor patch <<<"${last#v}"
case $bump in
  major) major=$((major + 1)); minor=0; patch=0 ;;
  minor) minor=$((minor + 1)); patch=0 ;;
  patch) patch=$((patch + 1)) ;;
esac
echo "v$major.$minor.$patch"
