// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package rules

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/itiquette/gommitlint/internal/config"
	"github.com/itiquette/gommitlint/internal/domain"
	appErrors "github.com/itiquette/gommitlint/internal/errors"
	"github.com/itiquette/gommitlint/internal/infrastructure/log"
)

// CommitBodyRule validates the commit body structure.
type CommitBodyRule struct {
	baseRule         BaseRule
	requireBody      bool
	allowSignOffOnly bool
	minimumLines     int
}

// CommitBodyOption configures a CommitBodyRule.
type CommitBodyOption func(CommitBodyRule) CommitBodyRule

// WithRequireBody requires commits to have a non-empty body.
func WithRequireBody(required bool) CommitBodyOption {
	return func(r CommitBodyRule) CommitBodyRule {
		result := r
		result.requireBody = required

		return result
	}
}

// WithAllowSignOffOnly allows commit bodies that only contain sign-off information.
func WithAllowSignOffOnly(allow bool) CommitBodyOption {
	return func(r CommitBodyRule) CommitBodyRule {
		result := r
		result.allowSignOffOnly = allow

		return result
	}
}

// WithMinimumLines sets the minimum number of lines required in the commit body.
func WithMinimumLines(lines int) CommitBodyOption {
	return func(r CommitBodyRule) CommitBodyRule {
		result := r
		result.minimumLines = lines

		return result
	}
}

// NewCommitBodyRule creates a new CommitBodyRule with the given options.
func NewCommitBodyRule(options ...CommitBodyOption) CommitBodyRule {
	// Create rule with default options
	rule := CommitBodyRule{
		baseRule:         NewBaseRule("CommitBody"),
		requireBody:      true,  // Default: body is required
		allowSignOffOnly: false, // Default: sign-off only is not sufficient
		minimumLines:     3,     // Default: body should have at least 3 lines
	}

	// Apply all options
	for _, option := range options {
		rule = option(rule)
	}

	return rule
}

// Name returns the rule name.
func (r CommitBodyRule) Name() string {
	return r.baseRule.Name()
}

// Validate performs validation against a commit.
func (r CommitBodyRule) Validate(ctx context.Context, commit domain.CommitInfo) []appErrors.ValidationError {
	// Get logger from context
	logger := log.Logger(ctx)
	logger.Trace().
		Str("rule", r.Name()).
		Str("commit_hash", commit.Hash).
		Msg("Validating commit body")

	// Create a new rule with context configuration
	rule := r.withContextConfig(ctx)

	// Run validation and return the results
	_, updatedRule := validateBodyWithState(rule, commit)

	// Return any validation errors
	return updatedRule.Errors()
}

// withContextConfig creates a new rule with configuration from context.
func (r CommitBodyRule) withContextConfig(ctx context.Context) CommitBodyRule {
	// Get config from context
	cfg := config.GetConfig(ctx)

	// Get logger from context
	logger := log.Logger(ctx)

	// Create a copy of the rule
	result := r

	// Override settings if specified in context config
	if cfg.Body.Required != result.requireBody {
		result.requireBody = cfg.Body.Required
		logger.Debug().
			Bool("require_body", result.requireBody).
			Msg("Body requirement from context")
	}

	if cfg.Body.AllowSignOffOnly != result.allowSignOffOnly {
		result.allowSignOffOnly = cfg.Body.AllowSignOffOnly
		logger.Debug().
			Bool("allow_sign_off_only", result.allowSignOffOnly).
			Msg("Sign-off only setting from context")
	}

	if cfg.Body.MinimumLines > 0 && cfg.Body.MinimumLines != result.minimumLines {
		result.minimumLines = cfg.Body.MinimumLines
		logger.Debug().
			Int("minimum_lines", result.minimumLines).
			Msg("Minimum lines setting from context")
	}

	return result
}

// Errors returns all validation errors found by this rule.
func (r CommitBodyRule) Errors() []appErrors.ValidationError {
	return r.baseRule.Errors()
}

// HasErrors returns true if the rule has found any errors.
func (r CommitBodyRule) HasErrors() bool {
	return r.baseRule.HasErrors()
}

// RequireBody returns whether a commit body is required.
func (r CommitBodyRule) RequireBody() bool {
	return r.requireBody
}

// AllowSignOffOnly returns whether a commit body can be just a sign-off.
func (r CommitBodyRule) AllowSignOffOnly() bool {
	return r.allowSignOffOnly
}

// MinimumLines returns the minimum number of lines required in the commit body.
func (r CommitBodyRule) MinimumLines() int {
	return r.minimumLines
}

// validateBodyWithState validates the commit body and returns both the errors and an updated rule.
func validateBodyWithState(rule CommitBodyRule, commit domain.CommitInfo) ([]appErrors.ValidationError, CommitBodyRule) {
	result := rule
	result.baseRule = rule.baseRule.WithClearedErrors().WithRun()

	// Check for proper commit message format with an empty line between subject and body
	if commit.Message != "" {
		lines := strings.Split(commit.Message, "\n")

		// If there's more than one line, check for the empty line separator
		if len(lines) > 1 {
			// Check if the second line is empty (if there's a second line)
			if len(lines) >= 2 && strings.TrimSpace(lines[1]) != "" {
				// Missing empty line between subject and body
				helpText := `Missing Empty Line Error: No blank line between subject and body.

Your commit message doesn't have an empty line separating the subject and body.

✅ CORRECT FORMAT:
This is the subject line

This is the body text that should be separated from the subject
with a blank line. This makes the commit message more readable
and allows tools to properly parse the subject line.

❌ INCORRECT FORMAT:
This is the subject line
This is the body without a blank line after the subject.

WHY THIS MATTERS:
- The empty line helps separate the summary from the detailed description
- Many Git tools rely on this format to display commit messages properly
- It improves readability in commit logs and history

NEXT STEPS:
1. Edit your commit message to add a blank line after the subject
2. Use 'git commit --amend' to modify your most recent commit`

				contextMap := map[string]string{
					"message": commit.Message,
				}

				validationErr := appErrors.BodyError(
					result.baseRule.Name(),
					"Commit message is missing empty line between subject and body",
					helpText,
					contextMap)

				result.baseRule = result.baseRule.WithError(validationErr)
			}
		}
	}

	// Parse body from full message if it isn't already set
	body := commit.Body
	if body == "" && commit.Message != "" {
		_, body = domain.SplitCommitMessage(commit.Message)
	}

	// Check if a body is required
	if rule.requireBody && body == "" {
		// Check if it's a merge commit (these often don't have bodies)
		if commit.IsMergeCommit {
			// We'll allow merge commits without bodies
			return result.baseRule.Errors(), result
		}

		// Non-merge commits must have a body if required
		helpText := `Missing Body Error: Commit message body is required.

Your commit is missing a descriptive body section.

✅ CORRECT FORMAT:
This is the subject line

This is the body with a detailed explanation of the changes.
It provides context about what and why the changes were made.

❌ INCORRECT FORMAT:
This is the subject line with no body

WHY THIS MATTERS:
- A descriptive body provides context and explains the reasoning behind changes
- It helps reviewers and future contributors understand your changes
- It documents important decisions for future reference

NEXT STEPS:
1. Edit your commit to add a detailed body explaining:
   - Why the change was needed
   - What approach you took
   - Any important implementation details
   - Any limitations or follow-up work needed
   
2. Use 'git commit --amend' to modify your most recent commit`

		contextMap := map[string]string{
			"commit_sha": commit.Hash,
			"subject":    commit.Subject,
		}

		validationErr := appErrors.BodyError(
			result.baseRule.Name(),
			"Commit message body is required",
			helpText,
			contextMap)

		result.baseRule = result.baseRule.WithError(validationErr)

		return result.baseRule.Errors(), result
	}

	// If body is empty but not required, that's fine
	if body == "" {
		return result.baseRule.Errors(), result
	}

	// Check if the body is just sign-off lines and whether that's allowed
	if IsSignOffOnly(body) && !rule.allowSignOffOnly {
		helpText := `Sign-off Only Error: Commit message body contains only sign-off information.

Your commit body only contains sign-off lines without any substantive content.

✅ CORRECT FORMAT:
This is the subject line

This is the body with a detailed explanation of the changes.
It provides context and reasoning for the changes.

Signed-off-by: Your Name <your.email@example.com>

❌ INCORRECT FORMAT:
This is the subject line

Signed-off-by: Your Name <your.email@example.com>

WHY THIS MATTERS:
- Sign-off lines alone don't explain the purpose or impact of changes
- A descriptive body provides context and reasoning for reviewers
- Future maintainers need to understand why changes were made

NEXT STEPS:
1. Edit your commit to add meaningful content to the body:
   - Explain the purpose of the changes
   - Describe your approach or implementation details
   - Note any relevant context or considerations
   
2. Keep the sign-off line(s) at the end of the body
3. Use 'git commit --amend' to update your commit message`

		contextMap := map[string]string{
			"commit_sha": commit.Hash,
			"subject":    commit.Subject,
			"body":       body,
		}

		validationErr := appErrors.BodyError(
			result.baseRule.Name(),
			"commit message body contains only sign-off information, substantive content is required",
			helpText,
			contextMap)

		result.baseRule = result.baseRule.WithError(validationErr)
	}

	// Check minimum lines
	lines := CountBodyLines(body)
	if rule.minimumLines > 0 && lines < rule.minimumLines {
		helpText := fmt.Sprintf(`Insufficient Body Content Error: Commit message body is too short.

Your commit body doesn't have enough content lines to meet the minimum requirement.

✅ CORRECT FORMAT:
This is the subject line

This is a body with sufficient content.
It should contain enough information to explain the changes.
Including the context, approach, and any important details.

❌ INCORRECT FORMAT:
This is the subject line

Too short.

WHY THIS MATTERS:
- A thorough commit message helps others understand your changes
- Short messages often lack necessary context and details
- Proper documentation in commits reduces the need for external explanations

NEXT STEPS:
1. Edit your commit message to expand the body:
   - Include more detail about what changes were made
   - Explain why the changes were necessary
   - Describe the approach or implementation details
   - Note any limitations or future work
   
2. Aim for at least %d substantive lines in your commit message body
3. Use 'git commit --amend' to update your commit message`, rule.minimumLines)

		contextMap := map[string]string{
			"commit_sha":    commit.Hash,
			"subject":       commit.Subject,
			"body":          body,
			"actual_lines":  strconv.Itoa(lines),
			"minimum_lines": strconv.Itoa(rule.minimumLines),
		}

		validationErr := appErrors.BodyError(
			result.baseRule.Name(),
			fmt.Sprintf("commit message body has insufficient lines (has %d, needs %d)", lines, rule.minimumLines),
			helpText,
			contextMap)

		result.baseRule = result.baseRule.WithError(validationErr)
	}

	return result.baseRule.Errors(), result
}

// Result returns a concise validation result.
func (r CommitBodyRule) Result(errors []appErrors.ValidationError) string {
	if len(errors) > 0 {
		return "❌ Invalid commit body"
	}

	return "✓ Valid commit body"
}

// VerboseResult returns a more detailed explanation for verbose mode.
func (r CommitBodyRule) VerboseResult(errors []appErrors.ValidationError) string {
	if len(errors) > 0 {
		return "❌ Commit message body has formatting issues - check for missing content, sign-off requirements, or insufficient detail"
	}

	return "✓ Commit message body format is valid and meets all requirements"
}

// Help returns guidance for fixing rule violations.
func (r CommitBodyRule) Help(errors []appErrors.ValidationError) string {
	if len(errors) == 0 {
		return ""
	}

	return "Your commit message body has formatting issues.\n" +
		"Ensure there's an empty line between subject and body, and include meaningful content."
}

// Test code should be in test files with the "testing" build tag.
