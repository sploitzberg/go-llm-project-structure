# Internal

This directory contains all private application code following **Hexagonal Architecture** (also known as Ports & Adapters).

## Architecture Layers

| Layer      | Responsibility                          | Can Depend On      | Must Not Depend On      |
| ---------- | --------------------------------------- | ------------------ | ----------------------- |
| `domain/`  | Business entities and rules             | Nothing            | Anything else           |
| `port/`    | Interfaces (contracts)                  | `domain/`          | `adapter/`, `service/`  |
| `service/` | Application use cases and orchestration | `domain/`, `port/` | `adapter/`              |
| `adapter/` | Concrete implementations                | `port/`, `domain/` | Other adapters (mostly) |
| `config/`  | Configuration structures                | Nothing            | -                       |

**Core Rule**: All dependencies must point **inward** toward the `domain`.
