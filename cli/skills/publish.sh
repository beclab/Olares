#!/usr/bin/env bash
# Publish the 9 olares-cli skills under cli/skills/ to ClawHub (clawhub.ai).
#
# Prerequisites:
#   1. Node.js 22+ (clawhub uses ES2025 import attributes).
#   2. npm i -g clawhub
#   3. Either run `clawhub login` interactively first, OR export CLAWHUB_TOKEN
#      (a non-interactive token from https://clawhub.ai/account/tokens) so
#      this script can call `clawhub login --token "$CLAWHUB_TOKEN"`.
#
# Usage:
#   ./publish.sh                  # publish all 9 skills (real upload)
#   ./publish.sh --dry-run        # local validation only — no network
#   ./publish.sh olares-files     # publish a single skill by slug
#   ./publish.sh --dry-run olares-shared olares-files
#
# Note: `clawhub skill publish` does NOT support a server-side --dry-run.
# The --dry-run flag here only does local frontmatter checks + prints the
# commands that *would* run. For a real server-side preview, use:
#   clawhub sync --workdir cli --dry-run     # run from repo root
# (NOT plain `clawhub sync --dry-run` — that falls back to the OpenClaw
# user workspace and would mass-republish unrelated skills.)
#
# The CLI binary `olares-cli` itself is NOT published here — it ships with
# Olares. Each skill declares it as a host requirement via
# metadata.openclaw.requires.bins: ["olares-cli"].

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# slug | display name
SKILLS=(
  "olares-shared|Olares Shared (olares-cli foundation)"
  "olares-files|Olares Files (olares-cli files)"
  "olares-search|Olares Search (olares-cli search)"
  "olares-market|Olares Market (olares-cli market)"
  "olares-settings|Olares Settings (olares-cli settings)"
  "olares-dashboard|Olares Dashboard (olares-cli dashboard)"
  "olares-cluster|Olares Cluster (olares-cli cluster)"
  "olares-doctor|Olares Doctor (runtime diagnosis)"
  "olares-chart|Olares Chart (olares-cli chart)"
  "olares-publish|Olares Publish (Olares Market distribution)"
)

DRY_RUN=""
TARGETS=()
CHANGELOG="${CLAWHUB_CHANGELOG:-Automated publish from cli/skills/publish.sh}"
TAGS="${CLAWHUB_TAGS:-latest}"

for arg in "$@"; do
  case "$arg" in
    --dry-run) DRY_RUN="--dry-run" ;;
    -h|--help)
      sed -n '2,24p' "$0"
      exit 0
      ;;
    *) TARGETS+=("$arg") ;;
  esac
done

if [[ -z "$DRY_RUN" ]]; then
  if ! command -v clawhub >/dev/null 2>&1; then
    echo "ERROR: clawhub CLI not found. Run: npm i -g clawhub" >&2
    exit 1
  fi
  if [[ -n "${CLAWHUB_TOKEN:-}" ]]; then
    echo ">> logging in via CLAWHUB_TOKEN"
    clawhub login --token "$CLAWHUB_TOKEN" >/dev/null
  fi
fi

# Read `version: <x.y.z>` from a SKILL.md frontmatter.
read_version() {
  local file="$1"
  awk '/^---[[:space:]]*$/{n++; next} n==1 && /^version:[[:space:]]*/ {sub(/^version:[[:space:]]*/, ""); gsub(/[[:space:]"]/, ""); print; exit}' "$file"
}

# Read `name: <slug>` from a SKILL.md frontmatter.
read_name() {
  local file="$1"
  awk '/^---[[:space:]]*$/{n++; next} n==1 && /^name:[[:space:]]*/ {sub(/^name:[[:space:]]*/, ""); gsub(/[[:space:]"]/, ""); print; exit}' "$file"
}

# Read quoted `description:` value length from frontmatter (OpenCode max 1024).
read_description_len() {
  local file="$1"
  awk '/^---[[:space:]]*$/{n++; next} n==1 && /^description:[[:space:]]*"/ {
    sub(/^description:[[:space:]]*"/, ""); sub(/"[[:space:]]*$/, ""); print length($0); exit
  }' "$file"
}

# Local validation: check required frontmatter fields + semver.
validate_skill() {
  local dir="$1" slug="$2" file="$dir/SKILL.md"
  local errs=()

  [[ -f "$file" ]] || { echo "  - SKILL.md missing"; return 1; }
  head -n1 "$file" | grep -q '^---[[:space:]]*$' || errs+=("frontmatter does not start with --- on line 1")

  local fm_name fm_version desc_len
  fm_name="$(read_name "$file")"
  fm_version="$(read_version "$file")"
  desc_len="$(read_description_len "$file")"
  [[ -n "$fm_name" ]] || errs+=("missing name: in frontmatter")
  [[ "$fm_name" == "$slug" ]] || errs+=("frontmatter name '$fm_name' != folder slug '$slug'")
  [[ -n "$fm_version" ]] || errs+=("missing version: in frontmatter")
  [[ "$fm_version" =~ ^[0-9]+\.[0-9]+\.[0-9]+([+-][0-9A-Za-z.-]+)?$ ]] \
    || errs+=("version '$fm_version' is not valid semver")
  grep -qE '^description:' "$file" || errs+=("missing description: in frontmatter")
  if [[ -z "$desc_len" ]]; then
    errs+=("description must be a single quoted line in frontmatter")
  elif [[ "$desc_len" -gt 1024 ]]; then
    errs+=("description length $desc_len exceeds OpenCode limit of 1024")
  fi
  grep -qE '^compatibility:' "$file" || errs+=("missing compatibility: in frontmatter")
  grep -qE '^metadata:' "$file" || errs+=("missing metadata: block in frontmatter")
  grep -qE '^[[:space:]]+openclaw:' "$file" || errs+=("missing metadata.openclaw block in frontmatter")
  grep -qE '^[[:space:]]+- olares-cli[[:space:]]*$' "$file" || errs+=("metadata.openclaw.requires.bins must list olares-cli")
  if grep -qE '^  requires:[[:space:]]*$' "$file"; then
    errs+=("deprecated top-level metadata.requires — use metadata.openclaw.requires.bins")
  fi

  if [[ ${#errs[@]} -gt 0 ]]; then
    for e in "${errs[@]}"; do echo "  - $e"; done
    return 1
  fi
  return 0
}

# Decide whether a slug is in the user-supplied target list (empty list = all).
is_target() {
  local slug="$1"
  if [[ ${#TARGETS[@]} -eq 0 ]]; then
    return 0
  fi
  for t in "${TARGETS[@]}"; do
    [[ "$t" == "$slug" ]] && return 0
  done
  return 1
}

FAILED=()
for entry in "${SKILLS[@]}"; do
  slug="${entry%%|*}"
  name="${entry#*|}"
  is_target "$slug" || continue

  dir="$SCRIPT_DIR/$slug"
  if [[ ! -f "$dir/SKILL.md" ]]; then
    echo "ERROR: $dir/SKILL.md not found" >&2
    FAILED+=("$slug")
    continue
  fi

  version="$(read_version "$dir/SKILL.md")"
  if [[ -z "$version" ]]; then
    echo "ERROR: could not read 'version:' from $dir/SKILL.md frontmatter" >&2
    FAILED+=("$slug")
    continue
  fi

  echo
  echo "=================================================="
  if [[ -n "$DRY_RUN" ]]; then
    echo ">> [dry-run] $slug v$version"
  else
    echo ">> publishing $slug v$version"
  fi
  echo "=================================================="

  if [[ -n "$DRY_RUN" ]]; then
    if validate_skill "$dir" "$slug"; then
      echo "  local validation OK"
      echo "  would run: clawhub skill publish $dir \\"
      echo "               --slug $slug --name \"$name\" --version $version \\"
      echo "               --tags \"$TAGS\" --changelog \"$CHANGELOG\""
    else
      echo "!! FAILED: $slug v$version (local validation)" >&2
      FAILED+=("$slug")
    fi
    continue
  fi

  if clawhub skill publish "$dir" \
    --slug "$slug" \
    --name "$name" \
    --version "$version" \
    --tags "$TAGS" \
    --changelog "$CHANGELOG"; then
    echo ">> OK: $slug v$version"
  else
    echo "!! FAILED: $slug v$version" >&2
    FAILED+=("$slug")
  fi
done

if [[ ${#FAILED[@]} -gt 0 ]]; then
  echo
  echo "Failed slugs: ${FAILED[*]}" >&2
  exit 1
fi

echo
echo "All requested skills published successfully."
