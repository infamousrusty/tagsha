#!/usr/bin/env bash
set -euo pipefail

git config user.name "github-actions[bot]"
git config user.email "41898282+github-actions[bot]@users.noreply.github.com"

git add docs/decisions/

if git diff --cached --quiet; then
  exit 0
fi

git commit -m "docs(adr): auto-generate ADRs [skip ci]"

git fetch origin main
git rebase origin/main

git push origin HEAD:main