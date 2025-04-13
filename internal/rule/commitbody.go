// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package rule

import (
	"strconv"
	"strings"

	"github.com/itiquette/gommitlint/internal/model"
)

// CommitBodyConfig provides configuration for the CommitBodyRule.
type CommitBodyConfig struct {
	// RequireBody indicates whether a commit message must have a body
	RequireBody bool

	// AllowSignOffOnly determines if a message with only sign-off lines is valid
	AllowSignOffOnly bool
}

// DefaultCommitBodyConfig returns the default configuration.
func DefaultCommitBodyConfig() CommitBodyConfig {
	return CommitBodyConfig{
		RequireBody:      true,
		AllowSignOffOnly: false,
	}
}

// CommitBodyRule validates the presence and format of a commit message body.
// This rule helps teams maintain high-quality Git history by ensuring commit
// messages have properly formatted bodies that explain the changes in detail.
//
// The rule enforces that commit messages have a clear structure with a subject line,
// a blank line separator, and a meaningful body. It can be configured to require
// a body and to specify whether sign-off-only messages are acceptable.
//
// This ensures that commit history remains useful for future developers by providing
// context about why changes were made, not just what changes occurred.
//
// Example usage:
//
//	rule := ValidateCommitBody(message, WithRequireBody(true), WithAllowSignOffOnly(false))
//	if len(rule.Errors()) > 0 {
//	    fmt.Println(rule.Help())
//	}
type CommitBodyRule struct {
	config CommitBodyConfig
	errors []*model.ValidationError
}

// Name returns the rule identifier.
func (c *CommitBodyRule) Name() string {
	return "CommitBodyRule"
}

// Result returns a concise string representation of the validation result.
func (c *CommitBodyRule) Result() string {
	if len(c.errors) > 0 {
		// Return a concise error message
		return "Invalid commit message body"
	}

	return "Commit body is valid"
}

// VerboseResult returns a more detailed explanation for verbose mode.
func (c *CommitBodyRule) VerboseResult() string {
	if len(c.errors) > 0 {
		// Return a more detailed error message in verbose mode
		switch c.errors[0].Code {
		case "missing_body":
			return "Commit message lacks a body. A well-formed commit should have a subject, blank line, and descriptive body."
		case "missing_blank_line":
			return "Missing blank line between subject and body. Commit format must include one empty line after the subject."
		case "empty_body":
			return "Commit body is empty. The body must contain meaningful description of changes."
		case "signoff_first_line":
			return "Body starts with a sign-off line. Start with content explaining your changes, then add sign-offs at the end."
		case "only_signoff":
			return "Body contains only sign-off lines. Include actual content explaining the purpose and impact of changes."
		default:
			return c.errors[0].Error()
		}
	}

	return "Commit body follows proper format with subject, blank line separator, and meaningful content"
}

// Errors returns all validation errors found.
func (c *CommitBodyRule) Errors() []*model.ValidationError {
	return c.errors
}

// Help returns guidance on how to fix rule violations.
func (c *CommitBodyRule) Help() string {
	if len(c.errors) == 0 {
		return "No errors to fix"
	}

	// Use the error code for more precise help messages
	switch c.errors[0].Code {
	case "missing_body":
		return `Add a descriptive body to your commit message that explains:
- Why the change was made
- What problem it solves
- Any important implementation details

Separate the subject from the body with one blank line.`

	case "missing_blank_line":
		return `Your commit message should have exactly one blank line between the subject and body.

Example format:
feat: your subject line here

Your body text starts here and can
span multiple lines.`

	case "empty_body":
		return `Your commit message body must contain actual content.
Add descriptive text explaining the purpose and impact of your changes.`

	case "signoff_first_line":
		return `Your commit message body should not start with a sign-off line.
Begin with actual content explaining your changes, then add sign-off lines at the end.`

	case "only_signoff":
		return `Your commit message must include meaningful content beyond just sign-off lines.
Explain the reasons and details of your changes.`

	default:
		return `Ensure your commit message follows this structure:
1. Subject line (brief summary)
2. One blank line
3. Body with detailed explanation`
	}
}

// addError adds a structured validation error.
func (c *CommitBodyRule) addError(code, message string, context map[string]string) {
	err := model.NewValidationError("CommitBodyRule", code, message)

	// Add any context values
	for key, value := range context {
		_ = err.WithContext(key, value)
	}

	c.errors = append(c.errors, err)
}

// CommitBodyOption configures a CommitBodyConfig.
type CommitBodyOption func(*CommitBodyConfig)

// WithRequireBody sets whether a body is required in commit messages.
func WithRequireBody(require bool) CommitBodyOption {
	return func(c *CommitBodyConfig) {
		c.RequireBody = require
	}
}

// WithAllowSignOffOnly sets whether messages with only sign-off lines are valid.
func WithAllowSignOffOnly(allow bool) CommitBodyOption {
	return func(c *CommitBodyConfig) {
		c.AllowSignOffOnly = allow
	}
}

// ValidateCommitBody checks that the commit message has a proper body.
func ValidateCommitBody(message string, opts ...CommitBodyOption) *CommitBodyRule {
	// Apply configuration options
	config := DefaultCommitBodyConfig()
	for _, opt := range opts {
		opt(&config)
	}

	rule := &CommitBodyRule{
		config: config,
	}

	// Skip validation if body is not required
	if !config.RequireBody {
		return rule
	}

	// Split message into lines
	lines := strings.Split(message, "\n")

	// Check minimum structure
	if len(lines) < 3 {
		rule.addError(
			"missing_body",
			"Commit message requires a body explaining the changes",
			map[string]string{
				"actual_lines": strconv.Itoa(len(lines)),
				"min_lines":    "3",
			},
		)

		return rule
	}

	// Check for blank line after subject
	if lines[1] != "" {
		rule.addError(
			"missing_blank_line",
			"Commit message must have exactly one empty line between the subject and the body",
			map[string]string{
				"found": lines[1],
			},
		)

		return rule
	}

	// Check if first body line is a sign-off
	firstBodyLine := strings.TrimSpace(lines[2])
	if firstBodyLine == "" {
		rule.addError(
			"empty_body",
			"Commit message must have a non-empty body text",
			nil,
		)

		return rule
	}

	if SignOffRegex.MatchString(firstBodyLine) && !config.AllowSignOffOnly {
		rule.addError(
			"signoff_first_line",
			"Commit message body should not start with a sign-off line",
			map[string]string{
				"found": firstBodyLine,
			},
		)

		return rule
	}

	// Check for meaningful content beyond sign-off lines
	if !config.AllowSignOffOnly {
		hasContent := false

		for _, line := range lines[2:] {
			trimmed := strings.TrimSpace(line)
			if trimmed != "" && !SignOffRegex.MatchString(trimmed) {
				hasContent = true

				break
			}
		}

		if !hasContent {
			rule.addError(
				"only_signoff",
				"Commit message body is required with meaningful content beyond sign-off lines",
				nil,
			)

			return rule
		}
	}

	return rule
}
