#!/usr/bin/env bash
# scripts/ci/pre-commit/03-tests.sh
# Run tests with configurable flags

set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../../.." && pwd)"
cd "$ROOT"

if ! command -v go >/dev/null 2>&1; then
    echo "error: go not found"
    exit 1
fi

echo "> Running tests"

# Use GO_TEST_FLAGS env var or default to -short for pre-commit
GO_TEST_FLAGS="${GO_TEST_FLAGS:--short}"

go test $GO_TEST_FLAGS ./...
echo "Tests: OK"
echo
