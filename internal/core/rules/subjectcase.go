// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package rules

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/itiquette/gommitlint/internal/domain"
	appErrors "github.com/itiquette/gommitlint/internal/errors"
)

// SubjectCaseRule validates the case style of commit subjects.
type SubjectCaseRule struct {
	name          string
	caseChoice    string
	checkCommit   bool
	allowNonAlpha bool
}

// SubjectCaseOption configures a SubjectCaseRule.
type SubjectCaseOption func(SubjectCaseRule) SubjectCaseRule

// WithCaseChoice sets the required case style for subjects.
func WithCaseChoice(caseStyle string) SubjectCaseOption {
	return func(r SubjectCaseRule) SubjectCaseRule {
		result := r
		result.caseChoice = caseStyle

		return result
	}
}

// WithSubjectCaseCommitFormat enables checking conventional commit format.
func WithSubjectCaseCommitFormat(check bool) SubjectCaseOption {
	return func(r SubjectCaseRule) SubjectCaseRule {
		result := r
		result.checkCommit = check

		return result
	}
}

// WithAllowNonAlpha allows non-alphabetic characters at the start.
func WithAllowNonAlpha(allow bool) SubjectCaseOption {
	return func(r SubjectCaseRule) SubjectCaseRule {
		result := r
		result.allowNonAlpha = allow

		return result
	}
}

// NewSubjectCaseRule creates a new rule for validating subject case.
func NewSubjectCaseRule(options ...SubjectCaseOption) SubjectCaseRule {
	// Create a rule with default values
	rule := SubjectCaseRule{
		name:          "SubjectCase",
		caseChoice:    "sentence", // Default: Sentence case (first letter uppercase, rest lowercase)
		checkCommit:   true,       // Default: Check conventional commits
		allowNonAlpha: false,      // Default: Don't allow non-alpha chars at start
	}

	// Apply all options
	for _, option := range options {
		rule = option(rule)
	}

	return rule
}

// Validate checks that commit subjects follow the required case style.
func (r SubjectCaseRule) Validate(_ context.Context, commit domain.CommitInfo) []appErrors.ValidationError {
	// Special handling for "ignore" or "any" case choice - always valid
	if r.caseChoice == "ignore" || r.caseChoice == "any" {
		return nil
	}

	// Extract subject
	subject := commit.Subject

	// Check for empty subject first
	if subject == "" {
		return []appErrors.ValidationError{
			appErrors.NewCaseError(
				appErrors.ErrEmptyDescription,
				"SubjectCase",
				"Commit message must have a description",
				"Add a meaningful description after your commit type",
			),
		}
	}

	// For conventional commits, need to extract the description part
	var textToCheck string

	if r.checkCommit {
		// Try to parse as conventional commit
		// Format: type(scope)!: description
		conventionalRegex := regexp.MustCompile(`^(?:\w+)(?:\([^)]*\))?(?:!)?:\s*(.*)$`)
		matches := conventionalRegex.FindStringSubmatch(subject)

		// Check for invalid or empty conventional commit format
		if len(matches) > 1 {
			// Found conventional format, extract description
			if matches[1] == "" {
				// Conventional format but empty description
				return []appErrors.ValidationError{
					appErrors.NewCaseError(
						appErrors.ErrEmptyDescription,
						"SubjectCase",
						"Conventional commit requires a description after the type",
						"Format: type(scope): description",
					),
				}
			}

			// Use the description part after the conventional commit prefix
			textToCheck = matches[1]
		} else if isConventionalCommitLike(subject) {
			// It's trying to be a conventional commit but the format is invalid
			return []appErrors.ValidationError{
				appErrors.NewCaseError(
					appErrors.ErrInvalidFormat,
					"SubjectCase",
					"Invalid conventional commit format",
					"Use format: type(scope): description",
				),
			}
		} else {
			// Not a conventional commit, check whole subject
			textToCheck = subject
		}
	} else {
		// Always check the whole subject
		textToCheck = subject
	}

	// Special handling for non-alphabetic starts with allowNonAlpha option
	if r.allowNonAlpha && !startsWithAlpha(textToCheck) {
		return nil
	}

	// Get the first word to check its case
	firstWord := extractFirstWordForCase(textToCheck, r.allowNonAlpha)
	if firstWord == "" {
		return []appErrors.ValidationError{
			appErrors.NewCaseError(
				appErrors.ErrSubjectCase,
				"SubjectCase",
				"Unable to extract a word to validate case",
				"Ensure your commit message starts with an alphabetic character",
			).WithContextMap(map[string]string{"subject": textToCheck}),
		}
	}

	// Check the case
	actualCase, isValid := checkCase(firstWord, r.caseChoice)

	if !isValid {
		return []appErrors.ValidationError{
			appErrors.NewCaseError(
				appErrors.ErrSubjectCase,
				"SubjectCase",
				fmt.Sprintf("First word '%s' should be in %s case", firstWord, r.caseChoice),
				fmt.Sprintf("Change '%s' to %s case", firstWord, r.caseChoice),
			).WithContextMap(map[string]string{
				"word":          firstWord,
				"required_case": r.caseChoice,
				"actual_case":   actualCase,
			}),
		}
	}

	return nil
}

// Name returns the rule name.
func (r SubjectCaseRule) Name() string {
	return r.name
}

// isConventionalCommitLike checks if a string looks like it's trying to be a conventional commit
// but doesn't match the full pattern.
func isConventionalCommitLike(subject string) bool {
	// Look for partial conventional commit pattern (has type and colon but missing description)
	partialPattern := regexp.MustCompile(`^(?:\w+)(?:\([^)]*\))?(?:!)?:`)
	// Also detect incorrect format with something after the colon
	invalidPattern := regexp.MustCompile(`^(?:[^: ]+):`)

	return partialPattern.MatchString(subject) || invalidPattern.MatchString(subject)
}

// startsWithAlpha checks if a string starts with an alphabetic character.
func startsWithAlpha(text string) bool {
	if text == "" {
		return false
	}

	return unicode.IsLetter(rune(text[0]))
}

// extractFirstWordForCase extracts the first word from text, optionally allowing non-alphabetic characters.
func extractFirstWordForCase(text string, allowNonAlpha bool) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return ""
	}

	// Split into words
	words := strings.Fields(text)
	if len(words) == 0 {
		return ""
	}

	firstWord := words[0]

	// Skip leading non-alpha characters if not allowed
	if !allowNonAlpha {
		for len(firstWord) > 0 && !unicode.IsLetter(rune(firstWord[0])) {
			firstWord = firstWord[1:]
		}
	}

	return firstWord
}

// checkCase determines the case of a word and checks if it matches the required case.
//
//nolint:gocyclo // This function is inherently complex due to the many different case types
func checkCase(word string, requiredCase string) (string, bool) {
	if word == "" {
		return "empty", false
	}

	// If the word starts with a non-alphabetic character, we can't check case properly
	// In this case, we just get the first alphabetic character and check its case
	var firstChar rune

	if !unicode.IsLetter(rune(word[0])) {
		// Find the first alphabetic character
		for _, r := range word {
			if unicode.IsLetter(r) {
				firstChar = r

				break
			}
		}
		// If no alphabetic character is found, can't determine case
		if firstChar == 0 {
			return "non-alpha", true // Special case - non-alphabetic words are always valid
		}
	} else {
		firstChar = rune(word[0])
	}

	isFirstUpper := unicode.IsUpper(firstChar)

	// The rest of the characters
	var restAllUpper, restAllLower = true, true

	if len(word) > 1 {
		// Find rest after the first alphabetic character
		restIndex := strings.IndexFunc(word, unicode.IsLetter) + 1
		if restIndex < len(word) {
			rest := word[restIndex:]
			restAllUpper = strings.ToUpper(rest) == rest
			restAllLower = strings.ToLower(rest) == rest
		}
	}

	// Determine the actual case
	var actualCase string
	if isFirstUpper && restAllUpper {
		actualCase = "upper"
	} else if isFirstUpper && restAllLower {
		actualCase = "sentence"
	} else if isFirstUpper && !restAllLower {
		actualCase = "camel" // Mix of upper/lower
	} else if !isFirstUpper && restAllLower {
		actualCase = "lower"
	} else {
		actualCase = "mixed"
	}

	// Handle special case choice values
	if requiredCase == "ignore" || requiredCase == "any" {
		return actualCase, true
	}

	// Check if the case matches the required case
	var isValid bool

	switch requiredCase {
	case "upper":
		isValid = isFirstUpper && restAllUpper
	case "lower":
		isValid = !isFirstUpper && restAllLower
	case "sentence":
		isValid = isFirstUpper && restAllLower
	case "camel":
		isValid = (isFirstUpper && !restAllUpper) || actualCase == "sentence"
	case "title":
		// For title case, we'll accept both "Sentence" and "Title Case" formats
		isValid = (isFirstUpper && restAllLower) || (isFirstUpper && !restAllLower && !restAllUpper)
	default:
		// Default to sentence case (first uppercase, rest lowercase) if invalid case choice
		// This matches the default in NewSubjectCaseRule
		isValid = isFirstUpper && restAllLower
	}

	return actualCase, isValid
}
