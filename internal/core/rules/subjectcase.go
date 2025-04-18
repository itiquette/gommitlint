// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package rules

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/itiquette/gommitlint/internal/domain"
	appErrors "github.com/itiquette/gommitlint/internal/errors"
)

// subjectCaseFirstWordRegex is the regular expression used to find the first word in a commit.
var subjectCaseFirstWordRegex = regexp.MustCompile(`^\s*([a-zA-Z0-9]+)`)

// Format: type(scope)!: description.
var subjectCaseRegex = regexp.MustCompile(`^(\w+)(?:\(([\w,/-]+)\))?(!)?:[ ](.+)$`)

// SubjectCaseRule enforces the case of the first word in the subject.
// This rule helps ensure commit messages follow a consistent style by validating
// the first letter's case based on the project's convention.
//
// For conventional commits (e.g., "feat(scope): add feature"), it checks the
// capitalization of the first word after the "type(scope): " prefix.
//
// For non-conventional commits, it simply checks the first word of the subject.
type SubjectCaseRule struct {
	*BaseRule
	isConventional bool   // Whether to treat as a conventional commit format
	caseChoice     string // The desired case ("upper", "lower", or "ignore")
	allowNonAlpha  bool   // Whether to allow non-alphabetic first characters
	firstWord      string // Store for verbose output
	firstLetter    rune   // Store for verbose output
}

// SubjectCaseOption is a function that modifies a SubjectCaseRule.
type SubjectCaseOption func(*SubjectCaseRule)

// WithCaseChoice sets the desired case for the subject.
func WithCaseChoice(caseChoice string) SubjectCaseOption {
	return func(rule *SubjectCaseRule) {
		if caseChoice == "upper" || caseChoice == "lower" || caseChoice == "ignore" {
			rule.caseChoice = caseChoice
		}
	}
}

// WithSubjectCaseCommitFormat configures whether to treat as a conventional commit.
func WithSubjectCaseCommitFormat(isConventional bool) SubjectCaseOption {
	return func(rule *SubjectCaseRule) {
		rule.isConventional = isConventional
	}
}

// WithAllowNonAlpha sets whether to allow non-alphabetic first characters.
func WithAllowNonAlpha(allow bool) SubjectCaseOption {
	return func(rule *SubjectCaseRule) {
		rule.allowNonAlpha = allow
	}
}

// NewSubjectCaseRule creates a new SubjectCaseRule with the specified options.
func NewSubjectCaseRule(options ...SubjectCaseOption) *SubjectCaseRule {
	rule := &SubjectCaseRule{
		BaseRule:       NewBaseRule("SubjectCase"),
		isConventional: false,
		caseChoice:     "lower", // Default to lowercase
		allowNonAlpha:  false,   // Default to requiring alphabetic first characters
	}

	// Apply options
	for _, option := range options {
		option(rule)
	}

	return rule
}

// Name method is inherited from BaseRule.

// addError adds a structured validation error.
func (r *SubjectCaseRule) addError(code appErrors.ValidationErrorCode, message string, context map[string]string) {
	// Add the error with context if provided
	if context != nil {
		r.AddErrorWithContext(code, message, context)
	} else {
		r.AddErrorWithCode(code, message)
	}
}

// The Errors method is inherited from BaseRule.

// Validate validates the commit subject case.
func (r *SubjectCaseRule) Validate(commit *domain.CommitInfo) []appErrors.ValidationError {
	// Reset errors and state
	r.ClearErrors()

	subject := commit.Subject

	// Check for empty message first
	if subject == "" {
		r.AddErrorWithCode(
			appErrors.ErrEmptyDescription,
			"subject is empty",
		)

		return r.Errors()
	}

	// Extract first word
	firstWord, err := extractSubjectCaseFirstWord(r.isConventional, subject)
	if err != nil {
		// Determine the specific error type
		var errorCode appErrors.ValidationErrorCode

		if r.isConventional {
			if strings.Contains(err.Error(), "missing subject after type") {
				errorCode = appErrors.ErrMissingSubject
			} else {
				errorCode = appErrors.ErrInvalidFormat
			}
		} else {
			errorCode = appErrors.ErrInvalidFormat
		}

		r.addError(
			errorCode,
			err.Error(),
			map[string]string{
				"subject":         subject,
				"is_conventional": strconv.FormatBool(r.isConventional),
			},
		)

		return r.Errors()
	}

	// Store first word for verbose output
	r.firstWord = firstWord

	// Decode first rune
	first, size := utf8.DecodeRuneInString(firstWord)
	r.firstLetter = first

	if first == utf8.RuneError && size == 0 {
		r.addError(
			appErrors.ErrUnknown,
			"subject does not start with valid UTF-8 text",
			map[string]string{
				"subject": subject,
			},
		)

		return r.Errors()
	}

	// If AllowNonAlpha is enabled and the first character isn't a letter, skip case check
	if r.allowNonAlpha && !unicode.IsLetter(first) {
		return r.Errors()
	}

	// Validate case
	var valid bool

	var errorCode = appErrors.ErrSubjectCase

	switch r.caseChoice {
	case "upper":
		valid = unicode.IsUpper(first)
	case "lower":
		valid = unicode.IsLower(first)
	case "ignore":
		valid = true
	default:
		valid = unicode.IsLower(first) // Default to lowercase if unspecified
	}

	if !valid {
		r.addError(
			errorCode,
			"commit subject case is not "+r.caseChoice,
			map[string]string{
				"expected_case": r.caseChoice,
				"actual_case":   map[bool]string{true: "upper", false: "lower"}[unicode.IsUpper(first)],
				"first_word":    firstWord,
				"subject":       subject,
			},
		)
	}

	return r.Errors()
}

// Result returns a concise result message.
func (r *SubjectCaseRule) Result() string {
	if r.HasErrors() {
		// Check for case-specific error
		errors := r.Errors()
		if len(errors) > 0 {
			//nolint:exhaustive // Only handling relevant error codes
			switch appErrors.ValidationErrorCode(errors[0].Code) {
			case appErrors.ErrSubjectCase:
				return "Subject should start with " + r.caseChoice
			case appErrors.ErrEmptyDescription, appErrors.ErrEmptyMessage:
				return "Subject is empty"
			case appErrors.ErrInvalidFormat:
				return "Invalid format"
			default:
				return "Subject case validation failed"
			}
		}

		return "Invalid subject case"
	}

	return "Subject case is correct"
}

// VerboseResult returns a more detailed explanation for verbose mode.
func (r *SubjectCaseRule) VerboseResult() string {
	if r.HasErrors() {
		// Get errors
		errors := r.Errors()
		if len(errors) == 0 {
			return "Unknown error"
		}

		// errors[0] is already a ValidationError, so no need for type assertion
		validationErr := errors[0]

		// We're deliberately not handling all possible validation error codes here,
		// just the ones that can be generated by this specific rule.
		//nolint:exhaustive
		switch appErrors.ValidationErrorCode(validationErr.Code) {
		case appErrors.ErrEmptyDescription, appErrors.ErrEmptyMessage:
			return "Commit subject is empty. Cannot validate case."

		case appErrors.ErrUnknown:
			errMsg := validationErr.Message
			if strings.Contains(errMsg, "UTF-8") {
				return "Subject doesn't start with valid UTF-8 text. Check for encoding issues."
			}

		case appErrors.ErrInvalidFormat:
			errMsg := validationErr.Message
			if strings.Contains(errMsg, "conventional commit format") {
				return "Invalid conventional commit format. Expected 'type(scope): subject'."
			} else if strings.Contains(errMsg, "no valid first word") {
				return "Invalid commit format. Subject should start with a valid word."
			}

		case appErrors.ErrMissingSubject:
			return "Missing subject after conventional commit prefix."

		case appErrors.ErrSubjectCase:
			if r.caseChoice == "upper" {
				return "First letter should be uppercase. Found '" + string(r.firstLetter) + "' in '" + r.firstWord + "'."
			} else if r.caseChoice == "lower" {
				return "First letter should be lowercase. Found '" + string(r.firstLetter) + "' in '" + r.firstWord + "'."
			}
		}

		// Default case
		return validationErr.Error()
	}

	// Construct a detailed success message
	var formatType string
	if r.isConventional {
		formatType = "conventional commit"
	} else {
		formatType = "standard commit"
	}

	return "Subject has correct " + r.caseChoice + "case for " + formatType + ": '" + r.firstWord + "'"
}

// Help returns a description of how to fix the rule violation.
func (r *SubjectCaseRule) Help() string {
	if !r.HasErrors() {
		return "No errors to fix"
	}

	// Get errors
	errors := r.Errors()
	if len(errors) == 0 {
		return "No specific guidance available"
	}

	// errors[0] is already a ValidationError, so no need for type assertion
	validationErr := errors[0]
	// We're deliberately not handling all possible validation error codes here,
	// just the ones that can be generated by this specific rule.
	//nolint:exhaustive
	switch appErrors.ValidationErrorCode(validationErr.Code) {
	case appErrors.ErrEmptyDescription, appErrors.ErrEmptyMessage:
		return "Provide a non-empty commit message subject with appropriate capitalization."

	case appErrors.ErrUnknown:
		if strings.Contains(validationErr.Message, "UTF-8") {
			return "Ensure your commit message begins with valid UTF-8 text. Remove any invalid characters from the start."
		}

	case appErrors.ErrInvalidFormat:
		if strings.Contains(validationErr.Message, "conventional commit format") {
			return "Format your commit message according to the Conventional Commits specification: type(scope): subject\n" +
				"Example: feat(auth): Add login feature"
		}

		return "Ensure your commit message starts with a valid word following proper capitalization rules."

	case appErrors.ErrMissingSubject:
		return "Add a subject after the type(scope): prefix in your conventional commit message.\n" +
			"Example: fix(api): Resolve timeout issue"

	case appErrors.ErrSubjectCase:
		if r.caseChoice == "upper" {
			return "Capitalize the first letter of your commit subject.\n" +
				"Example for conventional commit: feat(auth): Add feature (not 'add feature')\n" +
				"Example for standard commit: Add feature (not 'add feature')"
		} else if r.caseChoice == "lower" {
			return "Use lowercase for the first letter of your commit subject.\n" +
				"Example for conventional commit: feat(auth): add feature (not 'Add feature')\n" +
				"Example for standard commit: add feature (not 'Add feature')"
		}
	}

	// Default help
	return "Check the capitalization of the first letter in your commit message subject according to your project's guidelines."
}

// extractSubjectCaseFirstWord extracts the first word from the commit message.
//
// Parameters:
//   - isConventional: Whether to parse as a conventional commit
//   - subject: The commit subject line
//
// For conventional commits, it extracts the first word after the "type(scope): " prefix.
// For standard commits, it extracts the first word of the subject.
//
// Returns:
//   - The first word to validate
//   - Any error encountered during extraction
func extractSubjectCaseFirstWord(isConventional bool, subject string) (string, error) {
	if isConventional {
		// For conventional commits, try to extract the part after type(scope):
		matches := subjectCaseRegex.FindStringSubmatch(subject)

		// Validate conventional commit format
		if len(matches) != 5 {
			return "", errors.New("invalid conventional commit format")
		}

		msg := matches[4]
		if msg == "" {
			return "", errors.New("missing subject after type")
		}

		matches = subjectCaseFirstWordRegex.FindStringSubmatch(msg)
		if len(matches) == 0 {
			return "", errors.New("no valid first word found")
		}

		return matches[1], nil
	}

	// For non-conventional commits
	matches := subjectCaseFirstWordRegex.FindStringSubmatch(subject)
	if len(matches) == 0 {
		return "", errors.New("no valid first word found")
	}

	return matches[1], nil
}
