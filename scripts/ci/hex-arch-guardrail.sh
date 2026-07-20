#!/usr/bin/env bash
# Enforce the dependency matrix documented in AGENTS.md and docs/architecture/architecture.md.

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT"

if ! command -v go >/dev/null 2>&1; then
    echo "error: missing required command: go" >&2
    exit 1
fi

if [[ ! -f go.mod ]]; then
    echo "error: go.mod not found at repository root: $ROOT" >&2
    exit 1
fi

echo "> Checking package loading and active-build import cycles"
go list -deps -test ./... >/dev/null

echo "> Checking hexagonal architecture dependencies in all Go files"
go run ./scripts/ci/hexarch
