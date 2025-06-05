// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package rules

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/itiquette/gommitlint/internal/domain/config"
)

// SubjectRule validates commit subject length, case, and suffix.
type SubjectRule struct {
	maxLength       int
	caseChoice      string
	invalidSuffixes string
	checkCommit     bool
	allowNonAlpha   bool
}

// NewSubjectRule creates a new SubjectRule from config.
func NewSubjectRule(cfg config.Config) SubjectRule {
	maxLength := cfg.Message.Subject.MaxLength
	if maxLength <= 0 {
		maxLength = 72 // Default
	}

	caseChoice := cfg.Message.Subject.Case
	if caseChoice == "" {
		caseChoice = "sentence" // Default
	}

	invalidSuffixes := ".,"

	if len(cfg.Message.Subject.ForbidEndings) > 0 {
		var builder strings.Builder
		for _, suffix := range cfg.Message.Subject.ForbidEndings {
			builder.WriteString(suffix)
		}

		if builder.Len() > 0 {
			invalidSuffixes = builder.String()
		}
	}

	isConventionalEnabled := domain.ShouldRunRule("ConventionalCommit", cfg.Rules.Enabled, cfg.Rules.Disabled)

	return SubjectRule{
		maxLength:       maxLength,
		caseChoice:      caseChoice,
		invalidSuffixes: invalidSuffixes,
		checkCommit:     isConventionalEnabled,
		allowNonAlpha:   cfg.Message.Subject.RequireImperative,
	}
}

// Name returns the rule name.
func (r SubjectRule) Name() string {
	return "Subject"
}

// Validate performs validation against a commit.
func (r SubjectRule) Validate(commit domain.Commit, _ domain.Repository, _ *config.Config) []domain.ValidationError {
	var errors []domain.ValidationError

	// Length validation
	if len(commit.Subject) > r.maxLength {
		errors = append(errors,
			domain.New(r.Name(), domain.ErrSubjectTooLong,
				fmt.Sprintf("subject exceeds %d characters (actual: %d)", r.maxLength, len(commit.Subject))).
				WithHelp(fmt.Sprintf("Keep subject under %d characters", r.maxLength)))
	}

	// Case validation
	if caseErrors := r.validateCase(commit.Subject); len(caseErrors) > 0 {
		errors = append(errors, caseErrors...)
	}

	// Suffix validation
	if suffixErrors := r.validateSuffix(commit.Subject); len(suffixErrors) > 0 {
		errors = append(errors, suffixErrors...)
	}

	return errors
}

// validateCase validates the case style of commit subjects.
func (r SubjectRule) validateCase(subject string) []domain.ValidationError {
	// Special handling for "ignore" or "any" case choice - always valid
	if r.caseChoice == "ignore" || r.caseChoice == "any" {
		return nil
	}

	// Check for empty subject first
	if subject == "" {
		return []domain.ValidationError{
			domain.New(r.Name(), domain.ErrMissingSubject, "Commit message must have a description").
				WithHelp("Add a meaningful description after your commit type"),
		}
	}

	// For conventional commits, need to extract the description part
	var textToCheck string

	if r.checkCommit {
		// Try to parse as conventional commit
		conventionalRegex := regexp.MustCompile(`^(?:\w+)(?:\([^)]*\))?(?:!)?:\s*(.*)$`)
		matches := conventionalRegex.FindStringSubmatch(subject)

		if len(matches) > 1 {
			if matches[1] == "" {
				return []domain.ValidationError{
					domain.New(r.Name(), domain.ErrEmptyDescription, "Conventional commit requires a description after the type").
						WithHelp("Format: type(scope): description"),
				}
			}

			textToCheck = matches[1]
		} else if isConventionalCommitLike(subject) {
			return []domain.ValidationError{
				domain.New(r.Name(), domain.ErrInvalidConventionalFormat, "Invalid conventional commit format").
					WithHelp("Use format: type(scope): description"),
			}
		} else {
			textToCheck = subject
		}
	} else {
		textToCheck = subject
	}

	// Special handling for non-alphabetic starts with allowNonAlpha option
	if r.allowNonAlpha && !startsWithAlpha(textToCheck) {
		return nil
	}

	// Get the first word to check its case
	firstWord := extractFirstWordForCase(textToCheck, r.allowNonAlpha)
	if firstWord == "" {
		return []domain.ValidationError{
			domain.New(r.Name(), domain.ErrNoFirstWord, "Unable to extract a word to validate case").
				WithHelp("Ensure your commit message starts with an alphabetic character"),
		}
	}

	// Check the case
	_, isValid := checkCase(firstWord, r.caseChoice)

	if !isValid {
		return []domain.ValidationError{
			domain.New(r.Name(), domain.ErrSubjectCase,
				fmt.Sprintf("First word '%s' should be in %s case", firstWord, r.caseChoice)).
				WithHelp(fmt.Sprintf("Change '%s' to %s case", firstWord, r.caseChoice)),
		}
	}

	return nil
}

// validateSuffix validates that the commit subject doesn't end with invalid characters.
func (r SubjectRule) validateSuffix(subject string) []domain.ValidationError {
	// Empty subject is always an error
	if len(subject) == 0 {
		return []domain.ValidationError{
			domain.New(r.Name(), domain.ErrMissingSubject, "Commit subject is missing").
				WithHelp("Add a descriptive subject line to your commit"),
		}
	}

	// Check if the subject ends with any of the invalid suffixes
	if len(subject) > 0 {
		subjectRunes := []rune(subject)
		if len(subjectRunes) > 0 {
			lastRune := subjectRunes[len(subjectRunes)-1]
			lastChar := string(lastRune)

			// Check if the last character is in the invalid suffixes
			suffixContainsLastChar := false

			for _, suffixRune := range r.invalidSuffixes {
				if suffixRune == lastRune {
					suffixContainsLastChar = true

					break
				}
			}

			if suffixContainsLastChar {
				return []domain.ValidationError{
					domain.New(r.Name(), domain.ErrSubjectSuffix,
						fmt.Sprintf("Subject ends with invalid character '%s'", lastChar)).
						WithHelp(fmt.Sprintf("Remove the trailing '%s' from your commit subject", lastChar)),
				}
			}
		}
	}

	return nil
}

// Helper functions

// isConventionalCommitLike checks if a string looks like it's trying to be a conventional commit
// but doesn't match the full pattern.
func isConventionalCommitLike(subject string) bool {
	partialPattern := regexp.MustCompile(`^(?:\w+)(?:\([^)]*\))?(?:!)?:`)
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

	words := strings.Fields(text)
	if len(words) == 0 {
		return ""
	}

	firstWord := words[0]

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

	var firstChar rune

	if !unicode.IsLetter(rune(word[0])) {
		for _, r := range word {
			if unicode.IsLetter(r) {
				firstChar = r

				break
			}
		}

		if firstChar == 0 {
			return "non-alpha", true
		}
	} else {
		firstChar = rune(word[0])
	}

	isFirstUpper := unicode.IsUpper(firstChar)

	var restAllUpper, restAllLower = true, true

	if len(word) > 1 {
		restIndex := strings.IndexFunc(word, unicode.IsLetter) + 1
		if restIndex < len(word) {
			rest := word[restIndex:]
			restAllUpper = strings.ToUpper(rest) == rest
			restAllLower = strings.ToLower(rest) == rest
		}
	}

	var actualCase string
	if isFirstUpper && restAllUpper {
		actualCase = "upper"
	} else if isFirstUpper && restAllLower {
		actualCase = "sentence"
	} else if isFirstUpper && !restAllLower {
		actualCase = "camel"
	} else if !isFirstUpper && restAllLower {
		actualCase = "lower"
	} else {
		actualCase = "mixed"
	}

	if requiredCase == "ignore" || requiredCase == "any" {
		return actualCase, true
	}

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
		isValid = (isFirstUpper && restAllLower) || (isFirstUpper && !restAllLower && !restAllUpper)
	default:
		isValid = isFirstUpper && restAllLower
	}

	return actualCase, isValid
}
