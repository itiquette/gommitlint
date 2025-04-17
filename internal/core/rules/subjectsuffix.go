// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package rules

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/itiquette/gommitlint/internal/domain"
)

// DefaultInvalidSuffixes is the default set of characters that should not appear
// at the end of a commit subject line.
const DefaultInvalidSuffixes = ".,;:!?"

// SubjectSuffixRule enforces that the last character of the commit subject line
// is not in a specified set of invalid suffixes.
//
// This rule helps ensure commit messages maintain a consistent format by
// preventing subjects from ending with unwanted characters like periods,
// commas, or other punctuation marks that can affect readability and
// automated processing of commit messages.
type SubjectSuffixRule struct {
	errors          []*domain.ValidationError
	lastChar        rune
	invalidSuffixes string
}

// SubjectSuffixOption is a function that modifies a SubjectSuffixRule.
type SubjectSuffixOption func(*SubjectSuffixRule)

// WithInvalidSuffixes sets custom invalid suffix characters.
func WithInvalidSuffixes(suffixes string) SubjectSuffixOption {
	return func(rule *SubjectSuffixRule) {
		rule.invalidSuffixes = suffixes
	}
}

// NewSubjectSuffixRule creates a new SubjectSuffixRule with the specified options.
func NewSubjectSuffixRule(options ...SubjectSuffixOption) *SubjectSuffixRule {
	rule := &SubjectSuffixRule{
		errors:          make([]*domain.ValidationError, 0),
		invalidSuffixes: DefaultInvalidSuffixes,
	}

	// Apply options
	for _, option := range options {
		option(rule)
	}

	// If invalid suffixes is empty, use the default
	if rule.invalidSuffixes == "" {
		rule.invalidSuffixes = DefaultInvalidSuffixes
	}

	return rule
}

// Name returns the rule identifier.
func (r *SubjectSuffixRule) Name() string {
	return "SubjectSuffix"
}

// Validate validates that the subject doesn't end with invalid characters.
func (r *SubjectSuffixRule) Validate(commit *domain.CommitInfo) []*domain.ValidationError {
	// Reset errors
	r.errors = make([]*domain.ValidationError, 0)

	subject := commit.Subject

	if subject == "" {
		r.addError(
			"subject_empty",
			"subject is empty",
			map[string]string{
				"subject": subject,
			},
		)

		return r.errors
	}

	lastChar, size := utf8.DecodeLastRuneInString(subject)
	r.lastChar = lastChar

	// Check for invalid UTF-8
	if lastChar == utf8.RuneError && size == 0 {
		r.addError(
			"invalid_utf8",
			"subject does not end with valid UTF-8 text",
			map[string]string{
				"subject": subject,
			},
		)

		return r.errors
	}

	// Check if the last character is in the invalid suffix set
	if strings.ContainsRune(r.invalidSuffixes, lastChar) {
		r.addError(
			"invalid_suffix",
			fmt.Sprintf("subject has invalid suffix %q (invalid suffixes: %q)", string(lastChar), r.invalidSuffixes),
			map[string]string{
				"subject":          subject,
				"last_char":        string(lastChar),
				"invalid_suffixes": r.invalidSuffixes,
			},
		)
	}

	return r.errors
}

// Result returns a concise result message.
func (r *SubjectSuffixRule) Result() string {
	if len(r.errors) > 0 {
		return "Invalid subject suffix"
	}

	return "Valid subject suffix"
}

// VerboseResult returns a more detailed explanation for verbose mode.
func (r *SubjectSuffixRule) VerboseResult() string {
	if len(r.errors) > 0 {
		switch r.errors[0].Code {
		case "subject_empty":
			return "Subject is empty. Cannot validate suffix."
		case "invalid_utf8":
			return "Subject ends with invalid UTF-8 character sequence."
		case "invalid_suffix":
			return fmt.Sprintf("Subject ends with forbidden character '%s'. Avoid ending with any of these: %s",
				string(r.lastChar),
				strings.Join(strings.Split(r.invalidSuffixes, ""), ", "))
		default:
			return r.errors[0].Error()
		}
	}

	return "Subject ends with valid character '" + string(r.lastChar) + "' (not in invalid set: " +
		strings.Join(strings.Split(r.invalidSuffixes, ""), ", ") + ")"
}

// Help returns a description of how to fix the rule violation.
func (r *SubjectSuffixRule) Help() string {
	if len(r.errors) == 0 {
		return "No errors to fix"
	}

	// Check for specific error codes
	if len(r.errors) > 0 {
		switch r.errors[0].Code {
		case "subject_empty":
			return "Provide a non-empty subject line for your commit message"
		case "invalid_utf8":
			return "Ensure your commit message contains only valid UTF-8 characters"
		case "invalid_suffix":
			var invalidSuffixes string
			if suffixes, ok := r.errors[0].Context["invalid_suffixes"]; ok {
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

// Errors returns all validation errors.
func (r *SubjectSuffixRule) Errors() []*domain.ValidationError {
	return r.errors
}

// addError adds a structured validation error.
func (r *SubjectSuffixRule) addError(code, message string, context map[string]string) {
	err := domain.NewValidationError(r.Name(), code, message)

	// Add any context values
	for key, value := range context {
		_ = err.WithContext(key, value)
	}

	r.errors = append(r.errors, err)
}
