#!/usr/bin/env bash
# scripts/ci/pre-push/03-coverage.sh
# Check test coverage threshold

set -euo pipefail

if ! command -v go >/dev/null 2>&1; then
    echo "error: go not found"
    exit 1
fi

echo "> Running test coverage check"

COVERAGE_THRESHOLD=${COVERAGE_THRESHOLD:-80}
coverage_file=$(mktemp)
trap 'rm -f "$coverage_file"' EXIT INT TERM

# Measure shipped application packages. CI/setup tools have focused regression suites.
if ! go test -coverprofile="$coverage_file" -covermode=atomic ./cmd/... ./internal/...; then
    echo "error: Tests failed during coverage check"
    exit 1
fi

# Get total coverage.
coverage=$(go tool cover -func="$coverage_file" | awk '/^total:/ {gsub(/%/, "", $3); print $3}')

if [ -z "$coverage" ]; then
    echo "error: Could not determine coverage"
    exit 1
fi

echo "Current coverage: ${coverage}%"
echo "Required coverage: ${COVERAGE_THRESHOLD}%"

if ! awk -v coverage="$coverage" -v threshold="$COVERAGE_THRESHOLD" 'BEGIN { exit !(coverage >= threshold) }'; then
    echo "error: Coverage below threshold. Current: ${coverage}%, Required: ${COVERAGE_THRESHOLD}%"
    exit 1
fi

echo "Coverage: OK (${coverage}%)"
