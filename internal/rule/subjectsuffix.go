// SPDX-FileCopyrightText: 2024 Sidero Labs, Inc.
// SPDX-FileCopyrightText: 2025 itiquette/gommitlint
//
// SPDX-License-Identifier: EUPL-1.2
package rule

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/itiquette/gommitlint/internal/interfaces"
	"github.com/pkg/errors"
)

// SubjectSuffix enforces that the last character of the subject isn't in a specified set.
type SubjectSuffix struct {
	errors []error
}

// Name returns the name of the rule.
func (h *SubjectSuffix) Name() string {
	return "Subject Last Character"
}

// Result returns the check message.
func (h *SubjectSuffix) Result() string {
	if len(h.errors) > 0 {
		return h.errors[0].Error()
	}

	return "Subject last character is valid"
}

// Errors returns any violations of the check.
func (h *SubjectSuffix) Errors() []error {
	return h.errors
}

// ValidateSubjectSuffix checks the last character of the subject.
func ValidateSubjectSuffix(subject, invalidSuffixes string) interfaces.CommitRule {
	rule := &SubjectSuffix{}

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
