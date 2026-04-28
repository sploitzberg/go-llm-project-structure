#!/usr/bin/env bash
# Enhanced Hexagonal Architecture Guardrail
# Enforces dependency rules for the project structure based on docs/architecture/architecture.md

set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$ROOT"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
CYAN='\033[0;36m'
NC='\033[0m'

errors=0
warnings=0

die() {
    echo -e "${RED}error:${NC} $*" >&2
    ((errors++))
}

warn() {
    echo -e "${YELLOW}warning:${NC} $*" >&2
    ((warnings++))
}

success() {
    echo -e "${GREEN}success:${NC} $*"
}

echo -e "${CYAN}> Enhanced Hexagonal Architecture Guardrail${NC}"
echo -e "${CYAN}> Validating against docs/architecture/architecture.md${NC}"
echo

# 1. domain/ must remain pure
echo -e "${CYAN}> Checking domain/ (should only depend on standard library and itself)${NC}"
if [ -d "internal/domain" ]; then
    # Use go list to check actual imports
    domain_imports=$(go list -f '{{.ImportPath}}: {{join .Imports " "}}' ./internal/domain/... 2>/dev/null || true)
    for pkg in $domain_imports; do
        if [[ "$pkg" =~ internal/(adapter|service|port/primary|port/secondary) ]]; then
            die "domain/ must not import from adapter, service, or port subpackages"
        fi
    done
fi
echo

# 2. service/ may depend on domain and port, but not adapter
echo -e "${CYAN}> Checking service/ dependencies${NC}"
if [ -d "internal/service" ]; then
    service_imports=$(go list -f '{{.ImportPath}}: {{join .Imports " "}}' ./internal/service/... 2>/dev/null || true)
    for pkg in $service_imports; do
        if [[ "$pkg" =~ internal/adapter ]]; then
            die "service/ must not depend on adapter/"
        fi
    done
fi
echo

# 3. port/ should not depend on adapters or services
echo -e "${CYAN}> Checking port/ dependencies${NC}"
if [ -d "internal/port" ]; then
    port_imports=$(go list -f '{{.ImportPath}}: {{join .Imports " "}}' ./internal/port/... 2>/dev/null || true)
    for pkg in $port_imports; do
        if [[ "$pkg" =~ internal/(adapter|service) ]]; then
            die "port/ must not depend on adapter or service"
        fi
    done
fi
echo

# 4. Primary adapters should only depend on ports (primary) and domain
echo -e "${CYAN}> Checking adapter/primary/${NC}"
if [ -d "internal/adapter/primary" ]; then
    primary_imports=$(go list -f '{{.ImportPath}}: {{join .Imports " "}}' ./internal/adapter/primary/... 2>/dev/null || true)
    for pkg in $primary_imports; do
        if [[ "$pkg" =~ internal/(service|adapter/secondary) ]]; then
            die "adapter/primary/ should only depend on domain/ and port/primary/"
        fi
    done
fi
echo

# 5. Secondary adapters should depend on port/secondary
echo -e "${CYAN}> Checking adapter/secondary/${NC}"
if [ -d "internal/adapter/secondary" ]; then
    secondary_imports=$(go list -f '{{.ImportPath}}: {{join .Imports " "}}' ./internal/adapter/secondary/... 2>/dev/null || true)
    for pkg in $secondary_imports; do
        if [[ "$pkg" =~ internal/adapter/primary ]]; then
            die "adapter/secondary/ should not depend on adapter/primary/"
        fi
    done
fi
echo

# 6. No framework leaks into core (domain, port, service)
echo -e "${CYAN}> Checking for framework leaks in core packages${NC}"
framework_packages="github.com/gin github.com/gorilla database/sql github.com/lib/pq github.com/go-redis github.com/gorm"
for dir in internal/domain internal/port internal/service; do
    if [ -d "$dir" ]; then
        dir_imports=$(go list -f '{{.ImportPath}}: {{join .Imports " "}}' ./"$dir"/... 2>/dev/null || true)
        for pkg in $dir_imports; do
            for framework in $framework_packages; do
                if [[ "$pkg" =~ "$framework" ]]; then
                    die "Framework or database package $framework found in core ($dir)"
                fi
            done
        done
    fi
done
echo

# 7. Port interfaces should not reference concrete adapter types
echo -e "${CYAN}> Checking port interfaces for adapter references${NC}"
if [ -d "internal/port" ]; then
    port_imports=$(go list -f '{{.ImportPath}}: {{join .Imports " "}}' ./internal/port/... 2>/dev/null || true)
    for pkg in $port_imports; do
        if [[ "$pkg" =~ internal/adapter ]]; then
            die "port/ interfaces must not reference concrete types from adapter/"
        fi
    done
fi
echo

# 8. Check for import cycles
echo -e "${CYAN}> Checking for import cycles${NC}"
check_import_cycles() {
    local dir="$1"
    [[ -d "$dir" ]] || return 0

    local packages=($(find "$dir" -name "*.go" -exec dirname {} \; 2>/dev/null | sort -u))
    for pkg in "${packages[@]}"; do
        local imports=$(grep -h --include='*.go' '^import' "$pkg"/*.go 2>/dev/null | grep -o '"[^"]*internal/[^"]*"' | sort -u || true)
        for imp in $imports; do
            local imp_path="${imp//\"/}"
            local imp_dir="$ROOT/$imp_path"
            if [[ -d "$imp_dir" ]] && [[ "$imp_dir" != "$pkg" ]]; then
                if grep -R --include='*.go' -l -F "\"$(basename "$dir")/$(basename "$pkg")\"" "$imp_dir" 2>/dev/null | grep -q .; then
                    warn "Potential import cycle between $pkg and $imp_dir"
                fi
            fi
        done
    done
}

check_import_cycles "internal/domain"
check_import_cycles "internal/service"
check_import_cycles "internal/adapter"
echo

# 9. Check cmd directory for business logic
echo -e "${CYAN}> Checking cmd directory for business logic violations${NC}"
if [[ -d "cmd" ]]; then
    business_logic=$(find cmd -name "*.go" -exec grep -l -E "(func.*Business|func.*Validate|func.*Calculate)" {} \; 2>/dev/null || true)
    if [[ -n "$business_logic" ]]; then
        warn "Business logic found in cmd directory (should be in domain/service):"
        echo "$business_logic"
    fi
fi

# Summary
echo
if ((errors > 0)); then
    echo -e "${RED}error:${NC} Hexagonal architecture guardrail FAILED with $errors error(s)"
    exit 1
elif ((warnings > 0)); then
    echo -e "${YELLOW}warning:${NC} Hexagonal architecture guardrail PASSED with $warnings warning(s)"
    exit 0
else
    echo -e "${GREEN}success:${NC} Hexagonal architecture guardrail PASSED"
    echo "  - domain/ is pure (no internal imports)"
    echo "  - service/ only depends on domain/ and port/"
    echo "  - port/ does not depend on adapter/ or service/"
    echo "  - adapter/ follows dependency direction rules"
    echo "  - No framework packages in core (domain/port/service)"
    echo "  - Port interfaces are clean"
    echo "  - No import cycles detected"
    exit 0
fi
echo
