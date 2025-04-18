// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package rules

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/itiquette/gommitlint/internal/domain"
	appErrors "github.com/itiquette/gommitlint/internal/errors"
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
	errors          []appErrors.ValidationError
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
		errors:          make([]appErrors.ValidationError, 0),
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
func (r *SubjectSuffixRule) Validate(commit *domain.CommitInfo) []appErrors.ValidationError {
	// Reset errors
	r.errors = make([]appErrors.ValidationError, 0)

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

// Result returns a concise validation result.
func (r *SubjectSuffixRule) Result() string {
	if len(r.errors) == 0 {
		return "Valid subject suffix"
	}

	return "Invalid subject suffix"
}

// VerboseResult returns a more detailed result message.
func (r *SubjectSuffixRule) VerboseResult() string {
	if len(r.errors) == 0 {
		return "Subject ends with valid character"
	}

	// If we have an error, provide details based on the error type
	if len(r.errors) > 0 {
		code := r.errors[0].Code
		if code == "subject_empty" || code == string(appErrors.ErrMissingSubject) {
			return "Subject is empty"
		}

		if code == "invalid_utf8" || code == string(appErrors.ErrInvalidFormat) {
			return "Subject contains invalid UTF-8 characters"
		}

		// If we have a more specific error message from the validation, use it
		message := r.errors[0].Message
		if message != "" {
			return message
		}
	}

	// Default message
	return fmt.Sprintf("Subject ends with invalid character (invalid suffixes: %s)", r.invalidSuffixes)
}

// Help returns guidance on how to fix rule violations.
func (r *SubjectSuffixRule) Help() string {
	if len(r.errors) == 0 {
		return "No errors to fix"
	}

	// Check for specific error codes and provide appropriate help messages
	if len(r.errors) > 0 {
		code := r.errors[0].Code

		// Check for missing subject errors
		if code == string(appErrors.ErrMissingSubject) || code == "subject_empty" {
			return "Provide a non-empty subject line for your commit message"
		}

		// Check for invalid UTF-8 errors
		if code == string(appErrors.ErrInvalidFormat) || code == "invalid_utf8" {
			return "Ensure your commit message contains only valid UTF-8 characters"
		}

		// Check for invalid suffix errors
		if code == string(appErrors.ErrSubjectSuffix) || code == "invalid_suffix" {
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
func (r *SubjectSuffixRule) Errors() []appErrors.ValidationError {
	return r.errors
}

// addError adds a structured validation error.
func (r *SubjectSuffixRule) addError(code, message string, context map[string]string) {
	// Create the validation error directly with the app errors package
	var err appErrors.ValidationError

	switch code {
	case "invalid_suffix":
		// Map to appropriate error code
		err = appErrors.New(
			r.Name(),
			appErrors.ErrSubjectSuffix,
			message,
			appErrors.WithContextMap(context),
		)
	case "subject_empty":
		// Map to missing subject code
		err = appErrors.New(
			r.Name(),
			appErrors.ErrMissingSubject,
			message,
			appErrors.WithHelp("Provide a non-empty subject"),
			appErrors.WithContextMap(context),
		)
	case "invalid_utf8":
		// Map to invalid format code
		err = appErrors.New(
			r.Name(),
			appErrors.ErrInvalidFormat,
			message,
			appErrors.WithHelp("Ensure your subject contains valid UTF-8 characters"),
			appErrors.WithContextMap(context),
		)
	default:
		// Fall back to unknown error code
		err = appErrors.New(
			r.Name(),
			appErrors.ErrUnknown,
			message,
			appErrors.WithContextMap(context),
		)
	}

	r.errors = append(r.errors, err)
}
