#!/usr/bin/env bash
# scripts/ci/pre-push/02-hex-arch-guardrail.sh
# Check hexagonal architecture rules

set -euo pipefail

echo "> Checking hexagonal architecture rules"

if [[ -f ./scripts/ci/hex-arch-guardrail.sh ]]; then
    ./scripts/ci/hex-arch-guardrail.sh
else
    echo "warning: hex-arch-guardrail.sh not found (skipped)"
fi
echo
