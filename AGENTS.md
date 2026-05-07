# Project Instructions

You are working on a **production-grade Go codebase** that follows strict Hexagonal Architecture (Ports & Adapters).

## Your Responsibilities as an Agent

- **Mosaic Retention Policy:** Follow the configured retention policy (see `docs/mosaic/AGENT_QUICK_REFERENCE.md` for details). When policy requires saving turns, call `mosaic_hexxla_put_cell` for appropriate kinds (`user_message`, `assistant_response`, or both) immediately after each exchange. Include relevant topic-specific tags when calling put_cell. Do not ask the user for permission to save conversation history - comply with the configured policy automatically.
- **ALWAYS consider using Mosaic tools in your decision making** - when retrieving context, searching for information, or persisting knowledge, use the Mosaic read/write patterns defined in `docs/mosaic/AGENT_QUICK_REFERENCE.md`.
- Strictly adhere to architectural rules, coding standards, and best practices. Never cut corners.
- Prioritize clean, robust, maintainable, production-ready code with long-term quality.
- Refactor messy or unclear solutions properly. Security, testability, and separation of concerns are non-negotiable.

---

## Hexagonal Architecture

This project uses **Hexagonal Architecture** (also known as **Ports & Adapters**), originally introduced by Alistair Cockburn.

### Core Concepts

- The **domain** is the heart of the application and must remain completely independent.
- All dependencies point **inward** toward the domain.
- External concerns (databases, HTTP, third-party APIs, etc.) are pushed to the edges as adapters.
- The core business logic is decoupled from frameworks, databases, and delivery mechanisms.

### Layer Responsibilities & Rules

- **`core/domain/`** — Pure business entities, value objects, and domain rules.
  **Must have zero dependencies** on any other internal package.

- **`core/ports/`** — Defines clean interfaces (contracts).
  - `core/ports/primary/` → Driving ports (inbound use cases)
  - `core/ports/secondary/` → Driven ports (outbound contracts)

- **`core/services/`** — Application services that implement use cases.
  Depends **only** on `core/domain/` and `core/ports/`.

- **`adapter/`** — Concrete implementations of ports.
  - `adapter/primary/` → Inbound (HTTP, CLI, gRPC, etc.)
  - `adapter/secondary/` → Outbound (databases, external APIs, cache, etc.)

- **`config/`** — Centralized configuration structures and loading.

### Strict Rules

- Never import `adapter/` from `core/domain/`, `core/ports/`, or `core/services/`
- Never import `core/services/` from `core/domain/` or `core/ports/`
- `core/domain/` must remain pure at all times
- All external interactions must go through ports
- Follow `.golangci.yml` (especially `depguard` rules)

**Security is mandatory** — see `rules/security.mdc`

### Full Documentation

- Detailed layer explanations → [`internal/README.md`](internal/README.md)
- Core architecture overview → [`internal/core/README.md`](internal/core/README.md)
- Visual dependency graph + deeper explanation → [`docs/architecture/architecture.md`](docs/architecture/architecture.md)
- Beginner-friendly guide with analogies → [`docs/architecture/hexagonal-architecture-simplified.md`](docs/architecture/hexagonal-architecture-simplified.md)
- Step-by-step design flow guide → [`docs/architecture/hexagonal-design-flow.md`](docs/architecture/hexagonal-design-flow.md)

---

## Mosaic Documentation

This project uses Mosaic (HexxlaDB) as an external workflow tool for AI agents. Mosaic is NOT part of the project architecture - it's a separate tool used in the development workflow for long-term memory and context retrieval. Comprehensive documentation is available in `docs/mosaic/`:

- [`docs/mosaic/AGENT_QUICK_REFERENCE.md`](docs/mosaic/AGENT_QUICK_REFERENCE.md) — Quick reference for AI agents working with Mosaic
- [`docs/mosaic/PROJECT_INTEGRATION.md`](docs/mosaic/PROJECT_INTEGRATION.md) — How Mosaic is used in the development workflow
- [`docs/mosaic_retention_compliance.md`](docs/mosaic_retention_compliance.md) — Retention policy compliance documentation

### Retention Policy

- **capture_mode**: `save_all_turns`
- **enforcement**: `true` — server returns error for conflicting put_cell kinds
- **Requirement**: Persist every user_message and assistant_response automatically
- **Critical**: Never ask user for permission — comply automatically

---

## Go Style & Best Practices

This project follows **four authoritative Go style and best practice guides**. Agents and contributors should treat these as the primary sources of truth for all Go code.

### Authoritative References

1. [Effective Go](https://go.dev/doc/effective_go)
2. [Google Go Style Guide](https://google.github.io/styleguide/go/guide)
3. [Google Go Style Decisions](https://google.github.io/styleguide/go/decisions)
4. [Google Go Best Practices](https://google.github.io/styleguide/go/best-practices)

### Summary of Core Expectations

- **Clarity & Simplicity**: Prioritize readability over cleverness. Choose the simplest solution.
- **Consistency**: Follow `gofmt`, `goimports`, and surrounding codebase style.
- **Naming & Comments**: `MixedCaps` for exported, `mixedCaps` for unexported. Exported identifiers need sentence comments.
- **Error Handling**: Always check errors, wrap with `%w`, use `errors.Is`/`errors.As`.
- **Interfaces**: Keep small and focused, named after behavior.
- **Testing**: Table-driven with `t.Run`, maintain 80%+ coverage.
- **Concurrency**: Use `context.Context` as first parameter when appropriate.
- **Architecture**: Maintain strict hexagonal boundaries, keep `core/domain/` pure.

**Precedence**: Effective Go → Google Style Guide → Style Decisions → Best Practices

---

## Testing

This project separates unit and integration tests to maintain fast CI feedback.

### Test Types

- **Unit tests** — Run on commit/PR (`make test`). Test logic in isolation with mocks.
- **Integration tests** — Run on push (`make integration`). Add `//go:build integration` tag. Test with real dependencies.

### Test Guidelines

- **Prefer table-driven tests** for multiple test cases
- Use `t.Run` for subtests with descriptive names
- Test file naming: `mycode_test.go` in same package as `mycode.go`
- Exported test functions must start with `Test`
- Maintain test coverage above 80% (configurable via `COVERAGE_THRESHOLD`)

### Running Tests

```bash
make test                    # Unit tests
go test -race -count=1 ./...
make integration             # Integration tests
go test -race -tags=integration ./...
go test -coverprofile=coverage.out ./...  # With coverage
```

---

## Go Conventions Validation

The `scripts/ci/pre-commit/18-go-conventions.sh` script validates adherence to modern Go conventions from Effective Go and Google Style Guide (no context.Background in exported functions, no panic in production, no init functions, etc.). See the script for the full list of checks.

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

- **pre-commit**: gofmt, goimports, golangci-lint, tests, hex-arch-guardrail, gosec, go-vet, secrets, license headers, file quality, complexity (changed files)
- **pre-push**: build, tests, hex-arch-guardrail, coverage, outdated dependencies, complexity (full with CRAP)
- **commit-msg**: conventional commits format, message length
- **pre-rebase**: protect main/master branches
- **prepare-commit-msg**: add branch name to commit message

### Configuration

Environment variables control behavior:

- `COVERAGE_THRESHOLD` — Default 80%
- `GOLANGCI_LINT_TIMEOUT` — Default 2m (pre-commit), 5m (CI)
- `GO_TEST_FLAGS` — Default `-short` (pre-commit), `-race` (CI)

---

## Complexity Guardrails & Semantic Stability

This project enforces code complexity limits to maintain maintainability and testability. Complexity analysis runs automatically in CI and git hooks.

### Complexity Metrics

- **Cyclomatic Complexity** (gocyclo) - McCabe complexity metric, measures decision points
- **Cognitive Complexity** (gocognit) - Measures nesting, control flow jumps, and mental effort
- **CRAP Score** - Combines cyclomatic complexity with test coverage: `CRAP = (cyclomatic² × (1 - coverage/100)³) + cyclomatic`
  - High CRAP = complex AND poorly tested = dangerous to change

### Layer-Specific Thresholds

Complexity expectations differ by hexagonal architecture layer:

| Layer                | Cyclomatic Max | Cognitive Max | Rationale                                       |
| -------------------- | -------------- | ------------- | ----------------------------------------------- |
| `core/domain/`       | 5              | 10            | Pure business logic, must be simple and focused |
| `core/ports/`        | 3              | 5             | Interface definitions, minimal logic            |
| `core/services/`     | 10             | 15            | Orchestration allowed, but keep manageable      |
| `adapter/primary/`   | 15             | 20            | HTTP handlers, CLI — external concerns          |
| `adapter/secondary/` | 12             | 18            | Database, API clients — external integrations   |

### Configuration

Complexity thresholds are configurable via `.complexity.yml`:

- Enable/disable checks globally
- Adjust thresholds per layer
- Configure CRAP scoring threshold (default: 30)

### Guidance for AI Agents

**When writing code:**

- Keep functions small and focused
- Prefer simple control flow over nested conditions
- Extract complex logic into smaller, testable functions
- Domain layer code should be the simplest (pure business rules)
- If complexity approaches threshold, consider refactoring

**When CRAP score is high:**

- High CRAP indicates complex code with poor test coverage
- Options:
  1. **Refactor** - Break down complex function into smaller ones
  2. **Improve tests** - Add specific assertions to catch mutations
  3. **Property-based tests** - For pure functions in domain layer
  4. **Contract tests** - For port interfaces

**Semantic Stability:**

- High complexity = harder to write semantically stable tests
- Mutation testing (Gremlins) validates test effectiveness
- Aim for high mutation score (80%+) on critical code paths
- Property-based testing helps ensure invariants hold across inputs

### Mutation Testing with Gremlins

This project uses Gremlins for mutation testing to ensure semantic stability of critical code.

**What is Mutation Testing?**

- Gremlins introduces small mutations (bugs) into code
- Tests are run against mutated code
- If tests fail → mutation is "killed" (good)
- If tests pass → mutation "lived" (test missed the bug)

**Configuration:**

- Configuration file: `.gremlins.yml`
- Target: Domain layer by default (`internal/core/domain/`)
- Easily changeable to test other layers
- Mutation threshold: 80% killed mutants
- Runs in dry-run mode during CI for speed

**When it Runs:**

- Local: Part of `make ci` pipeline (dry-run for speed)
- CI: Runs on every push and PR (dry-run mode)
- Full testing: `make mutation-test` — thorough but slow (runs actual mutations)
- Quick check: `make mutation-test-dry` — fast dry-run, same as CI

**Mutation Results:**

- **RUNNABLE**: Mutation can be tested (dry-run mode)
- **NOT COVERED**: Mutation not covered by tests
- **KILLED**: Test caught the mutation (good)
- **LIVED**: Test missed the mutation (improve tests)
- **TIMED OUT**: Tests took too long (mutation broke performance)
- **NOT VIABLE**: Mutation made build fail

**Guidance for AI Agents:**

**When writing tests for domain logic:**

- Write tests that assert specific behavior, not just execution
- Use property-based testing for pure functions
- Test invariants that should hold across inputs
- Consider edge cases and boundary conditions
- Avoid overly permissive assertions (e.g., `err == nil` without checking result)

**When mutation score is low:**

- Add specific assertions for expected behavior
- Test error cases explicitly
- Add property-based tests for pure functions
- Consider if the code is over-engineered (simplify if possible)
- Extract complex logic into smaller, testable functions

**Before committing domain layer changes:**

- Run `make ci` to trigger Gremlins dry-run
- If mutations live, improve tests to catch them
- Aim for 80%+ mutation score on critical paths
- Document any intentional test exclusions in `.gremlins.yml`

**Before committing complex code:**

- Run `./scripts/ci/pre-push/05-complexity.sh` to check full analysis
- Run `make ci` to trigger Gremlins mutation testing
- If CRAP is high, either refactor or improve test coverage
- Consider adding property-based tests for domain logic
- Add contract tests for port implementations

---

## Dependency & Coupling Analysis

This project tracks fan-out (number of external packages imported) per hexagonal layer to prevent tight coupling.

**Configuration:** `.coupling.yml` — per-layer thresholds, easily adjustable.

**When it Runs:** Part of `make ci`, pre-push hook, and GitHub Actions.

**Layer Fan-Out Thresholds:**

| Layer                | Max Fan-Out | Rationale                       |
| -------------------- | ----------- | ------------------------------- |
| `core/domain/`       | 5           | Must have minimal external deps |
| `core/ports/`        | 5           | Interface definitions only      |
| `core/services/`     | 10          | Orchestration layer             |
| `adapter/primary/`   | 20          | HTTP/CLI — more allowed         |
| `adapter/secondary/` | 15          | DB/API clients — bounded        |

**Guidance for AI Agents:**

- Before adding a new import to `core/domain/` or `core/ports/`, consider whether it's truly necessary
- Prefer passing values/interfaces over importing packages in core layers
- Use value objects or domain types to avoid pulling in external packages
- High fan-out in domain/ports is a strong signal of leaky abstractions
- If thresholds need adjusting for the project, update `.coupling.yml` — not the script

---

## Leveraging Go Package Documentation

Use [pkg.go.dev](https://pkg.go.dev) for official Go package documentation. Visit `https://pkg.go.dev/<import-path>` for any package (e.g., https://pkg.go.dev/net/http). Check API docs before using unfamiliar packages to ensure idiomatic usage.
