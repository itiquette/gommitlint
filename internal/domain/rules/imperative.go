// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package rules

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/kljensen/snowball"

	"github.com/itiquette/gommitlint/internal/domain"
)

// ImperativeValidator provides sophisticated imperative mood validation using Snowball stemming.
// This is a modular component that can be used by any rule that needs imperative validation.
type ImperativeValidator struct{}

// NewImperativeValidator creates a new imperative validator.
func NewImperativeValidator() *ImperativeValidator {
	return &ImperativeValidator{}
}

// ValidateImperative performs sophisticated imperative mood validation using linguistic analysis.
// It returns validation errors if the text is not in imperative mood.
func (v *ImperativeValidator) ValidateImperative(text, originalSubject string, isConventional bool, ruleName string) []domain.ValidationError {
	// Extract the text to validate based on commit type
	textToCheck, validationErr := v.extractValidationText(text, isConventional, ruleName)
	if validationErr != nil {
		return []domain.ValidationError{*validationErr}
	}

	// Extract first word for validation
	firstWord, validationErr := v.extractFirstWord(textToCheck, ruleName)
	if validationErr != nil {
		return []domain.ValidationError{*validationErr}
	}

	// Validate the first word using linguistic analysis
	return v.analyzeWord(firstWord, originalSubject, ruleName)
}

// extractValidationText extracts the text to validate based on commit format.
func (v *ImperativeValidator) extractValidationText(subject string, isConventional bool, ruleName string) (string, *domain.ValidationError) {
	if !isConventional {
		return subject, nil
	}

	// Use shared conventional commit parser
	parsed := domain.ParseConventionalCommit(subject)
	if !parsed.IsValid {
		// Check if it looks like an attempted conventional commit
		if v.isConventionalLike(subject) {
			errorPtr := domain.New(ruleName, domain.ErrInvalidConventionalFormat,
				"Invalid conventional commit format").
				WithContextMap(map[string]string{
					"subject":  subject,
					"expected": "type(scope): description",
				}).
				WithHelp("Use format: type(scope): description (e.g., 'feat: add login')")

			return "", &errorPtr
		}
		// Not conventional format, validate entire subject
		return subject, nil
	}

	description := strings.TrimSpace(parsed.Description)
	if description == "" {
		errorPtr := domain.New(ruleName, domain.ErrMissingConventionalSubject,
			"Missing subject after conventional commit type").
			WithContextMap(map[string]string{
				"subject": subject,
				"type":    parsed.Type,
			}).
			WithHelp("Add a descriptive subject after the type: prefix")

		return "", &errorPtr
	}

	return description, nil
}

// extractFirstWord extracts the first word from the text to validate.
func (v *ImperativeValidator) extractFirstWord(text, ruleName string) (string, *domain.ValidationError) {
	// Use regex to extract first word (alphanumeric characters)
	firstWordRegex := regexp.MustCompile(`^\s*([a-zA-Z0-9]+)`)
	matches := firstWordRegex.FindStringSubmatch(text)

	if len(matches) < 2 {
		errorPtr := domain.New(ruleName, domain.ErrNoFirstWord,
			"Cannot extract first word from commit message").
			WithContextMap(map[string]string{
				"text": text,
			}).
			WithHelp("Start your commit message with a verb")

		return "", &errorPtr
	}

	return matches[1], nil
}

// analyzeWord performs linguistic analysis of the word to determine if it's imperative.
func (v *ImperativeValidator) analyzeWord(word, subject, ruleName string) []domain.ValidationError {
	wordLower := strings.ToLower(word)

	// Check for non-imperative starters (articles, pronouns, etc.)
	if v.isNonImperativeStarter(wordLower) {
		return []domain.ValidationError{
			domain.New(ruleName, domain.ErrNonVerb,
				fmt.Sprintf("'%s' is not a verb", word)).
				WithContextMap(map[string]string{
					"actual":  word,
					"type":    "non_verb",
					"subject": subject,
				}).
				WithHelp("Start with an imperative verb like: add, fix, update, remove, implement"),
		}
	}

	// Use Snowball stemming for sophisticated analysis
	stem, err := snowball.Stem(wordLower, "english", true)
	if err != nil {
		// Fallback to simple pattern matching if stemming fails
		return v.validateWithSimpleRules(wordLower, word, subject, ruleName)
	}

	// Analyze using stemming results
	return v.validateWithStemming(wordLower, word, stem, subject, ruleName)
}

// validateWithStemming validates using Snowball stemming results.
func (v *ImperativeValidator) validateWithStemming(wordLower, originalWord, stem, subject, ruleName string) []domain.ValidationError {
	// Check for past tense (word ends with 'ed' and stem is different)
	if strings.HasSuffix(wordLower, "ed") && stem != wordLower {
		// Check if it's actually a base form ending with 'ed'
		if !v.isBaseFormEndingWithED(wordLower) {
			// Create proper imperative suggestion by capitalizing the stem
			suggestion := v.createImperativeSuggestion(stem, originalWord)

			return []domain.ValidationError{
				domain.New(ruleName, domain.ErrPastTense,
					fmt.Sprintf("'%s' is past tense", originalWord)).
					WithContextMap(map[string]string{
						"actual":     originalWord,
						"subject":    subject,
						"suggestion": suggestion,
						"type":       "past_tense",
					}).
					WithHelp(fmt.Sprintf("Use the base form '%s' instead of past tense '%s'", suggestion, originalWord)),
			}
		}
	}

	// Check for gerunds (word ends with 'ing' and stem is different)
	if strings.HasSuffix(wordLower, "ing") && stem != wordLower {
		suggestion := v.createImperativeSuggestion(stem, originalWord)

		return []domain.ValidationError{
			domain.New(ruleName, domain.ErrGerund,
				fmt.Sprintf("'%s' is a gerund (present participle)", originalWord)).
				WithContextMap(map[string]string{
					"actual":     originalWord,
					"subject":    subject,
					"suggestion": suggestion,
					"type":       "gerund",
				}).
				WithHelp(fmt.Sprintf("Use the base form '%s' instead of gerund '%s'", suggestion, originalWord)),
		}
	}

	// Check for third person singular (word ends with 's' and stem is different)
	if strings.HasSuffix(wordLower, "s") && stem != wordLower {
		// Check if it's actually a base form ending with 's'
		if !v.isBaseFormEndingWithS(wordLower) {
			suggestion := v.createImperativeSuggestion(stem, originalWord)

			return []domain.ValidationError{
				domain.New(ruleName, domain.ErrThirdPerson,
					fmt.Sprintf("'%s' is third person singular", originalWord)).
					WithContextMap(map[string]string{
						"actual":     originalWord,
						"subject":    subject,
						"suggestion": suggestion,
						"type":       "third_person",
					}).
					WithHelp(fmt.Sprintf("Use the base form '%s' instead of third person '%s'", suggestion, originalWord)),
			}
		}
	}

	// Word appears to be in imperative form
	return nil
}

// validateWithSimpleRules provides fallback validation when stemming fails.
func (v *ImperativeValidator) validateWithSimpleRules(wordLower, originalWord, subject, ruleName string) []domain.ValidationError {
	// Simple pattern-based detection as fallback
	if strings.HasSuffix(wordLower, "ed") && !v.isBaseFormEndingWithED(wordLower) {
		suggestion := strings.TrimSuffix(wordLower, "ed")
		suggestion = strings.TrimSuffix(suggestion, "d")

		return []domain.ValidationError{
			domain.New(ruleName, domain.ErrPastTense,
				fmt.Sprintf("'%s' appears to be past tense", originalWord)).
				WithContextMap(map[string]string{
					"actual":     originalWord,
					"type":       "past_tense",
					"subject":    subject,
					"suggestion": suggestion,
				}).
				WithHelp(fmt.Sprintf("Try using '%s' instead of '%s'", suggestion, originalWord)),
		}
	}

	if strings.HasSuffix(wordLower, "ing") {
		suggestion := strings.TrimSuffix(wordLower, "ing")

		return []domain.ValidationError{
			domain.New(ruleName, domain.ErrGerund,
				fmt.Sprintf("'%s' appears to be a gerund", originalWord)).
				WithContextMap(map[string]string{
					"actual":     originalWord,
					"type":       "gerund",
					"subject":    subject,
					"suggestion": suggestion,
				}).
				WithHelp(fmt.Sprintf("Try using '%s' instead of '%s'", suggestion, originalWord)),
		}
	}

	if strings.HasSuffix(wordLower, "s") && !v.isBaseFormEndingWithS(wordLower) {
		suggestion := strings.TrimSuffix(wordLower, "s")

		return []domain.ValidationError{
			domain.New(ruleName, domain.ErrThirdPerson,
				fmt.Sprintf("'%s' appears to be third person", originalWord)).
				WithContextMap(map[string]string{
					"actual":     originalWord,
					"type":       "third_person",
					"subject":    subject,
					"suggestion": suggestion,
				}).
				WithHelp(fmt.Sprintf("Try using '%s' instead of '%s'", suggestion, originalWord)),
		}
	}

	return nil
}

// isConventionalLike checks if text looks like an attempted conventional commit.
func (v *ImperativeValidator) isConventionalLike(subject string) bool {
	// Check for common patterns that suggest conventional commit attempt
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`^[a-zA-Z]+\(`), // type(
		regexp.MustCompile(`^[a-zA-Z]+:`),  // type:
		regexp.MustCompile(`^(feat|fix|docs|style|refactor|test|chore|perf|ci|build|revert)\s+`), // common types with space
	}

	for _, pattern := range patterns {
		if pattern.MatchString(subject) {
			return true
		}
	}

	return false
}

// isNonImperativeStarter checks if a word is a non-imperative starter.
func (v *ImperativeValidator) isNonImperativeStarter(word string) bool {
	nonImperativeStarters := map[string]bool{
		// Pronouns
		"i": true, "we": true, "they": true, "he": true, "she": true, "it": true,
		// Articles
		"the": true, "a": true, "an": true,
		// Demonstratives
		"this": true, "that": true, "these": true, "those": true,
		// Possessives
		"my": true, "your": true, "our": true, "his": true, "her": true, "its": true,
		// Other non-verb starters
		"all": true, "some": true, "every": true, "no": true, "any": true,
	}

	return nonImperativeStarters[word]
}

// isBaseFormEndingWithED checks if a word ending with 'ed' is actually a base form.
func (v *ImperativeValidator) isBaseFormEndingWithED(word string) bool {
	baseFormsEndingWithED := map[string]bool{
		// Words that naturally end with 'ed' in their base form
		"shed": true, "embed": true, "speed": true, "proceed": true,
		"exceed": true, "succeed": true, "feed": true, "need": true,
		"breed": true, "seed": true, "bleed": true, "freed": true,
		"greed": true, "creed": true, "deed": true, "weed": true,
	}

	return baseFormsEndingWithED[word]
}

// isBaseFormEndingWithS checks if a word ending with 's' is actually a base form.
func (v *ImperativeValidator) isBaseFormEndingWithS(word string) bool {
	baseFormsEndingWithS := map[string]bool{
		// Words that naturally end with 's' in their base form
		"focus": true, "process": true, "pass": true, "address": true,
		"express": true, "dismiss": true, "access": true, "press": true,
		"cross": true, "miss": true, "toss": true, "guess": true,
		"dress": true, "bless": true, "stress": true, "class": true,
		"mass": true, "bass": true, "glass": true, "grass": true,
	}

	return baseFormsEndingWithS[word]
}

// createImperativeSuggestion creates a proper imperative suggestion from a stem,
// preserving the original word's capitalization pattern.
func (v *ImperativeValidator) createImperativeSuggestion(stem, originalWord string) string {
	if stem == "" {
		return originalWord
	}

	// Handle common irregular stems
	switch stem {
	case "ad":
		return "Add"
	case "updat":
		return "Update"
	case "creat":
		return "Create"
	case "delet":
		return "Delete"
	case "fix":
		return "Fix"
	case "remov":
		return "Remove"
	case "chang":
		return "Change"
	case "modifi":
		return "Modify"
	}

	// For regular stems, capitalize first letter
	if len(stem) > 0 {
		return strings.ToUpper(stem[:1]) + stem[1:]
	}

	return stem
}
