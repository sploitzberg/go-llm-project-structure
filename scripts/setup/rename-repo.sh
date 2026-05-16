#!/usr/bin/env bash
# scripts/setup/rename-repo.sh
# Automates renaming the repository when forking this template
# Updates all references from the upstream repo to the new repo URL

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
cd "$PROJECT_ROOT"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Current repo URL (from go.mod)
CURRENT_MODULE=$(go list -m 2>/dev/null || echo "github.com/sploitzberg/go-llm-project-structure")
CURRENT_REPO=$(echo "$CURRENT_MODULE" | sed 's/^[^/]*\///')

echo "=== Repository Rename Script ==="
echo ""
echo "Current module: $CURRENT_MODULE"
echo "Current repo: $CURRENT_REPO"
echo ""

# Get new repo info
if [ -z "${1:-}" ]; then
    read -p "Enter new GitHub username/org: " NEW_OWNER
else
    NEW_OWNER="$1"
fi

if [ -z "${2:-}" ]; then
    read -p "Enter new repository name: " NEW_REPO_NAME
else
    NEW_REPO_NAME="$2"
fi

NEW_MODULE="github.com/${NEW_OWNER}/${NEW_REPO_NAME}"

echo ""
echo "New module: $NEW_MODULE"
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

# Function to replace in file
replace_in_file() {
    local file="$1"
    local old_pattern="$2"
    local new_text="$3"

    if [ -f "$file" ]; then
        sed -i.bak "s|$old_pattern|$new_text|g" "$file"
        rm -f "${file}.bak"
        echo "✓ Updated: $file"
    fi
}

# Update go.mod
replace_in_file "go.mod" "$CURRENT_MODULE" "$NEW_MODULE"

# Update .golangci.yml
replace_in_file ".golangci.yml" "$CURRENT_MODULE" "$NEW_MODULE"

# Update README.md
replace_in_file "README.md" "$CURRENT_REPO" "${NEW_OWNER}/${NEW_REPO_NAME}"
replace_in_file "README.md" "$CURRENT_MODULE" "$NEW_MODULE"

# Update CHANGELOG.md (comparison URLs)
replace_in_file "CHANGELOG.md" "$CURRENT_REPO" "${NEW_OWNER}/${NEW_REPO_NAME}"

# Update test validators (test code)
replace_in_file "scripts/test-validators.sh" "$CURRENT_MODULE" "$NEW_MODULE"

# Update LLM platform configs
replace_in_file "scripts/llm/platforms/windsurf/config.yaml" "$CURRENT_MODULE" "$NEW_MODULE"
replace_in_file "scripts/llm/platforms/cursor/config.yaml" "$CURRENT_MODULE" "$NEW_MODULE" 2>/dev/null || true
replace_in_file "scripts/llm/platforms/claude/config.yaml" "$CURRENT_MODULE" "$NEW_MODULE" 2>/dev/null || true
replace_in_file "scripts/llm/platforms/codex/config.yaml" "$CURRENT_MODULE" "$NEW_MODULE" 2>/dev/null || true
replace_in_file "scripts/llm/platforms/continue/config.yaml" "$CURRENT_MODULE" "$NEW_MODULE" 2>/dev/null || true
replace_in_file "scripts/llm/platforms/copilot/config.yaml" "$CURRENT_MODULE" "$NEW_MODULE" 2>/dev/null || true

# Update AGENTS.md (root level)
replace_in_file "AGENTS.md" "$CURRENT_MODULE" "$NEW_MODULE"

echo ""
echo "Running go mod tidy..."
go mod tidy

echo ""
echo "${GREEN}✓ Repository renamed successfully!${NC}"
echo ""
echo "Next steps:"
echo "1. Update the 'repository' URL in .github/workflows/ci.yml if needed"
echo "2. Update the 'homepage' and 'repository' URLs in go.mod if needed"
echo "3. Commit the changes"
echo "4. Push to your new repository"
echo ""
