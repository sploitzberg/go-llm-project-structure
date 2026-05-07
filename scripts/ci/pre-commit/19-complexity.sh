#!/usr/bin/env bash
# scripts/ci/pre-commit/19-complexity.sh
# Fast complexity check on changed Go files only
# Runs gocyclo and gocognit with layer-specific thresholds

set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../../.." && pwd)"
cd "$ROOT"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
CYAN='\033[0;36m'
NC='\033[0m'

errors=0

# Check if complexity checking is enabled
if [[ -f ".complexity.yml" ]]; then
    enabled=$(grep "enabled:" .complexity.yml | head -1 | awk '{print $2}' || echo "true")
    if [[ "$enabled" == "false" ]]; then
        echo -e "${YELLOW}> Complexity checks disabled in .complexity.yml${NC}"
        exit 0
    fi
fi

# Check for required tools
if ! command -v gocyclo >/dev/null 2>&1; then
    echo -e "${RED}error:${NC} gocyclo not found. Install with: go install github.com/fzipp/gocyclo/cmd/gocyclo@latest"
    exit 1
fi

if ! command -v gocognit >/dev/null 2>&1; then
    echo -e "${RED}error:${NC} gocognit not found. Install with: go install github.com/uudashr/gocognit/cmd/gocognit@latest"
    exit 1
fi

# Get changed Go files
changed_files=$(git diff --cached --name-only --diff-filter=ACM | grep '\.go$' || true)

if [[ -z "$changed_files" ]]; then
    echo -e "${GREEN}> No Go files changed - skipping complexity check${NC}"
    exit 0
fi

echo -e "${CYAN}> Running complexity check on changed files${NC}"

# Layer detection and threshold application
detect_layer() {
    local file="$1"
    if [[ "$file" == internal/core/domain/* ]]; then
        echo "core_domain"
    elif [[ "$file" == internal/core/ports/* ]]; then
        echo "core_ports"
    elif [[ "$file" == internal/core/services/* ]]; then
        echo "core_services"
    elif [[ "$file" == internal/adapter/primary/* ]]; then
        echo "adapter_primary"
    elif [[ "$file" == internal/adapter/secondary/* ]]; then
        echo "adapter_secondary"
    elif [[ "$file" == cmd/* ]]; then
        echo "cmd"
    else
        echo "default"
    fi
}

# Get threshold from .complexity.yml or use default
get_threshold() {
    local layer="$1"
    local metric="$2"  # cyclomatic or cognitive
    
    if [[ -f ".complexity.yml" ]]; then
        # Extract threshold using simple grep/awk
        # Format: cyclomatic: 5 (indented under layer)
        local val=$(grep -A 5 "${layer}:" .complexity.yml 2>/dev/null | grep "${metric}:" | head -1 | awk '{print $2}' || true)
        if [[ -n "$val" ]]; then
            echo "$val"
            return
        fi
    fi
    
    # Defaults
    case "$layer" in
        core_domain) echo 5 ;;
        core_ports) echo 3 ;;
        core_services) echo 10 ;;
        adapter_primary) echo 15 ;;
        adapter_secondary) echo 12 ;;
        cmd) echo 15 ;;
        *) echo 10 ;;
    esac
}

# Check cyclomatic complexity
check_cyclomatic() {
    local file="$1"
    local layer=$(detect_layer "$file")
    local max=$(get_threshold "$layer" "cyclomatic")
    
    # gocyclo outputs: <complexity> <package> <function> <file:line>
    local violations=$(gocyclo "$file" 2>/dev/null | awk -v max="$max" '$1 > max {print}' || true)
    
    if [[ -n "$violations" ]]; then
        echo -e "${RED}Cyclomatic complexity violation in $file (layer: $layer, max: $max):${NC}"
        echo "$violations" | while read -r line; do
            local comp=$(echo "$line" | awk '{print $1}')
            local func=$(echo "$line" | awk '{print $3}')
            local loc=$(echo "$line" | awk '{print $4}')
            echo "  $func: $comp (max: $max) at $loc"
        done
        ((errors++))
    fi
}

# Check cognitive complexity
check_cognitive() {
    local file="$1"
    local layer=$(detect_layer "$file")
    local max=$(get_threshold "$layer" "cognitive")
    
    # gocognit outputs similar format
    local violations=$(gocognit "$file" 2>/dev/null | awk -v max="$max" '$1 > max {print}' || true)
    
    if [[ -n "$violations" ]]; then
        echo -e "${RED}Cognitive complexity violation in $file (layer: $layer, max: $max):${NC}"
        echo "$violations" | while read -r line; do
            local comp=$(echo "$line" | awk '{print $1}')
            local func=$(echo "$line" | awk '{print $3}')
            local loc=$(echo "$line" | awk '{print $4}')
            echo "  $func: $comp (max: $max) at $loc"
        done
        ((errors++))
    fi
}

# Process each changed file
for file in $changed_files; do
    if [[ -f "$file" ]]; then
        check_cyclomatic "$file"
        check_cognitive "$file"
    fi
done

if ((errors > 0)); then
    echo
    echo -e "${RED}error:${NC} Complexity check FAILED with $errors violation(s)"
    echo "  Fix the violations or adjust thresholds in .complexity.yml"
    exit 1
else
    echo -e "${GREEN}Complexity check: OK${NC}"
    echo
    exit 0
fi
