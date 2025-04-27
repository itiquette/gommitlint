# Enhanced Error Context System with Functional Programming Principles

This document outlines a plan to improve error context in validation paths throughout Gommitlint using functional programming principles with value semantics.

## Current Challenges

1. **Inconsistent Error Context**: Some validation paths provide detailed context while others are minimal
2. **Limited Actionability**: Errors often describe what went wrong but not how to fix it
3. **Missing Metadata**: Context about file locations, rule configurations, and inputs often missing
4. **Context Propagation**: Error context is sometimes lost when errors are wrapped or propagated
5. **Inconsistent Structure**: Different error types have different context structures

## Proposed Solution: Immutable Validation Error System

### 1. Create an Immutable Rich Error Type

Define an immutable validation error type with comprehensive context:

```go
// ValidationError represents an immutable validation error with rich context.
type ValidationError struct {
    // Core error information
    Rule    string // Rule that found the error
    Code    string // Error code for programmatic handling
    Message string // Human-readable error message
    
    // Detailed context
    Context map[string]string // Flexible key-value context
    
    // Actionable guidance
    Help     string   // Human-readable help message
    Examples []string // Examples of valid input
    
    // Diagnostic information
    Location ValidationErrorLocation
    
    // Rule configuration
    Config map[string]string // Configuration that led to the error
}

// ValidationErrorLocation contains location information for errors.
type ValidationErrorLocation struct {
    File      string // File that caused the error
    Line      int    // Line number in the file
    CommitSHA string // Related commit SHA
}
```

### 2. Pure Functions for Error Creation and Transformation

Implement pure functions to create and transform errors with value semantics:

```go
// New creates a new ValidationError with the minimum required fields.
func New(rule, code, message string) ValidationError {
    return ValidationError{
        Rule:     rule,
        Code:     code,
        Message:  message,
        Context:  make(map[string]string),
        Config:   make(map[string]string),
        Examples: make([]string, 0),
    }
}

// WithContext returns a new ValidationError with the added context key-value pair.
func (e ValidationError) WithContext(key, value string) ValidationError {
    result := e.copy()
    
    // Create a copy of the context map
    newContext := make(map[string]string, len(result.Context)+1)
    for k, v := range result.Context {
        newContext[k] = v
    }
    newContext[key] = value
    
    result.Context = newContext
    return result
}

// WithHelp returns a new ValidationError with the specified help text.
func (e ValidationError) WithHelp(help string) ValidationError {
    result := e.copy()
    result.Help = help
    return result
}

// WithExample returns a new ValidationError with an additional example.
func (e ValidationError) WithExample(example string) ValidationError {
    result := e.copy()
    
    // Create a copy of the examples slice
    newExamples := make([]string, len(result.Examples), len(result.Examples)+1)
    copy(newExamples, result.Examples)
    newExamples = append(newExamples, example)
    
    result.Examples = newExamples
    return result
}

// WithLocation returns a new ValidationError with the specified location information.
func (e ValidationError) WithLocation(file string, line int, commit string) ValidationError {
    result := e.copy()
    result.Location = ValidationErrorLocation{
        File:      file,
        Line:      line,
        CommitSHA: commit,
    }
    return result
}

// WithConfig returns a new ValidationError with the added configuration key-value pair.
func (e ValidationError) WithConfig(key, value string) ValidationError {
    result := e.copy()
    
    // Create a copy of the config map
    newConfig := make(map[string]string, len(result.Config)+1)
    for k, v := range result.Config {
        newConfig[k] = v
    }
    newConfig[key] = value
    
    result.Config = newConfig
    return result
}

// copy creates a shallow copy of the ValidationError.
// This is a private helper method used internally.
func (e ValidationError) copy() ValidationError {
    return e
}
```

### 3. Error Usage and Propagation Patterns

Define consistent patterns for creating and propagating errors:

```go
// Example of creating a rich error in a rule
func (r SubjectLengthRule) Validate(commit domain.CommitInfo) []errors.ValidationError {
    if len(commit.Subject) > r.maxLength {
        return []errors.ValidationError{
            errors.New(r.Name(), "subject_too_long", fmt.Sprintf("Subject length (%d) exceeds maximum allowed (%d)", len(commit.Subject), r.maxLength)).
                WithContext("subject_length", strconv.Itoa(len(commit.Subject))).
                WithContext("max_length", strconv.Itoa(r.maxLength)).
                WithContext("subject", commit.Subject).
                WithHelp(fmt.Sprintf("Ensure the subject line is at most %d characters long", r.maxLength)).
                WithExample(fmt.Sprintf("%.20s... (%d chars)", commit.Subject, r.maxLength)).
                WithConfig("max_length", strconv.Itoa(r.maxLength)),
        }
    }
    return []errors.ValidationError{}
}

// Example of propagating error context when wrapping errors
func (s *Service) ValidateCommit(ctx context.Context, hash string) (domain.ValidationResult, error) {
    commit, err := s.repository.GetCommit(ctx, hash)
    if err != nil {
        // Create a new error with additional context about the operation
        return domain.ValidationResult{}, errors.New(
            "ValidationService", 
            "repo_access_error",
            fmt.Sprintf("Failed to access commit: %s", err),
        ).
            WithContext("commit_hash", hash).
            WithContext("error_details", err.Error()).
            WithLocation("", 0, hash).
            WithHelp("Check that the repository exists and the commit hash is valid.")
    }
    
    // Continue with validation...
}
```

### 4. Error Formatting for Different Outputs

Implement formatters for different output formats:

```go
// FormatAsText formats the error as text for CLI output.
func (e *ValidationError) FormatAsText(verbose bool) string {
    if !verbose {
        return fmt.Sprintf("[%s] %s", e.Rule, e.Message)
    }
    
    // More detailed output for verbose mode
    var b strings.Builder
    fmt.Fprintf(&b, "Rule:    %s\n", e.Rule)
    fmt.Fprintf(&b, "Code:    %s\n", e.Code)
    fmt.Fprintf(&b, "Message: %s\n", e.Message)
    
    if e.Help != "" {
        fmt.Fprintf(&b, "Help:    %s\n", e.Help)
    }
    
    if len(e.Examples) > 0 {
        fmt.Fprintf(&b, "Example: %s\n", e.Examples[0])
    }
    
    if len(e.Context) > 0 {
        fmt.Fprintf(&b, "Context:\n")
        for k, v := range e.Context {
            fmt.Fprintf(&b, "  %s: %s\n", k, v)
        }
    }
    
    // Add other fields for verbose output...
    
    return b.String()
}

// FormatAsJSON formats the error as JSON.
func (e *ValidationError) FormatAsJSON() ([]byte, error) {
    return json.Marshal(e)
}

// FormatAsMarkdown formats the error as Markdown for documentation.
func (e *ValidationError) FormatAsMarkdown() string {
    var b strings.Builder
    fmt.Fprintf(&b, "### %s: %s\n\n", e.Rule, e.Message)
    fmt.Fprintf(&b, "**Code:** `%s`\n\n", e.Code)
    
    if e.Help != "" {
        fmt.Fprintf(&b, "**Help:** %s\n\n", e.Help)
    }
    
    if len(e.Examples) > 0 {
        fmt.Fprintf(&b, "**Examples:**\n\n")
        for _, example := range e.Examples {
            fmt.Fprintf(&b, "```\n%s\n```\n\n", example)
        }
    }
    
    // Add other fields...
    
    return b.String()
}
```

### 5. Error Context Collection for Rules

Add a mechanism for collecting rule state information:

```go
// ValidationErrorContext is a container for collecting context information.
type ValidationErrorContext struct {
    // Common fields used across multiple rules
    CommitHash       string
    CommitMessage    string
    SubjectLine      string
    BodyText         string
    Configuration    map[string]string
    RelatedCommits   []string
    ValidationPhase  string
    ValidationOptions map[string]string
}

// NewContext creates a new validation error context.
func NewContext() *ValidationErrorContext {
    return &ValidationErrorContext{
        Configuration:    make(map[string]string),
        ValidationOptions: make(map[string]string),
    }
}

// WithCommit adds commit information to the context.
func (c *ValidationErrorContext) WithCommit(commit domain.CommitInfo) *ValidationErrorContext {
    result := *c // Create a copy
    result.CommitHash = commit.Hash
    result.CommitMessage = commit.Message
    result.SubjectLine = commit.Subject
    result.BodyText = commit.Body
    return &result
}

// Use when creating errors
func CreateError(rule domain.Rule, code, message string, ctx *ValidationErrorContext) errors.ValidationError {
    err := errors.New(rule.Name(), code, message)
    
    // Add available context
    if ctx.CommitHash != "" {
        err = err.WithContext("commit_hash", ctx.CommitHash)
    }
    
    if ctx.SubjectLine != "" {
        err = err.WithContext("subject", ctx.SubjectLine)
    }
    
    // Add more context as needed...
    
    return *err
}
```

## Benefits of this Approach

1. **Richer Diagnostics**: Users get more detailed information about what went wrong
2. **Increased Actionability**: Clear guidance on how to fix issues
3. **Consistent Structure**: All errors follow the same pattern for easier processing
4. **Structured Data**: Errors contain machine-readable data for tools
5. **Improved Debugging**: More context makes debugging easier
6. **Better UX**: End users get more helpful error messages
7. **Extensibility**: Easy to add new error types and context fields

## Migration Strategy

We'll implement this new error system incrementally:

### Phase 1: Core Error Types and Helpers

1. Enhance the ValidationError structure with additional context fields
2. Implement helper methods for creating rich errors
3. Create formatters for different output formats
4. Implement context collection helpers

### Phase 2: Update Core Rules

1. Identify high-priority rules for enhanced error context
2. Update one rule at a time with enhanced context
3. Update tests to verify the new error context

### Phase 3: Add Support to Infrastructure

1. Enhance repository access errors with context
2. Improve configuration loading error context
3. Add context to CLI command errors

### Phase 4: Enhance Reporting

1. Update report generators to use rich error context
2. Improve CLI output formatting for errors
3. Enhance JSON output with detailed error context
4. Update documentation with examples of the new error context

### Phase 5: Final Testing and Refinement

1. Gather feedback on error messages
2. Refine error context based on feedback
3. Ensure consistent context across all errors
4. Document all error codes and context fields

## Examples of Enhanced Errors

### Before

```
Error: subject too long (100 > 80)
```

### After

```
Rule:    SubjectLength
Code:    subject_too_long
Message: Subject line exceeds maximum length (100 characters, max is 80)
Context:
  subject_length: 100
  max_length: 80
  subject: This is a very long subject line that exceeds the maximum length and will cause validation to fail
Help:    Shorten the subject line to at most 80 characters
Example: This is a very long sub... (80 chars)
Configuration:
  max_length: 80
Location:
  CommitSHA: abc1234
```

## Implementation Timeline

- **Week 1**: Define enhanced error types and helpers
- **Week 2**: Update high-priority rules with enhanced context
- **Week 3**: Enhance infrastructure and reporting components
- **Week 4**: Testing and refinement