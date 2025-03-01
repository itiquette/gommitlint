// SPDX-FileCopyrightText: 2024 Sidero Labs, Inc.
// SPDX-FileCopyrightText: 2025 itiquette/gommitlint
//
// SPDX-License-Identifier: EUPL-1.2
package rule

import (
	"fmt"
	"unicode"
	"unicode/utf8"

	"github.com/pkg/errors"
)

// SubjectCase enforces the case of the first word in the subject.
type SubjectCase struct {
	subjectCase string
	RuleErrors  []error
}

// Name returns the name of the check.
func (h *SubjectCase) Name() string {
	return "Subject case"
}

// Result returns the check message.
func (h *SubjectCase) Result() string {
	if len(h.RuleErrors) > 0 {
		return h.RuleErrors[0].Error()
	}

	return "Subject case is valid"
}

// Errors returns any violations of the check.
func (h *SubjectCase) Errors() []error {
	return h.RuleErrors
}

// ValidateSubjectCase checks the subject case based on the specified case choice.
func ValidateSubjectCase(subject, caseChoice string, isConventional bool) *SubjectCase {
	rule := &SubjectCase{subjectCase: caseChoice}

	// Extract first word
	firstWord, err := extractFirstWord(isConventional, subject)
	if err != nil {
		rule.RuleErrors = append(rule.RuleErrors, err)

		return rule
	}

	// Decode first rune
	first, _ := utf8.DecodeRuneInString(firstWord)
	if first == utf8.RuneError {
		rule.RuleErrors = append(rule.RuleErrors, errors.New("subject does not start with valid UTF-8 text"))

		return rule
	}

	// Validate case
	var valid bool

	switch caseChoice {
	case "upper":
		valid = unicode.IsUpper(first)
	default:
		valid = unicode.IsLower(first)
	}

	if !valid {
		rule.RuleErrors = append(rule.RuleErrors, fmt.Errorf("commit subject case is not %s", caseChoice))
	}

	return rule
}
