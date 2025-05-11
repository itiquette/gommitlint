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

	"github.com/itiquette/gommitlint/internal/config"
	"github.com/itiquette/gommitlint/internal/domain"
	appErrors "github.com/itiquette/gommitlint/internal/errors"
	"github.com/itiquette/gommitlint/internal/infrastructure/log"
)

// SubjectCaseRule validates the case style of commit subjects.
type SubjectCaseRule struct {
	baseRule      BaseRule
	caseChoice    string
	checkCommit   bool
	allowNonAlpha bool
	firstWord     string
	foundCase     string
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
		baseRule:      NewBaseRule("SubjectCase"),
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

// Validate checks that commit subjects follow the required case style
// using configuration from context.
func (r SubjectCaseRule) Validate(ctx context.Context, commit domain.CommitInfo) []appErrors.ValidationError {
	logger := log.Logger(ctx)
	logger.Trace().
		Str("rule", r.Name()).
		Str("commit_hash", commit.Hash).
		Msg("Validating subject case using context configuration")

	// Create a new rule with context configuration
	rule := r.withContextConfig(ctx)

	// Use the existing validation logic
	errors, _ := ValidateWithState(rule, commit)

	return errors
}

// withContextConfig creates a new rule with configuration from context.
func (r SubjectCaseRule) withContextConfig(ctx context.Context) SubjectCaseRule {
	// Get configuration from context
	cfg := config.GetConfig(ctx)

	// Extract configuration values
	caseChoice := cfg.Subject.Case
	isConventional := cfg.Conventional.Required
	allowNonAlpha := cfg.Subject.RequireImperative // Allow non-alpha if imperative is required

	// Create a copy of the rule
	result := r

	// Update settings from context
	if caseChoice != "" {
		result.caseChoice = caseChoice
	}

	result.checkCommit = isConventional
	result.allowNonAlpha = allowNonAlpha

	// Log configuration at debug level
	logger := log.Logger(ctx)
	logger.Debug().
		Str("case_choice", caseChoice).
		Bool("is_conventional", isConventional).
		Bool("allow_non_alpha", allowNonAlpha).
		Msg("Subject case rule configuration from context")

	return result
}

// Name returns the rule name.
func (r SubjectCaseRule) Name() string {
	return r.baseRule.Name()
}

// WithCaseChoice returns a new rule with the specified case choice.
func (r SubjectCaseRule) WithCaseChoice(caseStyle string) SubjectCaseRule {
	result := r
	result.caseChoice = caseStyle

	return result
}

// ValidateWithState validates the subject case and returns both the errors and an updated rule.
func ValidateWithState(rule SubjectCaseRule, commit domain.CommitInfo) ([]appErrors.ValidationError, SubjectCaseRule) {
	// Special handling for "ignore" case choice - always valid
	if rule.caseChoice == "ignore" || rule.caseChoice == "any" {
		return []appErrors.ValidationError{}, rule
	}

	result := rule
	result.baseRule = rule.baseRule.WithClearedErrors().WithRun()

	// Extract subject
	subject := commit.Subject

	// Check for empty subject first
	if subject == "" {
		validationErr := appErrors.CreateBasicError(
			result.baseRule.Name(),
			appErrors.ErrEmptyDescription,
			"commit message description is empty",
		)
		result.baseRule = result.baseRule.WithError(validationErr)

		return result.baseRule.Errors(), result
	}

	// For conventional commits, need to extract the description part
	var textToCheck string

	if rule.checkCommit {
		// Try to parse as conventional commit
		// Format: type(scope)!: description
		conventionalRegex := regexp.MustCompile(`^(?:\w+)(?:\([^)]*\))?(?:!)?:\s*(.*)$`)
		matches := conventionalRegex.FindStringSubmatch(subject)

		// Check for invalid or empty conventional commit format
		if len(matches) > 1 {
			// Found conventional format, extract description
			if matches[1] == "" {
				// Conventional format but empty description
				validationErr := appErrors.CreateBasicError(
					result.baseRule.Name(),
					appErrors.ErrEmptyDescription,
					"commit message description is empty",
				)
				result.baseRule = result.baseRule.WithError(validationErr)

				return result.baseRule.Errors(), result
			}

			textToCheck = matches[1]
		} else if isConventionalCommitLike(subject) {
			// It's trying to be a conventional commit but the format is invalid
			validationErr := appErrors.CreateBasicError(
				result.baseRule.Name(),
				appErrors.ErrInvalidFormat,
				"commit message does not follow conventional commit format",
			)
			result.baseRule = result.baseRule.WithError(validationErr)

			return result.baseRule.Errors(), result
		} else {
			// Not a conventional commit, check whole subject
			textToCheck = subject
		}
	} else {
		// Always check the whole subject
		textToCheck = subject
	}

	// Special handling for non-alphabetic starts with allowNonAlpha option
	if rule.allowNonAlpha && !startsWithAlpha(textToCheck) {
		return []appErrors.ValidationError{}, result
	}

	// Get the first word to check its case
	firstWord := extractFirstWordForCase(textToCheck, rule.allowNonAlpha)
	if firstWord == "" {
		validationErr := appErrors.CreateBasicError(
			result.baseRule.Name(),
			appErrors.ErrSubjectCase,
			"could not find a word to check case",
		).WithContext("subject", textToCheck)

		result.baseRule = result.baseRule.WithError(validationErr)

		return result.baseRule.Errors(), result
	}

	// Store the first word for result reporting
	result.firstWord = firstWord

	// Check the case
	actualCase, isValid := checkCase(firstWord, rule.caseChoice)
	result.foundCase = actualCase

	if !isValid {
		validationErr := appErrors.CreateBasicError(
			result.baseRule.Name(),
			appErrors.ErrSubjectCase,
			fmt.Sprintf("first word '%s' is not in %s case", firstWord, rule.caseChoice),
		).
			WithContext("word", firstWord).
			WithContext("required_case", rule.caseChoice).
			WithContext("actual_case", actualCase)

		result.baseRule = result.baseRule.WithError(validationErr)
	}

	return result.baseRule.Errors(), result
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

// isConventionalCommitLike checks if a string looks like it's trying to be a conventional commit
// but doesn't match the full pattern.
func isConventionalCommitLike(subject string) bool {
	// Look for partial conventional commit pattern (has type and colon but missing description)
	partialPattern := regexp.MustCompile(`^(?:\w+)(?:\([^)]*\))?(?:!)?:`)
	// Also detect incorrect format with something after the colon
	invalidPattern := regexp.MustCompile(`^(?:[^: ]+):`)

	return partialPattern.MatchString(subject) || invalidPattern.MatchString(subject)
}

// checkCase determines the case of a word and checks if it matches the required case.
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
	default:
		// Default to sentence case (first uppercase, rest lowercase) if invalid case choice
		// This matches the default in NewSubjectCaseRule
		isValid = isFirstUpper && restAllLower
	}

	return actualCase, isValid
}

// SetErrors sets the errors for this rule and returns an updated rule.
func (r SubjectCaseRule) SetErrors(errors []appErrors.ValidationError) SubjectCaseRule {
	result := r
	result.baseRule = result.baseRule.WithClearedErrors()

	for _, err := range errors {
		result.baseRule = result.baseRule.WithError(err)
	}

	return result
}

// Errors returns all validation errors found by this rule.
func (r SubjectCaseRule) Errors() []appErrors.ValidationError {
	return r.baseRule.Errors()
}

// HasErrors returns true if the rule has found any errors.
func (r SubjectCaseRule) HasErrors() bool {
	return r.baseRule.HasErrors()
}

// Result returns a concise validation result.
func (r SubjectCaseRule) Result(errors []appErrors.ValidationError) string {
	if len(errors) == 0 {
		return "Subject case is correct"
	}

	// Return based on error code
	errorCode := appErrors.ValidationErrorCode(errors[0].Code)
	if errorCode == appErrors.ErrEmptyMessage || errorCode == appErrors.ErrEmptyDescription {
		return "Subject is empty"
	}

	if errorCode == appErrors.ErrInvalidFormat {
		return "Invalid format"
	}

	return "Subject should start with " + r.caseChoice
}

// VerboseResult returns a more detailed explanation for verbose mode.
func (r SubjectCaseRule) VerboseResult(errors []appErrors.ValidationError) string {
	if len(errors) == 0 {
		var result string

		switch r.caseChoice {
		case "upper":
			result = "Subject starts with an uppercase letter"
		case "lower":
			result = "Subject starts with a lowercase letter"
		case "ignore":
			result = "Subject case check is disabled"
		default:
			result = fmt.Sprintf("Subject correctly uses %s case", r.caseChoice)
		}

		return result
	}

	// Return based on error code
	errorCode := appErrors.ValidationErrorCode(errors[0].Code)
	if errorCode == appErrors.ErrEmptyMessage || errorCode == appErrors.ErrEmptyDescription {
		return "Subject is empty and cannot be checked for case"
	}

	if errorCode == appErrors.ErrInvalidFormat {
		return "Invalid conventional commit format - cannot check subject case"
	}

	expected := "lowercase"
	if r.caseChoice == "upper" {
		expected = "uppercase"
	}

	return "Subject should start with an " + expected + " letter"
}

// Help returns guidance for fixing rule violations.
func (r SubjectCaseRule) Help(errors []appErrors.ValidationError) string {
	if len(errors) == 0 {
		return "No errors to fix"
	}

	// Based on error code
	errorCode := appErrors.ValidationErrorCode(errors[0].Code)
	if errorCode == appErrors.ErrEmptyMessage || errorCode == appErrors.ErrEmptyDescription {
		return "Provide a non-empty commit message"
	}

	if errorCode == appErrors.ErrInvalidFormat {
		return "Format your commit message according to the Conventional Commits specification"
	}

	help := fmt.Sprintf("The first word in your commit message must be in %s case.\n\n", r.caseChoice)

	if r.caseChoice == "sentence" {
		help += "Sentence case means the first letter is uppercase and the rest are lowercase.\n"
		if r.firstWord != "" {
			help += fmt.Sprintf("Instead of '%s', use proper sentence case", r.firstWord)
		}

		help += "\nCapitalize the first letter"
	} else if r.caseChoice == "upper" {
		help += "Upper case means all letters are uppercase.\n"
		if r.firstWord != "" {
			help += fmt.Sprintf("Instead of '%s', use UPPERCASE", r.firstWord)
		}

		help += "\nCapitalize the first letter"
	} else if r.caseChoice == "lower" {
		help += "Lower case means all letters are lowercase.\n"
		if r.firstWord != "" {
			help += fmt.Sprintf("Instead of '%s', use lowercase", r.firstWord)
		}

		help += "\nUse lowercase for the first letter"
	}

	return help
}
