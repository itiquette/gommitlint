// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
// Package rules provides validation rules for Git commits.
package rules

import (
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
		// Create context for the error
		context := map[string]string{
			"actual_length": strconv.Itoa(subjectLength),
			"max_length":    strconv.Itoa(rule.maxLength),
		}

		// Update rule with error using value semantics
		updatedRule.BaseRule = updatedRule.BaseRule.WithErrorWithContext(
			appErrors.ErrSubjectTooLong,
			fmt.Sprintf("Subject length (%d) exceeds maximum allowed (%d)", subjectLength, rule.maxLength),
			context,
		)
	}

	return updatedRule.Errors(), updatedRule
}

// Validate validates the commit subject length.
// This uses value semantics and returns the errors.
func (r SubjectLengthRule) Validate(commit domain.CommitInfo) []appErrors.ValidationError {
	// Call the pure function implementation
	errors, _ := validateSubjectLengthWithState(r, commit)

	return errors
}

// Result returns a concise result message.
func (r SubjectLengthRule) Result() string {
	if len(r.Errors()) > 0 {
		return "Subject too long"
	}

	return "Subject length OK"
}

// VerboseResult returns a detailed result message.
func (r SubjectLengthRule) VerboseResult() string {
	if len(r.Errors()) > 0 {
		// Get context values
		actualLength := "unknown"
		maxLength := strconv.Itoa(r.maxLength)

		if errors := r.Errors(); len(errors) > 0 {
			if ctx := errors[0].Context; ctx != nil {
				if val, ok := ctx["actual_length"]; ok {
					actualLength = val
				}

				if val, ok := ctx["max_length"]; ok {
					maxLength = val
				}
			}
		}

		return fmt.Sprintf("Subject length (%s characters) exceeds maximum length (%s characters)", actualLength, maxLength)
	}

	return fmt.Sprintf("Subject length is within the limit of %d characters", r.maxLength)
}

// Help returns guidance on how to fix rule violations.
func (r SubjectLengthRule) Help() string {
	if len(r.Errors()) == 0 {
		return "No help needed"
	}

	// Get max length from context
	maxLength := r.maxLength

	if errors := r.Errors(); len(errors) > 0 {
		if ctx := errors[0].Context; ctx != nil {
			if val, ok := ctx["max_length"]; ok {
				if parsedVal, err := strconv.Atoi(val); err == nil {
					maxLength = parsedVal
				}
			}
		}
	}

	// Use the template-based help message
	return appErrors.FormatHelp(appErrors.ErrSubjectTooLong, maxLength)
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
