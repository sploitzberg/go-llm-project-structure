#!/usr/bin/env bash
# scripts/ci/pre-commit/12-file-quality.sh
# Check for file quality issues (trailing whitespace, merge conflicts, large files)

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
cd "$ROOT"

echo "> Running file quality checks"

# Check for trailing whitespace.
trailing_whitespace=$(find . \
    \( -path './.git' -o -path './.tmp' -o -path './bin' -o -path './dist' -o -path './vendor' -o -path './node_modules' \) -prune -o \
    -type f \( -name '*.go' -o -name '*.md' -o -name '*.yaml' -o -name '*.yml' -o -name '*.json' \) \
    -exec grep -Il '[[:blank:]]$' {} + 2>/dev/null || true)
if [ -n "$trailing_whitespace" ]; then
    echo "error: Files with trailing whitespace:"
    echo "$trailing_whitespace"
    echo "Fix by running: find . -type f -exec sed -i 's/[[:space:]]*$//' {} +"
    exit 1
fi

# Check for all Git merge conflict marker forms.
merge_conflicts=$(find . \
    \( -path './.git' -o -path './.tmp' -o -path './bin' -o -path './dist' -o -path './vendor' -o -path './node_modules' \) -prune -o \
    -type f \( -name '*.go' -o -name '*.md' -o -name '*.yaml' -o -name '*.yml' -o -name '*.json' \) \
    -exec grep -EIl '^(<{7}|={7}|>{7})( |$)' {} + 2>/dev/null || true)
if [ -n "$merge_conflicts" ]; then
    echo "error: Files with merge conflict markers:"
    echo "$merge_conflicts"
    echo "Resolve merge conflicts before committing"
    exit 1
fi

# Check for large files (max 512kb) - only check source/text files
large_files=$(find . \
    \( -path './.git' -o -path './.tmp' -o -path './bin' -o -path './dist' -o -path './vendor' -o -path './node_modules' \) -prune -o \
    -type f -size +512k \( -name '*.go' -o -name '*.md' -o -name '*.yaml' -o -name '*.yml' -o -name '*.json' -o -name '*.txt' \) \
    -print 2>/dev/null || true)
if [ -n "$large_files" ]; then
    echo "error: Large files found (>512kb):"
    echo "$large_files"
    echo "Consider using git-lfs or removing these files"
    exit 1
fi

echo "File quality: OK"
echo
