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

// SubjectCaseRule enforces the case of the first word in the subject.
type SubjectCaseRule struct {
	subjectCase string
	RuleErrors  []error
}

// Name returns the validation rule name.
func (h *SubjectCaseRule) Name() string {
	return "SubjectCaseRule"
}

// Result returns the rule message.
func (h *SubjectCaseRule) Result() string {
	if len(h.RuleErrors) > 0 {
		return h.RuleErrors[0].Error()
	}

	return "SubjectCaseRule is valid"
}

func (h *SubjectCaseRule) Errors() []error {
	return h.RuleErrors
}

// ValidateSubjectCaseRule checks the subject case based on the specified case choice.
func ValidateSubjectCaseRule(subject, caseChoice string, isConventional bool) *SubjectCaseRule {
	rule := &SubjectCaseRule{subjectCase: caseChoice}

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
