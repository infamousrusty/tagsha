#!/usr/bin/env bash
set -euo pipefail

git config user.name "github-actions[bot]"
git config user.email "41898282+github-actions[bot]@users.noreply.github.com"

# Ensure working tree is clean before proceeding
if ! git diff --quiet || ! git diff --cached --quiet; then
  git add docs/decisions/
fi

if git diff --cached --quiet; then
  exit 0
fi

# Commit locally (no push yet)
git commit -m "docs(adr): auto-generate ADRs [skip ci]"

# Rebase against latest main
git fetch origin main
git rebase origin/main

# Push after successful rebase
git push origin HEAD:main