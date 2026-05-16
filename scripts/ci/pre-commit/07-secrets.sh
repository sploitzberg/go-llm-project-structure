#!/usr/bin/env bash
# scripts/ci/pre-commit/07-secrets.sh
# Run secret scanning

set -euo pipefail

echo "> Scanning for secrets"

if [[ -f ./scripts/ci/check-secrets.sh ]]; then
    ./scripts/ci/check-secrets.sh
else
    echo "Warning: check-secrets.sh not found (skipped)"
fi
echo
