# Architecture

Gommitlint implements **functional hexagonal architecture** with value semantics, optimizing for testability, maintainability, and reasoning simplicity.

## Executive Summary

- **Pattern**: Hexagonal Architecture (Ports & Adapters)
- **Paradigm**: Functional programming with pure functions
- **Dependencies**: Explicit injection, no service locators
- **Testing**: Table-driven with >80% coverage
- **Performance**: Immutable data structures, zero-allocation paths

## Core Architectural Decisions

### 1. Functional Over Object-Oriented
**Decision**: Pure functions instead of stateful services  
**Rationale**: Eliminates state management complexity, improves testability  
**Trade-off**: Slightly more explicit parameter passing

### 2. Value Semantics Throughout
**Decision**: All domain types use value receivers  
**Rationale**: Memory safety, immutability guarantees, goroutine safety  
**Trade-off**: Small memory overhead for large structs

### 3. No Central Dependency Container
**Decision**: Direct constructor composition in CLI layer  
**Rationale**: Explicit dependencies, compile-time safety  
**Trade-off**: Manual wiring vs. reflection-based frameworks

### 4. Domain-Defined Interfaces
**Decision**: Interfaces defined where consumed (domain)  
**Rationale**: Dependency inversion, testability  
**Implementation**: `Repository`, `Logger`, `SignatureVerifier` in domain

## Architecture Overview

```
┌─────────────────────────────────────────────────┐
│                 Adapters                        │
│  cli/ → config/ → git/ → logging/ → signing/    │
├─────────────────────────────────────────────────┤
│              Domain Interfaces                  │
│   Repository • Logger • SignatureVerifier       │
├─────────────────────────────────────────────────┤
│                Core Domain                      │
│  Validation • Rules • Entities • Value Objects  │
└─────────────────────────────────────────────────┘
```

**Dependency Direction**: Always inward (adapters → domain)

## Directory Structure

```
gommitlint/
├── main.go                    # Composition root
├── internal/
│   ├── domain/               # Business logic (pure functions)
│   │   ├── *.go             # Core types and validation functions
│   │   ├── config/          # Configuration types
│   │   └── rules/           # Validation rule implementations
│   ├── adapters/            # Infrastructure implementations
│   │   ├── cli/            # Command-line interface
│   │   ├── config/         # Configuration loading (was loader/)
│   │   ├── git/            # Git operations
│   │   ├── logging/        # Structured logging
│   │   ├── output/         # Report formatting
│   │   └── signing/        # Cryptographic verification
│   └── integrationtest/    # End-to-end tests
└── docs/                   # Architecture documentation
```

## Key Components

### Core Domain (`internal/domain/`)
- **Pure functions**: `ValidateCommit`, `ValidateCommits`
- **Value objects**: `Commit`, `ValidationError`, `Signature`
- **Interfaces**: `Repository`, `Logger`, `SignatureVerifier`
- **Rules**: Pluggable validation implementations

### Adapters (`internal/adapters/`)
- **CLI**: Cobra-based command interface
- **Git**: go-git repository operations
- **Config**: Koanf-based configuration loading
- **Signing**: GPG/SSH signature verification
- **Output**: Text, JSON, GitHub, GitLab formatters

## Functional Programming Patterns

### Value Semantics
```go
// All transformations return new values
func (c Commit) WithAuthor(author string) Commit {
    c.Author = author
    return c
}
```

### Pure Functions
```go
// Business logic has no side effects
func ValidateCommit(commit Commit, rules []Rule, repo Repository, cfg Config) ValidationResult {
    var errors []ValidationError
    for _, rule := range rules {
        errors = append(errors, rule.Validate(commit, cfg)...)
    }
    return ValidationResult{Commit: commit, Errors: errors}
}
```

### Explicit Dependencies
```go
// No hidden dependencies, all parameters explicit
func RunValidation(ctx context.Context, ref string, rules []Rule, 
    repo Repository, cfg Config) error {
    // Implementation
}
```

## Testing Strategy

### Unit Tests
- **Pattern**: Table-driven tests with `testCase` variable
- **Coverage**: >80% across all packages
- **Isolation**: Mock domain interfaces

### Integration Tests
- **Location**: `internal/integrationtest/`
- **Scope**: End-to-end command validation
- **Data**: Real Git repositories in `testdata/`

### Performance Tests
- **Benchmarks**: Critical path validation functions
- **Memory**: Zero-allocation goals for hot paths

## Configuration Management

### Rule Priority System
1. **Explicitly enabled** (highest priority)
2. **Explicitly disabled** 
3. **Default disabled** (`jirareference`, `commitbody`, `spell`)
4. **Default enabled** (all others)

### Access Pattern
- **Explicit parameters**: No configuration in context
- **Constructor injection**: Rules receive config during creation
- **Immutable**: Configuration never modified after load

## Performance Characteristics

### Memory Usage
- **Value semantics**: Predictable allocation patterns
- **Immutable data**: Safe concurrent access
- **Copy optimization**: Structural sharing where possible

### Execution Model
- **Stateless**: No global state, thread-safe by design
- **Functional**: Deterministic execution paths
- **Parallel**: Rules can be validated concurrently

## Security Considerations

### Cryptographic Verification
- **GPG/SSH**: Pluggable signature verification
- **Key management**: Secure file operations
- **TOCTOU prevention**: File descriptor-based operations

### Input Validation
- **Path traversal**: Secure path joining and validation
- **Injection**: No shell command construction
- **Resource limits**: Bounded operations

## Extension Points

### Adding New Rules
1. Implement `CommitRule` or `RepositoryRule` interface
2. Add to factory in `rules/factory.go`
3. Configure in rule priority system

### Adding New Adapters
1. Implement domain interfaces
2. Add to adapter composition in CLI layer
3. Register in dependency wiring

### Adding New Output Formats
1. Implement formatter function signature
2. Register in output registry
3. Add CLI flag support

## Deployment Architecture

### Single Binary
- **Static linking**: No external dependencies
- **Configuration**: File-based with environment overrides
- **Installation**: Git hooks or CI/CD integration

### Resource Requirements
- **Memory**: <50MB typical usage
- **CPU**: Validation is CPU-bound, rules run in parallel
- **I/O**: Git repository access, minimal file operations

## Monitoring and Observability

### Structured Logging
- **Levels**: Error, Warn, Info, Debug
- **Format**: JSON in production, human-readable in development
- **Context**: Request correlation through context propagation

### Metrics
- **Validation time**: Per-rule and total execution time
- **Rule failures**: Frequency and types of violations
- **Repository operations**: Git operation performance

## Quality Attributes

### Maintainability
- **Separation of concerns**: Clear architectural boundaries
- **Testability**: High test coverage with simple mocking
- **Readability**: Functional style reduces cognitive load

### Performance
- **Startup time**: <100ms for typical repositories
- **Memory efficiency**: Bounded by repository size
- **Concurrency**: Safe parallel execution

### Reliability
- **Error handling**: Explicit error returns, no panics
- **Resource cleanup**: Proper resource management
- **Graceful degradation**: Partial validation on failures

## Anti-Patterns Avoided

- **Service objects**: Direct function composition instead
- **Global state**: Explicit dependency injection
- **Complex factories**: Simple constructor functions
- **Hidden dependencies**: All parameters explicit
- **Mutable configuration**: Immutable after initialization

---

This architecture optimizes for **simplicity**, **testability**, and **maintainability** while maintaining high performance and extensibility.