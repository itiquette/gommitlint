// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package rules

import (
	"context"
	"fmt"
	"strconv"

	"github.com/itiquette/gommitlint/internal/common/contextx"
	"github.com/itiquette/gommitlint/internal/domain"
	appErrors "github.com/itiquette/gommitlint/internal/errors"
)

// SubjectLengthRule validates the length of commit subjects.
// This rule uses configuration from context rather than embedded fields.
type SubjectLengthRule struct {
	name      string
	verbosity string
}

// NewSubjectLengthRule creates a new SubjectLengthRule.
func NewSubjectLengthRule() SubjectLengthRule {
	return SubjectLengthRule{
		name: "SubjectLength",
	}
}

// Name returns the rule name.
func (r SubjectLengthRule) Name() string {
	return r.name
}

// Validate performs validation against a commit using configuration from context.
func (r SubjectLengthRule) Validate(ctx context.Context, commit domain.CommitInfo) []appErrors.ValidationError {
	// Log validation at trace level
	logger := contextx.GetLogger(ctx)
	logger.Debug("Validating subject length",
		"rule", r.Name(),
		"commit_hash", commit.Hash)

	// Get configuration from context
	cfg := contextx.GetConfig(ctx)
	if cfg == nil {
		// If no config is available, return nil (no validation)
		return nil
	}

	// Get max length from configuration with default
	maxLength := cfg.GetInt("subject.max_length")
	if maxLength == 0 {
		maxLength = 72 // Default max length
	}

	// Log configuration at debug level
	logger.Debug("Subject length rule configuration from context", "max_length", maxLength)

	// Validate the subject length
	subject := commit.Subject

	// Check if subject length exceeds maximum
	if len(subject) <= maxLength {
		return nil
	}

	// Create validation error
	err := appErrors.New(
		"SubjectLength",
		appErrors.ErrMaxLengthExceeded,
		fmt.Sprintf(
			"Commit subject exceeds %d characters. Current length: %d - Subject: %s",
			maxLength,
			len(subject),
			subject,
		),
	).WithContext("actual_length", strconv.Itoa(len(subject))).
		WithContext("max_length", strconv.Itoa(maxLength)).
		WithContext("subject", subject).
		WithContext("help", fmt.Sprintf("Shorten your commit subject to %d characters or less. Current length: %d characters.",
			maxLength, len(subject)))

	// Log the error at error level
	logger.Debug("Subject length validation failed", "error", err.Error())

	return []appErrors.ValidationError{err}
}

// WithVerbosity returns a new SubjectLengthRule with the specified verbosity.
func (r SubjectLengthRule) WithVerbosity(verbosity string) SubjectLengthRule {
	newRule := r
	newRule.verbosity = verbosity

	return newRule
}
