// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package rules

import (
	"context"
	"fmt"

	"github.com/itiquette/gommitlint/internal/common/contextx"
	"github.com/itiquette/gommitlint/internal/domain"
	appErrors "github.com/itiquette/gommitlint/internal/errors"
)

// SubjectLengthRule validates the length of commit subjects.
// This rule uses configuration from context rather than embedded fields.
type SubjectLengthRule struct {
	name      string
	maxLength int
}

// NewSubjectLengthRule creates a new SubjectLengthRule.
func NewSubjectLengthRule() SubjectLengthRule {
	return SubjectLengthRule{
		name:      "SubjectLength",
		maxLength: 72, // Default max length
	}
}

// Name returns the rule name.
func (r SubjectLengthRule) Name() string {
	return r.name
}

// WithContext implements the ConfigurableRule interface for SubjectLengthRule.
// It returns a new rule with configuration from the provided context.
func (r SubjectLengthRule) WithContext(ctx context.Context) domain.Rule {
	// Get configuration from context
	cfg := contextx.GetConfig(ctx)
	if cfg == nil {
		return r
	}

	// Get max length from configuration
	maxLength := cfg.GetInt("message.subject.max_length")

	// Create a copy of the rule
	result := r

	// Only update if a value is explicitly set in config
	if maxLength > 0 {
		result.maxLength = maxLength
	}

	return result
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
	).WithContext("subject", subject)

	return []appErrors.ValidationError{err}
}
