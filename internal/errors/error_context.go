// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package errors

// ValidationErrorContext is a container for collecting context information.
type ValidationErrorContext struct {
	// Common fields used across multiple rules
	CommitHash    string
	CommitMessage string
	SubjectLine   string
	BodyText      string
}

// NewContext creates a new validation error context.
func NewContext() ValidationErrorContext {
	return ValidationErrorContext{}
}

// WithCommit adds commit information to the context and returns a new context.
func (c ValidationErrorContext) WithCommit(commitHash, message, subject, body string) ValidationErrorContext {
	result := c // Create a copy
	result.CommitHash = commitHash
	result.CommitMessage = message
	result.SubjectLine = subject
	result.BodyText = body

	return result
}

// CreateRichError creates an enhanced validation error from a context.
func CreateRichError(rule string, code ValidationErrorCode, message string, help string, ctx ValidationErrorContext) ValidationError {
	err := New(rule, code, message)

	// Add help
	err = err.WithContext("help", help)

	// Add available context
	if ctx.CommitHash != "" {
		err = err.WithContext("commit_hash", ctx.CommitHash)
		err = err.WithCommitSHA(ctx.CommitHash)
	}

	if ctx.SubjectLine != "" {
		err = err.WithContext("subject", ctx.SubjectLine)
	}

	if ctx.BodyText != "" {
		err = err.WithContext("body", ctx.BodyText)
	}

	return err
}
