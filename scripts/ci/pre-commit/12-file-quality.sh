#!/usr/bin/env bash
# scripts/ci/pre-commit/12-file-quality.sh
# Check for file quality issues (trailing whitespace, merge conflicts, large files)

set -euo pipefail

echo "> Running file quality checks"

# Check for trailing whitespace
trailing_whitespace=$(grep -Rl ' $' --include='*.go' --include='*.md' --include='*.yaml' --include='*.yml' --include='*.json' . 2>/dev/null || true)
if [ -n "$trailing_whitespace" ]; then
    echo "error: Files with trailing whitespace:"
    echo "$trailing_whitespace"
    echo "Fix by running: find . -type f -exec sed -i 's/[[:space:]]*$//' {} +"
    exit 1
fi

# Check for merge conflict markers
merge_conflicts=$(grep -Rl '^[<>]{7}' --include='*.go' --include='*.md' --include='*.yaml' --include='*.yml' --include='*.json' . 2>/dev/null || true)
if [ -n "$merge_conflicts" ]; then
    echo "error: Files with merge conflict markers:"
    echo "$merge_conflicts"
    echo "Resolve merge conflicts before committing"
    exit 1
fi

# Check for large files (max 512kb) - only check source/text files
large_files=$(find . -type f -size +512k \( -name '*.go' -o -name '*.md' -o -name '*.yaml' -o -name '*.yml' -o -name '*.json' -o -name '*.txt' \) ! -path './.git/*' ! -path './.tmp/*' 2>/dev/null || true)
if [ -n "$large_files" ]; then
    echo "error: Large files found (>512kb):"
    echo "$large_files"
    echo "Consider using git-lfs or removing these files"
    exit 1
fi

echo "File quality: OK"
echo
