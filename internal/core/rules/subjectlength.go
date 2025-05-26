// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package rules

import (
	"context"
	"fmt"

	"github.com/itiquette/gommitlint/internal/domain"
	appErrors "github.com/itiquette/gommitlint/internal/errors"
)

// SubjectLengthRule validates the length of commit subjects.
type SubjectLengthRule struct {
	name      string
	maxLength int
}

// SubjectLengthOption configures a SubjectLengthRule.
type SubjectLengthOption func(SubjectLengthRule) SubjectLengthRule

// WithMaxLength sets the maximum length for the subject.
func WithMaxLength(maxLength int) SubjectLengthOption {
	return func(r SubjectLengthRule) SubjectLengthRule {
		if maxLength > 0 {
			r.maxLength = maxLength
		}

		return r
	}
}

// NewSubjectLengthRule creates a new SubjectLengthRule.
func NewSubjectLengthRule(options ...SubjectLengthOption) SubjectLengthRule {
	rule := SubjectLengthRule{
		name:      "SubjectLength",
		maxLength: 72, // Default max length
	}

	for _, option := range options {
		rule = option(rule)
	}

	return rule
}

// Name returns the rule name.
func (r SubjectLengthRule) Name() string {
	return r.name
}

// Validate performs validation against a commit.
func (r SubjectLengthRule) Validate(_ context.Context, commit domain.CommitInfo) []appErrors.ValidationError {
	// Validate the subject length
	subject := commit.Subject

	// Check if subject length exceeds maximum
	if len(subject) <= r.maxLength {
		return nil
	}

	// Create validation error with improved user message
	err := appErrors.NewLengthError(
		appErrors.ErrMaxLengthExceeded,
		"SubjectLength",
		fmt.Sprintf("Commit subject is %d characters too long", len(subject)-r.maxLength),
		len(subject),
		r.maxLength,
	).WithContextMap(map[string]string{"subject": subject})

	return []appErrors.ValidationError{err}
}
