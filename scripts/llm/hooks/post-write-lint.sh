#!/usr/bin/env bash
# scripts/llm/hooks/post-write-lint.sh
# Post-write hook: Fast lint feedback after LLM writes code
# Reports lint issues back to the LLM so they can be fixed immediately

set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../../.." && pwd)"
cd "$ROOT"

file_path=$(echo "$1" | jq -r '.file_path // empty' 2>/dev/null || echo "")

if [[ -z "$file_path" ]] || [[ "$file_path" != *.go ]]; then
    exit 0
fi

# Only lint files in the project
if [[ "$file_path" != ./* ]] && [[ "$file_path" != internal/* ]] && [[ "$file_path" != cmd/* ]]; then
    exit 0
fi

RED='\033[0;31m'
YELLOW='\033[1;33m'
GREEN='\033[0;32m'
NC='\033[0m'

# Fast auto-fix for formatting and imports
gofmt -w "$file_path" 2>/dev/null || true
goimports -w "$file_path" 2>/dev/null || true

# Run all linters with project config
issues=$(golangci-lint run --config .golangci.yml "$file_path" 2>&1) || true

if [[ -n "$issues" ]] && [[ "$issues" != *"no go files"* ]]; then
    echo -e "${RED}=== LINT ISSUES in $file_path ===${NC}"
    echo "$issues"
    echo -e "${YELLOW}Fix these issues before proceeding.${NC}"
    echo "Run: golangci-lint run --fix $file_path"
    echo -e "${RED}=== END LINT ISSUES ===${NC}"
    exit 1
fi

echo -e "${GREEN}Lint check passed for $file_path${NC}"
exit 0
