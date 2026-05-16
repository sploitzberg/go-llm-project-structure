#!/bin/bash
# Post-create script for dev container
# Sets up git hooks and validates environment

set -e

cd /workspace

echo "Setting up git hooks..."
mkdir -p .git/hooks
cp -r .githooks/* .git/hooks/
chmod +x .git/hooks/*

echo "Validating environment..."
command -v go >/dev/null 2>&1 || echo "Warning: go not found"
command -v task >/dev/null 2>&1 || echo "Warning: task not found"
command -v golangci-lint >/dev/null 2>&1 || echo "Warning: golangci-lint not found"
command -v jq >/dev/null 2>&1 || echo "Warning: jq not found"
command -v yq >/dev/null 2>&1 || echo "Warning: yq not found"

echo "Dev container setup complete!"
echo "Go version: $(go version)"
