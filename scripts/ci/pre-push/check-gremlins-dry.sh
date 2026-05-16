#!/usr/bin/env bash
# Check for gremlins installation and run dry-run mutation testing

set -euo pipefail

if ! command -v gremlins >/dev/null 2>&1; then
  echo "error: gremlins not found. Install with:"
  echo "  go install github.com/go-gremlins/gremlins/cmd/gremlins@v0.6.0"
  exit 1
fi

TARGET="${1:-internal/core/domain}"

echo "> Running Gremlins mutation testing (dry-run)..."
gremlins unleash --dry-run --tags=integration "$TARGET"
