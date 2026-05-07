#!/usr/bin/env bash
# scripts/ci/pre-push/05-complexity.sh
# Full complexity analysis with CRAP scoring
# Runs on all Go files before push

set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../../.." && pwd)"
cd "$ROOT"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
CYAN='\033[0;36m'
NC='\033[0m'

errors=0
total_funcs=0
violation_funcs=0

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

echo -e "${CYAN}> Running full complexity analysis${NC}"

# Layer detection
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

# Get threshold from config
get_threshold() {
    local layer="$1"
    local metric="$2"

    if [[ -f ".complexity.yml" ]]; then
        local val=$(grep -A 5 "${layer}:" .complexity.yml 2>/dev/null | grep "${metric}:" | head -1 | awk '{print $2}' || true)
        if [[ -n "$val" ]]; then
            echo "$val"
            return
        fi
    fi

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

# Get CRAP threshold
get_crap_threshold() {
    if [[ -f ".complexity.yml" ]]; then
        local val=$(grep -A 3 "crap:" .complexity.yml 2>/dev/null | grep "threshold:" | head -1 | awk '{print $2}' || true)
        if [[ -n "$val" ]]; then
            echo "$val"
            return
        fi
    fi
    echo 30
}

# Build coverage map if available
declare -A coverage_map

load_coverage() {
    local coverage_file="${1:-coverage.out}"

    if [[ ! -f "$coverage_file" ]]; then
        # Try to generate it
        echo -e "${YELLOW}  Generating coverage profile...${NC}"
        go test -coverprofile="$coverage_file" -covermode=atomic ./... 2>/dev/null || true
    fi

    if [[ -f "$coverage_file" ]]; then
        # Parse coverage.out format:
        # github.com/sploitzberg/go-llm-project-structure/internal/core/domain.User.Name:1.1,2.2 2 1
        # Format: <file:func>:<start>,<end> <statements> <covered>
        while IFS= read -r line; do
            if [[ "$line" =~ ^github\.com/[^/]+/[^/]+/(.+):(.+)\.(.+): ]]; then
                local pkg_func="${BASH_REMATCH[2]}.${BASH_REMATCH[3]}"
                local coverage_info=$(echo "$line" | awk '{print $3 "/" $4}')
                coverage_map["$pkg_func"]="$coverage_info"
            fi
        done < "$coverage_file" 2>/dev/null || true
    fi
}

# Calculate CRAP score: (cyclomatic² × (1 - coverage/100)³) + cyclomatic
calculate_crap() {
    local cyclomatic="$1"
    local coverage_pct="${2:-0}"

    # Convert to float for calculation
    local cov_frac=$(echo "scale=6; $coverage_pct / 100" | bc 2>/dev/null || echo "0")
    local one_minus_cov=$(echo "scale=6; 1 - $cov_frac" | bc 2>/dev/null || echo "1")

    # (1 - coverage)³
    local cubed=$(echo "scale=6; $one_minus_cov * $one_minus_cov * $one_minus_cov" | bc 2>/dev/null || echo "1")

    # cyclomatic² × cubed
    local sq=$(echo "scale=6; $cyclomatic * $cyclomatic" | bc 2>/dev/null || echo "$cyclomatic")
    local product=$(echo "scale=6; $sq * $cubed" | bc 2>/dev/null || echo "$sq")

    # + cyclomatic
    local crap=$(echo "scale=2; $product + $cyclomatic" | bc 2>/dev/null || echo "$cyclomatic")

    # Return integer part
    echo "${crap%.*}"
}

# Parse gocyclo output format: <complexity> <package> <function> <file:line>
# Function name may be "Type.Method" format
# We need to match with coverage format

get_coverage_for_function() {
    local pkg="$1"
    local func="$2"

    # Try exact match
    local key="${pkg}.${func}"
    if [[ -n "${coverage_map[$key]:-}" ]]; then
        local info="${coverage_map[$key]}"
        local total=$(echo "$info" | cut -d'/' -f1)
        local covered=$(echo "$info" | cut -d'/' -f2)
        if [[ "$total" -gt 0 ]]; then
            echo "scale=2; ($covered / $total) * 100" | bc 2>/dev/null || echo "0"
            return
        fi
    fi

    # Try without last part (for methods)
    if [[ "$func" =~ \. ]]; then
        local base=$(echo "$func" | rev | cut -d'.' -f2- | rev)
        key="${pkg}.${base}"
        if [[ -n "${coverage_map[$key]:-}" ]]; then
            local info="${coverage_map[$key]}"
            local total=$(echo "$info" | cut -d'/' -f1)
            local covered=$(echo "$info" | cut -d'/' -f2)
            if [[ "$total" -gt 0 ]]; then
                echo "scale=2; ($covered / $total) * 100" | bc 2>/dev/null || echo "0"
                return
            fi
        fi
    fi

    echo "0"
}

# Run analysis on all packages
echo -e "${CYAN}  Analyzing cyclomatic complexity...${NC}"

# Collect all violations - initialize with empty string to avoid unbound issues
cyclo_violations=()
cog_violations=()
crap_violations=()

# Process all Go files - cyclomatic
while IFS= read -r line; do
    [[ -z "$line" ]] && continue

    total_funcs=$((total_funcs + 1))

    # Parse: <complexity> <package> <function> <file:line>
    comp=$(echo "$line" | awk '{print $1}')
    pkg=$(echo "$line" | awk '{print $2}')
    func=$(echo "$line" | awk '{print $3}')
    loc=$(echo "$line" | awk '{print $4}')
    file=$(echo "$loc" | cut -d: -f1)

    layer=$(detect_layer "$file")
    max_cyclo=$(get_threshold "$layer" "cyclomatic")

    # Check cyclomatic
    if [[ "$comp" -gt "$max_cyclo" ]]; then
        cyclo_violations+=("$file|$func|$comp|$max_cyclo|$layer|$loc")
        violation_funcs=$((violation_funcs + 1))
    fi
done < <(find . -name '*.go' -not -path './vendor/*' -exec gocyclo {} + 2>/dev/null || true)

# Check cognitive complexity
echo -e "${CYAN}  Analyzing cognitive complexity...${NC}"

while IFS= read -r line; do
    [[ -z "$line" ]] && continue

    comp=$(echo "$line" | awk '{print $1}')
    pkg=$(echo "$line" | awk '{print $2}')
    func=$(echo "$line" | awk '{print $3}')
    loc=$(echo "$line" | awk '{print $4}')
    file=$(echo "$loc" | cut -d: -f1)

    layer=$(detect_layer "$file")
    max_cog=$(get_threshold "$layer" "cognitive")

    if [[ "$comp" -gt "$max_cog" ]]; then
        cog_violations+=("$file|$func|$comp|$max_cog|$layer|$loc")
    fi
done < <(find . -name '*.go' -not -path './vendor/*' -exec gocognit {} + 2>/dev/null || true)

# CRAP scoring
echo -e "${CYAN}  Calculating CRAP scores...${NC}"
load_coverage
crap_threshold=$(get_crap_threshold)

# Recalculate CRAP for functions with coverage data
while IFS= read -r line; do
    [[ -z "$line" ]] && continue

    cyclo=$(echo "$line" | awk '{print $1}')
    pkg=$(echo "$line" | awk '{print $2}')
    func=$(echo "$line" | awk '{print $3}')
    loc=$(echo "$line" | awk '{print $4}')
    file=$(echo "$loc" | cut -d: -f1)

    coverage=$(get_coverage_for_function "$pkg" "$func")
    crap=$(calculate_crap "$cyclo" "$coverage")

    if [[ "$crap" -gt "$crap_threshold" ]]; then
        crap_violations+=("$file|$func|$crap|$crap_threshold|$cyclo|$coverage|$loc")
    fi
done < <(find . -name '*.go' -not -path './vendor/*' -exec gocyclo {} + 2>/dev/null || true)

# Report violations
cyclo_count=${#cyclo_violations[@]}
if [[ $cyclo_count -gt 0 ]]; then
    echo
    echo -e "${RED}Cyclomatic Complexity Violations:${NC}"
    printf "%-40s %-30s %8s %8s %15s\n" "File" "Function" "Actual" "Max" "Layer"
    echo "-----------------------------------------------------------------------------------------------"
    for v in "${cyclo_violations[@]}"; do
        IFS='|' read -r file func comp max layer loc <<< "$v"
        printf "%-40s %-30s %8s %8s %15s\n" "$file" "$func" "$comp" "$max" "$layer"
    done
    errors=$((errors + cyclo_count))
fi

cog_count=${#cog_violations[@]}
if [[ $cog_count -gt 0 ]]; then
    echo
    echo -e "${RED}Cognitive Complexity Violations:${NC}"
    printf "%-40s %-30s %8s %8s %15s\n" "File" "Function" "Actual" "Max" "Layer"
    echo "-----------------------------------------------------------------------------------------------"
    for v in "${cog_violations[@]}"; do
        IFS='|' read -r file func comp max layer loc <<< "$v"
        printf "%-40s %-30s %8s %8s %15s\n" "$file" "$func" "$comp" "$max" "$layer"
    done
    errors=$((errors + cog_count))
fi

crap_count=${#crap_violations[@]}
if [[ $crap_count -gt 0 ]]; then
    echo
    echo -e "${RED}CRAP Score Violations (threshold: $crap_threshold):${NC}"
    printf "%-40s %-30s %8s %8s %12s\n" "File" "Function" "CRAP" "Cyclo" "Coverage"
    echo "-----------------------------------------------------------------------------------------------"
    for v in "${crap_violations[@]}"; do
        IFS='|' read -r file func crap cyclo coverage loc <<< "$v"
        printf "%-40s %-30s %8s %8s %11s%%\n" "$file" "$func" "$crap" "$cyclo" "$coverage"
    done
    errors=$((errors + crap_count))
fi

# Summary
echo
if ((errors > 0)); then
    echo -e "${RED}error:${NC} Complexity analysis FAILED"
    echo "  Total functions analyzed: $total_funcs"
    echo "  Functions with violations: $violation_funcs"
    echo "  Total violations: $errors"
    echo
    echo "  Fix the violations or adjust thresholds in .complexity.yml"
    echo "  CRAP = (cyclomatic² × (1 - coverage/100)³) + cyclomatic"
    exit 1
else
    echo -e "${GREEN}Complexity analysis: OK${NC}"
    echo "  Total functions analyzed: $total_funcs"
    echo "  All metrics within thresholds"
    echo
    exit 0
fi
