# Internal

This directory contains all private application code following **Hexagonal Architecture** (also known as Ports & Adapters).

## Architecture Layers

| Layer                       | Project packages it may import                         | External packages |
| --------------------------- | ------------------------------------------------------ | ----------------- |
| `core/domain/`              | None                                                   | Standard library only |
| `core/ports/`               | `core/domain/`                                         | Standard library only |
| `core/services/`            | `core/domain/`, `core/ports/`                          | Standard library only |
| `adapter/primary/`          | `core/domain/`, `core/ports/primary/`, `config/`       | Allowed |
| `adapter/secondary/`        | `core/domain/`, `core/ports/secondary/`                | Allowed |
| `config/`                   | None                                                   | Standard library only |

**Core Rule**: All dependencies must point inward. Adapters never import other adapters, and any new Go package under `internal/` requires an explicit architecture policy.
