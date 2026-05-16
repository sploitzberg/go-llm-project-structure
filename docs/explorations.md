# Explorations

This document tracks potential enhancements and explorations for the Go project template.

## High Priority (Recommended for Implementation)

### 1. Development Container (.devcontainer.json)

**Status:** Done
**Effort:** Medium
**Value:** High

**What it is:**

- VS Code Dev Container configuration for consistent development environment
- Containerized dev environment with all tools pre-installed
- Eliminates "works on my machine" issues

**Benefits:**

- Consistent environment across team members
- Pre-installed Go 1.26, golangci-lint, Task, jq, yq, and other tools
- Quick onboarding for new contributors
- Isolated from host system

**Implementation Approach:**

```
.devcontainer/
├── devcontainer.json
└── Dockerfile
```

**Dockerfile would include:**

- Go 1.26
- golangci-lint
- Task (task runner)
- jq, yq (for LLM hooks)
- Git hooks setup
- Go tools (gremlins, goda)

**Considerations:**

- Adds ~500MB-1GB to repository (Docker image)
- Requires Docker installed (but most Go devs have it)
- Great for templates/onboarding, optional for experienced devs
- Can be made lightweight with multi-stage builds

---

### 2. Benchmark Testing with CI Integration

**Status:** Done
**Effort:** Low
**Value:** Medium

**What it is:**

- Go benchmark tests (`func BenchmarkX(*testing.B)`)
- CI integration to detect performance regressions
- Baseline comparison and trend tracking

**Benefits:**

- Detect performance regressions early
- Quantitative performance metrics
- Helps optimize critical paths
- Standard Go feature (no external dependencies)

**Implementation Approach:**

```go
// internal/core/domain/entity_test.go
func BenchmarkEntityCreation(b *testing.B) {
    for i := 0; i < b.N; i++ {
        NewEntity("test", "description")
    }
}
```

**CI Integration:**

- Add `go test -bench=. -benchmem ./...` to CI
- Use `benchstat` tool for comparison
- Store benchmark results (GitHub Actions artifacts or separate service)
- Optional: `gobench` or `benchdiff` for regression detection

**Considerations:**

- Benchmarks need to be meaningful (not trivial code)
- CI noise (variability in cloud environments)
- Requires baseline establishment
- Adds CI time (~1-2 minutes)

---

### 3. Test Coverage Badge with Codecov

**Status:** Done
**Effort:** Very Low
**Value:** Medium

**What it is:**

- Test coverage reporting via Codecov
- Badge in README showing coverage percentage
- Coverage trends and PR comments

**Benefits:**

- Visibility into test coverage
- Encourages better testing
- PR comments show coverage impact
- Historical tracking

**Current State:**

- Already have coverage collection in CI: `go test -race -coverprofile=coverage.out`
- Already have Codecov upload in `.github/workflows/ci.yml`
- **Missing:** Badge in README, coverage threshold enforcement

**Implementation Approach:**

```markdown
# Add to README.md

[![codecov](https://codecov.io/gh/sploitzberg/go-llm-project-structure/branch/main/graph/badge.svg)](https://codecov.io/gh/sploitzberg/go-llm-project-structure)
```

**Additional Enhancements:**

- Add coverage threshold check in CI (currently configurable via COVERAGE_THRESHOLD)
- Fail CI if coverage drops below threshold
- Show coverage per package

**Considerations:**

- Already have Codecov configured - just need badge
- Coverage percentage can be gamed (tests without assertions)
- Threshold enforcement needs balance (too high = burden, too low = useless)
- Currently 80% threshold in scripts

---

## Medium Priority (Nice to Have)

### 4. SBOM Generation - Syft/Grype

**Status:** Done
**Effort:** Low
**Value:** Medium

**What it is:**

- Software Bill of Materials (SBOM) generation
- Vulnerability scanning of dependencies
- Supply chain security

**Benefits:**

- Complete inventory of dependencies
- Detects vulnerabilities in dependencies
- Compliance requirements (some industries require SBOM)
- Supply chain transparency

**Implementation Approach:**

- Add Syft to CI for SBOM generation
- Add Grype for vulnerability scanning
- Upload SBOM as artifact
- Optional: Integrate with GitHub Dependency Review

**Considerations:**

- Already have govulncheck for Go vulnerabilities
- SBOM adds supply chain visibility
- Adds ~1-2 minutes to CI
- Required for some compliance standards

---

### 5. Fuzzing Support - Go Native Fuzzing

**Status:** Done
**Effort:** Medium
**Value:** Medium

**What it is:**

- Go native fuzzing (Go 1.18+)
- Automated input generation to find edge cases
- Security vulnerability detection

**Benefits:**

- Finds edge cases and security bugs
- Automated testing with random inputs
- Standard Go feature (no external deps)
- Complements unit tests

**Implementation Approach:**

```go
// internal/core/domain/entity_test.go
func FuzzNewEntity(f *testing.F) {
    f.Add("test", "description")
    f.Fuzz(func(t *testing.T, name, desc string) {
        entity, err := NewEntity(name, desc)
        if err != nil {
            t.Errorf("NewEntity(%q, %q) failed: %v", name, desc, err)
        }
        if entity.Name != name {
            t.Errorf("Expected name %q, got %q", name, entity.Name)
        }
    })
}
```

**CI Integration:**

- Add `go test -fuzz=. -fuzztime=30s ./...` to CI
- Or run fuzz tests in separate job (longer runtime)
- Store crash inputs for regression testing

**Considerations:**

- Requires meaningful fuzz targets (not all code benefits)
- Can be slow in CI (needs time budget)
- Requires Go 1.18+
- Best for parsing, validation, input handling code

---

### 6. Performance Profiling - pprof Integration

**Status:** Done
**Effort:** Low
**Value:** Medium

**What it is:**

- Go's built-in profiling tool (pprof)
- CPU, memory, goroutine profiling
- Performance bottleneck identification

**Benefits:**

- Built into Go standard library
- No external dependencies
- Identifies performance bottlenecks
- Helps optimize critical paths

**Implementation Approach:**

```go
import _ "net/http/pprof"

// Add to main.go for dev/profiling builds
if os.Getenv("ENABLE_PPROF") == "true" {
    go func() {
        log.Println(http.ListenAndServe("localhost:6060", nil))
    }()
}
```

**Usage:**

- `go tool pprof http://localhost:6060/debug/pprof/profile`
- Add benchmark profiling to CI
- Optional: flamegraph generation

**CI Integration:**

- Add profiling to benchmark tests
- Store profiles as artifacts
- Optional: automated regression detection

**Considerations:**

- Should only be enabled in dev/profiling builds
- Adds overhead if left enabled
- Requires understanding of pprof output
- Complements benchmark testing

---

## Summary Table

| Component             | Effort   | Value  | Status | Priority |
| --------------------- | -------- | ------ | ------ | -------- |
| Dev Container         | Medium   | High   | Done   | High     |
| Benchmark Testing     | Low      | Medium | Done   | High     |
| Codecov Badge         | Very Low | Medium | Done   | High     |
| SBOM Generation       | Low      | Medium | Done   | Medium   |
| Fuzzing Support       | Medium   | Medium | Done   | Medium   |
| Performance Profiling | Low      | Medium | Done   | Medium   |
