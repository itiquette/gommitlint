# Gommitlint Architecture

Gommitlint follows a hexagonal architecture (also known as ports and adapters) to ensure a clean separation of concerns and to make the codebase more maintainable and testable.

## Overview

The hexagonal architecture divides the application into layers:

1. **Domain Layer** (center): Core business logic and entities
2. **Application Layer**: Use cases that orchestrate domain entities
3. **Ports Layer**: Interfaces that connect the application to the outside world
4. **Adapters/Infrastructure Layer**: Implementation of interfaces defined in the ports layer

This architectural approach provides several benefits:
- Clear separation of concerns
- Domain logic isolated from infrastructure details
- Testability through interface-based design
- Flexibility to change implementation details without affecting core business logic

```javascript
┌───────────────────────────────────────────────────────────────┐
│                      Infrastructure Layer                      │
│                                                               │
│   ┌───────────────────────────────────────────────────────┐   │
│   │                      Ports Layer                      │   │
│   │                                                       │   │
│   │   ┌───────────────────────────────────────────────┐   │   │
│   │   │                Application Layer              │   │   │
│   │   │                                               │   │   │
│   │   │   ┌───────────────────────────────────────┐   │   │   │
│   │   │   │             Domain Layer              │   │   │   │
│   │   │   │                                       │   │   │   │
│   │   │   │    - Domain Entities                  │   │   │   │
│   │   │   │    - Rule Interfaces                  │   │   │   │
│   │   │   │    - Value Objects                    │   │   │   │
│   │   │   │                                       │   │   │   │
│   │   │   └───────────────────────────────────────┘   │   │   │
│   │   │                                               │   │   │
│   │   │    - Validation Service                       │   │   │
│   │   │    - Report Generation                        │   │   │
│   │   │                                               │   │   │
│   │   └───────────────────────────────────────────────┘   │   │
│   │                                                       │   │
│   │    - CLI Commands                                     │   │
│   │    - Repository Interfaces                            │   │
│   │                                                       │   │
│   └───────────────────────────────────────────────────────┘   │
│                                                               │
│    - Git Repository Implementation                            │
│    - Configuration Provider                                   │
│    - Output Formatters                                        │
│                                                               │
└───────────────────────────────────────────────────────────────┘
```

## Project Structure

### Domain Layer

The domain layer is the heart of the application, containing the core business logic and entities:

- `/internal/domain/commit.go`: Commit-related domain entities
- `/internal/domain/rule.go`: Rule interfaces and validation errors
- `/internal/domain/result.go`: Validation result structures
- `/internal/domain/git_interfaces.go`: Segregated interfaces for Git operations
- `/internal/domain/commit_collection.go`: Domain collection for commit operations

Key interfaces:
- `Rule`: Interface that all validation rules implement
- `CommitReader`: Interface for reading individual commits
- `CommitHistoryReader`: Interface for accessing commit history
- `RepositoryInfoProvider`: Interface for repository information
- `CommitAnalyzer`: Interface for analyzing commits
- `GitRepositoryService`: Composite interface combining all Git interfaces

### Core Layer

The core layer contains the business rules implementation:

- `/internal/core/validation/engine.go`: Validation engine
- `/internal/core/validation/rule_provider.go`: Rule provider
- `/internal/core/rules/`: Rule implementations

All rules follow the same pattern and implement the `domain.Rule` interface.

### Application Layer

The application layer orchestrates the domain layer:

- `/internal/application/validate/service.go`: Validation service
- `/internal/application/report/generator.go`: Report generator

### Ports Layer

The ports layer provides interfaces to the outside world:

- `/internal/ports/cli/validate.go`: CLI validation command
- `/internal/ports/fileio/`: File I/O interfaces

### Infrastructure Layer (Adapters)

The infrastructure layer provides concrete implementations of interfaces:

- `/internal/infrastructure/git/repository.go`: Git repository adapter
- `/internal/infrastructure/git/repository_helpers.go`: Helper methods for common Git operations
- `/internal/infrastructure/git/repository_factory.go`: Factory for creating repository interfaces
- `/internal/infrastructure/config/provider.go`: Configuration provider
- `/internal/infrastructure/output/`: Output formatters (text, JSON)

## Interface Segregation

Following the Interface Segregation Principle, Git repository interfaces are segregated based on client needs:

```go
// CommitReader provides read access to individual commits.
type CommitReader interface {
    // GetCommit returns a commit by its hash.
    GetCommit(hash string) (*CommitInfo, error)
}

// CommitHistoryReader provides read access to commit history.
type CommitHistoryReader interface {
    // GetCommitRange returns all commits in the given range.
    GetCommitRange(fromHash, toHash string) ([]*CommitInfo, error)
    
    // GetHeadCommits returns the specified number of commits from HEAD.
    GetHeadCommits(count int) ([]*CommitInfo, error)
}

// RepositoryInfoProvider provides general information about the repository.
type RepositoryInfoProvider interface {
    // GetCurrentBranch returns the name of the current branch.
    GetCurrentBranch() (string, error)
    
    // GetRepositoryName returns the name of the repository.
    GetRepositoryName() string
    
    // IsValid checks if the repository is a valid Git repository.
    IsValid() bool
}

// CommitAnalyzer provides analysis functionality for commits.
type CommitAnalyzer interface {
    // GetCommitsAhead returns the number of commits ahead of the given reference.
    GetCommitsAhead(reference string) (int, error)
}

// GitRepositoryService combines all Git repository interfaces.
type GitRepositoryService interface {
    CommitReader
    CommitHistoryReader
    RepositoryInfoProvider
    CommitAnalyzer
}
```

These specialized interfaces allow clients to depend only on what they need, reducing coupling.

## Domain Collections

The `CommitCollection` type provides domain-specific operations on groups of commits:

```go
// CommitCollection represents a collection of commits with common operations.
type CommitCollection struct {
    commits []*CommitInfo
}

// FilterMergeCommits returns a new collection with merge commits filtered out.
func (c *CommitCollection) FilterMergeCommits() *CommitCollection {
    // Implementation...
}

// FilterByAuthor returns a new collection with commits filtered by author.
func (c *CommitCollection) FilterByAuthor(author string) *CommitCollection {
    // Implementation...
}
```

## Repository Pattern

For Git operations, Gommitlint uses:

1. Repository interfaces in the domain layer
2. A repository adapter in the infrastructure layer
3. A factory to create specific repository interfaces
4. Helper methods to reduce complexity

## Factory Pattern

The `RepositoryFactory` creates specialized interfaces for different use cases:

```go
// RepositoryFactory provides factory methods for creating Git repository services.
type RepositoryFactory struct {
    adapter *RepositoryAdapter
}

// CreateCommitReader creates a CommitReader.
func (f *RepositoryFactory) CreateCommitReader() domain.CommitReader {
    return f.adapter
}

// CreateHistoryReader creates a CommitHistoryReader.
func (f *RepositoryFactory) CreateHistoryReader() domain.CommitHistoryReader {
    return f.adapter
}
```

## Dependency Injection

Dependencies are injected through:

1. Constructor parameters
2. Factory pattern for creating repositories
3. Functional options pattern for configuration

## Error Handling

The application uses a structured approach to error handling:

1. Domain-specific errors with context information
2. Error wrapping for maintaining context
3. Standardized validation error format
4. Error templates with consistent formatting for messages and help text

The error template system centralizes error message formats:

1. Each error type has a predefined template with consistent wording
2. Templates include both error messages and help text
3. Helper functions ensure uniform context handling
4. Standardized error codes are mapped to templates
5. Consistent formatting across all validation rules

The template system is implemented in the errorx package:
- `ErrorTemplate` struct defines standardized formats for messages and help text
- `NewValidationError` creates templated errors with proper formatting
- `CreateValidationError` provides a standardized way to create errors with context
- `WithContext` provides a standardized way to add context information

## Validation Rule Design

Validation rules follow these principles:

1. **Interface-Based**: All rules implement the `domain.Rule` interface
2. **Functional Options**: Rules use the functional options pattern for configuration
3. **Immutability**: Rules are immutable after creation
4. **Self-Contained**: Each rule contains all logic needed for validation
5. **Plugin-Based**: Rules are registered through a provider
6. **Focused**: Each rule focuses on a specific aspect of commit message validation

### Rule Interface

```go
type Rule interface {
    // Name returns the rule's name.
    Name() string

    // Validate performs validation against a commit and returns any errors.
    Validate(commit *CommitInfo) []*ValidationError

    // Result returns a concise result message.
    Result() string

    // VerboseResult returns a detailed result message.
    VerboseResult() string

    // Help returns guidance on how to fix rule violations.
    Help() string

    // Errors returns all validation errors found by this rule.
    Errors() []*ValidationError
}
```

## Testing

The application is designed for testability:

1. **Unit Tests**: Each component is tested in isolation
2. **Table-Driven Tests**: Tests use the table-driven pattern for comprehensive coverage
3. **Mock Interfaces**: Tests use mocks for dependencies
4. **Co-located Tests**: Tests are next to the code they test
5. **Interface-Based Design**: Interface-based design allows for easy mocking

## Benefits of the Architecture

The hexagonal architecture provides several benefits:

1. **Cleaner Code**: Clear separation of concerns makes the code easier to understand
2. **Enhanced Testability**: Dependency inversion makes unit testing straightforward
3. **Improved Maintainability**: Each component has a single responsibility
4. **Better Extensibility**: Adding new rules or features follows a consistent pattern
5. **Reduced Coupling**: Components depend on abstractions, not concrete implementations


## Areas for Improvement


A  few done...

  3. Streamline Configuration System

  The configuration loading and management has multiple layers:
  - Config loader
  - Config provider
  - Config manager
  - Config validator

  Suggestion: Consider a simpler approach to configuration with fewer abstractions. A single configuration manager that handles loading and
  validation might be sufficient.

  4. Simplify Rule Registration

  The rule registration system is somewhat complex:

  Suggestion: Consider a more straightforward registration mechanism, possibly using Go's init pattern or a simpler factory function.

  5. Reduce Factory Proliferation

  There are multiple factory patterns which add indirection:

  Suggestion: Consolidate factories where possible and consider more direct instantiation patterns where appropriate.

### General Assessment

  - Idiomatic Go: Overall, yes. The code follows Go idioms well, with proper use of interfaces, error handling, and package structure.
  - Maintainability: Strong, due to clear separation of concerns and consistent patterns.
  - Testability: Excellent, with dependency injection and interface-based design enabling thorough testing.
  - Simplicity: Good but with room for improvement in some areas.
  - Coherence: Strong architectural boundaries make the system coherent.
  - Conciseness: Generally good, with the BaseRule pattern helping reduce dupl
