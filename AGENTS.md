# Project Instructions

You are working on a **production-grade Go codebase** that follows strict Hexagonal Architecture (Ports & Adapters).

## Your Responsibilities as an Agent

- **Strictly adhere** to all architectural rules, coding standards, and best practices defined in this repository.
- **Never cut corners** — even when making small changes or working under time pressure.
- Always prioritize **clean, robust, maintainable, and production-ready** code.
- Focus on **long-term code quality**, readability, and architectural integrity.
- If a solution feels messy, unclear, or "just works for now", refactor it properly.
- Security, testability, and strict separation of concerns are non-negotiable.

**Default to excellence.** Every change should improve the codebase or at least maintain its high standards.

---

## Hexagonal Architecture

This project uses **Hexagonal Architecture** (also known as **Ports & Adapters**), originally introduced by Alistair Cockburn.

### Core Concepts

- The **domain** is the heart of the application and must remain completely independent.
- All dependencies point **inward** toward the domain.
- External concerns (databases, HTTP, third-party APIs, etc.) are pushed to the edges as adapters.
- The core business logic is decoupled from frameworks, databases, and delivery mechanisms.

### Layer Responsibilities & Rules

- **`domain/`** — Pure business entities, value objects, and domain rules.
  **Must have zero dependencies** on any other internal package.

- **`port/`** — Defines clean interfaces (contracts).
  - `port/primary/` → Driving ports (inbound use cases)
  - `port/secondary/` → Driven ports (outbound contracts)

- **`service/`** — Application services that implement use cases.
  Depends **only** on `domain/` and `port/`.

- **`adapter/`** — Concrete implementations of ports.
  - `adapter/primary/` → Inbound (HTTP, CLI, gRPC, etc.)
  - `adapter/secondary/` → Outbound (databases, external APIs, cache, etc.)

- **`config/`** — Centralized configuration structures and loading.

### Strict Rules

- Never import `adapter/` from `domain/`, `port/`, or `service/`
- Never import `service/` from `domain/` or `port/`
- `domain/` must remain pure at all times
- All external interactions must go through ports
- Follow `.golangci.yml` (especially `depguard` rules)

**Security is mandatory** — see `rules/security.mdc`

### Full Documentation

- Detailed layer explanations → [`internal/README.md`](internal/README.md)
- Visual dependency graph + deeper explanation → [`docs/architecture.md`](docs/architecture.md)

---

## Go Style & Best Practices

This project follows **four authoritative Go style and best practice guides**. Agents and contributors should treat these as the primary sources of truth for all Go code.

### Authoritative References

1. **[Effective Go](https://go.dev/doc/effective_go)** — Official idiomatic Go guidance from the language team.
2. **[Google Go Style Guide](https://google.github.io/styleguide/go/guide)** — General style, naming, and formatting.
3. **[Google Go Style Decisions](https://google.github.io/styleguide/go/decisions)** — Specific practical decisions.
4. **[Google Go Best Practices](https://google.github.io/styleguide/go/best-practices)** — High-level guidance on design and maintainability.

### Summary of Core Expectations

- **Clarity First**: Code must be obvious to the reader at a glance. Prioritize readability over cleverness.
- **Simplicity**: Choose the simplest solution that works. "Less is more" — avoid unnecessary abstraction.
- **Consistency**: Follow `gofmt`, `goimports`, and the style of the surrounding codebase.
- **Naming**: Use `MixedCaps` for exported names and `mixedCaps` for unexported. Be concise but descriptive.
- **Comments**: Every exported identifier must have a complete sentence comment starting with its name.
- **Error Handling**: Always check errors. Wrap them with `%w`. Use `errors.Is` and `errors.As` appropriately.
- **Interfaces**: Keep them small and focused. Name them after behavior.
- **Testing**: Prefer table-driven tests. Use clear names and `t.Run` for subtests.
- **Concurrency**: Always use `context.Context` as the first parameter when appropriate.
- **Architecture**: Maintain strict hexagonal boundaries — keep `domain/` pure and dependencies flowing inward.

**When in doubt**, consult the references above in this order of precedence:

**Effective Go → Google Style Guide → Style Decisions → Best Practices**

---

## Testing

This project separates unit and integration tests to maintain fast CI feedback.

### Test Types

- **Unit tests** — Run on every commit/PR (`make test`)
  - Test business logic in isolation
  - Use mocks for external dependencies
  - No build tag required

- **Integration tests** — Run on every push (`make integration`)
  - Test with real external dependencies (databases, APIs)
  - Add `//go:build integration` tag at top of file
  - Use for end-to-end validation

### Test Guidelines

- **Prefer table-driven tests** for multiple test cases
- Use `t.Run` for subtests with descriptive names
- Test file naming: `mycode_test.go` in same package as `mycode.go`
- Exported test functions must start with `Test`
- Maintain test coverage above 80% (configurable via `COVERAGE_THRESHOLD`)

### Running Tests

```bash
# Run unit tests
make test
# or
go test -race -count=1 ./...

# Run integration tests
make integration
# or
go test -race -tags=integration ./...

# Run tests with coverage
go test -coverprofile=coverage.out ./...
```

### Example Integration Test

```go
//go:build integration

package mypackage

import "testing"

func TestMyIntegration(t *testing.T) {
    // Test with real external dependencies
}
```

---

## Go Conventions Validation

This project includes a dedicated script to validate adherence to modern Go conventions from Effective Go and Google Style Guide.

### Convention Checks

The `scripts/ci/pre-commit/18-go-conventions.sh` script validates:

- No `context.Background()` in exported functions (should accept `context.Context`)
- No TODO/FIXME/HACK comments in production code
- No `panic()` in production code (except in init or tests)
- No empty `struct{}` for channels (use `struct{}{}`)
- No `time.Sleep()` in production code (use proper timing/timeout)
- No bare returns in complex functions
- No exported errors without `Error()` method
- No `string()` conversion on errors (use `.Error()` or type assertion)
- No `os.Exit()` in non-main packages
- No `log.Fatal()` in production code
- No main packages outside `cmd/`
- No `init()` functions in production code (use proper initialization)

### Running Convention Checks

```bash
# Run via git hook (automatic on commit)
git commit

# Run standalone
./scripts/ci/pre-commit/18-go-conventions.sh
```

---

## CI & Quality Checks

This project uses centralized scripts for all quality checks, ensuring consistency between local git hooks and CI pipelines.

### Running CI Locally

```bash
# Run full CI pipeline (same as GitHub Actions)
make ci
# or
./scripts/ci/ci.sh
```

### Git Hooks

Git hooks run automatically on commits/push. All hook logic lives in `scripts/ci/` with thin wrappers in `.githooks/`:

- **pre-commit**: gofmt, goimports, golangci-lint, tests, hex-arch-guardrail, gosec, go-vet, secrets, license headers, file quality
- **pre-push**: build, tests, hex-arch-guardrail, coverage, outdated dependencies
- **commit-msg**: conventional commits format, message length
- **pre-rebase**: protect main/master branches
- **prepare-commit-msg**: add branch name to commit message

### CI Script Structure

All check logic is centralized in `scripts/ci/` organized by hook type:

- `scripts/ci/pre-commit/` — Pre-commit checks
- `scripts/ci/pre-push/` — Pre-push checks
- `scripts/ci/commit-msg/` — Commit message validation
- `scripts/ci/pre-rebase/` — Pre-rebase checks
- `scripts/ci/prepare-commit-msg/` — Prepare-commit-msg hooks

### Configuration

Environment variables control behavior:

- `COVERAGE_THRESHOLD` — Default 80%
- `GOLANGCI_LINT_TIMEOUT` — Default 2m (pre-commit), 5m (CI)
- `GO_TEST_FLAGS` — Default `-short` (pre-commit), `-race` (CI)

---

## Leveraging Go Package Documentation

One of the most powerful resources available to you is the official Go package documentation at [pkg.go.dev](https://pkg.go.dev).

### How to Use It

For any Go package, you can instantly access its full documentation by visiting: https://pkg.go.dev/ + the import path

### Examples

- Standard library: https://pkg.go.dev/net/http
- Third-party package: https://pkg.go.dev/github.com/gin-gonic/gin

### Why This Is Valuable

- Always up-to-date documentation with examples, function signatures, and source code.
- Excellent search and cross-referencing.
- Shows godoc comments, usage examples, and related packages.
- Helps you understand how to use a package correctly and idiomatically.

### Best Practice

When working with any unfamiliar package (standard library or external), first visit its pkg.go.dev page to understand its API before writing code. This significantly improves accuracy and adherence to Go idioms.
