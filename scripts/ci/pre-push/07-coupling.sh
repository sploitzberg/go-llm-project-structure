#!/usr/bin/env bash
# scripts/ci/pre-push/07-coupling.sh
# Dependency & coupling analysis using goda
# Checks fan-out per hexagonal architecture layer

set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../../.." && pwd)"
cd "$ROOT"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
CYAN='\033[0;36m'
NC='\033[0m'

errors=0

# Check if coupling checks are enabled
if [[ -f ".coupling.yml" ]]; then
    enabled=$(grep "^enabled:" .coupling.yml | awk '{print $2}' || echo "true")
    if [[ "$enabled" == "false" ]]; then
        echo -e "${YELLOW}> Coupling checks disabled in .coupling.yml${NC}"
        exit 0
    fi
fi

# Check for goda
if ! command -v goda >/dev/null 2>&1; then
    echo -e "${RED}error:${NC} goda not found."
    echo -e "${YELLOW}> Install with: go install github.com/loov/goda@latest${NC}"
    exit 1
fi

# Read thresholds from config (with defaults)
threshold_for_layer() {
    local layer="$1"
    local default="$2"
    if [[ -f ".coupling.yml" ]]; then
        local val
        val=$(awk "/^  ${layer}:/,/^  [a-z]/" .coupling.yml \
            | grep "max-fan-out:" | head -1 | awk '{print $2}')
        echo "${val:-$default}"
    else
        echo "$default"
    fi
}

THRESHOLD_DOMAIN=$(threshold_for_layer "core_domain" "5")
THRESHOLD_PORTS=$(threshold_for_layer "core_ports" "5")
THRESHOLD_SERVICES=$(threshold_for_layer "core_services" "10")
THRESHOLD_PRIMARY=$(threshold_for_layer "adapter_primary" "20")
THRESHOLD_SECONDARY=$(threshold_for_layer "adapter_secondary" "15")

MODULE=$(go list -m)

# Check fan-out for a given package path and threshold
# Fan-out = number of external (non-stdlib, non-self) imports
check_fanout() {
    local layer_name="$1"
    local pkg_path="$2"
    local threshold="$3"

    # Skip if directory does not exist
    if [[ ! -d "$pkg_path" ]]; then
        return 0
    fi

    # Skip if no Go packages in path
    if ! go list "./${pkg_path}/..." 2>/dev/null | grep -q .; then
        return 0
    fi

    # List all imports of this layer, subtract stdlib and self-module packages
    local fanout
    fanout=$(goda list "./${pkg_path}/...:import" 2>/dev/null \
        | grep -v "^ID$" \
        | grep -v "^${MODULE}" \
        | grep -v "^std " \
        | grep -v "^$" \
        | wc -l | tr -d ' ')

    if [[ "$fanout" -gt "$threshold" ]]; then
        echo -e "${RED}  FAIL${NC} [${layer_name}] fan-out=${fanout} exceeds max=${threshold}"
        echo -e "       Path: ${pkg_path}"
        echo -e "       Run: goda list \"./${pkg_path}/...:import\" to see dependencies"
        errors=$((errors + 1))
    else
        echo -e "${GREEN}  OK${NC}   [${layer_name}] fan-out=${fanout} (max=${threshold})"
    fi
}

echo -e "${CYAN}> Running coupling analysis (fan-out per layer)${NC}"

check_fanout "core/domain"       "internal/core/domain"       "$THRESHOLD_DOMAIN"
check_fanout "core/ports"        "internal/core/ports"        "$THRESHOLD_PORTS"
check_fanout "core/services"     "internal/core/services"     "$THRESHOLD_SERVICES"
check_fanout "adapter/primary"   "internal/adapter/primary"   "$THRESHOLD_PRIMARY"
check_fanout "adapter/secondary" "internal/adapter/secondary" "$THRESHOLD_SECONDARY"

if [[ "$errors" -gt 0 ]]; then
    echo -e "\n${RED}> Coupling analysis failed: ${errors} violation(s)${NC}"
    echo -e "${YELLOW}> To fix: reduce imports in the affected layer or raise thresholds in .coupling.yml${NC}"
    exit 1
fi

echo -e "\n${GREEN}> Coupling analysis passed${NC}"
