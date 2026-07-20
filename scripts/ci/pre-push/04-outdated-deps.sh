#!/usr/bin/env bash
# scripts/ci/pre-push/04-outdated-deps.sh
# Check for outdated dependencies

set -euo pipefail

if ! command -v go >/dev/null 2>&1; then
    echo "error: go not found"
    exit 1
fi

echo "> Checking for outdated dependencies"

# Check direct dependencies only (not transitive). Keep package-load failures fatal.
outdated=$(go list -u -m -f '{{if and (not .Main) (not .Indirect) .Update}}{{.Path}} {{.Version}} -> {{.Update.Version}}{{end}}' all)

if [ -n "$outdated" ]; then
    echo "error: Outdated dependencies found:"
    echo "$outdated"
    echo ""
    echo "Update with: go get -u ./..."
    exit 1
fi

echo "Dependencies: OK"
echo
