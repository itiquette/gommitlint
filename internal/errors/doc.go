// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

/*
Package errors provides a comprehensive error handling system for Gommitlint
with rich context and functional programming principles.

# Core Components

 1. ValidationError:
    A unified error type that includes rule name, code, message, and context.
    The errors follow immutable design principles - operations return new errors
    rather than modifying existing ones.

 2. Context System:
    Errors can include rich context information using a key-value map.
    Common context fields include commit information, rule configuration, etc.

 3. Help Messages:
    Errors include help text with guidance on how to fix issues.

 4. Formatters:
    Multiple formatters for different output formats (text, JSON, Markdown).

 5. Immutable Operations:
    All operations on errors follow functional programming principles and
    return new instances rather than modifying existing ones.

# Basic Usage

Creating errors:

	// Create a basic error
	baseErr := errors.New("RuleName", errors.ErrSubjectTooLong, "Subject too long")

	// Add help text
	withHelp := errors.EnhanceValidationError(baseErr, "Keep subject under 50 chars")

	// Add context
	enhancedErr := withHelp.WithContext("subject_length", "60")
	enhancedErr = enhancedErr.WithContext("max_length", "50")
	enhancedErr = enhancedErr.WithCommitSHA("abc123")

	// Get help message
	helpText := enhancedErr.GetHelp(errors []errors.ValidationError)

	// Format as text
	textOutput := enhancedErr.FormatAsText(true) // verbose

Using context collection:

	// Create a context with commit information
	ctx := errors.NewContext().WithCommit(
	    "abc123",     // commit hash
	    "Full message", // commit message
	    "Subject",    // subject line
	    "Body",       // body text
	)

	// Create an error with the context
	err := errors.CreateRichError(
	    "RuleName",
	    errors.ErrSubjectTooLong,
	    "Subject length exceeds maximum",
	    "Keep subject under 50 characters",
	    ctx,
	)

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

 1. Specialized Context Methods:
    Helper methods like WithCommitSHA for commonly used context values.

 2. Multiple Output Formats:
    Supports text (with verbose and concise modes), JSON, and Markdown.

 3. Code-Based Help:
    Help messages can be generated based on error codes.

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
	err = errors.EnhanceValidationError(err, rule.Help(errors []errors.ValidationError))
	err = err.WithContext("subject_length", fmt.Sprintf("%d", subjectLength))
	    .WithContext("max_length", fmt.Sprintf("%d", rule.MaxLength))
	    .WithContext("subject", commit.Subject)
	    .WithCommitSHA(commit.Hash)
*/
package errors
