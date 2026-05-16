#!/usr/bin/env bash
# scripts/ci/pre-push/00-build.sh
# Build the project

set -euo pipefail

if ! command -v go >/dev/null 2>&1; then
    echo "error: go not found"
    exit 1
fi

echo "> Running go build"

# Detect host OS and architecture for output directory naming
GOOS=$(go env GOOS)
GOARCH=$(go env GOARCH)
BINDIR="bin/${GOOS}-${GOARCH}"
EXE=""
if [[ "$GOOS" == "windows" ]]; then
    EXE=".exe"
fi

BINARY_NAME="go-llm-project-structure"

# Create output directory and build
mkdir -p "$BINDIR"
go build -o "${BINDIR}/${BINARY_NAME}${EXE}" ./cmd/go-llm-project-structure

echo "Build: OK"
echo
