#!/usr/bin/env bash
# Compile adapter packages so explicit port implementation assertions are verified.

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
cd "$ROOT"

echo "> Validating adapter port contracts"

if ! command -v go >/dev/null 2>&1; then
    echo "error: go not found" >&2
    exit 1
fi

if ! find internal/adapter -type f -name '*.go' -print -quit 2>/dev/null | grep -q .; then
    echo "No adapter packages found, skipping"
    exit 0
fi

# The architecture guardrail requires each secondary-adapter package to contain
# an assertion such as: var _ secondaryport.Repository = (*Repository)(nil).
# Compiling the packages makes Go verify those explicit contracts.
go test -run '^$' ./internal/adapter/...

echo "Adapter port contracts: OK"
echo
