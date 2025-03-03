// SPDX-FileCopyrightText: 2024 Sidero Labs, Inc.
// SPDX-FileCopyrightText: 2025 itiquette/gommitlint
//
// SPDX-License-Identifier: EUPL-1.2
package rule

import (
	"fmt"
	"unicode/utf8"
)

// DefaultMaxCommitSubjectLength is the default maximum number of characters
// allowed in a commit subject.
const DefaultMaxCommitSubjectLength = 100

// SubjectLengthRule enforces a maximum number of characters on the commit subject.
type SubjectLengthRule struct {
	subjectLength int
	errors        []error
}

// Name returns the rule name.
func (h *SubjectLengthRule) Name() string {
	return "SubjectLengthRule"
}

// Result returns the validation result.
func (h *SubjectLengthRule) Result() string {
	if len(h.errors) > 0 {
		return h.errors[0].Error()
	}

	return fmt.Sprintf("Subject is %d characters", h.subjectLength)
}

// Errors returns validation errors.
func (h *SubjectLengthRule) Errors() []error {
	return h.errors
}

// ValidateSubjectLengthRule checks the subject length.
func ValidateSubjectLengthRule(subject string, maxLength int) *SubjectLengthRule {
	if maxLength == 0 {
		maxLength = DefaultMaxCommitSubjectLength
	}

	subjectLength := utf8.RuneCountInString(subject)

	rule := &SubjectLengthRule{
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
