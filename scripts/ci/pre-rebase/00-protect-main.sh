#!/usr/bin/env bash
# scripts/ci/pre-rebase/00-protect-main.sh
# Prevent rebasing main/master branches

set -euo pipefail

branch="${2-}"
if [[ -z "$branch" ]]; then
    branch=$(git symbolic-ref --short HEAD 2>/dev/null || echo "")
fi
branch="${branch#refs/heads/}"

if [[ "$branch" == "main" ]] || [[ "$branch" == "master" ]]; then
    echo "error: Rebasing the $branch branch is not allowed."
    exit 1
fi
