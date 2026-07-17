# go-llm-project-structure

[![codecov](https://codecov.io/gh/sploitzberg/go-llm-project-structure/branch/main/graph/badge.svg)](https://codecov.io/gh/sploitzberg/go-llm-project-structure)

A Go project template implementing strict **Hexagonal Architecture** (Ports & Adapters pattern) with first-class Zed IDE configuration.

## Quick Start

```bash
# Clone the repository
git clone https://github.com/sploitzberg/go-llm-project-structure.git
cd go-llm-project-structure

# Install dependencies
go mod download

# Enable the version-controlled Git hooks
git config core.hooksPath .githooks

# Open the project in Zed
zed .

# Run the application or tests from Zed's task picker, or from the terminal
task run
task test
```

Zed automatically loads the project tasks and Go settings from `.zed/`. Git hooks and CI enforce the same production-grade Go and Hexagonal Architecture guardrails outside the editor.

## Guardrails & CI/CD

This template includes a comprehensive CI/CD pipeline with automated quality checks:

### Architecture Guardrails

- **Hexagonal Architecture Guardrail** - Enforces dependency rules (domain purity, correct layer dependencies)
- **Import Cycle Detection** - Prevents circular dependencies
- **Framework Leak Detection** - Ensures core packages don't depend on frameworks

### Code Quality Checks

- **Formatting** - gopls formatting on save in Zed, with gofmt and goimports enforced by CI
- **Linting** - golangci-lint with comprehensive rule set
- **Static Analysis** - go vet, gosec (security scanner)
- **Error Wrapping** - Validates proper error handling with `%w`
- **Struct Fields** - Ensures exported struct fields have proper tags
- **Import Order** - Enforces consistent import ordering
- **Complexity Analysis** - Enforces cyclomatic and cognitive complexity limits with layer-specific thresholds
  - **Cyclomatic** (gocyclo) - McCabe complexity metric
  - **Cognitive** (gocognit) - Measures nesting and control flow jumps
  - **Configurable** via golangci-lint configuration with architecture-aware thresholds

### Mutation Testing

- **Gremlins** - Validates semantic test quality by introducing small bugs into code
  - If tests pass with the bug → they are too weak (mutation "lived")
  - If tests fail → mutation "killed" (good — tests are meaningful)
  - Target: `core/domain/` by default, configurable via `.gremlins.yml`
  - Threshold: 80%+ mutants killed
  - CI: Dry-run on every push (`task mutation-test-dry`)
  - Full run: `task mutation-test` (slow, thorough)

### Development Container

This project includes a Dev Container configuration for a consistent development environment.

### Using the Dev Container

The development container is optional and editor-independent. To start it manually:

```bash
# Build and start container
cd .devcontainer
docker-compose build
docker-compose up -d

# Find container name
docker ps

# Attach to running container (replace with actual name)
docker exec -it <container-name> bash
```

### Container includes

- Go 1.26.5
- golangci-lint 2.12.2
- Task 3.52.0 and RTK 0.43.0
- gopls 0.23.0 and Delve 1.27.0
- Gremlins 0.6.0
- goda 0.9.4
- Pre-configured git hooks

### Docker (Optional)

This project includes optional Docker support for deployment:

**Build and run with Docker:**

```bash
docker compose up --build
```

**Build image only:**

```bash
docker build -t go-llm-project-structure .
```

See [README.Docker.md](README.Docker.md) for more details.

Note: Docker is optional for development. Use the dev container or local Go tooling instead.

### Dependency & Coupling Analysis

- **goda** - Tracks fan-out (number of imported packages) per hexagonal layer
  - Prevents tight coupling between layers
  - Per-layer thresholds configurable via `.coupling.yml`
  - Runs on every push (pre-push hook) and in CI

### Go Conventions Validation

- No `context.Background()` in exported functions
- No `panic()` in production code
- No `time.Sleep()` in production code
- No `os.Exit()` in non-main packages
- No `init()` functions in production code
- And more (enforced by golangci-lint linters: godox, gochecknoinits, noctx, etc.)

### Security Checks

- Secret scanning (AWS credentials, GitHub tokens, API keys)
- Dependency vulnerability scanning
- Input validation enforcement

### Git Hooks

- **pre-commit** - Fast checks (formatting, linting, unit tests, complexity on changed files)
- **pre-push** - Comprehensive checks (build, all tests, coverage, complexity full analysis, outdated dependencies)
- **commit-msg** - Conventional commits format validation

## Architecture

This project follows Hexagonal Architecture with these layers:

- `internal/core/domain/` — Pure business entities and rules (zero dependencies)
- `internal/core/ports/` — Interface definitions (primary/secondary)
- `internal/core/services/` — Application use case orchestration
- `internal/adapter/` — Concrete implementations (HTTP, database, etc.)
- `internal/config/` — Configuration structures

## Zed IDE Integration

Project-scoped Zed configuration lives in `.zed/`:

- [`tasks.json`](.zed/tasks.json) exposes build, test, lint, architecture, CI, formatting, integration, and run workflows through `task: spawn`.
- [`settings.json`](.zed/settings.json) enables consistent gopls formatting on save.
- [`.rules`](.rules) routes Zed agents to reusable project-local skills under `.agents/skills/`.
- The [`use-rtk`](.agents/skills/use-rtk/SKILL.md) skill documents optional token-efficient command execution.
- Zed provides native Go language-server, test runnable, and Delve integration; no editor lifecycle hooks or generated provider configuration are required.

The Zed tasks call the same Taskfile and CI scripts used by local Git hooks and GitHub Actions, so editor workflows do not bypass repository guardrails. Verbose tasks prefer RTK when it is installed and otherwise execute the underlying command directly.

## Documentation

- [`internal/README.md`](internal/README.md) — Layer responsibilities
- [`internal/core/README.md`](internal/core/README.md) — Core architecture overview
- [`docs/architecture/architecture.md`](docs/architecture/architecture.md) — Visual dependency graph
- [`AGENTS.md`](AGENTS.md) — Project invariants for coding assistants
- [`.rules`](.rules) — Zed-specific skill routing and workflow rules
- [`.agents/skills/`](.agents/skills/) — Reusable project-local agent skills
- [`use-hexagonal-architecture`](.agents/skills/use-hexagonal-architecture/SKILL.md) — Inside-out architecture implementation and review workflow
- [`docs/PROMPT_ENGINEERING.md`](docs/PROMPT_ENGINEERING.md) — Prompt engineering best practices with examples
- [`SECURITY.md`](SECURITY.md) — Security policy and vulnerability reporting
- [`CODE_OF_CONDUCT.md`](CODE_OF_CONDUCT.md) — Contributor code of conduct
- [`.gremlins.yml`](.gremlins.yml) — Mutation testing configuration (Gremlins)
- [`.coupling.yml`](.coupling.yml) — Dependency & coupling analysis thresholds (goda)

## Task Targets

- `task build` — Build the binary
- `task test` — Run unit tests
- `task test-all` — Run both unit and integration tests
- `task integration` — Run integration tests (slower, external dependencies)
- `task lint` — Run golangci-lint
- `task fmt` — Format code
- `task ci` — Run full CI checks
- `task install` — Install locally
- `task tools` — Install pinned local/CI quality tools
- `task mutation-test` — Full mutation testing with Gremlins (slow, thorough)
- `task mutation-test-dry` — Fast mutation dry-run (CI mode)
- `task benchmark` — Run benchmark tests
- `task sbom` — Generate Software Bill of Materials (SBOM)
- `task pprof` — Run application with pprof profiling enabled
- `task fuzz` — Run fuzz tests for security-critical code

## Integration Tests

Integration tests are separated from unit tests to keep CI fast:

- **Unit tests** (`task test`) — Run on every commit (pre-commit hook)
- **Integration tests** (`task integration`) — Run explicitly and in the scheduled/manual integration workflow

To create an integration test, add the build tag to the top of your test file:

```go
//go:build integration

package mypackage

import "testing"

func TestMyIntegration(t *testing.T) {
    // Test with external dependencies (databases, APIs, etc.)
}
```

Run integration tests locally:

```bash
task integration
# or
go test -tags=integration ./...
```
