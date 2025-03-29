// SPDX-FileCopyrightText: 2025 itiquette/gommitlint
//
// SPDX-License-Identifier: EUPL-1.2
package rule

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// DefaultMaxCommitSubjectLength is the default maximum number of characters
// allowed in a commit subject, following the common Git convention.
const DefaultMaxCommitSubjectLength = 100

// SubjectLength enforces a maximum number of characters on the commit subject line.
// This rule helps ensure commit messages remain readable in Git tools and UIs by
// preventing overly long subject lines that might get truncated. Most Git best practices
// recommend keeping subject lines to 50-72 characters, with an absolute maximum of 100.
// For example, with maxLength=50, a subject like "Implement new authentication system with support
// for multiple providers" would fail validation because it's too long, while
// "Implement multi-provider authentication" would pass.
type SubjectLength struct {
	subjectLength int
	errors        []error
}

// Name returns the rule name.
func (rule SubjectLength) Name() string {
	return "SubjectLength"
}

// Result returns the validation result.
func (rule SubjectLength) Result() string {
	if len(rule.errors) > 0 {
		return rule.errors[0].Error()
	}

	return fmt.Sprintf("Subject is %d characters", rule.subjectLength)
}

// addErrorf adds an error to the rule's errors slice.
func (rule *SubjectLength) addErrorf(format string, args ...interface{}) {
	rule.errors = append(rule.errors, fmt.Errorf(format, args...))
}

// Errors returns validation errors.
func (rule SubjectLength) Errors() []error {
	return rule.errors
}

// Help returns a description of how to fix the rule violation.
func (rule SubjectLength) Help() string {
	const noErrMsg = "No errors to fix"
	if len(rule.errors) == 0 {
		return noErrMsg
	}

	errMsg := rule.errors[0].Error()

	if strings.Contains(errMsg, "subject too long") {
		return "Shorten your commit message subject line to be more concise. " +
			"A good subject should be brief but descriptive, ideally under 50 characters " +
			"and no more than 100 characters. Consider using the commit body for additional details."
	}

	return "Review your commit message subject line length and ensure it follows project guidelines"
}

// ValidateSubjectLength checks the subject length against the specified maximum.
// If maxLength is 0, it uses the DefaultMaxCommitSubjectLength.
func ValidateSubjectLength(subject string, maxLength int) *SubjectLength {
	if maxLength == 0 {
		maxLength = DefaultMaxCommitSubjectLength
	}

	subjectLength := utf8.RuneCountInString(subject)
	rule := &SubjectLength{
		subjectLength: subjectLength,
	}

	// Validate length
	if subjectLength > maxLength {
		rule.addErrorf(
			"subject too long: %d characters (maximum allowed: %d)",
			subjectLength,
			maxLength,
		)
	}

	return rule
}
