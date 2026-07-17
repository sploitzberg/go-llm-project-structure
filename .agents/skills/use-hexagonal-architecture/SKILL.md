---
name: use-hexagonal-architecture
description: Apply the project's Hexagonal Architecture workflow when designing, implementing, refactoring, or reviewing domain logic, use cases, ports, services, adapters, and composition roots. Use for changes that introduce or move behavior across architectural boundaries.
---

# Use Hexagonal Architecture

Treat `AGENTS.md` as the source of truth for dependency invariants. Read the relevant guides under `docs/architecture/` when the task needs conceptual or design-flow detail; do not duplicate those documents in code or responses.

## Work inside-out

For a new capability, proceed in this order unless the existing code provides a justified variation:

1. Define domain concepts, invariants, and errors in `internal/core/domain/`.
2. Define the use case and its input/output types in `internal/core/ports/primary/`.
3. Define only the external capabilities the use case needs in `internal/core/ports/secondary/`.
4. Implement orchestration in `internal/core/services/` using domain types and ports.
5. Implement inbound adapters in `internal/adapter/primary/` and outbound adapters in `internal/adapter/secondary/`.
6. Wire concrete implementations only in the composition root under `cmd/`.
7. Add focused tests at the layer where behavior is owned.

Follow existing package patterns before creating new abstractions. Keep interfaces narrow and owned by the core capability that consumes them.

## Place behavior by responsibility

- Business invariant or entity behavior → domain.
- Use-case sequencing, authorization, or transaction coordination → service.
- Capability exposed to an external actor → primary port.
- Capability required from infrastructure → secondary port.
- Transport, persistence, serialization, or vendor-specific translation → adapter.
- Concrete dependency construction → composition root.

Do not pass HTTP, SQL, framework, or vendor types through core ports. Do not call concrete services from primary adapters when a primary port exists, and do not let core packages import adapters.

## Review before finishing

Confirm that:

- dependencies point inward;
- domain code has no internal or infrastructure dependencies;
- services depend on ports rather than adapters;
- adapters translate at the boundary instead of containing business rules;
- ports expose business-oriented contracts rather than infrastructure details;
- tests verify domain rules and service orchestration independently of real infrastructure.

## Validate

Run the narrowest relevant tests first, then:

```sh
rtk task test
bash scripts/ci/hex-arch-guardrail.sh
```

Load the `use-rtk` skill before verbose validation. Use raw commands when detailed failure output is needed.
