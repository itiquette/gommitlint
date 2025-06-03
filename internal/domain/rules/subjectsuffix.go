// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package rules

import (
	"fmt"
	"strings"

	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/itiquette/gommitlint/internal/domain/config"
)

// DefaultInvalidSuffixes contains the default invalid subject suffixes.
const DefaultInvalidSuffixes = ".,"

// SubjectSuffixRule validates that commit subjects don't end with invalid suffixes.
type SubjectSuffixRule struct {
	invalidSuffixes string
}

// NewSubjectSuffixRule creates a new SubjectSuffixRule from config.
func NewSubjectSuffixRule(cfg config.Config) SubjectSuffixRule {
	invalidSuffixes := DefaultInvalidSuffixes

	if len(cfg.Message.Subject.ForbidEndings) > 0 {
		// Join forbidden endings into a single string
		var builder strings.Builder
		for _, suffix := range cfg.Message.Subject.ForbidEndings {
			builder.WriteString(suffix)
		}

		if builder.Len() > 0 {
			invalidSuffixes = builder.String()
		}
	}

	return SubjectSuffixRule{
		invalidSuffixes: invalidSuffixes,
	}
}

// Validate checks that the commit subject doesn't end with invalid characters.
func (r SubjectSuffixRule) Validate(commit domain.Commit, _ domain.Repository, _ *config.Config) []domain.RuleFailure {
	// Empty subject is always an error
	if len(commit.Subject) == 0 {
		return []domain.RuleFailure{{
			Rule:    r.Name(),
			Message: "Commit subject is missing",
			Help:    "Add a descriptive subject line to your commit",
		}}
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
			for _, suffixRune := range r.invalidSuffixes {
				if suffixRune == lastRune {
					suffixContainsLastChar = true

					break
				}
			}

			// If the last character is an invalid suffix, create an error
			if suffixContainsLastChar {
				return []domain.RuleFailure{{
					Rule:    r.Name(),
					Message: fmt.Sprintf("Subject ends with invalid character '%s'", lastChar),
					Help:    fmt.Sprintf("Remove the trailing '%s' from your commit subject", lastChar),
				}}
			}
		}
	}

	// No errors found
	return nil
}

// Name returns the rule name.
func (r SubjectSuffixRule) Name() string {
	return "SubjectSuffix"
}
