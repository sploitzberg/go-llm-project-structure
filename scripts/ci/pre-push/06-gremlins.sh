#!/usr/bin/env bash
# scripts/ci/pre-push/06-gremlins.sh
# Run Gremlins mutation testing for semantic stability

set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../../.." && pwd)"
cd "$ROOT"

# Check if gremlins is enabled
if [[ -f ".gremlins.yml" ]]; then
    enabled=$(grep "enabled:" .gremlins.yml | head -1 | awk '{print $2}' || echo "true")
    if [[ "$enabled" == "false" ]]; then
        echo "> Gremlins disabled in .gremlins.yml"
        exit 0
    fi
fi

# Check for required tool
if ! command -v gremlins >/dev/null 2>&1; then
    echo "error: gremlins not found."
    echo "> Install with: go install github.com/go-gremlins/gremlins/cmd/gremlins@v0.6.0"
    exit 1
fi

target_dir="internal/core/domain"
echo "> Running Gremlins mutation testing (domain layer)"

# Check if target directory exists
if [[ ! -d "$target_dir" ]]; then
    echo "> Target directory $target_dir does not exist, skipping"
    exit 0
fi

# Check if there are Go files in the target directory
if ! find "$target_dir" -name "*.go" | grep -q .; then
    echo "> No Go packages found in $target_dir, skipping mutation testing"
    exit 0
fi

echo "> Target: $target_dir"

# Check if we should run in dry-run mode (CI default)
if [[ "${GREMLINS_MODE:-}" == "dry" ]] || [[ "${CI:-}" == "true" ]]; then
    echo "> Running in dry-run mode (fast)"
    gremlins unleash --dry-run --tags=integration "$target_dir"
    if [ $? -eq 0 ]; then
        echo "> Gremlins dry-run passed"
    else
        echo "> Gremlins dry-run failed"
        echo "> For full mutation testing, run: task mutation-test"
        exit 1
    fi
else
    echo "> Running full mutation testing (slow, thorough)"
    echo "> This may take several minutes..."
    gremlins unleash --tags=integration "$target_dir"
fi

echo "> Gremlins mutation testing: OK"
