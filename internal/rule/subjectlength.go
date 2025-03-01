// SPDX-FileCopyrightText: 2024 Sidero Labs, Inc.
// SPDX-FileCopyrightText: 2025 itiquette/gommitlint
//
// SPDX-License-Identifier: EUPL-1.2
package rule

import (
	"fmt"
	"unicode/utf8"

	"github.com/itiquette/gommitlint/internal/interfaces"
)

// DefaultMaxCommitSubjectLength is the default maximum number of characters
// allowed in a commit subject.
const DefaultMaxCommitSubjectLength = 100

// SubjectLength enforces a maximum number of characters on the commit subject.
type SubjectLength struct {
	subjectLength int
	errors        []error
}

// Name returns the name of the check.
func (h *SubjectLength) Name() string {
	return "Subject Length"
}

// Result returns the check message.
func (h *SubjectLength) Result() string {
	if len(h.errors) > 0 {
		return h.errors[0].Error()
	}

	return fmt.Sprintf("Subject is %d characters", h.subjectLength)
}

// Errors returns any violations of the check.
func (h *SubjectLength) Errors() []error {
	return h.errors
}

// ValidateSubjectLength checks the subject length.
func ValidateSubjectLength(subject string, maxLength int) interfaces.CommitRule {
	if maxLength == 0 {
		maxLength = DefaultMaxCommitSubjectLength
	}

	subjectLength := utf8.RuneCountInString(subject)

	rule := &SubjectLength{
		subjectLength: subjectLength,
	}

	// Validate length
	if subjectLength > maxLength {
		rule.errors = append(rule.errors, fmt.Errorf(
			"subject too long: %d characters (maximum allowed: %d)",
			subjectLength,
			maxLength,
		))
	}

	return rule
}
