#!/usr/bin/env bash
# scripts/setup/rename-repo.sh
# Automates renaming the repository when forking this template.
# Updates all references from the upstream repo to the new repo URL.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
cd "$PROJECT_ROOT"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Current identifiers (from go.mod)
CURRENT_MODULE=$(go list -m 2>/dev/null || echo "github.com/sploitzberg/go-llm-project-structure")
CURRENT_OWNER=$(echo "$CURRENT_MODULE" | cut -d'/' -f2)
CURRENT_REPO=$(echo "$CURRENT_MODULE" | sed 's/^[^/]*\///')
CURRENT_PROJECT_NAME=$(basename "$CURRENT_MODULE")

echo "=== Repository Rename Script ==="
echo ""
echo "Usage:    ./scripts/setup/rename-repo.sh github.com/<owner>/<repo>"
echo "Example:  ./scripts/setup/rename-repo.sh github.com/acme/my-service"
echo ""
echo "Current module:  $CURRENT_MODULE"
echo "Current owner:   $CURRENT_OWNER"
echo "Current repo:    $CURRENT_REPO"
echo ""

# Parse input: expects github.com/owner/repo, strips https:// if present
if [ $# -eq 0 ]; then
    read -p "Enter new module path (e.g. 'github.com/acme/my-service'): " raw_input
else
    raw_input="$1"
fi

# Strip https:// prefix if present, then validate
cleaned=$(printf '%s' "$raw_input" | sed 's|^https://||')

# Must be exactly github.com/owner/repo (three segments)
segments=$(printf '%s' "$cleaned" | awk -F'/' '{print NF}')
if [ "$segments" -ne 3 ] || [ "$(printf '%s' "$cleaned" | cut -d'/' -f1)" != "github.com" ]; then
    echo -e "${RED}error: input must be github.com/<owner>/<repo>${NC}"
    echo "  Example: github.com/acme/my-service"
    exit 1
fi

NEW_OWNER=$(printf '%s' "$cleaned" | cut -d'/' -f2)
NEW_REPO_NAME=$(printf '%s' "$cleaned" | cut -d'/' -f3)
NEW_MODULE="github.com/${NEW_OWNER}/${NEW_REPO_NAME}"

echo ""
echo "Resulting module path: $NEW_MODULE"
echo "Resulting clone URL:   https://github.com/${NEW_OWNER}/${NEW_REPO_NAME}.git"
echo ""

read -p "Continue? (y/N) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Aborted."
    exit 0
fi

echo ""
echo "Updating repository references..."
echo ""

# Escape sed replacement text (escapes & and \)
escape_sed_repl() {
    printf '%s' "$1" | sed 's/&/\\&/g; s/\\\\/\\\\\\\\/g'
}

# Replace pattern in a single file, reporting if changed
replace_in_file() {
    local file="$1"
    local old="$2"
    local new="$3"

    [ -f "$file" ] || return 0
    grep -q "$old" "$file" 2>/dev/null || return 0

    local escaped_new
    escaped_new=$(escape_sed_repl "$new")
    sed -i.bak "s|${old}|${escaped_new}|g" "$file"
    rm -f "${file}.bak"
    echo "  ✓ Updated: $file"
}

# Find tracked text files containing a pattern, excluding the script itself
find_files_with() {
    local pattern="$1"
    grep -rl --binary-files=without-match \
         --exclude-dir={.git,bin,dist,node_modules,vendor} \
         --exclude='*.png' --exclude='*.jpg' --exclude='*.jpeg' \
         --exclude='*.gif' --exclude='*.ico' --exclude='*.pdf' \
         --exclude='*.zip' --exclude='*.tar.gz' --exclude='*.tar' \
         "$pattern" . 2>/dev/null | grep -v 'scripts/setup/rename-repo.sh' || true
}

# 1. Replace full module path (most specific first)
echo "> Replacing module path"
while IFS= read -r file; do
    [ -n "$file" ] && replace_in_file "$file" "$CURRENT_MODULE" "$NEW_MODULE"
done < <(find_files_with "$CURRENT_MODULE")

# 2. Replace repo path (owner/repo) — used in clone URLs, CHANGELOG, etc.
echo "> Replacing repo path"
while IFS= read -r file; do
    [ -n "$file" ] && replace_in_file "$file" "$CURRENT_REPO" "${NEW_OWNER}/${NEW_REPO_NAME}"
done < <(find_files_with "$CURRENT_REPO")

# 3. Replace bare project name — used in binary names, container names, titles, paths
echo "> Replacing project name"
while IFS= read -r file; do
    [ -n "$file" ] && replace_in_file "$file" "$CURRENT_PROJECT_NAME" "$NEW_REPO_NAME"
done < <(find_files_with "$CURRENT_PROJECT_NAME")

# 4. Replace owner/author name in known contexts (copyright, goheader, codecov)
echo "> Replacing owner name"
while IFS= read -r file; do
    [ -n "$file" ] || continue
    grep -q "$CURRENT_OWNER" "$file" 2>/dev/null || continue

    sed -i.bak "s|Copyright (c) [0-9]\{4\} ${CURRENT_OWNER}|Copyright (c) $(date +%Y) ${NEW_OWNER}|g" "$file"
    sed -i.bak "s|author: ${CURRENT_OWNER}|author: ${NEW_OWNER}|g" "$file"
    sed -i.bak "s|codecov\.io/gh/${CURRENT_OWNER}/|codecov.io/gh/${NEW_OWNER}/|g" "$file"
    rm -f "${file}.bak"
    echo "  ✓ Updated: $file"
done < <(find_files_with "$CURRENT_OWNER")

# 5. Update devcontainer.json name only if it follows the auto-generated pattern
if [ -f ".devcontainer/devcontainer.json" ]; then
    old_dev_name=$(grep '"name"' .devcontainer/devcontainer.json | head -1 | sed 's/.*"name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/')
    expected_old_name="Go ${CURRENT_PROJECT_NAME}"
    if [ -n "$old_dev_name" ] && [ "$old_dev_name" = "$expected_old_name" ]; then
        escaped_old=$(escape_sed_repl "$old_dev_name")
        escaped_new=$(escape_sed_repl "Go ${NEW_REPO_NAME}")
        sed -i.bak "s|\"name\": \"${escaped_old}\"|\"name\": \"${escaped_new}\"|" .devcontainer/devcontainer.json
        rm -f .devcontainer/devcontainer.json.bak
        echo "  ✓ Updated: .devcontainer/devcontainer.json"
    fi
fi

# 6. Update docker-compose container name (auto, no prompt)
replace_in_file ".devcontainer/docker-compose.yml" "container_name: ${CURRENT_PROJECT_NAME}" "container_name: ${NEW_REPO_NAME}"

# 7. Rename cmd/ directory if it matches the old project name
old_cmd_dir="cmd/${CURRENT_PROJECT_NAME}"
new_cmd_dir="cmd/${NEW_REPO_NAME}"
if [ -d "$old_cmd_dir" ] && [ ! -d "$new_cmd_dir" ]; then
    mv "$old_cmd_dir" "$new_cmd_dir"
    echo "  ✓ Renamed: $old_cmd_dir → $new_cmd_dir"
elif [ -d "$new_cmd_dir" ]; then
    echo "  ℹ  cmd/${NEW_REPO_NAME} already exists, skipping rename"
fi

echo ""
echo "Running go mod tidy..."
go mod tidy

echo ""
echo -e "${GREEN}✓ Repository renamed successfully!${NC}"
echo ""
echo "Next steps:"
echo "1. Review changes with: git diff"
echo "2. Update README.md title and description if needed"
echo "3. Commit the changes"
echo "4. Push to your new repository"
echo ""
