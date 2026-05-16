#!/usr/bin/env bash
# Check for gremlins installation and run full mutation testing

set -euo pipefail

if ! command -v gremlins >/dev/null 2>&1; then
  echo "error: gremlins not found. Install with:"
  echo "  go install github.com/go-gremlins/gremlins/cmd/gremlins@v0.6.0"
  exit 1
fi

TARGET="${1:-internal/core/domain}"

echo "> Running Gremlins mutation testing..."
echo "    Target: $TARGET"
echo "    This may take several minutes. Use 'task mutation-test-dry' for a fast check."

gremlins unleash --tags=integration "$TARGET"
