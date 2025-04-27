# Gommitlint Architecture

Gommitlint follows a simplified hexagonal architecture (also known as ports and adapters) to ensure a clean separation of concerns and to make the codebase more maintainable and testable.

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
- `/internal/ports/cli/installhook.go`: CLI command for installing Git hooks
- `/internal/ports/cli/removehook.go`: CLI command for removing Git hooks

### Infrastructure Layer (Adapters)

The infrastructure layer provides concrete implementations of interfaces:

- `/internal/infrastructure/git/repository.go`: Git repository adapter
- `/internal/infrastructure/git/repository_helpers.go`: Helper methods for common Git operations
- `/internal/infrastructure/git/repository_factory.go`: Factory for creating repository interfaces
- `/internal/infrastructure/config/provider.go`: Configuration provider
- `/internal/infrastructure/output/`: Output formatters (text, JSON, GitHub, GitLab)

## Simplified Interfaces

The new architecture includes simplified interfaces based on the Interface Segregation Principle, which ensures that clients only depend on the methods they actually use.

### RuleProvider Interface

```go
// RuleProvider defines the interface for accessing validation rules.
type RuleProvider interface {
    // GetAvailableRules returns all registered rules.
    GetAvailableRules() []string
    
    // GetActiveRules returns currently active rules.
    GetActiveRules() []string
    
    // SetEnabledRules activates specific rules.
    SetEnabledRules(rules []string)
    
    // SetDisabledRules deactivates specific rules.
    SetDisabledRules(rules []string)
}
```

### ValidationConfigAdapter

This adapter implements configuration interfaces for rule validation:

```go
// ValidationConfigAdapter adapts the configuration to rule validation interfaces.
type ValidationConfigAdapter struct {
    config *Config
}

// The adapter implements various interfaces:
// - RuleProvider for rule activation/deactivation
// - SubjectConfigProvider for subject-related settings
// - BodyConfigProvider for body-related settings
// - ConventionalConfigProvider for conventional commit settings
// - ... and other configuration interfaces
```

### CommitService Interface

```go
// CommitService provides access to Git commit operations.
type CommitService interface {
    // GetCommit returns a specific commit by hash.
    GetCommit(ctx context.Context, hash string) (domain.CommitInfo, error)
    
    // GetHeadCommits returns the specified number of commits from HEAD.
    GetHeadCommits(ctx context.Context, count int) (domain.CommitCollection, error)
    
    // GetCommitRange returns commits in the given range.
    GetCommitRange(ctx context.Context, fromHash, toHash string) (domain.CommitCollection, error)
}
```

### ValidationService Interface

```go
// ValidationService provides commit validation functionality.
type ValidationService interface {
    // ValidateCommit validates a single commit.
    ValidateCommit(ctx context.Context, hash string) (domain.ValidationResult, error)
    
    // ValidateHeadCommits validates a number of HEAD commits.
    ValidateHeadCommits(ctx context.Context, count int, skipMergeCommits bool) (domain.ValidationResult, error)
    
    // ValidateCommitRange validates a range of commits.
    ValidateCommitRange(ctx context.Context, fromHash, toHash string, skipMergeCommits bool) (domain.ValidationResult, error)
    
    // ValidateMessageFile validates a commit message from a file.
    ValidateMessageFile(ctx context.Context, filePath string) (domain.ValidationResult, error)
    
    // ValidateMessage validates a commit message string.
    ValidateMessage(ctx context.Context, message string) (domain.ValidationResult, error)
}
```

## Functional Programming and Value Semantics

The updated architecture embraces functional programming principles and value semantics extensively for better immutability, predictability, and testability:

### Value Semantics

```go
// Using value semantics for immutable data structures
type ValidationResult struct {
    PassCount    int
    FailCount    int
    Commits      []CommitResult
    RuleResults  map[string]RuleResult
}

// Functions return new structures rather than modifying existing ones
func (c CommitCollection) FilterMergeCommits() CommitCollection {
    var filtered []CommitInfo
    for _, commit := range c.commits {
        if !commit.IsMergeCommit {
            filtered = append(filtered, commit)
        }
    }
    return CommitCollection{commits: filtered}
}
```

### Fluent Value Methods

Methods are designed to work with value semantics, returning new instances instead of modifying receivers:

```go
// Value receiver with new instance return
func (r SubjectCaseRule) AddError(err appErrors.ValidationError) SubjectCaseRule {
    rule := r
    rule.BaseRule = rule.BaseRule.WithError(err)
    return rule
}

// Chaining methods
result := rule.
    ClearErrors().
    AddError(newError).
    SetFoundKeys(keys)
```

### Immutable State Transformations

Business logic operates on state through transformations that return new state:

```go
// State transformation through chained methods
func (s *ValidationService) ValidateCommit(ctx context.Context, hash string) (domain.ValidationResult, error) {
    // Get the commit
    commit, err := s.commitService.GetCommit(ctx, hash)
    if err != nil {
        return domain.ValidationResult{}, fmt.Errorf("failed to get commit: %w", err)
    }
    
    // Transform through validation
    return s.engine.Validate(commit), nil
}
```

### Pure Functional Validation

Validation in rules follows a functional pattern where state is transformed rather than modified:

```go
// Validation with pure functions and value semantics
func (r SubjectCaseRule) Validate(commit domain.CommitInfo) []appErrors.ValidationError {
    // Validate and return errors without modifying state
    if !meetsCase(commit.Subject) {
        return []appErrors.ValidationError{
            appErrors.New(r.Name(), appErrors.ErrSubjectCase, "Subject case is incorrect"),
        }
    }
    return []appErrors.ValidationError{}
}

// State transformation function
func validateWithState(rule SubjectCaseRule, commit domain.CommitInfo) ([]appErrors.ValidationError, SubjectCaseRule) {
    errors := []appErrors.ValidationError{}
    updatedRule := rule
    
    // Add logic and transform rule as needed
    if !meetsCase(commit.Subject) {
        err := appErrors.New(rule.Name(), appErrors.ErrSubjectCase, "Subject case is incorrect")
        errors = append(errors, err)
        updatedRule = updatedRule.addError(err)
    }
    
    return errors, updatedRule
}
```

## Functional Options Pattern

Rule creation follows the functional options pattern for flexible configuration with value semantics:

```go
// Option type for configuring rules
type SubjectLengthOption func(SubjectLengthRule) SubjectLengthRule

// Options function with value semantics
func WithMaxLength(length int) SubjectLengthOption {
    return func(r SubjectLengthRule) SubjectLengthRule {
        result := r
        result.maxLength = length
        return result
    }
}

// Factory function with options returning a value
func NewSubjectLengthRule(options ...SubjectLengthOption) SubjectLengthRule {
    rule := SubjectLengthRule{
        BaseRule: NewBaseRule("SubjectLength"),
        maxLength: DefaultMaxLength,
    }
    
    // Apply options in sequence, each returning a new value
    for _, option := range options {
        rule = option(rule)
    }
    
    return rule
}

// Usage
rule := NewSubjectLengthRule(
    WithMaxLength(80),
    WithErrorTemplate(customTemplate),
)
```

## Integration Testing

The architecture includes a dedicated integration test package:

- `/internal/integtest/`: Contains integration tests that test multiple components together
  - `validation_workflow_test.go`: Tests the complete validation workflow
  - `cli_workflow_test.go`: Tests CLI commands
  - `comprehensive_test.go`: Tests various scenarios comprehensively
  - `gittest_helper.go`: Helpers for setting up test Git repositories

This approach provides more robust testing than unit tests alone, ensuring that components work together correctly.

## Example Usage Patterns

### Creating and Using the Validation Service

```go
// In application code
func CreateValidationService(cfg *config.Config) (*ValidationService, error) {
    // Create repository objects
    repoFactory := git.NewRepositoryFactory("/path/to/repo")
    commitService := repoFactory.CreateCommitService()
    infoProvider := repoFactory.CreateRepositoryInfoProvider() 
    
    // Create validation config adapter
    configAdapter := config.NewValidationConfigAdapter(cfg)
    
    // Create rule registry
    ruleRegistry := validation.NewRuleRegistry()
    
    // Register built-in rules
    ruleRegistry.RegisterRule(rules.NewSubjectLengthRule(
        rules.WithMaxLength(configAdapter.SubjectMaxLength()),
    ))
    ruleRegistry.RegisterRule(rules.NewConventionalCommitRule(
        rules.WithRequireConventional(configAdapter.ConventionalRequired()),
        rules.WithAllowedTypes(configAdapter.ConventionalTypes()),
    ))
    // Register other rules...
    
    // Create validation engine
    engine := validation.NewEngine(ruleRegistry, configAdapter)
    
    // Create and return the service
    return &ValidationService{
        commitService: commitService,
        infoProvider:  infoProvider,
        engine:        engine,
        config:        configAdapter,
    }, nil
}
```

### Validating a Commit

```go
func ValidateHeadCommit() error {
    // Create validation service
    validationService, err := CreateValidationService(config.Load())
    if err != nil {
        return fmt.Errorf("failed to create validation service: %w", err)
    }
    
    // Create context
    ctx := context.Background()
    
    // Validate HEAD commit
    result, err := validationService.ValidateCommit(ctx, "HEAD")
    if err != nil {
        return fmt.Errorf("validation failed: %w", err)
    }
    
    // Process result
    for _, commitResult := range result.Commits {
        fmt.Printf("Commit %s: %s\n", commitResult.Hash, commitResult.Subject)
        
        for ruleName, ruleResult := range commitResult.RuleResults {
            if ruleResult.Status == domain.RuleStatusPassed {
                fmt.Printf("✓ %s: passed\n", ruleName)
            } else {
                fmt.Printf("✗ %s: failed\n", ruleName)
                for _, err := range ruleResult.Errors {
                    fmt.Printf("  - %s\n", err.Message)
                }
            }
        }
    }
    
    return nil
}
```

### Creating a Custom Rule

```go
// Define your custom rule
type CustomRule struct {
    BaseRule      rules.BaseRule
    customConfig  string
}

// Implement the Validate method with value semantics
func (r CustomRule) Validate(commit domain.CommitInfo) []domain.ValidationError {
    // Your validation logic here
    if !strings.Contains(commit.Subject, r.customConfig) {
        return []domain.ValidationError{
            domain.NewValidationError(
                "CustomRule",
                "custom_rule_violation",
                fmt.Sprintf("Subject must contain '%s'", r.customConfig),
            ),
        }
    }
    return nil
}

// Create a factory function with options
func NewCustomRule(options ...CustomRuleOption) CustomRule {
    rule := CustomRule{
        BaseRule:     rules.NewBaseRule("CustomRule"),
        customConfig: "default",
    }
    
    // Apply options in sequence, each returning a new value
    for _, option := range options {
        rule = option(rule)
    }
    
    return rule
}

// Define options with value semantics
type CustomRuleOption func(CustomRule) CustomRule

func WithCustomConfig(config string) CustomRuleOption {
    return func(r CustomRule) CustomRule {
        result := r
        result.customConfig = config
        return result
    }
}

// Register your rule
func RegisterCustomRule(registry *validation.RuleRegistry, config string) {
    registry.RegisterRule(NewCustomRule(
        WithCustomConfig(config),
    ))
}
```

## Benefits of the Updated Architecture

The updated architecture provides several additional benefits:

1. **Simplified Interfaces**: More focused interfaces with fewer methods make the system easier to understand and implement
2. **Value Semantics**: Immutable data structures and functional programming patterns improve code safety and readability
3. **Dedicated Integration Testing**: Comprehensive integration tests ensure components work together correctly
4. **Explicit Dependencies**: Dependencies are clearly stated and injected, improving testability and flexibility
5. **Consistent Context Handling**: Context objects are propagated throughout the application for better cancellation and timeout handling

## Error Handling

The error handling approach has been improved with better context:

```go
// Domain-specific error with context
validationErr := domain.NewValidationError(
    "RuleName",           // Rule that found the error
    "error_code",         // Specific error code
    "error message",      // Human-readable message
    domain.WithContext("key", "value"),  // Additional context
)

// Error wrapping for maintaining context
if err != nil {
    return fmt.Errorf("failed to validate commit %s: %w", hash, err)
}
```

## Testing Strategy

The testing strategy now emphasizes integration testing while maintaining strong unit tests:

1. **Unit Tests**: Each component is tested in isolation
2. **Integration Tests**: Key workflows are tested end-to-end
3. **Table-Driven Tests**: Tests use the table-driven pattern for comprehensive coverage
4. **Test Helpers**: Dedicated helpers simplify test setup
5. **Realistic Test Data**: Tests use realistic data to better simulate actual usage

This approach provides more robust validation that the system works correctly as a whole.
