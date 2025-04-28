# Functional Architecture in Gommitlint

This document describes the functional architecture of Gommitlint, highlighting the principles, patterns, and implementation details that make it a functionally-oriented Go application.

## Architecture Overview

Gommitlint follows a hexagonal architecture (ports and adapters) with a strong emphasis on functional programming principles:

```
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
│   │   │   │    - Immutable Entities               │   │   │   │
│   │   │   │    - Pure Functions                   │   │   │   │
│   │   │   │    - Value Objects                    │   │   │   │
│   │   │   │                                       │   │   │   │
│   │   │   └───────────────────────────────────────┘   │   │   │
│   │   │                                               │   │   │
│   │   │    - Functional Transformations               │   │   │
│   │   │    - Pure Validation Logic                    │   │   │
│   │   │                                               │   │   │
│   │   └───────────────────────────────────────────────┘   │   │
│   │                                                       │   │
│   │    - Functional CLI Commands                          │   │
│   │    - Pure Repository Interfaces                       │   │
│   │                                                       │   │
│   └───────────────────────────────────────────────────────┘   │
│                                                               │
│    - Functional Adapters with Value Semantics                 │
│    - Immutable Configuration Providers                        │
│    - Functional Output Formatters                             │
│                                                               │
└───────────────────────────────────────────────────────────────┘
```

## Core Principles

The architecture is built on these functional programming principles:

1. **Immutability**: All data structures are treated as immutable
2. **Value Semantics**: Types use value receivers and return new instances
3. **Pure Functions**: Functions avoid side effects and return consistent results
4. **Function Composition**: Complex operations are built from simpler functions
5. **Explicit Dependencies**: Dependencies are passed explicitly, not through global state

## Layer-by-Layer Implementation

### Domain Layer

The domain layer contains the core business logic and entities. It's completely pure, with:

- **Immutable Entities**: All domain entities are immutable value types
- **Value Objects**: Small, self-contained value objects with value semantics
- **Pure Functions**: Functions that operate on domain entities without side effects

Example from `domain/commit.go`:

```go
// CommitInfo represents a Git commit with immutable fields
type CommitInfo struct {
    Hash          string
    Subject       string
    Body          string
    Message       string
    Author        string
    AuthorEmail   string
    IsMergeCommit bool
}

// SplitCommitMessage is a pure function that splits a commit message
func SplitCommitMessage(message string) (subject, body string) {
    parts := strings.SplitN(message, "\n\n", 2)
    subject = strings.TrimSpace(parts[0])
    
    if len(parts) > 1 {
        body = strings.TrimSpace(parts[1])
    }
    
    return subject, body
}
```

### Application Layer

The application layer orchestrates the domain layer with pure functions and transformations:

- **Stateless Services**: Services use value semantics for method receivers
- **Functional Transformations**: Operations return new states without mutation
- **Function Composition**: Complex operations are built by composing functions

Example from `application/validate/service.go`:

```go
// ValidationService orchestrates validation with value semantics
type ValidationService struct {
    engine        ValidationEngine
    commitService domain.GitCommitService
    infoProvider  domain.RepositoryInfoProvider
}

// WithEngine returns a new ValidationService with the engine replaced
func (s ValidationService) WithEngine(engine ValidationEngine) ValidationService {
    return ValidationService{
        engine:        engine,
        commitService: s.commitService,
        infoProvider:  s.infoProvider,
    }
}

// ValidateCommit validates a commit through function composition
func (s ValidationService) ValidateCommit(ctx context.Context, hash string) (domain.CommitResult, error) {
    // Get the commit from the repository
    commit, err := s.commitService.GetCommit(ctx, hash)
    if err != nil {
        return domain.CommitResult{}, fmt.Errorf("failed to get commit: %w", err)
    }

    // Validate through the engine (which uses pure functions)
    return s.engine.ValidateCommit(ctx, commit), nil
}
```

### Ports Layer

The ports layer provides interfaces to the outside world with functional patterns:

- **Pure Interfaces**: Interfaces favor transformation methods over mutation
- **Value-Based Commands**: CLI commands use value semantics
- **Function Option Parameters**: Commands use functional options for configuration

Example from CLI command design:

```go
// ValidateCommand implements a CLI command with value semantics
type ValidateCommand struct {
    service      validate.ValidationService
    outputFormat string
    commitCount  int
}

// WithOutputFormat returns a new command with the output format changed
func (c ValidateCommand) WithOutputFormat(format string) ValidateCommand {
    newCmd := c
    newCmd.outputFormat = format
    return newCmd
}

// Run executes the command using function composition
func (c ValidateCommand) Run(ctx context.Context, args []string) error {
    // Implementation using function composition...
}
```

### Infrastructure Layer

The infrastructure layer provides implementations of interfaces using functional patterns:

- **Value-Based Adapters**: Repository adapters use value semantics
- **Immutable Configuration**: Configuration providers are immutable
- **Functional Formatters**: Output formatters use functional transformations

Example from `infrastructure/output/text.go`:

```go
// TextFormatter formats validation results as text with value semantics
type TextFormatter struct {
    colorEnabled bool
    verbose      bool
}

// WithColor returns a new formatter with color setting changed
func (f TextFormatter) WithColor(enabled bool) TextFormatter {
    newFormatter := f
    newFormatter.colorEnabled = enabled
    return newFormatter
}

// Format creates a formatted string from validation results
// This is a pure function that doesn't modify state
func (f TextFormatter) Format(results domain.ValidationResults) string {
    // Pure function implementation...
}
```

## Key Functional Design Patterns

### 1. Value-Based Functional Options

Configuration is done through functions that transform values:

```go
type Option func(Config) Config

func WithVerbose(verbose bool) Option {
    return func(c Config) Config {
        newConfig := c // Create a copy
        newConfig.Verbose = verbose
        return newConfig
    }
}

func New(opts ...Option) Config {
    config := DefaultConfig()
    for _, opt := range opts {
        config = opt(config)
    }
    return config
}
```

### 2. Pure Function Extraction

Complex logic is extracted into pure functions:

```go
// Main method that uses pure functions
func (e Engine) ValidateCommit(ctx context.Context, commit domain.CommitInfo) domain.CommitResult {
    // Extract logic into pure functions
    ruleResults := validateWithRules(ctx, commit, e.provider.GetActiveRules())
    passed := allRulesPassed(ruleResults)
    
    return createResult(commit, ruleResults, passed)
}

// Pure functions that can be composed
func validateWithRules(ctx context.Context, commit domain.CommitInfo, rules []domain.Rule) []domain.RuleResult {
    // Implementation...
}

func allRulesPassed(results []domain.RuleResult) bool {
    // Implementation...
}

func createResult(commit domain.CommitInfo, ruleResults []domain.RuleResult, passed bool) domain.CommitResult {
    // Implementation...
}
```

### 3. Transformation Methods

Methods that modify state are replaced with transformation methods:

```go
// Instead of this:
func (g *Generator) SetVerbose(verbose bool) {
    g.options.Verbose = verbose
}

// Use this:
func (g Generator) WithVerbose(verbose bool) Generator {
    newOptions := copyOptions(g.options)
    newOptions.Verbose = verbose
    return Generator{
        options:   newOptions,
        formatter: g.formatter,
    }
}
```

### 4. Immutable Collections

Collections are treated as immutable with defensive copying:

```go
func (p DomainRuleProvider) copyRules() []domain.Rule {
    if p.rules == nil {
        return nil
    }
    
    rulesCopy := make([]domain.Rule, len(p.rules))
    copy(rulesCopy, p.rules)
    return rulesCopy
}
```

## Dependency Flow

In the functional architecture, dependencies flow inward with explicit passing:

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│  Domain Layer   │ ◄── │ Application     │ ◄── │ Ports Layer     │
│                 │     │ Layer           │     │                 │
└─────────────────┘     └─────────────────┘     └─────────────────┘
                                                        ▲
                                                        │
                                                        │
                                                ┌───────────────┐
                                                │Infrastructure │
                                                │Layer          │
                                                └───────────────┘
```

Composition roots (like main.go) create and wire dependencies:

```go
func main() {
    // Create infrastructure components
    repoAdapter := git.NewRepositoryAdapter("/path/to/repo")
    configProvider := config.NewProvider()
    
    // Create application services with dependencies
    validationService := validate.CreateValidationServiceWithDependencies(
        configProvider,
        repoAdapter,
        repoAdapter,
        repoAdapter,
    )
    
    // Create port (CLI) with dependencies
    validateCmd := cli.NewValidateCommand(validationService)
    
    // Execute
    // ...
}
```

## Concurrency Safety

The functional architecture makes concurrency naturally safer because:

1. **No shared mutable state**: Methods don't modify their receivers
2. **Immutable data structures**: Data is treated as immutable
3. **Value passing**: Values are passed and copied rather than shared

This means that most concurrency issues are eliminated by design - no locks or synchronization primitives are needed.

## Testing Advantages

Functional architecture provides testing advantages:

1. **Isolated testing**: Pure functions can be tested in isolation
2. **Predictable results**: Same input always produces same output
3. **No state management**: Tests don't need to set up and tear down state
4. **Simplified mocks**: Dependencies are explicit and easily replaceable

Example of isolated testing for pure functions:

```go
func TestValidateWithRules(t *testing.T) {
    // Create test data
    commit := domain.CommitInfo{
        Subject: "Test commit",
        Body:    "Test body",
    }
    
    rules := []domain.Rule{mockRule{}}
    ctx := context.Background()
    
    // Test the pure function directly
    results := validateWithRules(ctx, commit, rules)
    
    // Assert results
    require.Len(t, results, 1)
    require.Equal(t, "mockRule", results[0].RuleName)
}
```

## Conclusion

Gommitlint's functional architecture provides a robust foundation for a maintainable, testable, and concurrent-safe application. By following functional programming principles throughout the codebase, the application achieves:

1. **Simplicity**: Pure functions and immutability reduce complexity
2. **Composability**: Small, focused functions compose into complex operations
3. **Testability**: Pure functions are easy to test in isolation
4. **Predictability**: Immutable data and explicit state transitions make behavior predictable
5. **Concurrency safety**: Value semantics eliminate many concurrency issues

This architecture enables easier maintenance, clearer code understanding, and safer concurrent operation.