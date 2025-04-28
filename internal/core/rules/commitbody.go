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
	BaseRule
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
		BaseRule:         NewBaseRule("CommitBody"),
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

// NewCommitBodyRuleWithConfig creates a CommitBodyRule using a configuration provider.
func NewCommitBodyRuleWithConfig(config domain.BodyConfigProvider) CommitBodyRule {
	return NewCommitBodyRule(
		WithRequireBody(config.BodyRequired()),
		WithAllowSignOffOnly(config.BodyAllowSignOffOnly()),
	)
}

// addError adds an error to the rule and returns a new rule instance.
func (r CommitBodyRule) addError(err appErrors.ValidationError) CommitBodyRule {
	result := r
	result.BaseRule = r.BaseRule.WithError(err)

	return result
}

// SetErrors sets the validation errors for this rule.
// This method supports value semantics by returning a new instance.
func (r CommitBodyRule) SetErrors(errors []appErrors.ValidationError) CommitBodyRule {
	result := r

	// Update BaseRule
	baseRule := r.BaseRule
	for _, err := range errors {
		baseRule = baseRule.WithError(err)
	}

	result.BaseRule = baseRule

	return result
}

// withRun marks the rule as having been run and returns a new rule instance.
func (r CommitBodyRule) withRun() CommitBodyRule {
	result := r
	result.BaseRule = r.BaseRule.WithRun()

	return result
}

// validateWithState performs validation and returns both errors and updated rule state.
func validateWithState(rule CommitBodyRule, commit domain.CommitInfo) ([]appErrors.ValidationError, CommitBodyRule) {
	// Skip validation if body is not required
	updatedRule := rule.withRun() // Mark as run first

	if !rule.requireBody {
		return []appErrors.ValidationError{}, updatedRule
	}

	// Split message into lines
	lines := strings.Split(commit.Message, "\n")

	// Check minimum structure
	if len(lines) < 3 {
		// Create error context with rich information
		ctx := appErrors.NewContext().WithCommit(
			commit.Hash,    // commit hash
			commit.Message, // full commit message
			commit.Subject, // subject line
			commit.Body,    // body text
		)

		helpMessage := `Missing Body Error: Your commit requires a descriptive body.

A proper commit message should include a subject line followed by a blank line and a detailed body 
that explains the changes. The body provides important context about why the change was made 
and any relevant details that reviewers should know.

✅ CORRECT FORMAT:
<subject line>
<BLANK LINE>
<detailed body text>

Example of a good commit message:

feat: add user authentication

This commit implements user authentication with the following features:
- Email/password login
- Password reset functionality
- Session management
- Rate limiting for failed login attempts

The implementation follows security best practices with password hashing
and protection against common attacks.

❌ INCORRECT FORMAT:
<subject line only>

WHY A BODY MATTERS:
- Helps reviewers understand your changes
- Documents your intentions for future reference
- Makes the git history more useful for troubleshooting
- Provides context that code alone cannot express`

		err := appErrors.CreateRichError(
			rule.Name(),
			appErrors.ErrInvalidBody,
			"commit requires a body, but only subject was provided",
			helpMessage,
			ctx,
		)

		// Add additional context
		err = err.WithContext("lines", "1")

		updatedRule = updatedRule.addError(err)

		return []appErrors.ValidationError{err}, updatedRule
	}

	// Check that the second line is empty (proper separation)
	if len(lines) >= 2 && strings.TrimSpace(lines[1]) != "" {
		// Create error context with rich information
		ctx := appErrors.NewContext().WithCommit(
			commit.Hash,    // commit hash
			commit.Message, // full commit message
			commit.Subject, // subject line
			commit.Body,    // body text
		)

		helpMessage := `Missing Blank Line Error: Your commit is missing an empty line between the subject and body.

Git commit messages must follow a specific format with a blank line separating the subject from 
the body. This format allows tools to properly parse the commit and display it correctly in logs.

✅ CORRECT FORMAT:
<subject line>
<BLANK LINE>
<body text>

Example:
feat: add login functionality

This commit adds a new login screen with email and password fields.
The implementation includes validation and error handling.

❌ INCORRECT FORMAT:
feat: add login functionality
This commit adds a new login screen with email and password fields.
The implementation includes validation and error handling.

WHY THIS MATTERS:
- The blank line is required by Git's commit message format
- Without it, tools may not correctly separate subject from body
- The separation makes commit logs much more readable
- Many Git tools specifically look for this format`

		err := appErrors.CreateRichError(
			rule.Name(),
			appErrors.ErrInvalidBody,
			"missing empty line between subject and body",
			helpMessage,
			ctx,
		)

		// Add additional context
		err = err.WithContext("second_line", lines[1])

		updatedRule = updatedRule.addError(err)

		return []appErrors.ValidationError{err}, updatedRule
	}

	// Check that body is not empty
	bodyLines := lines[2:]

	trimmedBody := strings.TrimSpace(strings.Join(bodyLines, "\n"))
	if trimmedBody == "" {
		// Create error context with rich information
		ctx := appErrors.NewContext().WithCommit(
			commit.Hash,    // commit hash
			commit.Message, // full commit message
			commit.Subject, // subject line
			commit.Body,    // body text
		)

		helpMessage := `Empty Body Error: Your commit has a blank line after the subject but no actual body content.

While you've correctly included a blank line after your subject, you haven't provided any descriptive
text in the body. The body of a commit message should explain what changes were made and why.

✅ CORRECT FORMAT:
<subject line>
<BLANK LINE>
<non-empty body with meaningful content>

Example:
feat: add user profile page

This commit implements the user profile page with the following features:
- Display user information including name, email, and avatar
- Account settings section with editable fields
- Activity history with recent actions
- Responsive design for mobile devices

❌ INCORRECT FORMAT:
feat: add user profile page

[empty body or only whitespace]

HOW TO FIX:
Add descriptive text after the blank line explaining:
- What changes you made in detail
- Why you made these changes
- Any relevant context or references
- Implementation details that may not be obvious from the code`

		err := appErrors.CreateRichError(
			rule.Name(),
			appErrors.ErrInvalidBody,
			"commit has a blank line after subject but no non-empty body",
			helpMessage,
			ctx,
		)

		// Add additional context
		err = err.WithContext("total_lines", "0")

		updatedRule = updatedRule.addError(err)

		return []appErrors.ValidationError{err}, updatedRule
	}

	// If allowSignOffOnly is not enabled, check that body doesn't start with sign-off lines
	firstBodyLine := strings.TrimSpace(bodyLines[0])
	if !rule.allowSignOffOnly && isSignOffLine(firstBodyLine) {
		// Create error context with rich information
		ctx := appErrors.NewContext().WithCommit(
			commit.Hash,    // commit hash
			commit.Message, // full commit message
			commit.Subject, // subject line
			commit.Body,    // body text
		)

		helpMessage := `Sign-Off First Error: Your commit body starts with a sign-off line instead of meaningful content.

Sign-off lines like "Signed-off-by:" should appear at the end of the commit message, after the
descriptive body text. Starting the body with sign-off lines means you're missing actual content
that explains what your changes do and why they were made.

✅ CORRECT FORMAT:
<subject line>
<BLANK LINE>
<descriptive content about the changes>
<BLANK LINE>
<sign-off lines>

Example:
feat: add user registration

This commit implements a secure user registration flow with:
- Email validation
- Password strength requirements
- CAPTCHA protection
- Rate limiting to prevent abuse

Signed-off-by: Developer Name <dev@example.com>

❌ INCORRECT FORMAT:
feat: add user registration

Signed-off-by: Developer Name <dev@example.com>

HOW TO FIX:
1. Add descriptive content before any sign-off lines
2. Explain what changes were made and why
3. Place all sign-off information at the end of the message
4. Separate sign-offs from content with a blank line`

		err := appErrors.CreateRichError(
			rule.Name(),
			appErrors.ErrInvalidBody,
			"body starts with a sign-off line instead of meaningful content",
			helpMessage,
			ctx,
		)

		// Add additional context
		err = err.WithContext("first_line", firstBodyLine)

		updatedRule = updatedRule.addError(err)

		return []appErrors.ValidationError{err}, updatedRule
	}

	// If allowSignOffOnly is not enabled, check that body contains meaningful content, not just sign-offs
	if !rule.allowSignOffOnly {
		hasContent := false

		for _, line := range bodyLines {
			if strings.TrimSpace(line) != "" && !isSignOffLine(line) {
				hasContent = true

				break
			}
		}

		if !hasContent {
			// Create error context with rich information
			ctx := appErrors.NewContext().WithCommit(
				commit.Hash,    // commit hash
				commit.Message, // full commit message
				commit.Subject, // subject line
				commit.Body,    // body text
			)

			helpMessage := `Sign-Off Only Error: Your commit body contains only sign-off lines without any meaningful content.

A commit message body should explain what changes were made and why they were necessary.
Sign-off lines are important for attribution and certification, but they don't replace descriptive content.

✅ CORRECT FORMAT:
<subject line>
<BLANK LINE>
<descriptive content about the changes>
<BLANK LINE>
<sign-off lines>

Example:
fix: resolve memory leak in background processor

This commit fixes a memory leak that occurred when the background processor
was handling large file uploads. The issue was caused by an unclosed resource
that is now properly managed with try-with-resources pattern.

The fix includes:
- Proper resource cleanup in the FileProcessor class
- Additional unit tests to verify memory usage
- Logging improvements to help detect similar issues

Signed-off-by: Developer Name <dev@example.com>

❌ INCORRECT FORMAT:
fix: resolve memory leak in background processor

Signed-off-by: Developer Name <dev@example.com>
Co-authored-by: Another Dev <another@example.com>

HOW TO FIX:
1. Add descriptive content explaining what you changed
2. Explain why the change was necessary
3. Include any relevant details about the implementation
4. Keep the sign-offs at the end of the message`

			err := appErrors.CreateRichError(
				rule.Name(),
				appErrors.ErrInvalidBody,
				"body contains only sign-off lines without meaningful content",
				helpMessage,
				ctx,
			)

			updatedRule = updatedRule.addError(err)

			return []appErrors.ValidationError{err}, updatedRule
		}
	}

	return []appErrors.ValidationError{}, updatedRule
}

// Validate validates the commit body.
// This uses value semantics and returns the errors without modifying the rule's state.
func (r CommitBodyRule) Validate(commit domain.CommitInfo) []appErrors.ValidationError {
	// Use the pure functional approach
	errors, _ := validateWithState(r, commit)

	return errors
}

// VerboseResult returns a more detailed explanation for verbose mode.
func (r CommitBodyRule) VerboseResult() string {
	if r.HasErrors() {
		// Get errors from the BaseRule
		errors := r.BaseRule.Errors()

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
	if r.HasErrors() {
		return "Invalid commit body"
	}

	return "Valid commit body"
}

// Help returns guidance for fixing rule violations.
func (r CommitBodyRule) Help() string {
	if !r.HasErrors() {
		return "No errors to fix. This rule checks for proper commit message structure, including a blank line between subject and body, and appropriate body content length."
	}

	errors := r.Errors()
	if len(errors) > 0 {
		// Get help text from the enhanced error
		helpText := errors[0].GetHelp()
		if helpText != "" {
			return helpText
		}

		// If the enhanced help is not available, fall back to the old help system
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
	return r.BaseRule.Name()
}

// Errors returns all validation errors found by this rule.
func (r CommitBodyRule) Errors() []appErrors.ValidationError {
	return r.BaseRule.Errors()
}

// HasErrors returns true if the rule has found any errors.
func (r CommitBodyRule) HasErrors() bool {
	return r.BaseRule.HasErrors()
}
