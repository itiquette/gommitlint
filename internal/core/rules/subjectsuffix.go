// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package rules

import (
	"context"
	"fmt"
	"strings"

	"github.com/itiquette/gommitlint/internal/common/contextx"
	"github.com/itiquette/gommitlint/internal/domain"
	appErrors "github.com/itiquette/gommitlint/internal/errors"
)

// DefaultInvalidSuffixes contains the default invalid subject suffixes.
const DefaultInvalidSuffixes = ".,"

// SubjectSuffixRule validates that commit subjects don't end with invalid suffixes.
type SubjectSuffixRule struct {
	name            string
	invalidSuffixes string
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

// NewSubjectSuffixRule creates a new SubjectSuffixRule with the specified options.
func NewSubjectSuffixRule(options ...SubjectSuffixOption) SubjectSuffixRule {
	// Create a rule with default settings
	rule := SubjectSuffixRule{
		name:            "SubjectSuffix",
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

// Validate checks that the commit subject doesn't end with invalid characters.
func (r SubjectSuffixRule) Validate(ctx context.Context, commit domain.CommitInfo) []appErrors.ValidationError {
	logger := contextx.GetLogger(ctx)
	logger.Debug("Validating subject suffix",
		"rule", r.Name(),
		"commit_hash", commit.Hash)

	// Get invalid suffixes from context configuration
	invalidSuffixes := r.invalidSuffixes

	// Get config object directly from standard context key
	cfg := contextx.GetConfig(ctx)
	disallowedSuffixes := cfg.GetStringSlice("message.subject.forbid_endings")

	// Process disallowed suffixes if we have them
	if len(disallowedSuffixes) > 0 {
		// For multi-byte characters like emojis, we need to ensure they're not split
		// Build the string without using Join which can break multi-byte characters
		var sb strings.Builder
		for _, suffix := range disallowedSuffixes {
			sb.WriteString(suffix)
		}

		configInvalidSuffixes := sb.String()

		// Log the processed suffixes for debugging
		logger.Debug("Processing disallowed suffixes from context config",
			"disallowed_suffixes_slice", disallowedSuffixes,
			"processed_invalid_suffixes", configInvalidSuffixes)

		// Update rule setting
		if configInvalidSuffixes != "" {
			invalidSuffixes = configInvalidSuffixes
		}
	}

	// Log configuration at debug level
	logger.Debug("Subject suffix rule configuration",
		"invalid_suffixes", invalidSuffixes)

	// Empty subject is always an error
	if len(commit.Subject) == 0 {
		return []appErrors.ValidationError{
			appErrors.NewSuffixError(
				appErrors.ErrMissingSubject,
				"SubjectSuffix",
				"Commit subject is missing",
				"Add a descriptive subject line to your commit",
			).WithContext("subject", ""),
		}
	}

	// Real validation logic - check if the subject ends with any of the invalid suffixes
	if len(commit.Subject) > 0 {
		// Get the last character, properly handling multi-byte characters like emojis
		subjectRunes := []rune(commit.Subject)
		if len(subjectRunes) > 0 {
			lastRune := subjectRunes[len(subjectRunes)-1]
			lastChar := string(lastRune)

			// Check if the last character is in the invalid suffixes
			suffixContainsLastChar := false

			// Range directly over the string to handle multi-byte characters properly
			for _, suffixRune := range invalidSuffixes {
				if suffixRune == lastRune {
					suffixContainsLastChar = true

					break
				}
			}

			// If the last character is an invalid suffix, create an error
			if suffixContainsLastChar {
				return []appErrors.ValidationError{
					appErrors.NewSuffixError(
						appErrors.ErrSubjectSuffix,
						"SubjectSuffix",
						fmt.Sprintf("Subject ends with invalid character '%s'", lastChar),
						fmt.Sprintf("Remove the trailing '%s' from your commit subject", lastChar),
					).WithContextMap(map[string]string{
						"subject":          commit.Subject,
						"invalid_suffix":   lastChar,
						"last_char":        lastChar,
						"invalid_suffixes": invalidSuffixes,
					}),
				}
			}
		}
	}

	// No errors found
	return nil
}

// Name returns the rule name.
func (r SubjectSuffixRule) Name() string {
	return r.name
}
