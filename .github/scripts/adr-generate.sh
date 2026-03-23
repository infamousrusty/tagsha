#!/usr/bin/env bash
set -euo pipefail

# --- Config ---
BASE_BRANCH="${BASE_BRANCH:-main}"

# --- Ensure we have full history ---
git fetch origin "${BASE_BRANCH}:${BASE_BRANCH}" --depth=0 || git fetch origin "${BASE_BRANCH}"

# --- Resolve HEAD safely ---
HEAD_SHA="${GITHUB_SHA:-$(git rev-parse HEAD)}"

# --- Safety checks ---
if [ -z "${HEAD_SHA}" ]; then
  echo "ERROR: HEAD_SHA is empty"
  exit 1
fi

if [ -z "$(git rev-parse "$BASE_BRANCH" 2>/dev/null || true)" ]; then
  echo "ERROR: BASE_BRANCH '$BASE_BRANCH' not found locally"
  exit 1
fi

# --- Diff commits between base and head ---
COMMITS="$(git rev-list --reverse "${BASE_BRANCH}..${HEAD_SHA}" || true)"

# Nothing to process
[ -z "$COMMITS" ] && exit 0

mkdir -p docs/decisions

yaml_escape() {
  printf '%s' "$1" | sed 's/\\/\\\\/g; s/"/\\"/g'
}

slugify() {
  printf '%s' "$1" \
    | tr '[:upper:]' '[:lower:]' \
    | sed -E 's/[^a-z0-9]+/-/g; s/^-+|-+$//g' \
    | cut -c1-50
}

# --- Track existing ADR commits ---
has_adr_for_commit() {
  grep -Rqs "source_commit: $1" docs/decisions 2>/dev/null
}

# --- Get next ADR ID ---
max_id=0
while IFS= read -r file; do
  [ -n "$file" ] || continue
  id_part="$(basename "$file" | sed -E 's/^([0-9]{4})-.*/\1/')"
  if printf '%s' "$id_part" | grep -Eq '^[0-9]{4}$'; then
    n=$((10#$id_part))
    [ "$n" -gt "$max_id" ] && max_id="$n"
  fi
done < <(find docs/decisions -maxdepth 1 -type f -name '*.md' 2>/dev/null | sort)

next_id() {
  max_id=$((max_id + 1))
  printf '%04d' "$max_id"
}

# --- Generate ADRs ---
count=0

for sha in $COMMITS; do
  files="$(git show --format= --name-only "$sha" | sed '/^$/d' || true)"
  [ -z "$files" ] && continue

  # Skip if no relevant files
  if ! printf '%s\n' "$files" | grep -Eq \
    '(^|/)(terraform|config|\.github/workflows)/|(^|/)Dockerfile(\..*)?$|(^|/)[^/]+\.go$|^go\.mod$'; then
    continue
  fi

  has_adr_for_commit "$sha" && continue

  id="$(next_id)"
  title="$(git show -s --format=%s "$sha" | sed 's/[[:space:]]\+/ /g')"
  slug="$(slugify "$title")"
  file="docs/decisions/${id}-${slug}.md"

  printf '%s\n' \
"---" \
"adr_id: \"$id\"" \
"source_commit: \"$sha\"" \
"source_type: \"commit\"" \
"status: \"proposed\"" \
"tags: \"general\"" \
"summary: \"$(yaml_escape "$title")\"" \
"---" \
"" \
"# ADR $id: $(yaml_escape "$title")" \
"" \
"## Status" \
"proposed" \
"" \
"## Timestamp" \
"$(date -u '+%Y-%m-%d %H:%M UTC')" \
"" \
"## Context" \
"\`\`\`" \
"$files" \
"\`\`\`" \
"" \
"## Decision" \
"accepted" \
> "$file"

  count=$((count + 1))
done

echo "count=$count"