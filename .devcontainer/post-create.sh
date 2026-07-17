#!/bin/bash
# Post-create script for dev container
# Sets up git hooks and validates environment

set -e

cd /workspace

echo "Setting up git hooks..."
chmod +x .githooks/*
git config core.hooksPath .githooks

echo "Validating environment..."
command -v go >/dev/null 2>&1 || echo "Warning: go not found"
command -v task >/dev/null 2>&1 || echo "Warning: task not found"
command -v golangci-lint >/dev/null 2>&1 || echo "Warning: golangci-lint not found"
command -v gopls >/dev/null 2>&1 || echo "Warning: gopls not found"
command -v dlv >/dev/null 2>&1 || echo "Warning: dlv not found"

echo "Dev container setup complete!"
echo "Go version: $(go version)"
