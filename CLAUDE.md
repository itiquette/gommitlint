# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Functional Hexagonal Architecture

Gommitlint follows a **functional hexagonal architecture** with pure functions and explicit dependencies:

### Core Architecture Principles

1. **Pure Functions**: All validation logic is implemented as pure functions
2. **Explicit Dependencies**: Dependencies are passed as function parameters, never hidden
3. **Domain Independence**: Domain logic only depends on interfaces it defines
4. **Adapter Implementation**: Adapters implement domain interfaces

### Architecture Structure

```go
// Domain defines business logic and interfaces
func ValidateCommit(commit Commit, rules []Rule, repo Repository, cfg *config.Config) ValidationResult

// Domain defines required interfaces (ports)
type Repository interface {
    GetCommit(ctx context.Context, ref string) (Commit, error)
    // ... other methods
}

// Adapters implement domain interfaces
type git.Repository struct { /* implementation */ }
var _ domain.Repository = (*git.Repository)(nil)

// Pure functional composition
rules := rules.CreateEnabledRules(config)
result := domain.ValidateCommit(commit, rules, repo, config)
```

### Benefits of This Approach

- **Simple**: No complex service objects or dependency injection frameworks
- **Testable**: Easy to test with mock implementations
- **Functional**: Pure functions with immutable data
- **Clean**: Clear separation between domain and infrastructure

## Context Key Architecture

Gommitlint uses context only for cross-cutting concerns:

- **Context keys** (`internal/adapters/cli/contextkeys.go`): Used only for logging and CLI options
- **No configuration in context**: Configuration flows through explicit parameters

This design follows hexagonal architecture principles by:

- Making all dependencies explicit
- Maintaining clear boundaries between architectural layers
- Avoiding the anti-pattern of storing configuration in context

> **IMPORTANT**: Configuration must never be stored in or accessed from context. It should always be passed as an explicit parameter.

## Rule Priority System

Gommitlint uses a three-level priority system for rule activation:

### Priority Levels (highest to lowest)

1. **Explicitly enabled**: Rules in the `enabled` list always run
2. **Explicitly disabled**: Rules in the `disabled` list are skipped
3. **Default behavior**: Most rules are enabled by default

### How It Works

```yaml
gommitlint:
  rules:
    enabled:
      - spell           # Overrides default-disabled
      - subjectlength   # Explicitly enabled
    disabled:
      - commitbody      # Always disabled (unless also in enabled)
```

### Resolution Logic

```text
if rule in enabled:
    ✓ include rule (highest priority)
else if rule in disabled:
    ✗ exclude rule
else if rule in DefaultDisabledRules:
    ✗ exclude rule  
else:
    ✓ include rule (default enabled)
```

### Default-Disabled Rules

Only three rules are disabled by default:

- `jirareference` - Validates JIRA ticket references (organization-specific)
- `commitbody` - Requires commit body (not all projects need detailed bodies)
- `spell` - Spell checking (requires additional setup)

All other rules are enabled by default unless explicitly disabled.

## Build Commands

- Build: `make build/plain`
- Test all: `make test`
- Single test: `go test -v -count=1 ./internal/path/to/package/file_test.go -run TestSpecificFunction`
- Lint: `make quality/golangcilint`
- Format: `make quality/tidy`
- Quality checks: `make quality`

### Keep Conversations Focused

- Ask specific questions with clear objectives
- Break complex tasks into smaller, well-defined requests
- Start new conversations for unrelated topics

### Leverage Command Line Tools

- Use built-in tools like `ls`, `grep`, `find`, and `cat` for file operations
- Ask Claude to suggest relevant shell commands for tasks instead of custom scripts
- Combine simple commands with pipes (|) for more complex operations

### Avoid Overengineering

- Request simpler solutions first, then expand if needed
- Prefer existing tools over custom implementations
- Ask for small shell scripts instead of complex programs when appropriate

### Code Quality

- Always provide solutions that are simple, coherent, concise, and idiomatic
- Fix and update tests alongside code changes
- Prioritize readability and maintainability over clever solutions

## Code Style

- Follow Go standard formatting (`go fmt`)
- Imports: stdlib first, external next, internal last (with blank lines between groups)
- Error handling:
  - Use `fmt.Errorf("context: %w", err)` pattern
  - Validation errors: `domain.New("RuleName", "error_code", "message").WithHelp("help text")`
  - Build errors using domain error builder pattern
- Naming: PascalCase for exported, camelCase for non-exported identifiers
- Test-only code:
  - Test utilities are organized by domain within their respective packages
  - Integration tests should be in the `internal/integrationtest` package

### Working with Code

- Specify exact file paths within your project directory
- Always mention file extensions (.js, .py, etc.)
- Request relative paths only (no git operations)

### File Management

- When sharing files, include only what's necessary
- Use tools like `head`, `tail`, and `wc` to summarize file contents
- Let Claude know if output should be redirected to files (> or >>)

### Project Understanding

- Always read docs/ARCHITECTURE.md on startup to understand the project structure and principles
- Reference project documentation when discussing implementation
- Never perform git commit, add, or push operations unless explicitly requested by the user
- Apply deep thinking to all changes, focus on the specific task at hand
- Don't overengineer solutions - keep them simple, clear, and aligned with the project's architecture
- Ensure all contributions are coherent, concise, and follow functional programming principles where possible
- Strive for readability and maintainability in all code changes

### Debugging Help

- Include specific error messages
- Use tools like `grep` to isolate relevant log sections
- Share command output with context

### Error Handling

- Request appropriate error handling strategies based on context
- Specify how verbose error messages should be
- Consider logging needs for production vs. debugging

### Documentation

- Request updates to README or other docs when implementing new features
- Ensure doc.go files are present for all packages
- Follow godoc conventions for all public APIs

## Core Guidelines for Claude

When writing or reviewing Go code, Claude should prioritize functional programming principles with value semantics. This means:

1. **Always prefer value semantics over pointer semantics**
   - Use value receivers for methods, not pointer receivers
   - Pass arguments as values, not pointers
   - Return new values rather than modifying existing ones

2. **Treat data as immutable**
   - Never modify input parameters
   - Create and return new data structures with desired changes
   - Use copy operations when transforming slices or maps

3. **Write pure functions whenever possible**
   - Given the same input, always return the same output
   - Avoid side effects (no I/O, no global state modification)
   - Make dependencies explicit via parameters

4. **Use function values as first-class citizens**
   - Pass functions as arguments
   - Return functions from other functions
   - Store functions in variables and data structures

5. **Compose functionality from smaller functions**
   - Build complex operations by combining simpler functions
   - Use the functional options pattern with value semantics

## Preferred Patterns

### ✅ DO: Immutable Transformations

```go
// Return new structs instead of modifying
func WithName(user User, name string) User {
    result := user // Create a copy
    result.Name = name
    return result
}

// For slices, return new slices
func AppendItem(items []Item, newItem Item) []Item {
    result := make([]Item, len(items), len(items)+1)
    copy(result, items)
    return append(result, newItem)
}
```

### ✅ DO: Value-Based Methods

```go
// Use value receivers and return modified copies
func (c Config) WithTimeout(timeout time.Duration) Config {
    c.Timeout = timeout
    return c
}

// Usage creates a new value
newConfig := config.WithTimeout(5 * time.Second)
```

### ✅ DO: Value-Based Functional Options

```go
type Option func(Config) Config

func WithRetries(count int) Option {
    return func(c Config) Config {
        c.Retries = count
        return c
    }
}

func New(opts ...Option) Config {
    config := NewDefaultConfig()
    for _, opt := range opts {
        config = opt(config)
    }
    return config
}
```

### ✅ DO: Higher-Order Functions

```go
func Map[T, U any](items []T, fn func(T) U) []U {
    result := make([]U, len(items))
    for i, item := range items {
        result[i] = fn(item)
    }
    return result
}

func Filter[T any](items []T, predicate func(T) bool) []T {
    var result []T
    for _, item := range items {
        if predicate(item) {
            result = append(result, item)
        }
    }
    return result
}
```

## Anti-Patterns to Avoid

### ❌ AVOID: Pointer Receivers

```go
// AVOID: Modifying state with pointer receivers
func (u *User) SetName(name string) {
    u.Name = name
}

// PREFER: Value receivers that return new instances
func (u User) WithName(name string) User {
    u.Name = name
    return u
}
```

### ❌ AVOID: Mutating Arguments

```go
// AVOID: Modifying a slice in-place
func AddItem(items []Item, item Item) {
    items = append(items, item) // Modifies the original
}

// PREFER: Returning a new slice
func AddItem(items []Item, item Item) []Item {
    newItems := make([]Item, len(items), len(items)+1)
    copy(newItems, items)
    return append(newItems, item)
}
```

### ❌ AVOID: Side Effects

```go
// AVOID: Functions with side effects
func ProcessOrder(order Order) {
    database.Save(order) // Side effect
    emailService.Send(order) // Side effect
}

// PREFER: Pure functions with explicit returns
func ValidateOrder(order Order) (Order, error) {
    if !isValid(order) {
        return Order{}, errors.New("invalid order")
    }
    return order, nil
}
```

### ❌ AVOID: Global State

```go
// AVOID: Using or modifying global state
var globalConfig Config

func UpdateConfig(key string, value string) {
    globalConfig.Settings[key] = value
}

// PREFER: Explicit dependencies
func ProcessWithConfig(data Data, config Config) Result {
    // Use config parameter explicitly
    return process(data, config)
}
```

## Practical Examples

### Example 1: Functional Data Pipeline

```go
func ProcessData(input []int) []string {
    return Map(
        Filter(
            Map(input, func(i int) int { return i * 2 }),
            func(i int) bool { return i > 10 },
        ),
        func(i int) string { return fmt.Sprintf("Value: %d", i) },
    )
}
```

### Example 2: Builder with Value Semantics

```go
type UserBuilder struct {
    user User
}

func NewUserBuilder() UserBuilder {
    return UserBuilder{user: User{}}
}

func (b UserBuilder) WithName(name string) UserBuilder {
    b.user.Name = name
    return b
}

func (b UserBuilder) WithEmail(email string) UserBuilder {
    b.user.Email = email
    return b
}

func (b UserBuilder) Build() User {
    return b.user
}

// Usage
user := NewUserBuilder().
    WithName("John").
    WithEmail("john@example.com").
    Build()
```

### Example 3: Error Handling with Values

```go
type Result[T any] struct {
    Value T
    Error error
}

func (r Result[T]) Then(fn func(T) Result[T]) Result[T] {
    if r.Error != nil {
        return r
    }
    return fn(r.Value)
}

// Usage
result := ValidateName(user).
    Then(ValidateEmail).
    Then(ValidateAge)
```

## Key Benefits to Highlight

- **Reasoning**: Code is easier to reason about when functions don't modify state
- **Testing**: Pure functions are easier to test (no mocks needed for side effects)
- **Concurrency**: Value semantics make concurrent programming safer
- **Composability**: Small, pure functions compose well into larger systems
- **Maintainability**: Immutable data prevents unexpected modifications

## Final Notes for Claude

When discussing or implementing Go code, Claude should:

1. Default to value semantics in all examples
2. Avoid pointer receivers unless explicitly asked for them
3. Structure code around transformations that return new values
4. Emphasize function composition and higher-order functions
5. Suggest refactoring imperative code into functional style when appropriate

## Core Documentation Principles

1. **Document packages, exported types, functions, and constants**
2. **Use complete sentences with proper punctuation**
3. **Start function documentation with the function name**
4. **Be concise but complete**
5. **Be idiomatic**
6. **Include examples for non-obvious usage**

- Request updates to README or other docs when implementing new features

## Package Documentation

Package documentation should provide an overview of the package's purpose and usage patterns.

### Using doc.go

For substantial packages, place package documentation in a dedicated `doc.go` file:

```go
// Package validation provides tools for validating commit messages
// against configurable rule sets.
//
// Validation can be performed against single commits, ranges of commits,
// or commit message files. Each rule validates a specific aspect of the
// commit message, such as subject length, conventional format, etc.
//
// Basic usage:
//
// repo := git.NewRepository(ctx, ".")
// rules := rules.CreateEnabledRules(config)
// commit, _ := repo.GetCommit(ctx, "HEAD")
// result := domain.ValidateCommit(commit, rules, repo, config)
package validation
```

For smaller packages, package documentation can go in any .go file, typically the main file:

```go
// Package errors provides error types and utilities for structured error handling.
package errors
```

## Function Documentation

Document all exported functions. Start with the function name and use a complete sentence:

```go
// NewValidator creates a new validator with the given options.
// It returns an error if any option is invalid.
func NewValidator(opts ...Option) (*Validator, error) {
    // ...
}

// Validate checks a commit message against all active rules.
// An empty result slice indicates the commit is valid.
func Validate(commit CommitInfo) []ValidationError {
    // ...
}
```

## Type Documentation

Document all exported types:

```go
// ValidationError represents a single validation rule failure.
// It includes information about which rule failed and why.
type ValidationError struct {
    // Code identifies the specific error type
    Code string
    
    // Message provides a human-readable error description
    Message string
    
    // Location indicates where in the commit the error occurred
    Location string
}

// Option configures a Validator instance.
type Option func(*Validator) error
```

## Interface Documentation

Document the purpose of interfaces and their expected behavior:

```go
// CommitProvider defines methods for accessing Git commit information.
// Implementations should handle repository access details.
type CommitProvider interface {
    // GetCommit retrieves a commit by its hash.
    GetCommit(ctx context.Context, hash string) (CommitInfo, error)
    
    // GetHeadCommits retrieves the specified number of commits from HEAD.
    GetHeadCommits(ctx context.Context, count int) ([]CommitInfo, error)
}
```

## Constants and Variables

Document groups of related constants or variables:

```go
// Error codes for validation failures.
const (
    // ErrSubjectTooLong indicates the commit subject exceeds the maximum length.
    ErrSubjectTooLong = "subject_too_long"
    
    // ErrMissingBody indicates a required commit body is missing.
    ErrMissingBody = "missing_body"
    
    // ErrInvalidFormat indicates the commit does not follow the required format.
    ErrInvalidFormat = "invalid_format"
)
```

## Examples

Add examples for non-obvious usage patterns using the example naming convention:

```go
// This goes in example_test.go

func Example() {
    validator := validation.New()
    errors := validator.Validate(commit)
    if len(errors) > 0 {
        fmt.Println("Validation failed")
    }
    // Output: Validation failed
}

func ExampleValidator_WithCustomRules() {
    validator := validation.New(
        validation.WithRule(myCustomRule),
        validation.WithMaxLength(100),
    )
    // ...
}
```

## Implementation Comments

For complex internal logic, add explanatory comments:

```go
// Parse the commit message into subject and body.
// According to Git conventions, the first line is the subject,
// followed by an empty line and then the body.
func parseCommitMessage(message string) (subject, body string) {
    parts := strings.SplitN(message, "\n\n", 2)
    subject = strings.TrimSpace(parts[0])
    
    if len(parts) > 1 {
        body = strings.TrimSpace(parts[1])
    }
    
    return subject, body
}
```

## Complex Packages

For complex packages or domain-specific concepts, add explanatory sections in `doc.go`:

```go
// Package validation provides tools for validating commit messages
// against configurable rule sets.
//
// # Validation Rules
//
// Each rule implements the Rule interface and focuses on a specific aspect:
//   - SubjectLength: Enforces maximum length for commit subjects
//   - CommitBody: Requires commit messages to have a non-empty body
//   - JiraReference: Validates that commits reference a JIRA ticket
//
// # Rule Configuration
//
// Rules can be configured via the configuration system:
//
// config := validation.NewRuleConfig()
// config.MaxSubjectLength = 80
// config.RequireBody = true
//
// # Custom Rules
//
// Custom rules can be implemented by satisfying the Rule interface:
//
// type MyRule struct{}
//
// func (r MyRule) Validate(commit CommitInfo) []ValidationError {
//     // Validation logic here
// }
//
package validation
```

## Testing Documentation

Document test helpers and patterns in test files:

```go
// setupTestRepo creates a temporary Git repository for testing.
// It returns the repository path and a cleanup function.
func setupTestRepo(t *testing.T) (string, func()) {
    t.Helper()
    
    // Setup code...
    
    return path, cleanup
}

// TestValidator_ValidateCommit tests the basic validation flow.
// It covers valid commits, invalid commits, and error handling.
func TestValidator_ValidateCommit(t *testing.T) {
    // Test code...
}
```

## Best Practices

1. **Focus on why, not what**: Explain rationale, not obvious implementation details
2. **Document package APIs completely**: Every exported entity deserves documentation
3. **Keep comments up to date**: Outdated comments are worse than no comments
4. **Use godoc-compatible syntax**: Your documentation should render well in godoc
5. **Document limitations and assumptions**: Note edge cases and requirements
6. **Be brief but thorough**: Strike a balance between completeness and brevity

## Documentation Format

The Go community has specific documentation format preferences:

```go
// ValidateCommit validates a single commit against all active rules.
// It returns a ValidationResult containing any rule violations found.
//
// If the commit cannot be found, an error is returned.
// If skipMergeCommits is true, merge commits are considered valid.
//
// Deprecated: Use ValidateWithOptions instead.
func ValidateCommit(hash string, skipMergeCommits bool) (ValidationResult, error) {
    // Implementation...
}
```

Notice the format:

- First line gives the overview
- Blank line separates sections
- Parameters and returns are documented when needed
- Deprecation notices use the standard "Deprecated:" prefix
- Sections like "Example:", "Note:", or "Bug:" can be used for special cases

## What Not to Document

1. **Private functions** (unless complex)
2. **Obvious implementations**
3. **Implementation details that might change**
4. **Information better expressed in the code itself**

## Go Tool Documentation

Use the following tools to verify your documentation:

```bash
# Check documentation coverage
go doc -all .

# Run a local godoc server
godoc -http=:6060
```

## Checking Documentation on Interfaces

Remember to document interface fulfillment when non-obvious:

```go
// Ensure GitRepository implements CommitProvider.
var _ CommitProvider = (*GitRepository)(nil)
```

This pattern both documents and verifies that the type implements the interface.

## Package Organization

For larger projects, organize related packages in a structure that aids discoverability:

```txt
/pkg
  /validation      # Core validation components
    doc.go         # Package overview documentation
    engine.go      # Core validation engine
    rules.go       # Rule interfaces
  /validation/rules # Concrete rule implementations
    doc.go         # Subpackage documentation  
    subject.go     # Subject rules
    body.go        # Body rules
```

Each level should have appropriate documentation explaining its purpose in the overall system.

## Core Testing Principles

1. **Table-driven tests** are the preferred pattern
2. **One test file per source file** (`foo.go` → `foo_test.go`)
3. **Use testify/require** for clean assertions
4. **Test both happy paths and error cases**
5. **Aim for high test coverage** (>80% as a baseline)

## Test Organization

Every source file should have a corresponding test file:

```txt
validation.go → validation_test.go
rule.go → rule_test.go
engine.go → engine_test.go
```

Package tests should use the same package name with `_test` suffix:

```go
// source file
package validation

// test file
package validation_test
```

Using `package mypackage_test` encourages testing the public API as clients would use it.

## Table-Driven Tests

Structure tests as data tables to test multiple scenarios efficiently. Always use `testCase` (not `tt`) as the variable name for table entries:

```go
func TestValidateCommit(t *testing.T) {
    tests := []struct {
        name           string
        commitMessage  string
        wantErrCode    string
        wantErrMessage string
    }{
        {
            name:           "Valid commit",
            commitMessage:  "Add new feature",
            wantErrCode:    "",
            wantErrMessage: "",
        },
        {
            name:           "Empty commit",
            commitMessage:  "",
            wantErrCode:    "empty_commit",
            wantErrMessage: "commit message cannot be empty",
        },
        {
            name:           "Too long subject",
            commitMessage:  "This subject line is way too long and exceeds the maximum length allowed for a commit message subject line according to our rules",
            wantErrCode:    "subject_too_long",
            wantErrMessage: "subject exceeds maximum length of 50 characters",
        },
    }

    for _, testCase := range tests {
        t.Run(testCase.name, func(t *testing.T) {
            validator := NewValidator()
            err := validator.Validate(testCase.commitMessage)
            
            if testCase.wantErrCode == "" {
                require.NoError(t, err)
            } else {
                require.Error(t, err)
                validationErr, ok := err.(*ValidationError)
                require.True(t, ok, "expected ValidationError")
                require.Equal(t, testCase.wantErrCode, validationErr.Code)
                require.Contains(t, validationErr.Message, testCase.wantErrMessage)
            }
        })
    }
}
```

## Using testify/require

Use require instead of assert for immediate test failures:

```go
import (
    "testing"
    "github.com/stretchr/testify/require"
)

func TestSomething(t *testing.T) {
    result, err := Something()
    
    require.NoError(t, err)
    require.NotNil(t, result)
    require.Equal(t, expected, result.Value)
    require.Contains(t, result.Message, "expected substring")
    require.True(t, result.IsValid)
}
```

Common require functions:

- `require.NoError(t, err)` - Assert no error occurred
- `require.Error(t, err)` - Assert an error occurred
- `require.Equal(t, expected, actual)` - Assert values are equal
- `require.NotEqual(t, notExpected, actual)` - Assert values are not equal
- `require.True(t, condition)` - Assert condition is true
- `require.False(t, condition)` - Assert condition is false
- `require.Nil(t, value)` - Assert value is nil
- `require.NotNil(t, value)` - Assert value is not nil
- `require.Contains(t, string/slice/map, element)` - Assert contains element
- `require.Len(t, collection, length)` - Assert collection has specific length

## Test Helper Functions

Create helper functions for common test setup and mark them with `t.Helper()`:

```go
func setupValidator(t *testing.T, rules ...Rule) *Validator {
    t.Helper()
    
    config := NewDefaultConfig()
    validator, err := NewValidator(config)
    require.NoError(t, err)
    
    for _, rule := range rules {
        validator.AddRule(rule)
    }
    
    return validator
}

func createMockCommit(t *testing.T, message string) *Commit {
    t.Helper()
    
    return &Commit{
        Hash:    "abc123",
        Message: message,
        Author:  "Test User <test@example.com>",
    }
}
```

## Testing Errors

Test both success and error paths:

```go
func TestParseCommitMessage(t *testing.T) {
    // Happy path
    t.Run("valid message", func(t *testing.T) {
        subject, body, err := ParseCommitMessage("Add feature\n\nThis is the body")
        require.NoError(t, err)
        require.Equal(t, "Add feature", subject)
        require.Equal(t, "This is the body", body)
    })
    
    // Error paths
    t.Run("empty message", func(t *testing.T) {
        _, _, err := ParseCommitMessage("")
        require.Error(t, err)
        require.Equal(t, "empty commit message", err.Error())
    })
}
```

## Mock Implementations

Create mock implementations for interfaces:

```go
type MockCommitProvider struct {
    commits map[string]*Commit
}

func NewMockCommitProvider(commits map[string]*Commit) *MockCommitProvider {
    return &MockCommitProvider{commits: commits}
}

func (m *MockCommitProvider) GetCommit(hash string) (*Commit, error) {
    commit, ok := m.commits[hash]
    if !ok {
        return nil, fmt.Errorf("commit not found: %s", hash)
    }
    return commit, nil
}

// In tests
func TestValidateWithProvider(t *testing.T) {
    provider := NewMockCommitProvider(map[string]*Commit{
        "abc123": {Hash: "abc123", Message: "Valid commit"},
        "def456": {Hash: "def456", Message: ""},
    })
    
    validator := NewValidator(WithProvider(provider))
    
    tests := []struct {
        name    string
        hash    string
        wantErr bool
    }{
        {
            name:    "valid commit",
            hash:    "abc123",
            wantErr: false,
        },
        {
            name:    "invalid commit",
            hash:    "def456",
            wantErr: true,
        },
    }
    
    for _, testCase := range tests {
        t.Run(testCase.name, func(t *testing.T) {
            err := validator.Validate(testCase.hash)
            if testCase.wantErr {
                require.Error(t, err)
            } else {
                require.NoError(t, err)
            }
        })
    }
}
```

## Testing Exported vs Unexported

Test exported functionality from an external package perspective:

```go
// In package validation_test (not validation)
func TestValidation(t *testing.T) {
    // Test as a package user would
    validator := validation.New()
    result, err := validator.Validate(commit)
}
```

For internal functionality, use the same package:

```go
// In package validation (not validation_test)
func TestInternalFunction(t *testing.T) {
    // Can access unexported functions and fields
    result := parseCommitMessage("subject\n\nbody")
}
```

## Test Coverage

Aim for high coverage, focusing on critical code paths:

```bash
# Run tests with coverage
go test -cover ./...

# Generate coverage profile
go test -coverprofile=coverage.out ./...

# View coverage in browser
go tool cover -html=coverage.out
```

## Performance Tests

Use benchmarks for performance-critical code:

```go
func BenchmarkValidate(b *testing.B) {
    validator := NewValidator()
    commit := &Commit{
        Message: "Add new feature\n\nDetailed description",
    }
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        validator.Validate(commit)
    }
}
```

Run with:
```bash
go test -bench=. -benchmem
```

## Test Examples

Include runnable examples for complex APIs:

```go
func ExampleValidator_Validate() {
    validator := NewValidator()
    err := validator.Validate("Add feature")
    if err != nil {
        fmt.Println("Validation failed")
    } else {
        fmt.Println("Validation passed")
    }
    // Output: Validation passed
}
```

## Test Naming Conventions

- Test functions: `TestFunctionName`
- Subtests: Use clear descriptions of what's being tested
- Test helpers: `setupX`, `createX`, `mockX`
- Test files: `filename_test.go`

## Test Data Organization

Always use a `testdata` directory adjacent to your test files for test fixtures:

```txt
package/
  ├── file.go
  ├── file_test.go
  └── testdata/
      ├── input1.json
      ├── expected1.json
      ├── simple.golden
      └── complex.golden
```

The `testdata` directory is specially recognized by the Go tool and will be excluded from regular package builds.

## Golden File Testing

For testing complex output, use golden files stored in the `testdata` directory:

```go
func TestFormatter(t *testing.T) {
    tests := []struct {
        name  string
        input string
        file  string
    }{
        {
            name:  "simple format",
            input: "Add feature",
            file:  "testdata/simple.golden",
        },
        {
            name:  "complex format",
            input: "feat(api): add new endpoint\n\nThis adds a new endpoint for users",
            file:  "testdata/complex.golden",
        },
    }
    
    for _, testCase := range tests {
        t.Run(testCase.name, func(t *testing.T) {
            formatter := NewFormatter()
            result := formatter.Format(testCase.input)
            
            // Update golden files with -update flag
            if *update {
                err := os.WriteFile(testCase.file, []byte(result), 0644)
                require.NoError(t, err)
            }
            
            // Read golden file
            expected, err := os.ReadFile(testCase.file)
            require.NoError(t, err)
            
            require.Equal(t, string(expected), result)
        })
    }
}
```

## Testing Context-Aware Functions

For functions using context:

```go
func TestWithContext(t *testing.T) {
    // Create cancelable context
    ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
    defer cancel()
    
    result, err := FunctionWithContext(ctx, args)
    require.NoError(t, err)
    
    // Test cancellation
    t.Run("context canceled", func(t *testing.T) {
        ctx, cancel := context.WithCancel(context.Background())
        cancel() // Cancel immediately
        
        _, err := FunctionWithContext(ctx, args)
        require.Error(t, err)
        require.Equal(t, context.Canceled, err)
    })
}
```

## Integration Tests

Mark slow or external integration tests:

```go
// +build integration

package validation_test

func TestIntegrationWithGit(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test in short mode")
    }
    
    // Test with actual git repository
}
```

Run with:
```bash
go test -tags=integration ./...
```

## Testing Best Practices

1. **Test behavior, not implementation details**
2. **Use subtests for better organization**
3. **Keep tests independent and isolated**
4. **Clean up test resources properly**
5. **Test both success and failure cases**
6. **Use realistic test data**
7. **Keep assertions focused and minimal**
8. **Use table-driven tests for all tests with multiple cases**
9. **Always use `testCase` (not `tt`) for table test variables**
10. **Always place test files in a package-adjacent `testdata` directory**
11. **Have a one-to-one mapping between source and test files**
12. **Use `require` over `assert` for immediate test failure**

## Architecture Design Principles

1. **Keep it simple**: Favor straightforward, idiomatic Go over complex abstractions
2. **Hexagonal architecture**: Separate core domain from external concerns
3. **Explicit dependencies**: Pass dependencies explicitly, don't rely on global state
4. **Testable by design**: Structure code for easy testing without complex mocks
5. **Idiomatic Go**: Follow Go conventions rather than forcing patterns from other languages

## Hexagonal Architecture in Go

Organize code in layers with domain logic at the center:

```txt
gommitlint/
├── main.go                  # Application entrypoint
├── internal/                # Private application code
│   ├── domain/              # Core domain models, interfaces, and validation logic
│   │   ├── commit.go        # Commit type and Repository interface
│   │   ├── rules.go         # Rule interfaces
│   │   ├── validation.go    # Validation functions
│   │   ├── config/          # Configuration types
│   │   └── rules/           # Rule implementations
│   ├── adapters/            # External implementations (adapters)
│   │   ├── cli/             # Command-line interface
│   │   ├── config/          # Configuration loading
│   │   ├── git/             # Repository implementation
│   │   ├── logging/         # Logging adapter
│   │   ├── output/          # Formatters and reporting
│   │   └── signing/         # Signature verification
│   └── integrationtest/     # Integration tests
└── docs/                    # Documentation
    └── ARCHITECTURE.md
```

## Keep it Simple

Avoid unnecessary abstractions. Use pure functions instead of complex factories.

## Dependency Injection

Pass dependencies explicitly as function parameters, never use global state.

## Port and Adapter Pattern

Domain defines interfaces (ports). Adapters implement them. Wire together in main.go.

## Functional Composition

Build validation pipelines by composing pure functions. See `domain.ValidateCommit`.

## Domain-Driven Design

Use entities, value objects, and interfaces without ceremony. Keep domain logic pure.

## Interface Guidelines

Create interfaces only when you need abstraction for testing or multiple implementations.

## Testing

Mock interfaces for isolated unit tests. Use integration tests for end-to-end validation.

## Function Options Pattern

Use functional options for configuration. See `rules/factory.go` for examples.

## Composition over Inheritance

Build complex behavior through composition, not inheritance.

## Error Handling

Use `fmt.Errorf("context: %w", err)` for wrapping. Define domain errors in `errors.go`.

## Command-Query Separation

Separate functions that return data from those that modify state.

## Context Usage

Use context for cancellation and logging only. Never store configuration in context.

## Interface Location

Define interfaces where they're consumed (domain), not where they're implemented (adapters).

## Complexity Guidelines

Add architectural complexity only when needed. Start simple, evolve as required.

## Architecture Best Practices

1. **Start simple** - Add complexity only when needed
2. **Be explicit** - No global state or hidden dependencies  
3. **Follow Go idioms** - Don't force patterns from other languages
4. **Design for testing** - Easy to test code is usually well-designed
5. **Stay coherent** - Maintain consistency throughout

## Container Pattern

The project uses an immutable container pattern for dependency injection:

### Design Principles

- **Immutable state**: Container never mutates after creation  
- **Explicit dependencies**: All dependencies passed as parameters
- **No service locator**: No global registry lookups
- **Fresh instances**: Each service gets fresh dependencies

### Implementation
```go
type Container struct {
    logger       log.Logger
    actualConfig configTypes.Config
    factory      *AdapterFactory
}

// Services created with fresh dependencies
func (c Container) ValidationService() *validation.Service {
    return validation.NewService(
        c.actualConfig,
        c.factory.GitRepository(),
        c.createRegistry(),
        c.logger,
    )
}

// Fresh registry for each service
func (c Container) createRegistry() domain.RuleRegistry {
    registry := domain.NewRuleRegistry()
    // Register all rules
    return registry
}
```

### Benefits

- **Thread-safe**: No shared mutable state
- **Testable**: Easy to provide test implementations  
- **Predictable**: No hidden dependencies
- **Functional**: Aligns with functional programming principles