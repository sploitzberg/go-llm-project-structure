#!/usr/bin/env bash
# scripts/ci/pre-push/02-hex-arch-guardrail.sh
# Check hexagonal architecture rules

set -euo pipefail

echo "> Checking hexagonal architecture rules"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GUARDRAIL="$SCRIPT_DIR/../hex-arch-guardrail.sh"

if [[ ! -f "$GUARDRAIL" ]]; then
    echo "error: required architecture guardrail not found: $GUARDRAIL" >&2
    exit 1
fi

exec bash "$GUARDRAIL"
