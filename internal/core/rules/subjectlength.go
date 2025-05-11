// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package rules

import (
	"context"
	"fmt"
	"strings"

	"github.com/itiquette/gommitlint/internal/config"
	"github.com/itiquette/gommitlint/internal/domain"
	appErrors "github.com/itiquette/gommitlint/internal/errors"
	"github.com/itiquette/gommitlint/internal/infrastructure/log"
)

// SubjectLengthRule validates the length of commit subjects.
type SubjectLengthRule struct {
	baseRule  BaseRule
	maxLength int
}

// SubjectLengthOption configures a SubjectLengthRule.
type SubjectLengthOption func(SubjectLengthRule) SubjectLengthRule

// WithMaxLength sets the maximum subject length for the rule.
func WithMaxLength(length int) SubjectLengthOption {
	return func(r SubjectLengthRule) SubjectLengthRule {
		result := r
		result.maxLength = length

		return result
	}
}

// NewSubjectLengthRule creates a new SubjectLengthRule with the given options.
func NewSubjectLengthRule(options ...SubjectLengthOption) SubjectLengthRule {
	rule := SubjectLengthRule{
		baseRule:  NewBaseRule("SubjectLength"),
		maxLength: 72, // Default max length
	}

	// Apply all options
	for _, option := range options {
		rule = option(rule)
	}

	return rule
}

// Name returns the rule name.
func (r SubjectLengthRule) Name() string {
	return r.baseRule.Name()
}

// Validate performs validation against a commit.
func (r SubjectLengthRule) Validate(ctx context.Context, commit domain.CommitInfo) []appErrors.ValidationError {
	// Log validation at trace level
	logger := log.Logger(ctx)
	logger.Trace().
		Str("rule", r.Name()).
		Str("commit_hash", commit.Hash).
		Msg("Validating subject length")

	// Create a new rule with context configuration
	rule := r.withContextConfig(ctx)

	// Create a copy of the rule and run validation
	_, updatedRule := validateSubjectLengthWithState(rule, commit)

	// Return any validation errors
	return updatedRule.Errors()
}

// withContextConfig creates a new rule with configuration from context.
func (r SubjectLengthRule) withContextConfig(ctx context.Context) SubjectLengthRule {
	// Get configuration from context
	cfg := config.GetConfig(ctx)

	// Create a copy of the rule
	result := r

	// Override max length if specified in context config
	if cfg.Subject.MaxLength > 0 {
		result.maxLength = cfg.Subject.MaxLength

		// Log configuration at debug level
		logger := log.Logger(ctx)
		logger.Debug().
			Int("max_length", result.maxLength).
			Msg("Subject length rule configuration from context")
	}

	return result
}

// Errors returns all validation errors found by this rule.
func (r SubjectLengthRule) Errors() []appErrors.ValidationError {
	return r.baseRule.Errors()
}

// HasErrors returns true if the rule has found any errors.
func (r SubjectLengthRule) HasErrors() bool {
	return r.baseRule.HasErrors()
}

// validateSubjectLengthWithState validates the subject length and returns both the errors and an updated rule.
func validateSubjectLengthWithState(rule SubjectLengthRule, commit domain.CommitInfo) ([]appErrors.ValidationError, SubjectLengthRule) {
	result := rule
	result.baseRule = rule.baseRule.WithClearedErrors().WithRun()

	// Parse subject from commit
	subject := commit.Subject
	if subject == "" && commit.Message != "" {
		subject, _ = domain.SplitCommitMessage(commit.Message)
	}

	// Check subject length
	if len(subject) > rule.maxLength {
		helpText := fmt.Sprintf(`Subject Length Error: Your commit subject is too long.

Your commit subject exceeds the maximum allowed length of %d characters.

✅ CORRECT FORMAT:
A concise subject line that fits within %d characters

❌ INCORRECT FORMAT:
%s

WHY THIS MATTERS:
- Short subjects are easier to read in logs and commit history
- Many tools truncate long subjects when displaying commit history
- Enforcing a maximum length encourages clear, focused commits

NEXT STEPS:
1. Edit your commit message to shorten the subject line:
   - Focus on the core change, not every detail
   - Remove unnecessary words or context
   - Consider using a more concise verb
   - If needed, move details to the commit body

2. Your current subject is %d characters (maximum: %d)
3. Use 'git commit --amend' to update your commit message`,
			rule.maxLength, rule.maxLength, subject, len(subject), rule.maxLength)

		// We don't need a contextMap since LengthError accepts individual parameters
		// instead of a map for the standard fields

		validationErr := appErrors.LengthError(
			result.baseRule.Name(),
			fmt.Sprintf("subject exceeds maximum length (%d > %d)", len(subject), rule.maxLength),
			helpText,
			len(subject),
			rule.maxLength,
			subject)

		result.baseRule = result.baseRule.WithError(validationErr)
	}

	return result.baseRule.Errors(), result
}

// Result returns a concise validation result.
func (r SubjectLengthRule) Result(errors []appErrors.ValidationError) string {
	if len(errors) > 0 {
		return "❌ Subject length exceeds maximum"
	}

	return "✓ Subject length is valid"
}

// VerboseResult returns a more detailed explanation for verbose mode.
func (r SubjectLengthRule) VerboseResult(errors []appErrors.ValidationError) string {
	if len(errors) > 0 {
		var details []string

		for _, err := range errors {
			if subjectLen, ok := err.Context["subject_length"]; ok {
				if maxLen, ok := err.Context["max_length"]; ok {
					details = append(details, fmt.Sprintf("subject is %s characters (maximum allowed: %s)", subjectLen, maxLen))
				}
			}
		}

		if len(details) > 0 {
			return "❌ Subject length validation failed: " + strings.Join(details, ", ")
		}

		return "❌ Subject length validation failed"
	}

	return "✓ Subject length is within allowed limits"
}

// Help returns guidance for fixing rule violations.
func (r SubjectLengthRule) Help(errors []appErrors.ValidationError) string {
	if len(errors) == 0 {
		return ""
	}

	return "Your commit subject is too long. Edit your commit message with 'git commit --amend' to shorten the subject line."
}
