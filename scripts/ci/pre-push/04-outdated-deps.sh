#!/usr/bin/env bash
# scripts/ci/pre-push/04-outdated-deps.sh
# Check for outdated dependencies

set -euo pipefail

if ! command -v go >/dev/null 2>&1; then
    echo "error: go not found"
    exit 1
fi

echo "> Checking for outdated dependencies"

# Check for outdated direct dependencies
outdated=$(go list -u -m all 2>/dev/null | grep -v "^\[" | awk '{if ($2 != $3) print $0}')

if [ -n "$outdated" ]; then
    echo "error: Outdated dependencies found:"
    echo "$outdated"
    echo ""
    echo "Update with: go get -u ./..."
    echo "Or run: go list -u -m all"
    exit 1
fi

echo "Dependencies: OK"
echo
