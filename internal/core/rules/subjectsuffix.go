// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package rules

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/itiquette/gommitlint/internal/domain"
	appErrors "github.com/itiquette/gommitlint/internal/errors"
)

// DefaultInvalidSuffixes is the default set of characters that should not appear
// at the end of a commit subject line.
const DefaultInvalidSuffixes = ".,;:!?"

// SubjectSuffixRule enforces that the last character of the commit subject line
// is not in a specified set of invalid suffixes.
//
// This rule helps ensure commit messages maintain a consistent format by
// preventing subjects from ending with unwanted characters like periods,
// commas, or other punctuation marks that can affect readability and
// automated processing of commit messages.
type SubjectSuffixRule struct {
	BaseRule
	invalidSuffixes string
}

// SubjectSuffixOption is a function that modifies a SubjectSuffixRule.
type SubjectSuffixOption func(SubjectSuffixRule) SubjectSuffixRule

// WithInvalidSuffixes sets custom invalid suffix characters.
func WithInvalidSuffixes(suffixes string) SubjectSuffixOption {
	return func(rule SubjectSuffixRule) SubjectSuffixRule {
		result := rule
		result.invalidSuffixes = suffixes

		return result
	}
}

// NewSubjectSuffixRule creates a new SubjectSuffixRule with the specified options.
func NewSubjectSuffixRule(options ...SubjectSuffixOption) SubjectSuffixRule {
	rule := SubjectSuffixRule{
		BaseRule:        NewBaseRule("SubjectSuffix"),
		invalidSuffixes: DefaultInvalidSuffixes,
	}

	// Apply options
	for _, option := range options {
		rule = option(rule)
	}

	// If invalid suffixes is empty, use the default
	if rule.invalidSuffixes == "" {
		rule.invalidSuffixes = DefaultInvalidSuffixes
	}

	return rule
}

// NewSubjectSuffixRuleWithConfig creates a SubjectSuffixRule using configuration.
func NewSubjectSuffixRuleWithConfig(config domain.SubjectConfigProvider) SubjectSuffixRule {
	// Build options based on the configuration
	var options []SubjectSuffixOption

	// Set the invalid suffixes if provided
	if suffixes := config.SubjectInvalidSuffixes(); suffixes != "" {
		options = append(options, WithInvalidSuffixes(suffixes))
	}

	return NewSubjectSuffixRule(options...)
}

// Name returns the rule name.
func (r SubjectSuffixRule) Name() string {
	return r.BaseRule.Name()
}

// validateSubjectSuffixWithState validates the commit and returns an updated rule state and errors.
// The function is purposely named with a unique name to avoid conflicts with other rules.
func validateSubjectSuffixWithState(rule SubjectSuffixRule, commit domain.CommitInfo) ([]appErrors.ValidationError, SubjectSuffixRule) {
	// Start with a clean slate by creating a new rule with cleared errors
	updatedRule := rule
	updatedRule.BaseRule = updatedRule.BaseRule.WithClearedErrors().WithRun()

	subject := commit.Subject
	if subject == "" {
		context := map[string]string{
			"subject": subject,
		}
		updatedRule.BaseRule = updatedRule.BaseRule.WithErrorWithContext(
			appErrors.ErrMissingSubject,
			"subject is empty",
			context,
		)

		return updatedRule.Errors(), updatedRule
	}

	lastChar, size := utf8.DecodeLastRuneInString(subject)

	// Check for invalid UTF-8
	if lastChar == utf8.RuneError && size == 0 {
		context := map[string]string{
			"subject": subject,
		}
		updatedRule.BaseRule = updatedRule.BaseRule.WithErrorWithContext(
			appErrors.ErrInvalidFormat,
			"subject does not end with valid UTF-8 text",
			context,
		)

		return updatedRule.Errors(), updatedRule
	}

	// Check if the last character is in the invalid suffix set
	if strings.ContainsRune(rule.invalidSuffixes, lastChar) {
		context := map[string]string{
			"subject":          subject,
			"last_char":        string(lastChar),
			"invalid_suffixes": rule.invalidSuffixes,
		}
		updatedRule.BaseRule = updatedRule.BaseRule.WithErrorWithContext(
			appErrors.ErrSubjectSuffix,
			fmt.Sprintf("subject has invalid suffix %q (invalid suffixes: %q)", string(lastChar), rule.invalidSuffixes),
			context,
		)
	}

	return updatedRule.Errors(), updatedRule
}

// Validate validates that the subject doesn't end with invalid characters.
// This method follows functional programming principles and does not modify the rule's state.
func (r SubjectSuffixRule) Validate(commit domain.CommitInfo) []appErrors.ValidationError {
	errors, _ := validateSubjectSuffixWithState(r, commit)

	return errors
}

// Result returns a concise validation result.
func (r SubjectSuffixRule) Result() string {
	if r.HasErrors() {
		return "Invalid subject suffix"
	}

	return "Valid subject suffix"
}

// VerboseResult returns a more detailed result message.
func (r SubjectSuffixRule) VerboseResult() string {
	if !r.HasErrors() {
		return "Subject ends with valid character"
	}

	// If we have an error, provide details based on the error type
	if r.ErrorCount() > 0 {
		code := r.Errors()[0].Code
		if code == string(appErrors.ErrMissingSubject) {
			return "Subject is empty"
		}

		if code == string(appErrors.ErrInvalidFormat) {
			return "Subject contains invalid UTF-8 characters"
		}
		// If we have a more specific error message from the validation, use it
		message := r.Errors()[0].Message
		if message != "" {
			return message
		}
	}

	// Default message
	return fmt.Sprintf("Subject ends with invalid character (invalid suffixes: %s)", r.invalidSuffixes)
}

// Help returns guidance on how to fix rule violations.
func (r SubjectSuffixRule) Help() string {
	if !r.HasErrors() {
		return "No errors to fix"
	}

	// Check for specific error codes and provide appropriate help messages
	if r.ErrorCount() > 0 {
		code := r.Errors()[0].Code
		// Check for missing subject errors
		if code == string(appErrors.ErrMissingSubject) {
			return "Provide a non-empty subject line for your commit message"
		}
		// Check for invalid UTF-8 errors
		if code == string(appErrors.ErrInvalidFormat) {
			return "Ensure your commit message contains only valid UTF-8 characters"
		}
		// Check for invalid suffix errors
		if code == string(appErrors.ErrSubjectSuffix) {
			var invalidSuffixes string
			if suffixes, ok := r.Errors()[0].Context["invalid_suffixes"]; ok {
				invalidSuffixes = suffixes
			} else {
				invalidSuffixes = DefaultInvalidSuffixes
			}

			return fmt.Sprintf("Remove the punctuation or special character from the end of your subject line. "+
				"The subject should end with a letter or number, not punctuation like: %s", invalidSuffixes)
		}
	}

	return "Review and fix your commit message subject line according to the guidelines"
}

// Errors returns all validation errors.
func (r SubjectSuffixRule) Errors() []appErrors.ValidationError {
	return r.BaseRule.Errors()
}

// HasErrors returns true if the rule has found any errors.
func (r SubjectSuffixRule) HasErrors() bool {
	return len(r.Errors()) > 0
}

// ValidateSubjectSuffixWithState is the exported version of validateSubjectSuffixWithState.
// This is needed for testing but follows the same pure function approach.
// The function name is unique to avoid conflicts with similar functions in other rules.
func ValidateSubjectSuffixWithState(rule SubjectSuffixRule, commit domain.CommitInfo) ([]appErrors.ValidationError, SubjectSuffixRule) {
	return validateSubjectSuffixWithState(rule, commit)
}
