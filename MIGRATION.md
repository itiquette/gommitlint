# Context Propagation Migration

This document tracks our progress in implementing the context propagation pattern throughout the codebase to support structured logging and request tracing.

## Approach

We're adding `context.Context` as the first parameter to all significant functions in the codebase, especially those that:
1. Perform operations that might be logged
2. Call other functions that need context
3. Are part of public APIs

The context is initially created in `main.go` with `context.Background()` and passed down through the call stack.

## Migration Progress

### Entry Points

- [x] **main.go**
  - [x] Create initial context (`context.Background()`) *(already done)*
  - [x] Pass context to root command *(already done)*

- [x] **CLI Commands**
  - [x] `ports/cli/root.go`: Update command execution to use context
  - [x] `ports/cli/validate.go`: Propagate context to validation commands
  - [ ] `ports/cli/installhook.go`: Add context to hook installation
  - [ ] `ports/cli/removehook.go`: Add context to hook removal

### Core Domain Services

- [x] **Application Services**
  - [x] `application/validate/service.go`: Update service interface and implementation
  - [x] `application/report/generator.go`: Add context to report generator

- [x] **Domain Service Interfaces**
  - [x] `domain/rule.go`: Update RuleProvider interface and Rule interface with context
  - [x] `domain/output_interfaces.go`: Add context to ReportGenerator and ResultFormatter interfaces

- [x] **Validation Engine**
  - [x] `core/validation/engine.go`: Update validation engine to use context
  - [ ] `core/validation/config.go`: Add context to config methods
  - [x] `core/validation/service.go`: Update RulesManager to use context in GetActiveRules

### Rule Implementations

- [x] **Base Rules**
  - [x] `core/rules/base_rule.go`: Rule interface already accepts context
  - [x] `core/rules/commitbody.go`: Already accepts context to validate method
  - [x] `core/rules/commitsahead.go`: Already accepts context to validate method
  - [x] `core/rules/conventional.go`: Already accepts context to validate method
  - [x] `core/rules/imperative.go`: Already accepts context to validate method
  - [x] `core/rules/jira.go`: Already accepts context to validate method
  - [x] `core/rules/signature.go`: Already accepts context to validate method
  - [x] `core/rules/signedidentity.go`: Already accepts context to validate method
  - [x] `core/rules/signoff.go`: Already accepts context to validate method
  - [x] `core/rules/spell.go`: Already accepts context to validate method
  - [x] `core/rules/subjectcase.go`: Already accepts context to validate method
  - [x] `core/rules/subjectlength.go`: Already accepts context to validate method
  - [x] `core/rules/subjectsuffix.go`: Already accepts context to validate method

### Infrastructure Components

- [x] **Configuration**
  - [x] `config/manager.go`: Add context to LoadConfig and GetValidationConfig
  - [x] `config/validator.go`: Add context parameter (already context-aware)

- [x] **Git Repository**
  - [x] `infrastructure/git/repository.go`: Already implements context in repository methods
  - [x] `infrastructure/git/repository_factory.go`: No context needed in factory methods

- [x] **Output Formatters**
  - [x] `infrastructure/output/formatter_adapter.go`: Add context to format methods
  - [x] `infrastructure/output/github.go`: Add context to formatter
  - [x] `infrastructure/output/gitlab.go`: Add context to formatter
  - [x] `infrastructure/output/json.go`: Add context to formatter
  - [x] `infrastructure/output/text.go`: Add context to formatter

### Error Handling

- [x] **Error Creation**
  - [x] `errors/enhanced_errors.go`: Context not needed for error creation
  - [x] `errors/formatter.go`: Context not needed for error formatters

### Helper and Utility Functions

- [x] **Context Utilities**
  - [x] Add logging utilities to `contextx/contextx.go`
  - [x] Create helper functions for context-based logging

## Integration Tests

- [x] **Update Integration Tests**
  - [x] `integtest/cli_workflow_test.go`: Already uses context where needed
  - [x] `integtest/comprehensive_test.go`: Already uses context where needed
  - [x] `integtest/validation_workflow_test.go`: Already uses context extensively

## Implementation Notes

### Logging Pattern

Example logging pattern to be implemented:

```go
func createProvider(ctx context.Context, opt model.GitProviderClientOption, httpClient *http.Client) (interfaces.GitProvider, error) {
    logger := log.Logger(ctx)
    logger.Trace().Msg("Entering createProvider function")
    defer logger.Trace().Msg("Exiting createProvider function")
    
    // Function implementation
    // ...
    
    logger.Debug().Str("provider", "github").Msg("Provider created successfully")
    return provider, nil
}
```

### Context Utilities

We'll add the following context utilities to support logging:

```go
// Logger retrieves a logger from the context or returns a default logger
func Logger(ctx context.Context) *zerolog.Logger {
    if ctx == nil {
        return defaultLogger
    }
    
    if logger, ok := ctx.Value(LoggerKey).(*zerolog.Logger); ok {
        return logger
    }
    
    return defaultLogger
}

// WithLogger adds a logger to the context
func WithLogger(ctx context.Context, logger *zerolog.Logger) context.Context {
    return context.WithValue(ctx, LoggerKey, logger)
}
```

## Completion Status

- Total functions to modify: 0 (will be updated as we identify them)
- Completed: 0
- Remaining: 0