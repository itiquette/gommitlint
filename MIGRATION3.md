# Complete Functional Migration Plan

This document outlines a plan to convert the remaining pointer receivers and imperative patterns in Gommitlint to functional programming style with value semantics, completing the transition to a fully functional codebase.

## Current State Analysis

While the core rule system has been migrated to use value semantics, several areas of the application still use pointer receivers and imperative programming patterns:

1. **Application Services**: Some application services still use pointer receivers
2. **Infrastructure Adapters**: Repository and configuration adapters often use pointer receivers
3. **CLI Commands**: Command implementations typically use pointer receivers
4. **Report Generation**: Report generators may modify state
5. **Validation Engine**: Parts of the validation engine may use imperative patterns

## Guiding Principles for Completion

To complete the migration, we'll follow these core functional programming principles:

1. **All Types Use Value Semantics**: Replace all pointer receivers with value receivers
2. **Immutable Data**: Ensure no internal state is modified
3. **Pure Functions**: Functions return values without side effects
4. **Function Composition**: Build complex operations from simple functions
5. **Explicit State Transitions**: Pass state explicitly, don't modify it implicitly

## Specific Migration Areas

### 1. Application Services

Convert application services to use value semantics:

```go
// BEFORE:
type ValidationService struct {
    engine        domain.ValidationEngine
    commitService domain.GitCommitService
    // ...
}

func (s *ValidationService) ValidateCommit(ctx context.Context, hash string) (domain.CommitResult, error) {
    // Implementation with potential state modification
}

// AFTER:
type ValidationService struct {
    engine        domain.ValidationEngine
    commitService domain.GitCommitService
    // ...
}

func (s ValidationService) ValidateCommit(ctx context.Context, hash string) (domain.CommitResult, error) {
    // Implementation without state modification
}
```

### 2. Infrastructure Adapters

Convert infrastructure adapters to use value semantics:

```go
// BEFORE:
type RepositoryAdapter struct {
    repo *git.Repository
    path string
}

func (g *RepositoryAdapter) GetCommit(ctx context.Context, hash string) (domain.CommitInfo, error) {
    // Implementation potentially modifying internal state
}

// AFTER:
type RepositoryAdapter struct {
    repo git.Repository
    path string
}

func (g RepositoryAdapter) GetCommit(ctx context.Context, hash string) (domain.CommitInfo, error) {
    // Implementation without modifying state
}
```

### 3. CLI Commands

Convert CLI commands to use value semantics:

```go
// BEFORE:
type ValidateCommand struct {
    service *validation.Service
    output  *output.Formatter
    // ...
}

func (c *ValidateCommand) Run(args []string) error {
    // Implementation potentially modifying internal state
}

// AFTER:
type ValidateCommand struct {
    service validation.Service
    output  output.Formatter
    // ...
}

func (c ValidateCommand) Run(args []string) (ValidateCommand, error) {
    // Implementation returning new state
}

// At the cobra command level:
func runValidate(cmd *cobra.Command, args []string) error {
    validateCmd := NewValidateCommand(service, output)
    _, err := validateCmd.Run(args)
    return err
}
```

### 4. Report Generation

Convert report generators to use value semantics:

```go
// BEFORE:
type ReportGenerator struct {
    formatter *output.Formatter
    // ...
}

func (g *ReportGenerator) Generate(results domain.ValidationResults) (string, error) {
    // Implementation potentially modifying internal state
}

// AFTER:
type ReportGenerator struct {
    formatter output.Formatter
    // ...
}

func (g ReportGenerator) Generate(results domain.ValidationResults) (string, error) {
    // Implementation without modifying state
}
```

### 5. Validation Engine

Ensure validation engine uses functional patterns:

```go
// BEFORE:
type ValidationEngine struct {
    rules []*domain.Rule
    // ...
}

func (e *ValidationEngine) ValidateCommit(ctx context.Context, commit domain.CommitInfo) domain.CommitResult {
    // Implementation potentially modifying internal state
}

// AFTER:
type ValidationEngine struct {
    rules []domain.Rule
    // ...
}

func (e ValidationEngine) ValidateCommit(ctx context.Context, commit domain.CommitInfo) domain.CommitResult {
    // Implementation without modifying state
    results := make([]domain.RuleResult, 0, len(e.rules))
    
    for _, rule := range e.rules {
        result := executeRule(rule, commit)
        results = append(results, result)
    }
    
    return createFinalResult(commit, results)
}

// Helper pure functions
func executeRule(rule domain.Rule, commit domain.CommitInfo) domain.RuleResult {
    // Pure function to execute a rule
}

func createFinalResult(commit domain.CommitInfo, results []domain.RuleResult) domain.CommitResult {
    // Pure function to create the final result
}
```

## Implementation Strategy

### Step 1: Identify Remaining Pointer Receivers

Use a code scan to identify all remaining pointer receivers:

```bash
grep -r "func ([^V].*\*[A-Za-z0-9_]*)" --include="*.go" . | grep -v "_test.go" | sort
```

### Step 2: Prioritize Components for Migration

1. **Core Application Services**: These influence many other components
2. **Infrastructure Interfaces**: Repository and configuration adapters
3. **Validation Engine**: Complete the central engine migration
4. **Report Generation**: Convert output formatting components
5. **CLI Commands**: Update the user-facing commands

### Step 3: Create Value-Based Alternatives

For each component, create a value-based alternative:

1. Create the same type with value receivers
2. Implement the same interface
3. Create factory functions that return values
4. Update callers to use the new implementation

### Step 4: Handling External Dependencies

For external dependencies that require pointer receivers, use an adapter pattern:

```go
// For go-git Repository that requires pointer operations
type GitRepositoryAdapter struct {
    repo *git.Repository // Pointer to external dependency
}

// Public methods use value semantics
func (g GitRepositoryAdapter) GetCommit(hash string) (domain.CommitInfo, error) {
    // Call pointer-based methods internally
    // But maintain value semantics on the public API
    commit, err := g.repo.CommitObject(plumbing.NewHash(hash))
    if err != nil {
        return domain.CommitInfo{}, err
    }
    
    return convertCommit(commit), nil
}

// Pure function for conversion
func convertCommit(commit *object.Commit) domain.CommitInfo {
    // Convert from pointer-based object to value-based domain entity
}
```

### Step 5: Handling Mutable Map Types

For map types that need to be updated, use functional patterns:

```go
// BEFORE:
func (s *Service) AddConfig(key, value string) {
    s.config[key] = value
}

// AFTER:
func (s Service) WithConfig(key, value string) Service {
    result := s
    newConfig := make(map[string]string, len(s.config)+1)
    
    // Copy existing values
    for k, v := range s.config {
        newConfig[k] = v
    }
    
    // Add new value
    newConfig[key] = value
    
    result.config = newConfig
    return result
}
```

### Step 6: Composition Pattern for Complex Operations

For complex operations, use function composition:

```go
// BEFORE:
func (s *Service) ValidateAndReport(commit string) (string, error) {
    result, err := s.validator.Validate(commit)
    if err != nil {
        return "", err
    }
    
    return s.reporter.Generate(result)
}

// AFTER:
func (s Service) ValidateAndReport(commit string) (string, error) {
    // Compose operations without modifying state
    validateResult := func(c string) (domain.ValidationResult, error) {
        return s.validator.Validate(c)
    }
    
    generateReport := func(r domain.ValidationResult) (string, error) {
        return s.reporter.Generate(r)
    }
    
    // Compose the functions
    result, err := validateResult(commit)
    if err != nil {
        return "", err
    }
    
    return generateReport(result)
}
```

## Common Patterns to Apply

### Factory Functions

Prefer factory functions that return values:

```go
// BEFORE:
func NewService() *Service {
    return &Service{
        // ...
    }
}

// AFTER:
func NewService() Service {
    return Service{
        // ...
    }
}
```

### Transformation Methods

Use transformation methods that return new values:

```go
// BEFORE:
func (s *Service) Configure(config Config) {
    s.config = config
}

// AFTER:
func (s Service) WithConfig(config Config) Service {
    result := s
    result.config = config
    return result
}
```

### State Initialization

Initialize collections properly:

```go
// BEFORE:
type Service struct {
    rules []Rule // Might be nil
}

// AFTER:
type Service struct {
    rules []Rule // Never nil
}

func NewService() Service {
    return Service{
        rules: make([]Rule, 0), // Always initialized
    }
}
```

### Collection Operations

Use pure functions for collection operations:

```go
// BEFORE:
func (s *Service) AddRule(rule Rule) {
    s.rules = append(s.rules, rule)
}

// AFTER:
func (s Service) WithRule(rule Rule) Service {
    result := s
    newRules := make([]Rule, len(s.rules), len(s.rules)+1)
    copy(newRules, s.rules)
    result.rules = append(newRules, rule)
    return result
}
```

## Testing Approach

### Testing Value Semantics

Verify value semantics in tests:

```go
func TestServiceImmutability(t *testing.T) {
    // Create original service
    original := NewService()
    
    // Create modified copy
    modified := original.WithConfig(Config{Value: "test"})
    
    // Original should remain unchanged
    require.Empty(t, original.config.Value)
    require.Equal(t, "test", modified.config.Value)
}
```

### Testing Pure Functions

Test pure functions separately:

```go
func TestPureFunctions(t *testing.T) {
    // Test data
    input := "test input"
    
    // First pure function
    result1 := processInput(input)
    require.Equal(t, "expected output", result1)
    
    // Second pure function
    result2 := transformResult(result1)
    require.Equal(t, "transformed output", result2)
    
    // Composition
    result3 := transformResult(processInput(input))
    require.Equal(t, "transformed output", result3)
}
```

## Implementation Phases

### Phase 1: Infrastructure Layer

1. Convert repository adapters to value semantics
2. Convert configuration providers to value semantics
3. Convert output formatters to value semantics

### Phase 2: Application Layer

1. Convert validation service to value semantics
2. Convert report generators to value semantics
3. Convert rule engines to value semantics (if any remaining)

### Phase 3: Ports Layer

1. Convert CLI commands to value semantics
2. Convert web handlers to value semantics (if any)
3. Convert integration points to value semantics

### Phase 4: Testing and Verification

1. Update tests to verify immutability
2. Add property-based tests for pure functions
3. Verify composition patterns work correctly

### Phase 5: Documentation and Finalization

1. Document the functional architecture
2. Create examples of functional patterns
3. Update architecture documentation

## Completion Criteria

The migration will be considered complete when:

1. All user-defined types use value receivers (except where required by external interfaces)
2. All functions are pure (no side effects, return new values)
3. All data structures are immutable (modifications return new instances)
4. All tests pass and verify immutability properties
5. All documentation reflects the functional approach

## Examples of Complete Functional Patterns

### Example 1: Stateless Validation

```go
// Value-based types
type Validator struct {
    rules []Rule
}

// Factory function returning a value
func NewValidator(rules []Rule) Validator {
    rulesCopy := make([]Rule, len(rules))
    copy(rulesCopy, rules)
    
    return Validator{
        rules: rulesCopy,
    }
}

// Value receiver method
func (v Validator) Validate(commit CommitInfo) ValidationResult {
    results := make([]RuleResult, 0, len(v.rules))
    
    for _, rule := range v.rules {
        result := runRule(rule, commit)
        results = append(results, result)
    }
    
    return buildResult(commit, results)
}

// Pure function helpers
func runRule(rule Rule, commit CommitInfo) RuleResult {
    errors := rule.Validate(commit)
    return RuleResult{
        RuleName: rule.Name(),
        Passed:   len(errors) == 0,
        Errors:   errors,
    }
}

func buildResult(commit CommitInfo, results []RuleResult) ValidationResult {
    passed := true
    for _, result := range results {
        if !result.Passed {
            passed = false
            break
        }
    }
    
    return ValidationResult{
        CommitInfo:  commit,
        RuleResults: results,
        Passed:      passed,
    }
}
```

### Example 2: Functional Command Execution

```go
// Value-based command
type ValidateCommand struct {
    validator     Validator
    reporter      Reporter
    commitService CommitService
}

// Factory function
func NewValidateCommand(
    validator Validator,
    reporter Reporter,
    commitService CommitService,
) ValidateCommand {
    return ValidateCommand{
        validator:     validator,
        reporter:      reporter,
        commitService: commitService,
    }
}

// Command execution with value semantics
func (c ValidateCommand) Execute(hash string) (string, error) {
    // Get commit
    commit, err := c.commitService.GetCommit(hash)
    if err != nil {
        return "", fmt.Errorf("failed to get commit: %w", err)
    }
    
    // Validate
    result := c.validator.Validate(commit)
    
    // Generate report
    report, err := c.reporter.Generate(result)
    if err != nil {
        return "", fmt.Errorf("failed to generate report: %w", err)
    }
    
    return report, nil
}
```

### Example 3: Transformations with Value Semantics

```go
// Immutable configuration
type Config struct {
    Rules    []string
    MaxDepth int
    Verbose  bool
}

// Transformation functions
func (c Config) WithRules(rules []string) Config {
    result := c
    
    // Deep copy the slice
    result.Rules = make([]string, len(rules))
    copy(result.Rules, rules)
    
    return result
}

func (c Config) WithMaxDepth(depth int) Config {
    result := c
    result.MaxDepth = depth
    return result
}

func (c Config) WithVerbose(verbose bool) Config {
    result := c
    result.Verbose = verbose
    return result
}

// Usage
func buildConfig() Config {
    return Config{}.
        WithRules([]string{"rule1", "rule2"}).
        WithMaxDepth(10).
        WithVerbose(true)
}
```

## Benefits of Completion

Completing the functional migration will provide these benefits:

1. **Consistency**: Unified programming model across the codebase
2. **Predictability**: All functions are pure and have no side effects
3. **Testability**: Easier testing with immutable values
4. **Concurrency Safety**: No shared mutable state, safer concurrency
5. **Composability**: Functions can be composed more easily
6. **Reasoning**: Easier to reason about code behavior