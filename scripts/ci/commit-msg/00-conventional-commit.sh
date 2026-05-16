#!/usr/bin/env bash
# scripts/ci/commit-msg/00-conventional-commit.sh
# Validate commit message format (Conventional Commits)

set -euo pipefail

echo "> Validating conventional commit format"

msg_file="$1"
msg=$(cat "$msg_file")


if ! echo "$msg" | grep -qE '^(feat|fix|chore|docs|style|refactor|perf|test|build|ci|revert)(\([a-z0-9-]+\))?!?: '; then
    echo "error: Commit message does not follow Conventional Commits format."
    echo "Example: feat(auth): add login endpoint"
    echo "  - build: Changes that affect the build system or external dependencies"
    echo "  - ci: Changes to CI configuration files and scripts"
    echo "  - revert: Reverts a previous commit"
    exit 1
fi
echo
