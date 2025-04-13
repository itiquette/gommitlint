// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package rule

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/itiquette/gommitlint/internal/model"
)

// DefaultInvalidSuffixes is the default set of characters that should not appear
// at the end of a commit subject line.
const DefaultInvalidSuffixes = ".,;:!?"

// SubjectSuffix enforces that the last character of the commit subject line
// is not in a specified set of invalid suffixes.
//
// This rule helps ensure commit messages maintain a consistent format by
// preventing subjects from ending with unwanted characters like periods,
// commas, or other punctuation marks that can affect readability and
// automated processing of commit messages.
//
// The rule validates the last character of the subject line against a set of
// specified invalid characters. If the last character is found in this set,
// the rule fails.
//
// Examples:
//
//   - With invalidSuffixes=".,;:":
//     "Add new feature" would pass
//     "Add new feature." would fail (ends with period)
//     "Implement login," would fail (ends with comma)
//
//   - With invalidSuffixes="!?":
//     "Fix critical bug" would pass
//     "Fix critical bug!" would fail (ends with exclamation)
//
// The rule also handles empty subjects and invalid UTF-8 characters.
type SubjectSuffix struct {
	lastChar        rune
	errors          []*model.ValidationError
	invalidSuffixes string // Store for verbose output
}

// Name returns the name of the rule.
func (rule SubjectSuffix) Name() string {
	return "SubjectSuffix"
}

// Result returns a concise rule message.
func (rule SubjectSuffix) Result() string {
	if len(rule.errors) > 0 {
		return "Invalid subject suffix"
	}

	return "Valid subject suffix"
}

// VerboseResult returns a more detailed explanation for verbose mode.
func (rule SubjectSuffix) VerboseResult() string {
	if len(rule.errors) > 0 {
		switch rule.errors[0].Code {
		case "subject_empty":
			return "Subject is empty. Cannot validate suffix."
		case "invalid_utf8":
			return "Subject ends with invalid UTF-8 character sequence."
		case "invalid_suffix":
			return fmt.Sprintf("Subject ends with forbidden character '%s'. Avoid ending with any of these: %s",
				string(rule.lastChar),
				strings.Join(strings.Split(rule.invalidSuffixes, ""), ", "))
		default:
			return rule.errors[0].Error()
		}
	}

	return "Subject ends with valid character '" + string(rule.lastChar) + "' (not in invalid set: " +
		strings.Join(strings.Split(rule.invalidSuffixes, ""), ", ") + ")"
}

// addError adds a structured validation error.
func (rule *SubjectSuffix) addError(code, message string, context map[string]string) {
	err := model.NewValidationError("SubjectSuffix", code, message)

	// Add any context values
	for key, value := range context {
		_ = err.WithContext(key, value)
	}

	rule.errors = append(rule.errors, err)
}

// Errors returns validation errors.
func (rule SubjectSuffix) Errors() []*model.ValidationError {
	return rule.errors
}

// Help returns a description of how to fix the rule violation.
func (rule SubjectSuffix) Help() string {
	const noErrMsg = "No errors to fix"
	if len(rule.errors) == 0 {
		return noErrMsg
	}

	// Check for specific error codes
	if len(rule.errors) > 0 {
		switch rule.errors[0].Code {
		case "subject_empty":
			return "Provide a non-empty subject line for your commit message"
		case "invalid_utf8":
			return "Ensure your commit message contains only valid UTF-8 characters"
		case "invalid_suffix":
			var invalidSuffixes string
			if suffixes, ok := rule.errors[0].Context["invalid_suffixes"]; ok {
				invalidSuffixes = suffixes
			} else {
				invalidSuffixes = DefaultInvalidSuffixes
			}

			return fmt.Sprintf("Remove the punctuation or special character from the end of your subject line. "+
				"The subject should end with a letter or number, not punctuation like: %s", invalidSuffixes)
		}
	}

	return "Review and fix your commit message subject line according to the guidelines"
}

// ValidateSubjectSuffix checks if the subject ends with a character in the invalidSuffixes set.
// If invalidSuffixes is empty, it uses the DefaultInvalidSuffixes.
//
// Parameters:
//   - subject: The commit subject line to validate
//   - invalidSuffixes: A string containing characters considered invalid at the end of a subject
//
// Returns:
//   - A SubjectSuffix instance with validation results
func ValidateSubjectSuffix(subject, invalidSuffixes string) *SubjectSuffix {
	if invalidSuffixes == "" {
		invalidSuffixes = DefaultInvalidSuffixes
	}

	rule := &SubjectSuffix{
		invalidSuffixes: invalidSuffixes,
	}

	if subject == "" {
		rule.addError(
			"subject_empty",
			"subject is empty",
			map[string]string{
				"subject": subject,
			},
		)

		return rule
	}

	lastChar, size := utf8.DecodeLastRuneInString(subject)
	rule.lastChar = lastChar

	// Check for invalid UTF-8
	if lastChar == utf8.RuneError && size == 0 {
		rule.addError(
			"invalid_utf8",
			"subject does not end with valid UTF-8 text",
			map[string]string{
				"subject": subject,
			},
		)

		return rule
	}

	// Check if the last character is in the invalid suffix set
	if strings.ContainsRune(invalidSuffixes, lastChar) {
		rule.addError(
			"invalid_suffix",
			fmt.Sprintf("subject has invalid suffix %q (invalid suffixes: %q)", string(lastChar), invalidSuffixes),
			map[string]string{
				"subject":          subject,
				"last_char":        string(lastChar),
				"invalid_suffixes": invalidSuffixes,
			},
		)
	}

	return rule
}
