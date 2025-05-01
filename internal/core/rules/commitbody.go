// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package rules

import (
	"strings"

	"github.com/itiquette/gommitlint/internal/domain"
	appErrors "github.com/itiquette/gommitlint/internal/errors"
)

// CommitBodyRule validates the commit body structure and content.
// Using functional programming principles with value semantics.
type CommitBodyRule struct {
	baseRule         BaseRule
	requireBody      bool
	allowSignOffOnly bool
	minimumLines     int
}

// CommitBodyOption is a function that configures a CommitBodyRule.
type CommitBodyOption func(CommitBodyRule) CommitBodyRule

// WithRequireBody configures whether a commit body is required.
func WithRequireBody(require bool) CommitBodyOption {
	return func(r CommitBodyRule) CommitBodyRule {
		r.requireBody = require

		return r
	}
}

// WithAllowSignOffOnly configures whether a body with only sign-off information is allowed.
func WithAllowSignOffOnly(allow bool) CommitBodyOption {
	return func(r CommitBodyRule) CommitBodyRule {
		r.allowSignOffOnly = allow

		return r
	}
}

// WithMinimumLines sets the minimum number of lines required in the body.
func WithMinimumLines(lines int) CommitBodyOption {
	return func(r CommitBodyRule) CommitBodyRule {
		r.minimumLines = lines

		return r
	}
}

// NewCommitBodyRule creates a new CommitBodyRule with the specified options.
func NewCommitBodyRule(options ...CommitBodyOption) CommitBodyRule {
	// Create a rule with default settings
	rule := CommitBodyRule{
		baseRule:         NewBaseRule("CommitBody"),
		requireBody:      true,  // By default, require a body
		allowSignOffOnly: false, // By default, don't allow sign-off only
		minimumLines:     1,     // By default, require at least one line
	}

	// Apply options
	for _, option := range options {
		rule = option(rule)
	}

	return rule
}

// Validate validates the commit body.
// It implements the domain.Rule interface.
func (r CommitBodyRule) Validate(commit domain.CommitInfo) []appErrors.ValidationError {
	// Create a new rule with validation results
	updatedRule := r.validateCommit(commit)

	// Return the validation errors from the updated rule
	return updatedRule.baseRule.Errors()
}

// and returns a new rule with validation results.
func (r CommitBodyRule) validateCommit(commit domain.CommitInfo) CommitBodyRule {
	// Start with a clean rule
	result := r
	result.baseRule = r.baseRule.WithClearedErrors().WithRun()

	// Check for proper commit message format with an empty line between subject and body
	if commit.Message != "" {
		lines := strings.Split(commit.Message, "\n")

		// If there's more than one line, check for the empty line separator
		if len(lines) > 1 {
			// Check if the second line is empty (if there's a second line)
			if len(lines) >= 2 && strings.TrimSpace(lines[1]) != "" {
				// Missing empty line between subject and body
				result.baseRule = result.baseRule.WithErrorWithCode(
					appErrors.ErrInvalidBody,
					"commit message is missing empty line between subject and body",
				)
			}
		}
	}

	// Parse body from full message if it isn't already set
	_, body := commit.Subject, commit.Body
	if body == "" && commit.Message != "" {
		_, body = domain.SplitCommitMessage(commit.Message)
	}

	// Check if a body is required
	if r.requireBody && body == "" {
		// Check if it's a merge commit (these often don't have bodies)
		if commit.IsMergeCommit {
			// We'll allow merge commits without bodies
			return result
		}

		// Non-merge commits must have a body if required
		result.baseRule = result.baseRule.WithErrorWithCode(
			appErrors.ErrInvalidBody,
			"commit message body is required",
		)

		return result
	}

	// If body is empty but not required, that's fine
	if body == "" {
		return result
	}

	// Check if the body is just sign-off lines and whether that's allowed
	if IsSignOffOnly(body) && !r.allowSignOffOnly {
		result.baseRule = result.baseRule.WithErrorWithCode(
			appErrors.ErrInvalidBody,
			"commit message body contains only sign-off information, substantive content is required",
		)
	}

	// Check minimum lines
	lines := CountBodyLines(body)
	if lines < r.minimumLines {
		result.baseRule = result.baseRule.WithErrorWithFormatf(
			appErrors.ErrInvalidBody,
			"commit message body has insufficient lines (has %d, needs %d)",
			lines, r.minimumLines,
		)
	}

	return result
}

// Name returns the rule name.
func (r CommitBodyRule) Name() string {
	return r.baseRule.Name()
}

// Errors returns all validation errors found by this rule.
func (r CommitBodyRule) Errors() []appErrors.ValidationError {
	return r.baseRule.Errors()
}

// HasErrors returns true if the rule has found any errors.
func (r CommitBodyRule) HasErrors() bool {
	return r.baseRule.HasErrors()
}

// Result returns a concise validation result.
func (r CommitBodyRule) Result() string {
	if r.HasErrors() {
		return "Invalid commit body"
	}

	return "Valid commit body"
}

// VerboseResult returns a more detailed explanation for verbose mode.
func (r CommitBodyRule) VerboseResult() string {
	if r.HasErrors() {
		return "Commit message body has formatting issues - check for missing content, sign-off requirements, or insufficient detail"
	}

	return "Commit message body format is valid and meets all requirements"
}

// Help returns guidance for fixing rule violations.
func (r CommitBodyRule) Help() string {
	if !r.HasErrors() {
		return "No errors to fix in the commit message body"
	}

	return "Your commit message body has formatting issues.\n" +
		"Ensure there's an empty line between subject and body, and include meaningful content."
}

// IsSignOffOnly checks if a body contains only sign-off lines.
func IsSignOffOnly(body string) bool {
	// If body is empty, it doesn't contain only sign-off lines
	if strings.TrimSpace(body) == "" {
		return false
	}

	lines := strings.Split(body, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		// Skip empty lines
		if trimmed == "" {
			continue
		}

		// If any non-empty line is not a sign-off line, return false
		if !isSignOffLine(trimmed) {
			return false
		}
	}

	// All non-empty lines are sign-off lines
	return true
}

// CountBodyLines counts the number of non-empty lines in a body.
func CountBodyLines(body string) int {
	// If body is empty, return 0
	if strings.TrimSpace(body) == "" {
		return 0
	}

	lines := strings.Split(body, "\n")
	count := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			count++
		}
	}

	return count
}

// isSignOffLine checks if a line is a sign-off line.
func isSignOffLine(line string) bool {
	// Line should already be trimmed by the caller
	prefixes := []string{
		"Signed-off-by:",
		"Co-authored-by:",
		"Reviewed-by:",
		"Tested-by:",
		"Acked-by:",
		"Cc:",
		"Reported-by:",
	}

	for _, prefix := range prefixes {
		if strings.HasPrefix(line, prefix) {
			return true
		}
	}

	return false
}
