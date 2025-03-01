// SPDX-FileCopyrightText: 2024 Sidero Labs, Inc.
// SPDX-FileCopyrightText: 2025 itiquette/gommitlint
//
// SPDX-License-Identifier: EUPL-1.2
package rule

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/pkg/errors"
)

// SubjectSuffixRule enforces that the last character of the subject isn't in a specified set.
type SubjectSuffixRule struct {
	errors []error
}

// Name returns the name of the rule.
func (h *SubjectSuffixRule) Name() string {
	return "SubjectSuffixRule"
}

func (h *SubjectSuffixRule) Result() string {
	if len(h.errors) > 0 {
		return h.errors[0].Error()
	}

	return "Subject last character is valid"
}

func (h *SubjectSuffixRule) Errors() []error {
	return h.errors
}

func ValidateSubjectSuffix(subject, invalidSuffixes string) *SubjectSuffixRule {
	rule := &SubjectSuffixRule{}

	last, _ := utf8.DecodeLastRuneInString(subject)

	// Check for invalid UTF-8
	if last == utf8.RuneError {
		rule.errors = append(rule.errors, errors.New("subject does not end with valid UTF-8 text"))

		return rule
	}

	// Check if the last character is in the invalid suffix set
	if strings.ContainsRune(invalidSuffixes, last) {
		rule.errors = append(rule.errors, fmt.Errorf("subject has invalid suffix %q (invalid suffixes: '%s')", last, invalidSuffixes))
	}

	return rule
}
