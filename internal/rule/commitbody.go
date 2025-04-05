// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package rule

import (
	"errors"
	"fmt"
	"strings"
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
type CommitBodyRule struct {
	config CommitBodyConfig
	errors []error
}

// Name returns the rule identifier.
func (c *CommitBodyRule) Name() string {
	return "CommitBodyRule"
}

// Result returns a string representation of the validation result.
func (c *CommitBodyRule) Result() string {
	if len(c.errors) > 0 {
		return c.errors[0].Error()
	}

	return "Commit body is valid"
}

// Errors returns all validation errors found.
func (c *CommitBodyRule) Errors() []error {
	return c.errors
}

// Help returns guidance on how to fix rule violations.
func (c *CommitBodyRule) Help() string {
	if len(c.errors) == 0 {
		return "No errors to fix"
	}

	errMsg := c.errors[0].Error()

	if strings.Contains(errMsg, "requires a body") {
		return `Add a descriptive body to your commit message that explains:
- Why the change was made
- What problem it solves
- Any important implementation details

Separate the subject from the body with one blank line.`
	}

	if strings.Contains(errMsg, "one empty line") || strings.Contains(errMsg, "subject and the body") {
		return `Your commit message should have exactly one blank line between the subject and body.

Example format:
feat: your subject line here

Your body text starts here and can
span multiple lines.`
	}

	if strings.Contains(errMsg, "non-empty body") {
		return `Your commit message body must contain actual content.
Add descriptive text explaining the purpose and impact of your changes.`
	}

	return `Ensure your commit message follows this structure:
1. Subject line (brief summary)
2. One blank line
3. Body with detailed explanation`
}

// addErrorf adds an error to the rule's errors slice.
func (c *CommitBodyRule) addErrorf(format string) {
	c.errors = append(c.errors, fmt.Errorf("%s", format))
}

// AddError adds an error to the rule for testing.
func (c *CommitBodyRule) AddError(message string) {
	c.errors = append(c.errors, errors.New(message))
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
		rule.addErrorf("Commit message requires a body explaining the changes")

		return rule
	}

	// Check for blank line after subject
	if lines[1] != "" {
		rule.addErrorf("Commit message must have exactly one empty line between the subject and the body")

		return rule
	}

	// Check if first body line is a sign-off
	firstBodyLine := strings.TrimSpace(lines[2])
	if firstBodyLine == "" || (SignOffRegex.MatchString(firstBodyLine) && !config.AllowSignOffOnly) {
		rule.addErrorf("Commit message must have a non-empty body text")

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
			rule.addErrorf("Commit message body is required with meaningful content beyond sign-off lines")

			return rule
		}
	}

	return rule
}
