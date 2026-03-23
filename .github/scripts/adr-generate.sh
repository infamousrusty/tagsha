#!/usr/bin/env bash
set -euo pipefail

BASE_SHA="${BASE_SHA:-}"
HEAD_SHA="${HEAD_SHA:-}"
RUN_URL="${RUN_URL:-}"

mkdir -p docs/decisions

yaml_escape() {
  printf '%s' "$1" | sed 's/\\/\\\\/g; s/"/\\"/g'
}

classify_tags() {
  local title_lower
  declare -A tags=()

  title_lower="$(printf '%s' "$1" | tr '[:upper:]' '[:lower:]')"

  while IFS= read -r path; do
    [ -n "$path" ] || continue

    case "$path" in
      terraform/*|*/terraform/*) tags[infra]=1 ;;
      .github/workflows/*|*/.github/workflows/*) tags[ci]=1 ;;
      config/*|*/config/*) tags[config]=1 ;;
    esac

    case "$path" in
      *.go|go.mod) tags[runtime]=1 ;;
    esac
  done

  if printf '%s' "$title_lower" | grep -Eq '(auth|security|secret|token|jwt|tls|crypto|encrypt|password|credential)'; then
    tags[security]=1
  fi

  if [ "${#tags[@]}" -eq 0 ]; then
    tags[general]=1
  fi

  printf '%s\n' "${!tags[@]}" | sort | paste -sd ', ' -
}

# Resolve base
if [ -z "$BASE_SHA" ] || printf '%s' "$BASE_SHA" | grep -Eq '^0+$'; then
  BASE_SHA="$(git rev-list --max-parents=0 "$HEAD_SHA" | tail -n 1)"
fi

COMMITS="$(git rev-list --reverse "${BASE_SHA}..${HEAD_SHA}" || true)"

[ -z "$COMMITS" ] && exit 0

# Skip ADR-only changes
CHANGED_FILES=$(for sha in $COMMITS; do
  git show --name-only --format= "$sha"
done | sed '/^$/d')

if ! printf '%s\n' "$CHANGED_FILES" | grep -qv '^docs/decisions/'; then
  exit 0
fi

# Determine next ADR ID
max_id=0
while IFS= read -r file; do
  [ -n "$file" ] || continue
  id_part="$(basename "$file" | sed -E 's/^([0-9]{4})-.*/\1/')"
  if printf '%s' "$id_part" | grep -Eq '^[0-9]{4}$'; then
    n=$((10#$id_part))
    [ "$n" -gt "$max_id" ] && max_id="$n"
  fi
done < <(find docs/decisions -maxdepth 1 -type f -name '*.md' | sort)

next_id() {
  max_id=$((max_id + 1))
  printf '%04d' "$max_id"
}

slugify() {
  printf '%s' "$1" \
    | tr '[:upper:]' '[:lower:]' \
    | sed -E 's/[^a-z0-9]+/-/g; s/^-+|-+$//g' \
    | cut -c1-50
}

has_adr_for_commit() {
  grep -Rqs "source_commit: $1" docs/decisions 2>/dev/null
}

count=0

for sha in $COMMITS; do
  files="$(git show --format= --name-only "$sha" | sed '/^$/d' || true)"
  [ -z "$files" ] && continue

  if ! printf '%s\n' "$files" | grep -Eq \
    '(^|/)(terraform|config|\.github/workflows)/|(^|/)Dockerfile(\..*)?$|(^|/)[^/]+\.go$|^go\.mod$'; then
    continue
  fi

  has_adr_for_commit "$sha" && continue

  id="$(next_id)"
  title="$(git show -s --format=%s "$sha" | sed 's/[[:space:]]\+/ /g')"
  slug="$(slugify "$title")"
  file="docs/decisions/${id}-${slug}.md"

  tags="$(printf '%s\n' "$files" | classify_tags "$title")"

  printf '%s\n' \
"---" \
"adr_id: \"$id\"" \
"source_commit: \"$sha\"" \
"source_type: \"commit\"" \
"source_run: \"$RUN_URL\"" \
"status: \"proposed\"" \
"tags: \"$(yaml_escape "$tags")\"" \
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