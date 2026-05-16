#!/usr/bin/env bash
# scripts/ci/pre-push/03-coverage.sh
# Check test coverage threshold

set -euo pipefail

if ! command -v go >/dev/null 2>&1; then
    echo "error: go not found"
    exit 1
fi

if ! command -v bc >/dev/null 2>&1; then
    echo "error: bc not found. Install with your package manager"
    exit 1
fi

echo "> Running test coverage check"

COVERAGE_THRESHOLD=${COVERAGE_THRESHOLD:-80}

# Run tests with coverage (show output for debugging)
if ! go test -coverprofile=coverage.out -covermode=atomic ./...; then
    echo "error: Tests failed during coverage check"
    rm -f coverage.out
    exit 1
fi

# Get total coverage
coverage=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')

if [ -z "$coverage" ]; then
    echo "error: Could not determine coverage"
    rm -f coverage.out
    exit 1
fi

echo "Current coverage: ${coverage}%"
echo "Required coverage: ${COVERAGE_THRESHOLD}%"

# Compare with threshold (using bc for floating point comparison)
result=$(echo "$coverage >= $COVERAGE_THRESHOLD" | bc -l 2>/dev/null || echo "0")
if [ "$result" = "0" ]; then
    echo "error: Coverage below threshold. Current: ${coverage}%, Required: ${COVERAGE_THRESHOLD}%"
    rm -f coverage.out
    exit 1
fi

echo "Coverage: OK (${coverage}%)"

# Clean up
rm -f coverage.out
