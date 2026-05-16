#!/usr/bin/env bash
# scripts/ci/pre-push/01-tests.sh
# Run full test suite

set -euo pipefail

if ! command -v go >/dev/null 2>&1; then
    echo "error: go not found"
    exit 1
fi
echo "> Running full test suite"

go test ./...
echo "Tests: OK"
echo
