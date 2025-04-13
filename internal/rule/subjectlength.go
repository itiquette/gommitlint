// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package rule

import (
	"fmt"
	"strconv"
	"unicode/utf8"

	"github.com/itiquette/gommitlint/internal/model"
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
	maxLength     int
	errors        []*model.ValidationError
}

// Name returns the rule name.
func (rule SubjectLength) Name() string {
	return "SubjectLength"
}

// Result returns a concise validation result.
func (rule SubjectLength) Result() string {
	if len(rule.errors) > 0 {
		return "Subject too long"
	}

	return "Subject length OK"
}

// VerboseResult returns a more detailed explanation for verbose mode.
func (rule SubjectLength) VerboseResult() string {
	if len(rule.errors) > 0 {
		return fmt.Sprintf("Subject exceeds maximum length: %d characters (limit: %d characters)",
			rule.subjectLength, rule.maxLength)
	}

	// Calculate how close we are to the limit
	percentUsed := (float64(rule.subjectLength) / float64(rule.maxLength)) * 100

	if percentUsed < 50 {
		return fmt.Sprintf("Subject length (%d chars) is well within the limit of %d characters",
			rule.subjectLength, rule.maxLength)
	} else if percentUsed < 80 {
		return fmt.Sprintf("Subject length (%d chars) is within the limit of %d characters",
			rule.subjectLength, rule.maxLength)
	}

	return fmt.Sprintf("Subject length (%d chars) is close to the limit of %d characters",
		rule.subjectLength, rule.maxLength)
}

// addError adds a structured validation error.
func (rule *SubjectLength) addError(code, message string, context map[string]string) {
	err := model.NewValidationError("SubjectLength", code, message)

	// Add any context values
	for key, value := range context {
		_ = err.WithContext(key, value)
	}

	rule.errors = append(rule.errors, err)
}

// Errors returns validation errors.
func (rule SubjectLength) Errors() []*model.ValidationError {
	return rule.errors
}

// Help returns a description of how to fix the rule violation.
func (rule SubjectLength) Help() string {
	const noErrMsg = "No errors to fix"
	if len(rule.errors) == 0 {
		return noErrMsg
	}

	// Check for specific error codes
	if len(rule.errors) > 0 && rule.errors[0].Code == "subject_too_long" {
		var maxLength string
		if maxVal, ok := rule.errors[0].Context["max_length"]; ok {
			maxLength = maxVal
		} else {
			maxLength = strconv.Itoa(DefaultMaxCommitSubjectLength)
		}

		return fmt.Sprintf("Shorten your commit message subject line to be more concise.\n"+
			"A good subject should be brief but descriptive, ideally under 50 characters "+
			"and no more than %s characters.\nConsider using the commit body for additional details.", maxLength)
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
		maxLength:     maxLength,
	}

	// Validate length
	if subjectLength > maxLength {
		rule.addError(
			"subject_too_long",
			fmt.Sprintf("subject too long: %d characters (maximum allowed: %d)",
				subjectLength,
				maxLength),
			map[string]string{
				"actual_length": strconv.Itoa(subjectLength),
				"max_length":    strconv.Itoa(maxLength),
				"subject":       subject,
			},
		)
	}

	return rule
}
