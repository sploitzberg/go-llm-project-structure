# go-llm-project-structure

A Go project template implementing strict **Hexagonal Architecture** (Ports & Adapters pattern) with LLM coding assistant tooling.

## Quick Start

```bash
# Build the project
make build

# Run tests
make test

# Run linting
make lint

# Setup LLM tool configurations (Cursor, Claude, Windsurf, etc.)
make llm-setup
```

## Architecture

This project follows Hexagonal Architecture with these layers:

- `internal/domain/` — Pure business entities and rules (zero dependencies)
- `internal/port/` — Interface definitions (primary/secondary)
- `internal/service/` — Application use case orchestration
- `internal/adapter/` — Concrete implementations (HTTP, database, etc.)
- `internal/config/` — Configuration structures

## Documentation

- [`internal/README.md`](internal/README.md) — Layer responsibilities
- [`docs/architecture/architecture.md`](docs/architecture/architecture.md) — Visual dependency graph
- [`AGENTS.md`](AGENTS.md) — Instructions for AI coding assistants

## LLM Tool Setup

Run `make llm-setup` to generate configurations for your preferred LLM coding assistant (Cursor, Claude, Windsurf, Continue, Copilot, or Codex). This injects architecture rules, agents, and hooks specific to each platform.

## Makefile Targets

- `make build` — Build the binary
- `make test` — Run tests
- `make integration` — Run integration tests (slower, external dependencies)
- `make lint` — Run golangci-lint
- `make fmt` — Format code
- `make ci` — Run full CI checks
- `make install` — Install locally
- `make llm-setup` — Setup LLM tool configurations

## Integration Tests

Integration tests are separated from unit tests to keep CI fast:

- **Unit tests** (`make test`) — Run on every commit/PR
- **Integration tests** (`make integration`) — Run weekly or manually via GitHub Actions

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
make integration
# or
go test -tags=integration ./...
```
