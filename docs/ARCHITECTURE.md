# Architecture

Gommitlint implements **functional hexagonal architecture** optimized for simplicity, testability, and maintainability.

## Design Philosophy

**Core Principle**: Pure functions with explicit dependencies over complex object hierarchies.

- **Functional**: Pure functions, immutable data, value semantics
- **Hexagonal**: Domain at center, adapters at edges, interfaces defined by consumers
- **Simple**: Direct composition over dependency injection frameworks
- **Explicit**: All dependencies passed as parameters, no hidden state

## Architecture Overview

```txt
┌──────────────────────────────────────────┐
│              Adapters                    │  
│  CLI → Git → Config → Output → Signing   │
├──────────────────────────────────────────┤
│           Domain Interfaces              │
│    Repository • Logger • Verifier        │
├──────────────────────────────────────────┤
│             Core Domain                  │
│   Validation Functions • Rules • Types   │
└──────────────────────────────────────────┘
```

**Flow**: CLI creates adapters → Domain validates using interfaces → Results formatted by adapters

## Directory Structure

```txt
gommitlint/
├── main.go                    # Composition root
├── internal/
│   ├── domain/               # Pure business logic
│   │   ├── validation.go     # Core validation functions  
│   │   ├── types.go         # Domain entities and values
│   │   ├── config/          # Configuration types
│   │   └── rules/           # Validation rule implementations
│   ├── adapters/            # Infrastructure implementations
│   │   ├── cli/            # Command interface and dependency wiring
│   │   ├── git/            # Repository operations
│   │   ├── config/         # Configuration loading
│   │   ├── output/         # Report formatting  
│   │   ├── logging/        # Structured logging
│   │   └── signing/        # Signature verification
│   └── integrationtest/    # End-to-end tests
└── docs/                   # Documentation
```

## Core Patterns

### Pure Functions

All business logic implemented as pure functions:

```go
// Domain validation function - no side effects
func ValidateCommit(commit Commit, rules []Rule, repo Repository, cfg Config) ValidationResult {
    var errors []ValidationError
    for _, rule := range rules {
        errors = append(errors, rule.Validate(commit, cfg)...)
    }
    return ValidationResult{Commit: commit, Errors: errors}
}
```

### Value Semantics

All domain types use value receivers and return new instances:

```go
// Transformations return new values
func (c Commit) WithAuthor(author string) Commit {
    c.Author = author
    return c
}

// Configuration is immutable after creation
type Config struct {
    Rules     RuleConfig
    Output    OutputConfig
    // No setters - construct with NewConfig()
}
```

### Explicit Dependencies

Dependencies always passed as function parameters:

```go
// ❌ Avoid: Hidden dependencies
func ValidateCommit(commit Commit) ValidationResult {
    repo := globalRepo // Hidden dependency
    cfg := getConfig()  // Hidden dependency
}

// ✅ Prefer: Explicit dependencies  
func ValidateCommit(commit Commit, rules []Rule, repo Repository, cfg Config) ValidationResult {
    // All dependencies explicit
}
```

### Interface Location

Interfaces defined where consumed (domain), implemented by adapters:

```go
// Domain defines what it needs
type Repository interface {
    GetCommit(ctx context.Context, ref string) (Commit, error)
    ListCommits(ctx context.Context, from, to string) ([]Commit, error)
}

// Adapter implements domain interface
type GitRepository struct { /* implementation */ }
var _ domain.Repository = (*GitRepository)(nil)
```

## Key Components

### Domain Layer (`internal/domain/`)

**Purpose**: Business logic and core types

- **validation.go**: `ValidateCommit`, `ValidateCommits` functions
- **types.go**: `Commit`, `ValidationError`, `Report` value objects  
- **config/**: Configuration types and defaults
- **rules/**: Rule implementations and factory

**Characteristics**:
- No external dependencies except standard library
- All functions are pure (deterministic, no side effects)
- Defines interfaces for external dependencies

### Adapter Layer (`internal/adapters/`)

**Purpose**: Infrastructure implementations

- **cli/**: Command parsing, dependency wiring, main application flow
- **git/**: Git repository operations using go-git
- **config/**: Configuration file loading with koanf  
- **output/**: Report formatting (text, JSON, GitHub Actions)
- **logging/**: Structured logging with levels
- **signing/**: GPG/SSH signature verification

**Characteristics**:
- Implements domain interfaces
- Handles external system interactions
- Contains framework and library integrations

## Progressive Disclosure

Three-tier information architecture matching user needs:

### Default Output

```text
✗ Subject: First word 'add' should be 'Add'
✗ ConventionalCommit: Missing required scope
```
**For**: Quick problem identification

### Verbose (`-v`)  

```text
✗ Subject:
    Error Code: invalid_case
    Error Message: First word 'add' should be 'Add'  
    Expected: Add
    Actual: add
    Position: first word after type/scope
```
**For**: Technical details, CI/CD integration

### Extra Verbose (`-vv`)

```text  
✗ Subject:
    Error Code: invalid_case
    Error Message: First word 'add' should be 'Add'
    Expected: Add  
    Actual: add
    Position: first word after type/scope

    Help:
    Examples of correct sentence case:
    ✓ feat(auth): Add user authentication
    ✓ fix(api): Fix memory leak in handler
    
    Common fixes:
    • Check your editor's auto-capitalization settings
    • Remember: the first significant word sets the tone
```
**For**: Learning and guidance

## Configuration System

### Rule Priority (highest to lowest)

1. **Explicitly enabled** - Always run
2. **Explicitly disabled** - Never run  
3. **Default behavior** - Most rules enabled by default

### Implementation

```yaml
gommitlint:
  rules:
    enabled:
      - spell           # Override default-disabled
      - subjectlength   # Explicitly enabled
    disabled:
      - commitbody      # Always disabled
```

**Key Principle**: Configuration flows through explicit parameters, never stored in context.

## Testing Strategy

### Unit Tests

- **Pattern**: Table-driven with `testCase` variable
- **Coverage**: >80% across all packages
- **Mocking**: Simple interface implementations

```go
func TestValidateCommit(t *testing.T) {
    tests := []struct {
        name          string
        commit        Commit
        rules         []Rule
        expectErrors  int
    }{
        // Test cases...
    }
    
    for _, testCase := range tests {
        t.Run(testCase.name, func(t *testing.T) {
            result := ValidateCommit(testCase.commit, testCase.rules, mockRepo, config)
            require.Len(t, result.Errors, testCase.expectErrors)
        })
    }
}
```

### Integration Tests

- **Location**: `internal/integrationtest/`  
- **Scope**: End-to-end command validation
- **Data**: Real Git repositories in `testdata/`

## Extension Points

### Adding Rules

1. Implement `CommitRule` or `RepositoryRule` interface
2. Add to factory in `rules/factory.go`
3. Configure default enabled/disabled status

### Adding Output Formats  

1. Implement formatter function signature: `func(Report, Options) string`
2. Register in output adapter
3. Add CLI flag support

### Adding Adapters

1. Implement domain interfaces
2. Add to dependency composition in CLI layer
3. Wire in main application flow

## Quality Attributes

### Maintainability

- **Clear boundaries**: Domain logic separate from infrastructure
- **Simple dependencies**: Direct function composition
- **High testability**: Pure functions easy to test

### Performance  

- **Memory efficient**: Value semantics with predictable allocation
- **Concurrent safe**: Immutable data, no shared state
- **Fast startup**: <100ms for typical repositories

### Reliability

- **Explicit errors**: All errors returned, no panics
- **Resource safety**: Proper cleanup and bounded operations  
- **Graceful degradation**: Partial validation on failures

## Anti-Patterns Avoided

- **Service objects**: Use direct function composition
- **Global state**: Pass dependencies explicitly  
- **Hidden configuration**: All parameters visible
- **Complex factories**: Simple constructor functions
- **Mutable configuration**: Immutable after creation

---

This architecture prioritizes **simplicity** and **clarity** while maintaining the flexibility needed for a robust commit validation tool.