# go-llm-project-structure

[![codecov](https://codecov.io/gh/sploitzberg/go-llm-project-structure/branch/main/graph/badge.svg)](https://codecov.io/gh/sploitzberg/go-llm-project-structure)

A Go project template implementing strict **Hexagonal Architecture** (Ports & Adapters pattern) with LLM coding assistant tooling.

## Quick Start

```bash
# Clone the repository
git clone https://github.com/sploitzberg/go-llm-project-structure.git
cd go-llm-project-structure

# Install dependencies
go mod download

# Setup LLM tool configuration
task llm-setup

# Select your LLM provider (Cursor, Claude, Windsurf, Continue, Copilot, or Codex)
# Sensible defaults are applied automatically

# Run the application
go run cmd/go-llm-project-structure/main.go

# Run tests
task test
```

That's it! Your LLM coding assistant now has comprehensive Go-specific rules, skills, agents, and hooks configured to enforce clean, production-ready, robust Go code following Hexagonal Architecture.

## Guardrails & CI/CD

This template includes a comprehensive CI/CD pipeline with automated quality checks:

### Architecture Guardrails

- **Hexagonal Architecture Guardrail** - Enforces dependency rules (domain purity, correct layer dependencies)
- **Import Cycle Detection** - Prevents circular dependencies
- **Framework Leak Detection** - Ensures core packages don't depend on frameworks

### Code Quality Checks

- **Formatting** - gofmt, goimports (auto-format on save via LLM hooks)
- **Linting** - golangci-lint with comprehensive rule set
- **Static Analysis** - go vet, gosec (security scanner)
- **Error Wrapping** - Validates proper error handling with `%w`
- **Struct Fields** - Ensures exported struct fields have proper tags
- **Import Order** - Enforces consistent import ordering
- **Complexity Analysis** - Enforces cyclomatic and cognitive complexity limits with layer-specific thresholds
  - **Cyclomatic** (gocyclo) - McCabe complexity metric
  - **Cognitive** (gocognit) - Measures nesting and control flow jumps
  - **CRAP Scoring** - Combines complexity with test coverage to identify high-risk code
  - **Configurable** via `.complexity.yml` with architecture-aware thresholds

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

### Using Dev Container

**VS Code / VSCodium:**

1. Install the "Dev Containers" extension
2. Command Palette → "Dev Containers: Reopen in Container"
3. First build takes 2-5 minutes (cached after)

**Windsurf:**

1. Command Palette → "Dev Containers: Reopen in Container"
2. Note: Windsurf may show all parent folders in file explorer (known limitation)
3. Container is functional despite this UI issue

**Manual attachment (fallback):**

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

- Go 1.26
- golangci-lint
- Task (task runner)
- jq, yq (YAML/JSON processors)
- Gremlins (mutation testing)
- goda (dependency analysis)
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
- **pre-push** - Comprehensive checks (build, all tests, coverage, complexity full analysis with CRAP, outdated dependencies)
- **commit-msg** - Conventional commits format validation

## Architecture

This project follows Hexagonal Architecture with these layers:

- `internal/core/domain/` — Pure business entities and rules (zero dependencies)
- `internal/core/ports/` — Interface definitions (primary/secondary)
- `internal/core/services/` — Application use case orchestration
- `internal/adapter/` — Concrete implementations (HTTP, database, etc.)
- `internal/config/` — Configuration structures

## LLM Platform Integration

This project supports integration with various LLM platforms through platform-specific configuration files in `scripts/llm/platforms/`. These configurations provide:

- Platform-specific hooks for context injection
- Intelligent read/write patterns
- Tag reuse conventions

## Documentation

- [`internal/README.md`](internal/README.md) — Layer responsibilities
- [`internal/core/README.md`](internal/core/README.md) — Core architecture overview
- [`docs/architecture/architecture.md`](docs/architecture/architecture.md) — Visual dependency graph
- [`AGENTS.md`](AGENTS.md) — Instructions for AI coding assistants
- [`docs/PROMPT_ENGINEERING.md`](docs/PROMPT_ENGINEERING.md) — Prompt engineering best practices with examples
- [`SECURITY.md`](SECURITY.md) — Security policy and vulnerability reporting
- [`CODE_OF_CONDUCT.md`](CODE_OF_CONDUCT.md) — Contributor code of conduct
- [`.complexity.yml`](.complexity.yml) — Complexity thresholds and CRAP scoring configuration
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
- `task llm-setup` — Setup LLM tool configurations
- `task mutation-test` — Full mutation testing with Gremlins (slow, thorough)
- `task mutation-test-dry` — Fast mutation dry-run (CI mode)

## Integration Tests

Integration tests are separated from unit tests to keep CI fast:

- **Unit tests** (`task test`) — Run on every commit (pre-commit hook)
- **Integration tests** (`task integration`) — Run on every push (pre-push hook)

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
