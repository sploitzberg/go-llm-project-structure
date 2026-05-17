#!/usr/bin/env bash
# scripts/ci/pre-commit/02-golangci-lint.sh
# Run golangci-lint with configurable timeout

set -euo pipefail

if ! command -v golangci-lint >/dev/null 2>&1; then
    echo "error: golangci-lint not found. Install with: go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest"
    exit 1
fi
echo "> Running golangci-lint"

# Use GOLANGCI_LINT_TIMEOUT env var or default to 2m for pre-commit
TIMEOUT=${GOLANGCI_LINT_TIMEOUT:-2m}

golangci-lint run --timeout="$TIMEOUT" --fix
echo "golangci-lint: OK"
echo
