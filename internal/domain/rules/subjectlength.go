// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package rules

import (
	"context"
	"fmt"
	"strconv"

	"github.com/itiquette/gommitlint/internal/config"
	"github.com/itiquette/gommitlint/internal/domain"
)

// SubjectLengthRule validates the length of commit subjects.
type SubjectLengthRule struct {
	name      string
	maxLength int
}

// NewSubjectLengthRule creates a new SubjectLengthRule from config.
func NewSubjectLengthRule(cfg config.Config) SubjectLengthRule {
	maxLength := cfg.Message.Subject.MaxLength
	if maxLength <= 0 {
		maxLength = 72 // Default
	}

	return SubjectLengthRule{
		name:      "SubjectLength",
		maxLength: maxLength,
	}
}

// Name returns the rule name.
func (r SubjectLengthRule) Name() string {
	return r.name
}

// Validate performs validation against a commit.
func (r SubjectLengthRule) Validate(_ context.Context, commit domain.CommitInfo) []domain.ValidationError {
	// Validate the subject length
	subject := commit.Subject

	// Check if subject length exceeds maximum
	if len(subject) <= r.maxLength {
		return nil
	}

	// Create validation error with improved user message
	err := domain.New(
		"SubjectLength",
		domain.ErrMaxLengthExceeded,
		fmt.Sprintf("Commit subject is %d characters too long", len(subject)-r.maxLength),
	).WithHelp(fmt.Sprintf("Keep subject under %d characters (currently %d)", r.maxLength, len(subject))).WithContextMap(map[string]string{
		"actual":  strconv.Itoa(len(subject)),
		"max":     strconv.Itoa(r.maxLength),
		"subject": subject,
	})

	return []domain.ValidationError{err}
}
