// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package rules

import (
	"context"
	"strings"

	"github.com/itiquette/gommitlint/internal/config"
	"github.com/itiquette/gommitlint/internal/domain"
	appErrors "github.com/itiquette/gommitlint/internal/errors"
	"github.com/itiquette/gommitlint/internal/infrastructure/log"
)

// DefaultInvalidSuffixes contains the default invalid subject suffixes.
const DefaultInvalidSuffixes = ".,"

// SubjectSuffixRule validates that commit subjects don't end with invalid suffixes.
type SubjectSuffixRule struct {
	baseRule           BaseRule
	invalidSuffixes    string
	lastCheckedSubject string
}

// SubjectSuffixOption is a function that configures a SubjectSuffixRule.
type SubjectSuffixOption func(SubjectSuffixRule) SubjectSuffixRule

// WithInvalidSuffixes sets the suffixes that a commit subject should not end with.
func WithInvalidSuffixes(suffixes string) SubjectSuffixOption {
	return func(r SubjectSuffixRule) SubjectSuffixRule {
		result := r
		result.invalidSuffixes = suffixes

		return result
	}
}

// WithInvalidSuffixes returns a new rule with updated invalid suffixes.
func (r SubjectSuffixRule) WithInvalidSuffixes(suffixes string) SubjectSuffixRule {
	rule := r
	rule.invalidSuffixes = suffixes

	return rule
}

// NewSubjectSuffixRule creates a new SubjectSuffixRule with the specified options.
func NewSubjectSuffixRule(options ...SubjectSuffixOption) SubjectSuffixRule {
	// Create a rule with default settings
	rule := SubjectSuffixRule{
		baseRule:        NewBaseRule("SubjectSuffix"),
		invalidSuffixes: DefaultInvalidSuffixes,
	}

	// Apply options
	for _, option := range options {
		rule = option(rule)
	}

	// If options resulted in empty invalidSuffixes, revert to default
	if rule.invalidSuffixes == "" {
		rule.invalidSuffixes = DefaultInvalidSuffixes
	}

	return rule
}

// Validate checks that the commit subject doesn't end with invalid characters
// using configuration from context.
func (r SubjectSuffixRule) Validate(ctx context.Context, commit domain.CommitInfo) []appErrors.ValidationError {
	logger := log.Logger(ctx)
	logger.Trace().
		Str("rule", r.Name()).
		Str("commit_hash", commit.Hash).
		Msg("Validating subject suffix using context configuration")

	// Create a new rule with context configuration
	rule := r.withContextConfig(ctx)

	// Use the validation logic
	errors, _ := validateSubjectSuffixWithState(rule, commit)

	return errors
}

// withContextConfig creates a new rule with configuration from context.
func (r SubjectSuffixRule) withContextConfig(ctx context.Context) SubjectSuffixRule {
	// Get configuration from context
	cfg := config.GetConfig(ctx)

	// Extract configuration values - disallowed suffixes
	// Join the slice of disallowed suffixes into a string
	invalidSuffixes := DefaultInvalidSuffixes

	// Use DisallowedSuffixes from config if available
	if len(cfg.Subject.DisallowedSuffixes) > 0 {
		invalidSuffixes = ""
		for _, suffix := range cfg.Subject.DisallowedSuffixes {
			invalidSuffixes += suffix
		}
	}

	// Create a copy of the rule
	result := r

	// Update settings from context - always set invalidSuffixes even if empty
	// to ensure we're using the context configuration
	result.invalidSuffixes = invalidSuffixes

	// Default to DefaultInvalidSuffixes if ended up with empty string
	if result.invalidSuffixes == "" {
		result.invalidSuffixes = DefaultInvalidSuffixes
	}

	// Log configuration at debug level
	logger := log.Logger(ctx)
	logger.Debug().
		Str("invalid_suffixes", result.invalidSuffixes).
		Msg("Subject suffix rule configuration from context")

	return result
}

// validateSubjectSuffixWithState validates the subject suffix and returns both the errors and an updated rule.
func validateSubjectSuffixWithState(rule SubjectSuffixRule, commit domain.CommitInfo) ([]appErrors.ValidationError, SubjectSuffixRule) {
	// Start with a clean rule
	result := rule
	result.baseRule = rule.baseRule.WithClearedErrors().WithRun()

	// Save the commit subject for consistency
	result.lastCheckedSubject = commit.Subject

	// Empty subject is always an error
	if len(commit.Subject) == 0 {
		missingSubjectError := appErrors.CreateBasicError(
			result.baseRule.Name(),
			appErrors.ErrMissingSubject,
			"Commit subject cannot be empty",
		).WithContext("subject", "")

		result.baseRule = result.baseRule.WithError(missingSubjectError)

		return result.baseRule.Errors(), result
	}

	// Real validation logic - check if the subject ends with any of the invalid suffixes
	if len(commit.Subject) > 0 {
		lastChar := string(commit.Subject[len(commit.Subject)-1])

		// Check if the last character is in the invalid suffixes
		if strings.Contains(rule.invalidSuffixes, lastChar) {
			invalidSuffixError := appErrors.CreateBasicError(
				result.baseRule.Name(),
				appErrors.ErrSubjectSuffix,
				"Commit subject should not end with '"+lastChar+"'",
			).
				WithContext("subject", commit.Subject).
				WithContext("invalid_suffix", lastChar).
				WithContext("last_char", lastChar).
				WithContext("invalid_suffixes", rule.invalidSuffixes)

			result.baseRule = result.baseRule.WithError(invalidSuffixError)
		}
	}

	return result.baseRule.Errors(), result
}

// Name returns the rule name.
func (r SubjectSuffixRule) Name() string {
	return r.baseRule.Name()
}

// Errors returns all validation errors found by this rule.
func (r SubjectSuffixRule) Errors() []appErrors.ValidationError {
	return r.baseRule.Errors()
}

// HasErrors returns true if the rule has found any errors.
func (r SubjectSuffixRule) HasErrors() bool {
	// Standard implementation - check if we have any errors
	errors := r.baseRule.Errors()

	return len(errors) > 0
}

// SetErrors sets the errors for this rule and returns an updated rule.
func (r SubjectSuffixRule) SetErrors(errors []appErrors.ValidationError) SubjectSuffixRule {
	result := r
	result.baseRule = result.baseRule.WithClearedErrors()

	for _, err := range errors {
		result.baseRule = result.baseRule.WithError(err)
	}

	return result
}

// Result returns a concise validation result.
func (r SubjectSuffixRule) Result(errors []appErrors.ValidationError) string {
	if len(errors) > 0 {
		// Return a consistent message to maintain compatibility
		return "Invalid subject suffix"
	}

	return "Valid subject suffix"
}

// VerboseResult returns a more detailed explanation for verbose mode.
func (r SubjectSuffixRule) VerboseResult(errors []appErrors.ValidationError) string {
	if len(errors) > 0 {
		for _, err := range errors {
			if err.Code == string(appErrors.ErrMissingSubject) {
				return "❌ Commit subject cannot be empty"
			}

			if suffix, ok := err.Context["invalid_suffix"]; ok {
				return "❌ Commit subject should not end with '" + suffix + "'"
			}
		}

		return "❌ Commit subject ends with an invalid suffix"
	}

	return "✓ Commit subject does not end with any invalid suffixes"
}

// Help returns guidance for fixing rule violations.
func (r SubjectSuffixRule) Help(errors []appErrors.ValidationError) string {
	if len(errors) == 0 {
		return "No errors to fix"
	}

	for _, err := range errors {
		if err.Code == string(appErrors.ErrMissingSubject) {
			return "Provide a non-empty subject for your commit message. The subject should be a brief summary of the changes."
		}
	}

	help := "Your commit subject should not end with punctuation marks like "

	// Show a sample of disallowed suffixes
	for _, c := range r.invalidSuffixes {
		help += "'" + string(c) + "', "
	}

	// Remove the trailing comma and space
	if len(r.invalidSuffixes) > 0 {
		help = help[:len(help)-2]
	}

	help += ".\n\nThese suffixes are considered unnecessary in commit messages.\n\nRemove the punctuation mark at the end of your subject line."

	return help
}
