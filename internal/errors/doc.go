// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

/*
Package errors provides a comprehensive error handling system for Gommitlint
with rich context and functional programming principles.

The package structure consists of:
  - errors.go: Core types, error codes, and base functions
  - formatting.go: Formatting methods and helper functions
  - formatter.go: Output formatters for different formats

# Core Components

 1. ValidationError:
    A structured error type that includes rule name, code, message, and context.
    The errors follow immutable design principles - operations return new errors
    rather than modifying existing ones.

 2. Context System:
    Errors can include rich context information using a key-value map.
    All context is added through the WithContext method for consistency.

 3. Help Messages:
    Errors include a dedicated Help field for guidance on fixing issues.
    Help text is added using WithHelp or in error creation functions.

 4. Formatters:
    Multiple formatters for different output formats (text, JSON, Markdown).
    Each formatter is a concrete type for simplicity.

 5. Immutable Operations:
    All operations on errors follow functional programming principles and
    return new instances rather than modifying existing ones.

# Basic Usage

Creating errors:

	// Create a basic error
	baseErr := errors.New("RuleName", errors.ErrSubjectTooLong, "Subject too long")

	// Add help text
	errWithHelp := baseErr.WithHelp("Keep subject under 50 chars")

	// Add context
	enhancedErr := errWithHelp.WithContext("subject_length", "60")
		.WithContext("max_length", "50")
		.WithContext("commit_sha", "abc123")

	// Get help message
	helpText := enhancedErr.GetHelp() // Returns the help field directly

	// Format as text
	textOutput := enhancedErr.FormatAsText(true) // verbose

Using formatters:

	// Create formatters
	textFormatter := errors.NewTextFormatter(true) // verbose
	jsonFormatter := errors.NewJSONFormatter()
	mdFormatter := errors.NewMarkdownFormatter()

	// Format errors
	textOutput := textFormatter.FormatError(err)
	jsonOutput := jsonFormatter.FormatError(err)
	mdOutput := mdFormatter.FormatError(err)

# Advanced Features

 1. Multiple Output Formats:
    Supports text (with verbose and concise modes), JSON, and Markdown.

 2. Code-Based Error Types:
    Comprehensive set of error codes organized by category.

 3. NewFormatValidationError Helper:
    A convenience function that creates errors with common format validation patterns.

# Functional Programming Approach

This package follows functional programming principles with value semantics:

1. Immutability: All operations return new errors instead of modifying existing ones
2. Pure functions: Operations have no side effects
3. Value semantics: Errors are passed by value, not by reference
4. Method chaining: Operations can be chained for clean, readable code

Example

	rule := SubjectLengthRule{MaxLength: 50}
	commit := CommitInfo{
	    Hash:    "abc123",
	    Subject: "This is a very long subject line that exceeds the maximum length",
	}

	// Create an error with all context
	subjectLength := len(commit.Subject)
	errorMessage := fmt.Sprintf("Subject length (%d) exceeds maximum allowed (%d)",
	    subjectLength, rule.MaxLength)

	err := errors.New(rule.Name(), errors.ErrSubjectTooLong, errorMessage)
	err = err.WithHelp(rule.Help())
	err = err.WithContext("subject_length", fmt.Sprintf("%d", subjectLength))
	    .WithContext("max_length", fmt.Sprintf("%d", rule.MaxLength))
	    .WithContext("subject", commit.Subject)
	    .WithContext("commit_sha", commit.Hash)
*/
package errors
