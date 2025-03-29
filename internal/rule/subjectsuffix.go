// SPDX-FileCopyrightText: 2025 itiquette/gommitlint
//
// SPDX-License-Identifier: EUPL-1.2
package rule

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// SubjectSuffix enforces that the last character of the subject line isn't in a specified set
// of invalid suffixes. This rule helps ensure commit messages maintain a consistent format by
// preventing them from ending with characters like periods, commas, or other specified punctuation.
// For example, with invalidSuffixes=".,;:", a subject like "Add new feature." would fail validation
// because it ends with a period, while "Add new feature" would pass.
type SubjectSuffix struct {
	errors []error
}

// Name returns the name of the rule.
func (rule SubjectSuffix) Name() string {
	return "SubjectSuffixRule"
}

// Result returns the rule message.
func (rule SubjectSuffix) Result() string {
	if len(rule.errors) > 0 {
		return rule.errors[0].Error()
	}

	return "Subject last character is valid"
}

// Errors returns any violations of the rule.
func (rule SubjectSuffix) Errors() []error {
	return rule.errors
}

// addErrorf adds an error to the rule's errors slice.
func (rule *SubjectSuffix) addErrorf(format string, args ...interface{}) {
	rule.errors = append(rule.errors, fmt.Errorf(format, args...))
}

// ValidateSubjectSuffix checks if the subject ends with a character in the invalidSuffixes set.
// It returns a SubjectSuffix with any validation errors.
func ValidateSubjectSuffix(subject, invalidSuffixes string) SubjectSuffix {
	rule := SubjectSuffix{}

	if subject == "" {
		rule.addErrorf("subject is empty")

		return rule
	}

	lastChar, size := utf8.DecodeLastRuneInString(subject)

	// Check for invalid UTF-8
	if lastChar == utf8.RuneError && size == 0 {
		rule.addErrorf("subject does not end with valid UTF-8 text")

		return rule
	}

	// Check if the last character is in the invalid suffix set
	if strings.ContainsRune(invalidSuffixes, lastChar) {
		rule.addErrorf("subject has invalid suffix %q (invalid suffixes: %q)", lastChar, invalidSuffixes)
	}

	return rule
}

// Help returns a description of how to fix the rule violation.
func (rule SubjectSuffix) Help() string {
	if len(rule.errors) == 0 {
		return "No errors to fix"
	}

	errMsg := rule.errors[0].Error()

	if strings.Contains(errMsg, "subject is empty") {
		return "Provide a non-empty subject line for your commit message"
	}

	if strings.Contains(errMsg, "does not end with valid UTF-8") {
		return "Ensure your commit message contains only valid UTF-8 characters"
	}

	if strings.Contains(errMsg, "invalid suffix") {
		return "Remove the punctuation or special character from the end of your subject line. " +
			"The subject should end with a letter or number, not punctuation."
	}

	return "Review and fix your commit message subject line according to the guidelines"
}
