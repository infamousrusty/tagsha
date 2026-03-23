#!/usr/bin/env bash
set -euo pipefail

git config user.name "github-actions[bot]"
git config user.email "41898282+github-actions[bot]@users.noreply.github.com"

# Ensure we are on a branch (not detached HEAD)
BRANCH="adr/auto-${GITHUB_RUN_ID:-$(date +%s)}"

git fetch origin main:main
git checkout -B "$BRANCH" main

# Stage changes
git add docs/decisions/

# Exit if nothing to commit
if git diff --cached --quiet; then
  echo "No ADR changes"
  exit 0
fi

git commit -m "docs(adr): auto-generate ADRs [skip ci]"

# Optional: reconcile with latest main using 'ours' strategy
# This keeps our ADR changes if conflicts occur
git fetch origin main

if ! git merge -X ours origin/main --no-edit; then
  echo "Merge conflicts resolved using 'ours' strategy"
fi

# Push feature branch (NOT main)
git push -u origin "$BRANCH"

echo "Branch pushed: $BRANCH"