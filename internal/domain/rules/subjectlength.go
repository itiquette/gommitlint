// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package rules

import (
	"fmt"

	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/itiquette/gommitlint/internal/domain/config"
)

// SubjectLengthRule validates the length of commit subjects.
type SubjectLengthRule struct {
	maxLength int
}

// NewSubjectLengthRule creates a new SubjectLengthRule from config.
func NewSubjectLengthRule(cfg config.Config) SubjectLengthRule {
	maxLength := cfg.Message.Subject.MaxLength
	if maxLength <= 0 {
		maxLength = 72 // Default
	}

	return SubjectLengthRule{
		maxLength: maxLength,
	}
}

// Name returns the rule name.
func (r SubjectLengthRule) Name() string {
	return "SubjectLength"
}

// Validate performs validation against a commit.
func (r SubjectLengthRule) Validate(commit domain.Commit, _ domain.Repository, _ *config.Config) []domain.RuleFailure {
	if len(commit.Subject) <= r.maxLength {
		return nil
	}

	return []domain.RuleFailure{{
		Rule:    r.Name(),
		Message: fmt.Sprintf("subject exceeds %d characters (actual: %d)", r.maxLength, len(commit.Subject)),
		Help:    fmt.Sprintf("Keep subject under %d characters", r.maxLength),
	}}
}
