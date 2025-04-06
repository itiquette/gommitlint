// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
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
//
// This rule helps ensure commit messages remain readable and effective in Git tools and
// interfaces by preventing overly long subject lines that might get truncated or
// become difficult to scan. Many Git interfaces truncate subjects at 72 characters, and
// best practices recommend keeping subjects concise and focused.
//
// The rule validates that the number of characters (Unicode code points) in the subject
// does not exceed the specified maximum length. If no maximum is specified, it uses
// the DefaultMaxCommitSubjectLength of 100 characters.
//
// Examples:
//
//   - With maxLength=50:
//     "Add user authentication" would pass (28 characters)
//     "Implement comprehensive user authentication system with multi-factor capabilities" would fail (76 characters)
//
//   - With maxLength=72:
//     "Refactor database connection pool to improve resource utilization" would pass (65 characters)
//     "Completely redesign the application's architecture to support distributed processing across multiple geographic regions" would fail (110 characters)
//
// The rule properly handles Unicode characters by counting code points rather than bytes.
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
//
// Parameters:
//   - subject: The commit subject line to validate
//   - maxLength: The maximum allowed character count (0 means use default)
//
// Returns:
//   - A SubjectLength instance with validation results
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
