# CLAUDE.md

Guidance for AI assistants working with the gommitlint codebase.

## Architecture Principles

**Read `docs/ARCHITECTURE.md` first** - it contains the complete architectural overview.

Key principles when making changes:

1. **Pure Functions**: All business logic as pure functions with explicit parameters
2. **Value Semantics**: Use value receivers, return new instances instead of modifying
3. **No Hidden Dependencies**: Pass everything as function parameters, never use global state
4. **Hexagonal Boundaries**: Domain defines interfaces, adapters implement them

```go
// ✅ Good: Pure function with explicit dependencies
func ValidateCommit(commit Commit, rules []Rule, repo Repository, cfg Config) ValidationResult

// ❌ Bad: Hidden dependencies 
func ValidateCommit(commit Commit) ValidationResult // Where do rules and config come from?
```

**Critical**: Configuration must never be stored in context - always pass as explicit parameters.

## Rule Configuration

Three-level priority: `enabled` (highest) → `disabled` → `default enabled`

Only 3 rules disabled by default: `jirareference`, `commitbody`, `identity`

## Build Commands

- Test: `make test`
- Lint: `make quality`
- Single test: `go test -v -count=1 ./internal/path/to/package/file_test.go -run TestSpecificFunction`

## Code Style

- Go standard formatting (`go fmt`)
- Imports: stdlib → external → internal (with blank lines)
- Error handling: `fmt.Errorf("context: %w", err)` pattern
- Validation errors: `domain.New("RuleName", "error_code", "message").WithHelp("help text")`
- Testing: Table-driven tests with `testCase` variable, use `testify/require`

## Value Semantics Patterns

### ✅ Prefer: Value receivers that return new instances
```go
func (c Config) WithTimeout(timeout time.Duration) Config {
    c.Timeout = timeout
    return c
}
```

### ✅ Prefer: Pure functions with explicit dependencies  
```go
func ValidateCommit(commit Commit, rules []Rule, repo Repository, cfg Config) ValidationResult
```

### ❌ Avoid: Pointer receivers that mutate state
```go
func (u *User) SetName(name string) { u.Name = name } // Don't do this
```

### ❌ Avoid: Hidden dependencies or global state
```go
func ValidateCommit(commit Commit) ValidationResult // Where do rules come from?
```

## Testing Guidelines

- Table-driven tests with `testCase` variable
- Use `testify/require` for immediate test failure
- Test both success and error paths
- Place test data in adjacent `testdata/` directory
- One test file per source file: `foo.go` → `foo_test.go`

## Documentation Standards

- Document all exported packages, types, functions, constants
- Start function docs with the function name
- Use `doc.go` for substantial package documentation
- Include examples for non-obvious usage
- Follow godoc conventions