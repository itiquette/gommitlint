// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package rule

import (
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/itiquette/gommitlint/internal/model"
)

// SubjectCase enforces the case of the first word in the subject.
// This rule helps ensure commit messages follow a consistent style by validating
// the first letter's case based on the project's convention.
//
// For conventional commits (e.g., "feat(scope): add feature"), it checks the
// capitalization of the first word after the "type(scope): " prefix.
//
// For non-conventional commits, it simply checks the first word of the subject.
//
// Examples:
//
//   - With caseChoice="upper" for a conventional commit:
//     "feat(auth): Add new login" would pass
//     "feat(auth): add new login" would fail
//
//   - With caseChoice="lower" for a non-conventional commit:
//     "add new feature" would pass
//     "Add new feature" would fail
type SubjectCase struct {
	subjectCase    string
	errors         []*model.ValidationError
	isConventional bool   // Store for verbose output
	firstWord      string // Store for verbose output
	firstLetter    rune   // Store for verbose output
}

// Name returns the validation rule name.
func (SubjectCase) Name() string {
	return "SubjectCase"
}

// Result returns a concise rule message.
func (rule *SubjectCase) Result() string {
	if len(rule.errors) > 0 {
		return "Invalid subject case"
	}

	return "Valid subject case"
}

// VerboseResult returns a more detailed explanation for verbose mode.
func (rule *SubjectCase) VerboseResult() string {
	if len(rule.errors) > 0 {
		switch rule.errors[0].Code {
		case "empty_subject":
			return "Commit subject is empty. Cannot validate case."
		case "invalid_utf8":
			return "Subject doesn't start with valid UTF-8 text. Check for encoding issues."
		case "invalid_conventional_format":
			return "Invalid conventional commit format. Expected 'type(scope): subject'."
		case "invalid_format":
			return "Invalid commit format. Subject should start with a valid word."
		case "missing_conventional_subject":
			return "Missing subject after conventional commit prefix."
		case "wrong_case_upper":
			return "First letter should be uppercase. Found '" + string(rule.firstLetter) + "' in '" + rule.firstWord + "'."
		case "wrong_case_lower":
			return "First letter should be lowercase. Found '" + string(rule.firstLetter) + "' in '" + rule.firstWord + "'."
		default:
			return rule.errors[0].Error()
		}
	}

	// Construct a detailed success message
	var formatType string
	if rule.isConventional {
		formatType = "conventional commit"
	} else {
		formatType = "standard commit"
	}

	return "Subject has correct " + rule.subjectCase + "case for " + formatType + ": '" + rule.firstWord + "'"
}

// Errors returns any validation errors.
func (rule SubjectCase) Errors() []*model.ValidationError {
	return rule.errors
}

// addError adds a structured validation error.
func (rule *SubjectCase) addError(code, message string, context map[string]string) {
	err := model.NewValidationError("SubjectCase", code, message)

	// Add any context values
	for key, value := range context {
		_ = err.WithContext(key, value)
	}

	rule.errors = append(rule.errors, err)
}

// Help returns a description of how to fix the rule violation.
func (rule SubjectCase) Help() string {
	if len(rule.errors) == 0 {
		return "No errors to fix"
	}

	// Check for specific error codes
	if len(rule.errors) > 0 {
		switch rule.errors[0].Code {
		case "empty_subject":
			return "Provide a non-empty commit message subject with appropriate capitalization."

		case "invalid_utf8":
			return "Ensure your commit message begins with valid UTF-8 text. Remove any invalid characters from the start."

		case "invalid_conventional_format":
			return "Format your commit message according to the Conventional Commits specification: type(scope): subject\n" +
				"Example: feat(auth): Add login feature"

		case "invalid_format":
			return "Ensure your commit message starts with a valid word following proper capitalization rules."

		case "missing_conventional_subject":
			return "Add a subject after the type(scope): prefix in your conventional commit message.\n" +
				"Example: fix(api): Resolve timeout issue"

		case "wrong_case_upper":
			return "Capitalize the first letter of your commit subject.\n" +
				"Example for conventional commit: feat(auth): Add feature (not 'add feature')\n" +
				"Example for standard commit: Add feature (not 'add feature')"

		case "wrong_case_lower":
			return "Use lowercase for the first letter of your commit subject.\n" +
				"Example for conventional commit: feat(auth): add feature (not 'Add feature')\n" +
				"Example for standard commit: add feature (not 'Add feature')"
		}
	}

	// Fallback to checking the error message for backward compatibility
	errMsg := rule.errors[0].Message

	if strings.Contains(errMsg, "does not start with valid UTF-8") {
		return "Ensure your commit message begins with valid UTF-8 text. Remove any invalid characters from the start."
	}

	if strings.Contains(errMsg, "invalid conventional commit format") {
		return "Format your commit message according to the Conventional Commits specification: type(scope): subject\n" +
			"Example: feat(auth): Add login feature"
	}

	if strings.Contains(errMsg, "missing subject after type") {
		return "Add a subject after the type(scope): prefix in your conventional commit message.\n" +
			"Example: fix(api): Resolve timeout issue"
	}

	if strings.Contains(errMsg, "commit subject case is not upper") {
		if rule.subjectCase == "upper" {
			return "Capitalize the first letter of your commit subject.\n" +
				"Example for conventional commit: feat(auth): Add feature (not 'add feature')\n" +
				"Example for standard commit: Add feature (not 'add feature')"
		}
	}

	if strings.Contains(errMsg, "commit subject case is not lower") {
		if rule.subjectCase == "lower" {
			return "Use lowercase for the first letter of your commit subject.\n" +
				"Example for conventional commit: feat(auth): add feature (not 'Add feature')\n" +
				"Example for standard commit: add feature (not 'Add feature')"
		}
	}

	return "Check the capitalization of the first letter in your commit message subject according to your project's guidelines."
}

// ValidateSubjectCase checks the subject case based on the specified case choice.
// It enforces capitalization rules for both conventional and standard commit messages.
//
// Parameters:
//   - subject: The commit subject line to validate
//   - caseChoice: The desired case ("upper" or "lower")
//   - isConventional: Whether to treat as a conventional commit format
//
// For conventional commits (format: "type(scope): subject"), it validates the
// first letter of the subject part after the colon, ignoring the type and scope.
// For example, in "feat(auth): Add login", it validates "Add".
//
// For standard commits, it validates the first letter of the subject.
// For example, in "Add login feature", it validates "Add".
//
// The caseChoice parameter accepts:
//   - "upper": Requires the first letter to be uppercase
//   - "lower": Requires the first letter to be lowercase
//   - "ignore": Ignores rule
//
// If an invalid caseChoice is provided, it defaults to "lower".
//
// Returns:
//   - A SubjectCase instance with validation results
func ValidateSubjectCase(subject, caseChoice string, isConventional bool) *SubjectCase {
	rule := &SubjectCase{
		subjectCase:    caseChoice,
		isConventional: isConventional,
	}

	if subject == "" {
		rule.addError(
			"empty_subject",
			"subject is empty",
			nil,
		)

		return rule
	}

	// Extract first word
	firstWord, err := extractFirstWord(isConventional, subject)
	if err != nil {
		// Determine the specific error type
		errorCode := "invalid_format"

		if isConventional {
			if strings.Contains(err.Error(), "missing subject after type") {
				errorCode = "missing_conventional_subject"
			} else {
				errorCode = "invalid_conventional_format"
			}
		}

		rule.addError(
			errorCode,
			err.Error(),
			map[string]string{
				"subject":         subject,
				"is_conventional": strconv.FormatBool(isConventional),
			},
		)

		return rule
	}

	// Store first word for verbose output
	rule.firstWord = firstWord

	// Decode first rune
	first, size := utf8.DecodeRuneInString(firstWord)
	rule.firstLetter = first

	if first == utf8.RuneError && size == 0 {
		rule.addError(
			"invalid_utf8",
			"subject does not start with valid UTF-8 text",
			map[string]string{
				"subject": subject,
			},
		)

		return rule
	}

	// Validate case
	var valid bool

	var errorCode string

	switch caseChoice {
	case "upper":
		valid = unicode.IsUpper(first)
		errorCode = "wrong_case_upper"
	case "lower":
		valid = unicode.IsLower(first)
		errorCode = "wrong_case_lower"
	case "ignore":
		valid = true
	default:
		valid = unicode.IsLower(first) // Default to lowercase if unspecified
		errorCode = "wrong_case_lower"
	}

	if !valid {
		rule.addError(
			errorCode,
			"commit subject case is not "+caseChoice,
			map[string]string{
				"expected_case": caseChoice,
				"actual_case":   map[bool]string{true: "upper", false: "lower"}[unicode.IsUpper(first)],
				"first_word":    firstWord,
				"subject":       subject,
			},
		)
	}

	return rule
}
