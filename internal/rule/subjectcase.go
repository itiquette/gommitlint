// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package rule

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
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
	subjectCase string
	errors      []error
}

// Name returns the validation rule name.
func (SubjectCase) Name() string {
	return "SubjectCase"
}

// Result returns the rule message.
func (rule *SubjectCase) Result() string {
	if len(rule.errors) > 0 {
		return rule.errors[0].Error()
	}

	return "Subject case is valid"
}

// Errors returns any validation errors.
func (rule SubjectCase) Errors() []error {
	return rule.errors
}

// Help returns a description of how to fix the rule violation.
func (rule SubjectCase) Help() string {
	if len(rule.errors) == 0 {
		return "No errors to fix"
	}

	errMsg := rule.errors[0].Error()

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
	rule := &SubjectCase{subjectCase: caseChoice}

	if subject == "" {
		rule.addErrorf("subject is empty")

		return rule
	}

	// Extract first word
	firstWord, err := extractFirstWord(isConventional, subject)
	if err != nil {
		rule.addErrorf("%s", err.Error())

		return rule
	}

	// Decode first rune
	first, size := utf8.DecodeRuneInString(firstWord)
	if first == utf8.RuneError && size == 0 {
		rule.addErrorf("subject does not start with valid UTF-8 text")

		return rule
	}

	// Validate case
	var valid bool

	switch caseChoice {
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
		rule.addErrorf("commit subject case is not %s", caseChoice)
	}

	return rule
}

// addErrorf adds an error to the rule's errors slice.
func (rule *SubjectCase) addErrorf(format string, args ...interface{}) {
	rule.errors = append(rule.errors, fmt.Errorf(format, args...))
}
