// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package rules

import (
	"strings"

	"github.com/itiquette/gommitlint/internal/domain"
	appErrors "github.com/itiquette/gommitlint/internal/errors"
)

// CommitBodyRule represents the configuration for commit body validation.
type CommitBodyRule struct {
	baseRule         *BaseRule
	requireBody      bool
	minLines         int
	allowSignOffOnly bool
}

// CommitBodyOption is a function that configures a CommitBodyRule.
type CommitBodyOption func(CommitBodyRule) CommitBodyRule

// WithRequiredBody configures the rule to require a commit body.
func WithRequiredBody() CommitBodyOption {
	return func(r CommitBodyRule) CommitBodyRule {
		r.requireBody = true

		return r
	}
}

// WithRequireBody configures whether a commit body is required based on a boolean.
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
		r.minLines = lines

		return r
	}
}

// NewCommitBodyRule creates a new CommitBodyRule with the specified options.
func NewCommitBodyRule(options ...CommitBodyOption) CommitBodyRule {
	rule := CommitBodyRule{
		baseRule:         NewBaseRule("CommitBody"),
		requireBody:      false, // Default to not requiring body
		minLines:         1,     // Default to requiring at least 1 line
		allowSignOffOnly: false, // Default to not allowing only sign-off in body
	}

	// Apply options
	for _, option := range options {
		rule = option(rule)
	}

	return rule
}

// Validate validates the commit body.
func (r CommitBodyRule) Validate(commit domain.CommitInfo) []appErrors.ValidationError {
	// Reset errors
	r.baseRule.ClearErrors()

	// Mark the rule as having been run
	r.baseRule.MarkAsRun()

	// Skip validation if body is not required
	if !r.requireBody {
		return r.baseRule.Errors()
	}

	// Split message into lines
	lines := strings.Split(commit.Message, "\n")

	// Check minimum structure
	if len(lines) < 3 {
		r.baseRule.AddErrorWithContext(
			appErrors.ErrInvalidBody,
			"commit requires a body, but only subject was provided",
			map[string]string{
				"lines": "1",
			},
		)

		return r.baseRule.Errors()
	}

	// Check that the second line is empty (proper separation)
	if len(lines) >= 2 && strings.TrimSpace(lines[1]) != "" {
		r.baseRule.AddErrorWithContext(
			appErrors.ErrInvalidBody,
			"missing empty line between subject and body",
			map[string]string{
				"second_line": lines[1],
			},
		)

		return r.baseRule.Errors()
	}

	// Check that body is not empty
	bodyLines := lines[2:]

	trimmedBody := strings.TrimSpace(strings.Join(bodyLines, "\n"))
	if trimmedBody == "" {
		r.baseRule.AddErrorWithContext(
			appErrors.ErrInvalidBody,
			"commit has a blank line after subject but no non-empty body",
			map[string]string{
				"total_lines": "0",
			},
		)

		return r.baseRule.Errors()
	}

	// If allowSignOffOnly is not enabled, check that body doesn't start with sign-off lines
	firstBodyLine := strings.TrimSpace(bodyLines[0])
	if !r.allowSignOffOnly && isSignOffLine(firstBodyLine) {
		r.baseRule.AddErrorWithContext(
			appErrors.ErrInvalidBody,
			"body starts with a sign-off line instead of meaningful content",
			map[string]string{
				"first_line": firstBodyLine,
			},
		)

		return r.baseRule.Errors()
	}

	// If allowSignOffOnly is not enabled, check that body contains meaningful content, not just sign-offs
	if !r.allowSignOffOnly {
		hasContent := false

		for _, line := range bodyLines {
			if strings.TrimSpace(line) != "" && !isSignOffLine(line) {
				hasContent = true

				break
			}
		}

		if !hasContent {
			r.baseRule.AddErrorWithContext(
				appErrors.ErrInvalidBody,
				"body contains only sign-off lines without meaningful content",
				nil,
			)

			return r.baseRule.Errors()
		}
	}

	return r.baseRule.Errors()
}

// VerboseResult returns a more detailed explanation for verbose mode.
func (r CommitBodyRule) VerboseResult() string {
	if r.baseRule.HasErrors() {
		// Get errors from the BaseRule
		errors := r.baseRule.Errors()

		// All errors use ErrInvalidBody code, so differentiate by message content
		if len(errors) > 0 {
			msg := errors[0].Message
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
			// Default case
			return msg
		}
	}

	return "Commit body follows proper format with subject, blank line separator, and meaningful content"
}

// Result returns a concise validation result.
func (r CommitBodyRule) Result() string {
	if r.baseRule.HasErrors() {
		return "Invalid commit body"
	}

	return "Valid commit body"
}

// Errors returns the validation errors for the rule.
func (r CommitBodyRule) Errors() []appErrors.ValidationError {
	return r.baseRule.Errors()
}

func (r CommitBodyRule) Help() string {
	if !r.baseRule.HasErrors() {
		return "No errors to fix"
	}

	errors := r.baseRule.Errors()
	if len(errors) > 0 {
		msg := errors[0].Message

		// Provide specific help based on the error message
		if strings.Contains(msg, "requires a body") {
			return `Commit messages should include a descriptive body after the subject line.
A proper commit message format is:
<subject line>
<BLANK LINE>
<body content>
Example:
Fix bug in user authentication
This patch fixes an issue where login attempts with valid 
credentials would fail when the username contained special 
characters. The fix properly escapes special characters
in the username before validation.`
		} else if strings.Contains(msg, "empty line") {
			return `Ensure there is a blank line between the subject and body of your commit message.
A proper commit message format is:
<subject line>
<BLANK LINE>
<body content>
Example:
Fix bug in user authentication
This patch fixes an issue with the authentication module.`
		} else if strings.Contains(msg, "non-empty body") || strings.Contains(msg, "meaningful content") {
			return `Include meaningful content in your commit message body to explain:
- What changes were made and why
- Any important technical details
- Related issues or references
Avoid commits with empty bodies or bodies containing only sign-off information.`
		} else if strings.Contains(msg, "sign-off line") {
			return `Start your commit body with meaningful content that explains the changes.
A proper commit message format is:
<subject line>
<BLANK LINE>
<explanation of changes>
<BLANK LINE>
<Sign-off lines or other footers>
Example:
Fix bug in user authentication
This patch fixes an issue where login attempts would fail.
The root cause was improper validation of special characters.
Signed-off-by: Developer Name <dev@example.com>`
		}
	}

	// Default help
	return `Ensure your commit message has a proper structure with:
1. A concise subject line
2. A blank line
3. A descriptive body explaining what changes were made and why
4. Optional sign-off lines at the end of the message, not the beginning`
}

// isSignOffLine checks if a line is a sign-off line.
func isSignOffLine(line string) bool {
	line = strings.TrimSpace(line)
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

// Name returns the rule name.
func (r CommitBodyRule) Name() string {
	return "CommitBody"
}
