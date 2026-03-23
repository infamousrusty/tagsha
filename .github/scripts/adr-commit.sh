#!/usr/bin/env bash
set -euo pipefail

git config user.name "github-actions[bot]"
git config user.email "41898282+github-actions[bot]@users.noreply.github.com"

# Stage changes
git add docs/decisions/

# Exit if nothing to commit
if git diff --cached --quiet; then
  exit 0
fi

# Commit
git commit -m "docs(adr): auto-generate ADRs [skip ci]"

# Detect if working tree is dirty after commit
if ! git diff --quiet || ! git diff --cached --quiet; then
  echo "Stashing remaining changes before rebase"
  git stash push -u -m "adr-autogen-stash"
  STASHED=1
else
  STASHED=0
fi

# Sync with remote
git fetch origin main

# Rebase
git rebase origin/main

# Restore stash if it exists
if [ "$STASHED" -eq 1 ]; then
  git stash pop || {
    echo "Stash conflict detected — aborting"
    exit 1
  }
fi

# Push
git push origin HEAD:main