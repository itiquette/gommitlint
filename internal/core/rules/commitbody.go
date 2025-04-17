// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package rules

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/itiquette/gommitlint/internal/errorx"
)

// Signs off commit messages line.
var commitBodySignOffRegex = regexp.MustCompile(`^Signed-off-by: (.+) <(.+)>$`)

// CommitBodyRule validates the presence and format of a commit message body.
// This rule helps teams maintain high-quality Git history by ensuring commit
// messages have properly formatted bodies that explain the changes in detail.
type CommitBodyRule struct {
	errors           []*domain.ValidationError
	requireBody      bool
	allowSignOffOnly bool
}

// CommitBodyOption is a function that configures a CommitBodyRule.
type CommitBodyOption func(*CommitBodyRule)

// WithRequireBody configures whether the rule requires a commit body.
func WithRequireBody(require bool) CommitBodyOption {
	return func(r *CommitBodyRule) {
		r.requireBody = require
	}
}

// WithAllowSignOffOnly configures whether the rule allows bodies with only sign-offs.
func WithAllowSignOffOnly(allow bool) CommitBodyOption {
	return func(r *CommitBodyRule) {
		r.allowSignOffOnly = allow
	}
}

// NewCommitBodyRule creates a new CommitBodyRule with the given options.
func NewCommitBodyRule(options ...CommitBodyOption) *CommitBodyRule {
	rule := &CommitBodyRule{
		errors:           []*domain.ValidationError{},
		requireBody:      true,  // Default to requiring body
		allowSignOffOnly: false, // Default to not allowing sign-off only
	}

	// Apply options
	for _, option := range options {
		option(rule)
	}

	return rule
}

// Name returns the rule identifier.
func (r *CommitBodyRule) Name() string {
	return "CommitBodyRule"
}

// Validate validates the commit body.
func (r *CommitBodyRule) Validate(commit *domain.CommitInfo) []*domain.ValidationError {
	// Reset errors
	r.errors = []*domain.ValidationError{}

	// Skip validation if body is not required
	if !r.requireBody {
		return r.errors
	}

	// Split message into lines
	lines := strings.Split(commit.Message, "\n")

	// Check minimum structure
	if len(lines) < 3 {
		r.addError(
			domain.ValidationErrorInvalidBody,
			"Commit message requires a body explaining the changes",
			map[string]string{
				"actual_lines": strconv.Itoa(len(lines)),
				"min_lines":    "3",
			},
		)

		return r.errors
	}

	// Check for blank line after subject
	if lines[1] != "" {
		r.addError(
			domain.ValidationErrorInvalidBody,
			"Commit message must have exactly one empty line between the subject and the body",
			map[string]string{
				"found": lines[1],
			},
		)

		return r.errors
	}

	// Check if first body line is a sign-off
	firstBodyLine := strings.TrimSpace(lines[2])
	if firstBodyLine == "" {
		r.addError(
			domain.ValidationErrorInvalidBody,
			"Commit message must have a non-empty body text",
			nil,
		)

		return r.errors
	}

	if commitBodySignOffRegex.MatchString(firstBodyLine) && !r.allowSignOffOnly {
		r.addError(
			domain.ValidationErrorInvalidBody,
			"Commit message body should not start with a sign-off line",
			map[string]string{
				"found": firstBodyLine,
			},
		)

		return r.errors
	}

	// Check for meaningful content beyond sign-off lines
	if !r.allowSignOffOnly {
		hasContent := false

		for _, line := range lines[2:] {
			trimmed := strings.TrimSpace(line)
			if trimmed != "" && !commitBodySignOffRegex.MatchString(trimmed) {
				hasContent = true

				break
			}
		}

		if !hasContent {
			r.addError(
				domain.ValidationErrorInvalidBody,
				"Commit message body is required with meaningful content beyond sign-off lines",
				nil,
			)

			return r.errors
		}
	}

	return r.errors
}

// Result returns a concise string representation of the validation result.
func (r *CommitBodyRule) Result() string {
	if len(r.errors) > 0 {
		// Return a concise error message
		return "Invalid commit message body"
	}

	return "Commit body is valid"
}

// VerboseResult returns a more detailed explanation for verbose mode.
func (r *CommitBodyRule) VerboseResult() string {
	if len(r.errors) > 0 {
		// All errors use ValidationErrorInvalidBody code, so differentiate by message content
		if domain.ValidationErrorCode(r.errors[0].Code) == domain.ValidationErrorInvalidBody {
			msg := r.errors[0].Message
			if strings.Contains(msg, "requires a body") {
				return "Commit message lacks a body. A well-formed commit should have a subject, blank line, and descriptive body."
			} else if strings.Contains(msg, "empty line") {
				return "Missing blank line between subject and body. Commit format must include one empty line after the subject."
			} else if strings.Contains(msg, "non-empty body") {
				return "Commit body is empty. The body must contain meaningful description of changes."
			} else if strings.Contains(msg, "sign-off line") {
				return "Body starts with a sign-off line. Start with content explaining your changes, then add sign-offs at the end."
			} else if strings.Contains(msg, "meaningful content") {
				return "Body contains only sign-off lines. Include actual content explaining the purpose and impact of changes."
			}
		}

		// Default case
		return r.errors[0].Error()
	}

	return "Commit body follows proper format with subject, blank line separator, and meaningful content"
}

// Help returns guidance on how to fix rule violations.
func (r *CommitBodyRule) Help() string {
	if len(r.errors) == 0 {
		return "No errors to fix"
	}

	// Use templated help messages based on the error type
	var helpMessage string

	if domain.ValidationErrorCode(r.errors[0].Code) == domain.ValidationErrorInvalidBody {
		// Extract helpful details from the error message
		details := "format the body correctly"
		if r.errors[0].Message != "" {
			details = r.errors[0].Message
		}

		// Use the error template helper
		helpMessage = errorx.FormatHelp(errorx.ErrInvalidBody, details)
	} else {
		// Generic help for unknown errors
		helpMessage = r.errors[0].Error()
	}

	// Return structured help message
	return `Ensure your commit message follows this structure:
1. Subject line (brief summary)
2. One blank line
3. Body with detailed explanation

` + helpMessage
}

// Errors returns all validation errors found.
func (r *CommitBodyRule) Errors() []*domain.ValidationError {
	return r.errors
}

// addError adds a structured validation error.
func (r *CommitBodyRule) addError(_ domain.ValidationErrorCode, message string, context map[string]string) {
	// Create a validation error with context in one step
	err := domain.NewValidationErrorWithContext(r.Name(), string(domain.ValidationErrorInvalidBody), message, context)
	r.errors = append(r.errors, err)
}
