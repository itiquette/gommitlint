// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
// Package rules provides validation rules for Git commits.
package rules

import (
	"context"
	"fmt"
	"strconv"
	"unicode/utf8"

	"github.com/itiquette/gommitlint/internal/domain"
	appErrors "github.com/itiquette/gommitlint/internal/errors"
)

// SubjectLengthRule validates that the commit subject is not too long.
type SubjectLengthRule struct {
	BaseRule
	maxLength int
}

// SubjectLengthOption is a function that configures a SubjectLengthRule.
type SubjectLengthOption func(SubjectLengthRule) SubjectLengthRule

// WithMaxLength sets the maximum allowed length for a commit subject.
func WithMaxLength(maxLength int) SubjectLengthOption {
	return func(rule SubjectLengthRule) SubjectLengthRule {
		if maxLength > 0 {
			rule.maxLength = maxLength
		}

		return rule
	}
}

// NewSubjectLengthRule creates a new SubjectLengthRule with options.
func NewSubjectLengthRule(options ...SubjectLengthOption) SubjectLengthRule {
	// Create base rule with default values
	rule := SubjectLengthRule{
		BaseRule:  NewBaseRule("SubjectLength"),
		maxLength: 100, // Default max length
	}

	// Apply options
	for _, option := range options {
		rule = option(rule)
	}

	return rule
}

// NewSubjectLengthRuleWithConfig creates a rule using configuration.
func NewSubjectLengthRuleWithConfig(config domain.SubjectConfigProvider) SubjectLengthRule {
	maxLength := config.SubjectMaxLength()

	// Use the options pattern
	return NewSubjectLengthRule(WithMaxLength(maxLength))
}

// NewSubjectLengthRuleWithConfig creates a rule using the unified configuration.

// Name returns the rule name.
func (r SubjectLengthRule) Name() string {
	return r.BaseRule.Name()
}

// validateSubjectLengthWithState validates a commit and returns both errors and an updated rule state.
// The function is purposely named with a unique name to avoid conflicts with other rules.
func validateSubjectLengthWithState(rule SubjectLengthRule, commit domain.CommitInfo) ([]appErrors.ValidationError, SubjectLengthRule) {
	// Start with clean slate and mark the rule as run
	updatedRule := rule
	updatedRule.BaseRule = updatedRule.BaseRule.WithClearedErrors().WithRun()

	// Calculate subject length
	subjectLength := utf8.RuneCountInString(commit.Subject)

	// Check if subject is too long
	if subjectLength > rule.maxLength {
		// Generate a rich help message with examples
		helpMessage := fmt.Sprintf(`Subject Length Error: Subject line is too long.

Your commit subject is %d characters long, which exceeds the maximum limit of %d characters.

✅ CORRECT FORMAT:
- Add user authentication feature (%d chars)
- Fix timeout issue in API requests (%d chars)
- Update documentation for deployment process (%d chars)

❌ INCORRECT FORMAT:
- %s (%d chars)

WHY THIS MATTERS:
- Shorter subjects are more readable in Git logs and UIs
- Many Git tools truncate subjects longer than 50-72 characters
- Short subjects force you to be concise and clearly describe the change
- Long subjects may indicate you're trying to include too many changes in one commit

NEXT STEPS:
1. Reduce the length of your subject line by:
   - Removing unnecessary details (save them for the body)
   - Using more concise wording
   - Focusing on the core change, not implementation details
   
2. If your subject is too complex to shorten, consider splitting 
   the commit into multiple smaller, focused commits

3. Use 'git commit --amend' to modify your most recent commit`,
			subjectLength, rule.maxLength,
			len("Add user authentication feature"),
			len("Fix timeout issue in API requests"),
			len("Update documentation for deployment process"),
			commit.Subject, subjectLength)

		// Create an enhanced validation error using the helper function
		errorMessage := fmt.Sprintf("Subject length (%d) exceeds maximum allowed (%d)", subjectLength, rule.maxLength)

		validationErr := appErrors.LengthError(
			rule.Name(),
			errorMessage,
			helpMessage,
			subjectLength,
			rule.maxLength,
			commit.Subject,
		)

		// Update rule with error using value semantics
		updatedRule.BaseRule = updatedRule.BaseRule.WithError(validationErr)
	}

	return updatedRule.Errors(), updatedRule
}

// Validate validates the commit subject length.
// This uses value semantics and returns the errors.
func (r SubjectLengthRule) Validate(_ context.Context, commit domain.CommitInfo) []appErrors.ValidationError {
	// Call the pure function implementation
	errors, _ := validateSubjectLengthWithState(r, commit)

	return errors
}

// Result returns a concise result message.
func (r SubjectLengthRule) Result(_ []appErrors.ValidationError) string {
	if len(r.Errors()) > 0 {
		return "Subject too long"
	}

	return "Subject length OK"
}

// VerboseResult returns a detailed result message.
func (r SubjectLengthRule) VerboseResult(_ []appErrors.ValidationError) string {
	if len(r.Errors()) > 0 {
		// Get the first error
		if errors := r.Errors(); len(errors) > 0 {
			// Use the enhanced error formatting
			formatter := appErrors.NewTextFormatter(true) // verbose mode

			return formatter.FormatError(errors[0])
		}

		// Fallback if no specific errors
		return fmt.Sprintf("Subject length exceeds maximum length of %d characters", r.maxLength)
	}

	return fmt.Sprintf("Subject length is within the limit of %d characters", r.maxLength)
}

// Help returns guidance on how to fix rule violations.
func (r SubjectLengthRule) Help(_ []appErrors.ValidationError) string {
	if len(r.Errors()) == 0 {
		return fmt.Sprintf("No errors to fix. This rule checks that commit subject lines don't exceed %d characters to ensure readability in various Git tools.", r.maxLength)
	}

	// If we have errors with the enhanced format, use their built-in help
	if errors := r.Errors(); len(errors) > 0 {
		// Get help from the enhanced error
		helpText := errors[0].GetHelp()
		if helpText != "" {
			return helpText
		}

		// Get current length if available
		currentLength := 0

		if ctx := errors[0].Context; ctx != nil {
			if val, ok := ctx["subject_length"]; ok {
				if parsedVal, err := strconv.Atoi(val); err == nil {
					currentLength = parsedVal
				}
			}
		}

		// Detailed help with examples
		if currentLength > 0 {
			return fmt.Sprintf(`Subject line is too long (%d characters). It should be at most %d characters.

Why this matters:
- Shorter subjects are more readable in Git logs and UIs 
- Many Git tools truncate subjects longer than 50-72 characters
- Short subjects force you to be concise and descriptive

Examples of good subject lines:
- Add user authentication feature (%d chars)
- Fix timeout issue in API requests (%d chars)
- Update documentation for deployment process (%d chars)

How to fix:
1. Remove unnecessary details (save them for the body)
2. Use more concise wording
3. Focus on the core change, not implementation details`,
				currentLength, r.maxLength,
				len("Add user authentication feature"),
				len("Fix timeout issue in API requests"),
				len("Update documentation for deployment process"))
		}

		// Fallback to template-based help message
		return fmt.Sprintf(`Ensure the subject line is at most %d characters long.

Why this matters:
- Shorter subjects are more readable in Git logs and UIs
- Many Git tools truncate subjects longer than 50-72 characters
- Short subjects force you to be concise and descriptive

Examples of good subject lines:
- Add user authentication feature
- Fix timeout issue in API requests
- Update documentation for deployment process

How to fix:
1. Remove unnecessary details (save them for the body)
2. Use more concise wording
3. Focus on the core change, not implementation details`, r.maxLength)
	}

	// Default help message
	return fmt.Sprintf("Keep the subject line under %d characters for better readability.", r.maxLength)
}

// Errors returns all validation errors.
func (r SubjectLengthRule) Errors() []appErrors.ValidationError {
	return r.BaseRule.Errors()
}

// HasErrors returns true if the rule has found any errors.
func (r SubjectLengthRule) HasErrors() bool {
	return len(r.Errors()) > 0
}

// ValidateSubjectLengthWithState is the exported version of validateSubjectLengthWithState.
// This is needed for testing but follows the same pure function approach.
// The function name is unique to avoid conflicts with similar functions in other rules.
func ValidateSubjectLengthWithState(rule SubjectLengthRule, commit domain.CommitInfo) ([]appErrors.ValidationError, SubjectLengthRule) {
	return validateSubjectLengthWithState(rule, commit)
}
