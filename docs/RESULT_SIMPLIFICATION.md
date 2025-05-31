# Result Type Simplification

## Overview

The domain result types have been simplified to focus on their core purpose while maintaining all necessary functionality for validation reporting.

## Changes Made

### 1. RuleResult Simplification

**Before:**
```go
type RuleResult struct {
    RuleID         string
    RuleName       string
    Status         ValidationStatus
    Message        string
    VerboseMessage string
    HelpMessage    string
    Errors         []ValidationError
}
```

**After:**
```go
type RuleResult struct {
    RuleName string
    Status   ValidationStatus
    Errors   []ValidationError
}
```

**Rationale:**
- Removed redundant `RuleID` field (same as RuleName)
- Removed pre-formatted message fields (Message, VerboseMessage, HelpMessage)
- Messages are now generated on-the-fly by formatters using the `FormatResult`, `FormatVerboseResult`, and `FormatHelp` functions

### 2. CommitResult Simplification

**Before:**
```go
type CommitResult struct {
    CommitInfo   CommitInfo
    RuleResults  []RuleResult
    RuleErrorMap map[string][]ValidationError
    Metadata     map[string]string
    Passed       bool
}
```

**After:**
```go
type CommitResult struct {
    CommitInfo  CommitInfo
    RuleResults []RuleResult
    Passed      bool
}
```

**Rationale:**
- Removed redundant `RuleErrorMap` (errors are already in RuleResults)
- Removed unused `Metadata` field
- Kept only essential fields for commit validation results

### 3. ValidationResults Simplification

**Before:**
```go
type ValidationResults struct {
    Results       []CommitResult
    RuleSummary   map[string]int
    TotalCommits  int
    PassedCommits int
    Metadata      map[string]string
}
```

**After:**
```go
type ValidationResults struct {
    Results       []CommitResult
    TotalCommits  int
    PassedCommits int
}
```

**Rationale:**
- Removed `RuleSummary` map in favor of `GetFailedRules()` method
- Removed unused `Metadata` field
- Summary data is calculated on-demand rather than stored

### 4. Method Simplifications

**Before:**
- `WithRuleResult(ruleName string, errs []ValidationError) CommitResult`
- `WithFormattedRuleResult(ruleResult RuleResult) CommitResult`
- `WithResult(result CommitResult) ValidationResults`

**After:**
- `AddRuleResult(ruleName string, errs []ValidationError) CommitResult`
- `AddResult(result CommitResult) ValidationResults`

**Rationale:**
- Simplified method names (Add instead of With)
- Consolidated two rule result methods into one
- Maintained functional immutability pattern

## Benefits

1. **Reduced Complexity**: Fewer fields and methods to maintain
2. **Better Separation of Concerns**: Formatting logic stays in formatters
3. **Less Memory Usage**: No duplicate data storage
4. **Cleaner API**: Simpler method names and fewer redundant methods
5. **Flexibility**: Formatters can generate different messages without changing domain types

## Migration Impact

### Formatters Updated
All formatters have been updated to generate messages on-the-fly:
- `TextFormatter`: Uses `domain.FormatResult/FormatVerboseResult/FormatHelp`
- `JSONFormatter`: Generates messages when transforming to output format
- `GitHubActionsFormatter`: Generates messages based on verbosity settings
- `GitLabCIFormatter`: Generates messages based on verbosity settings

### Validation Service
The validation service now uses the simplified `AddRuleResult` method, removing the need to pre-format messages.

## Functional Programming Alignment

The simplification maintains and enhances functional programming principles:
- All methods return new instances (immutability)
- No side effects or state mutations
- Pure functions for message formatting
- On-demand calculation instead of stored state