#!/usr/bin/env bash
# Single entry point for all quality checks.
# Run locally with: ./scripts/ci/ci.sh or `task ci`

set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$ROOT"

die() {
    echo "error: $*" >&2
    exit 1
}

need_cmd() {
    command -v "$1" >/dev/null 2>&1 || die "missing required command: $1"
}

need_cmd go
need_cmd git

# Run pre-commit checks with CI-specific configurations
echo "> Running pre-commit checks"
./scripts/ci/pre-commit/02-golangci-lint.sh
./scripts/ci/pre-commit/03-tests.sh
./scripts/ci/pre-commit/04-hex-arch-guardrail.sh
./scripts/ci/pre-commit/07-secrets.sh
./scripts/ci/pre-commit/12-file-quality.sh
./scripts/ci/pre-commit/13-interface-impl.sh
./scripts/ci/pre-commit/14-exported-symbols.sh
./scripts/ci/pre-commit/17-struct-fields.sh

# CI-specific: go mod tidy validation
echo "> Checking dependencies"
go mod tidy

if [[ "${CI:-}" == "true" ]] || git ls-files --error-unmatch go.mod >/dev/null 2>&1; then
    if [[ -n "$(git diff --name-only -- go.mod go.sum 2>/dev/null)" ]]; then
        echo "go.mod or go.sum is not up to date"
        git diff -- go.mod go.sum >&2 || true
        die "Run: go mod tidy && git add go.mod go.sum"
    fi
fi

# CI-specific: vulnerability check
echo "> Checking for known vulnerabilities"
go run golang.org/x/vuln/cmd/govulncheck@v1.6.0 ./...

# Run pre-push checks for CI
echo "> Running pre-push checks"
./scripts/ci/pre-push/00-build.sh
./scripts/ci/pre-push/04-outdated-deps.sh

# Run mutation testing (gremlins) for semantic stability
echo "> Running mutation testing (gremlins)"
./scripts/ci/pre-push/06-gremlins.sh

# Run coupling analysis
echo "> Running coupling analysis"
./scripts/ci/pre-push/07-coupling.sh

echo "> OK"
