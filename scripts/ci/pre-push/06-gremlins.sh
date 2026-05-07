#!/usr/bin/env bash
# scripts/ci/pre-push/06-gremlins.sh
# Mutation testing using Gremlins for semantic stability
# Runs on domain layer before push

set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../../.." && pwd)"
cd "$ROOT"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
CYAN='\033[0;36m'
NC='\033[0m'

# Check if gremlins is enabled
if [[ -f ".gremlins.yml" ]]; then
    enabled=$(grep "enabled:" .gremlins.yml 2>/dev/null | head -1 | awk '{print $2}' || echo "true")
    if [[ "$enabled" == "false" ]]; then
        echo -e "${YELLOW}> Gremlins disabled in .gremlins.yml${NC}"
        exit 0
    fi
fi

# Check for gremlins tool
if ! command -v gremlins >/dev/null 2>&1; then
    echo -e "${RED}error:${NC} gremlins not found."
    echo -e "${YELLOW}> Install with: go install github.com/go-gremlins/gremlins/cmd/gremlins@v0.6.0${NC}"
    exit 1
fi

echo -e "${CYAN}> Running Gremlins mutation testing (domain layer)${NC}"

# Get target directory from config or default to domain layer
target_dir="internal/core/domain"
if [[ -f ".gremlins.yml" ]]; then
    configured_target=$(grep "target:" .gremlins.yml 2>/dev/null | head -1 | awk '{print $2}' || echo "")
    if [[ -n "$configured_target" ]]; then
        target_dir="$configured_target"
    fi
fi

# Check if target directory exists
if [[ ! -d "$target_dir" ]]; then
    echo -e "${YELLOW}> Target directory $target_dir does not exist, skipping${NC}"
    exit 0
fi

# Check if there are any Go packages in the target directory
if ! go list "$target_dir/..." 2>/dev/null | grep -q .; then
    echo -e "${YELLOW}> No Go packages found in $target_dir, skipping mutation testing${NC}"
    exit 0
fi

echo -e "${CYAN}> Target: $target_dir${NC}"

# Run gremlins with dry-run mode for speed
# Dry-run reports which mutants would be tested without actually running tests
echo -e "${CYAN}> Running in dry-run mode (fast)${NC}"
if gremlins unleash --dry-run --tags=integration "$target_dir"; then
    echo -e "${GREEN}> Gremlins dry-run passed${NC}"
else
    echo -e "${RED}> Gremlins dry-run failed${NC}"
    echo -e "${YELLOW}> For full mutation testing, run: make mutation-test${NC}"
    exit 1
fi
