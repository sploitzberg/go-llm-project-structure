# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Common Changelog](https://common-changelog.org/)
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Initial project structure with hexagonal architecture (domain, port, service, adapter)
- Taskfile with common development targets
- CI pipeline with formatting, linting, tests, and architecture guardrails
- Comprehensive documentation for each architectural layer
- Project-scoped Zed tasks and Go editor settings
- Project-local Zed skill routing with Hexagonal Architecture and token-efficient RTK guidance

### Changed

- Updated folder structure to clearly separate primary and secondary ports/adapters
- Tailored the development container for Zed's native Go tooling
- Updated agent workflows to use version-aware Go and token-efficient RTK skills
- Updated Go, development tools, CI actions, and container images to current stable releases

### Removed

- Legacy multi-platform LLM setup scripts, manifests, hooks, and Taskfile targets

## [0.1.0] - 2025-04-25

### Added

- Bootstrapper CLI skeleton (`cmd/go-llm-project-structure/main.go`)
- Hexagonal architecture guardrail script
- AGENTS.md with instructions for LLMs and contributors
- Layer-specific README.md files explaining responsibilities

### Changed

- Standardized project layout for LLM-friendly hexagonal architecture templates

[Unreleased]: https://github.com/sploitzberg/go-llm-project-structure/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/sploitzberg/go-llm-project-structure/releases/tag/v0.1.0
