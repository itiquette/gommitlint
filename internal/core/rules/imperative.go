// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package rules

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/itiquette/gommitlint/internal/domain"
	appErrors "github.com/itiquette/gommitlint/internal/errors"
)

// ImperativeVerbRule validates that commit messages use imperative mood.
type ImperativeVerbRule struct {
	name                        string
	checkConventionalCommits    bool
	verbCategories              map[string][]string
	conventionalDescriptionOnly bool
	baseFormsEndingWithED       map[string]bool // Words ending with 'ed' that are already base forms
}

// ImperativeVerbOption configures an ImperativeVerbRule.
type ImperativeVerbOption func(ImperativeVerbRule) ImperativeVerbRule

// WithImperativeConventionalCommit configures the rule to check conventional commits.
func WithImperativeConventionalCommit(check bool) ImperativeVerbOption {
	return func(rule ImperativeVerbRule) ImperativeVerbRule {
		result := rule
		result.checkConventionalCommits = check
		result.conventionalDescriptionOnly = check

		return result
	}
}

// WithCustomNonImperativeStarters configures the rule with custom non-imperative starter words.
func WithCustomNonImperativeStarters(words map[string][]string) ImperativeVerbOption {
	return func(rule ImperativeVerbRule) ImperativeVerbRule {
		result := rule

		// Merge with existing categories or replace them
		if len(words) > 0 {
			for category, categoryWords := range words {
				if existing, ok := result.verbCategories[category]; ok {
					// Extend existing category with copy to avoid modifying original
					newWords := make([]string, len(existing), len(existing)+len(categoryWords))
					copy(newWords, existing)
					result.verbCategories[category] = append(newWords, categoryWords...)
				} else {
					// Create new category with copy
					newWords := make([]string, len(categoryWords))
					copy(newWords, categoryWords)
					result.verbCategories[category] = newWords
				}
			}
		}

		return result
	}
}

// WithAdditionalBaseFormsEndingWithED adds words that end with "ed" but are already in base form.
func WithAdditionalBaseFormsEndingWithED(words []string) ImperativeVerbOption {
	return func(rule ImperativeVerbRule) ImperativeVerbRule {
		result := rule

		// Initialize the map if it doesn't exist
		if result.baseFormsEndingWithED == nil {
			result.baseFormsEndingWithED = make(map[string]bool)
		} else {
			// Create a copy of the existing map
			newMap := make(map[string]bool, len(result.baseFormsEndingWithED))
			for k, v := range result.baseFormsEndingWithED {
				newMap[k] = v
			}

			result.baseFormsEndingWithED = newMap
		}

		// Add all the words to our exclusion list
		for _, word := range words {
			result.baseFormsEndingWithED[strings.ToLower(word)] = true
		}

		// Also remove these words from past_tense category if they exist there
		if pastTense, ok := result.verbCategories["past_tense"]; ok {
			// Create a set of words to exclude
			excludeSet := make(map[string]bool)
			for _, word := range words {
				excludeSet[strings.ToLower(word)] = true
			}

			// Filter out words that should be considered base forms
			filtered := make([]string, 0)

			for _, word := range pastTense {
				if !excludeSet[strings.ToLower(word)] {
					filtered = append(filtered, word)
				}
			}

			// Update the category
			result.verbCategories["past_tense"] = filtered
		}

		return result
	}
}

// NewImperativeVerbRule creates a new rule for validating imperative verbs.
func NewImperativeVerbRule(options ...ImperativeVerbOption) ImperativeVerbRule {
	// Create rule with default values
	rule := ImperativeVerbRule{
		name:                     "ImperativeVerb",
		checkConventionalCommits: true,
		baseFormsEndingWithED:    make(map[string]bool), // Initialize empty map for base forms with 'ed'
		// Initialize verb categories for validation
		verbCategories: map[string][]string{
			"past_tense": {
				"added", "fixed", "changed", "updated", "removed", "refactored",
				"improved", "implemented", "enhanced", "resolved", "corrected",
			},
			"gerund": {
				"adding", "fixing", "changing", "updating", "removing", "refactoring",
				"improving", "implementing", "enhancing", "resolving", "correcting",
			},
			"third_person": {
				"adds", "fixes", "changes", "updates", "removes", "refactors",
				"improves", "implements", "enhances", "resolves", "corrects",
			},
		},
	}

	// Apply all options
	for _, option := range options {
		rule = option(rule)
	}

	return rule
}

// Name returns the rule name.
func (r ImperativeVerbRule) Name() string {
	return r.name
}

// Validate checks if the commit message uses imperative mood.
func (r ImperativeVerbRule) Validate(_ context.Context, commit domain.CommitInfo) []appErrors.ValidationError {
	// Validate imperative mood
	subject := strings.TrimSpace(commit.Subject)
	if subject == "" {
		return nil
	}

	// Handle conventional commits if needed
	if r.checkConventionalCommits && r.conventionalDescriptionOnly {
		// Extract description from conventional commit format
		re := regexp.MustCompile(`^[a-z]+(?:\([a-zA-Z0-9/-]+\))?!?:\s*(.*)$`)

		matches := re.FindStringSubmatch(subject)
		if len(matches) > 1 {
			subject = matches[1]
		}
	}

	// Extract first word from subject
	firstWord := extractFirstWord(subject)
	if firstWord == "" {
		return []appErrors.ValidationError{
			appErrors.NewValidationError(
				appErrors.ErrNoFirstWord,
				"ImperativeVerb",
				"Cannot extract first word from commit message",
				"Ensure your commit message starts with a verb",
			).WithContextMap(map[string]string{"subject": subject}),
		}
	}

	firstWord = strings.ToLower(firstWord)

	// Check for imperative mood violations
	category, isViolation := categorizeVerb(firstWord, r.verbCategories, r.baseFormsEndingWithED)

	if isViolation {
		// Build suggestions based on category
		var suggestions []string

		switch category {
		case "past_tense":
			// Convert past tense to imperative
			if strings.HasSuffix(firstWord, "ed") {
				base := strings.TrimSuffix(firstWord, "ed")
				base = strings.TrimSuffix(base, "d")

				suggestions = []string{base}
			} else {
				suggestions = []string{"add", "fix", "update"}
			}
		case "gerund":
			// Convert gerund to imperative
			if strings.HasSuffix(firstWord, "ing") {
				base := strings.TrimSuffix(firstWord, "ing")
				if strings.HasSuffix(base, "nn") {
					base = strings.TrimSuffix(base, "n") // running -> run
				}

				suggestions = []string{base}
			} else {
				suggestions = []string{"add", "fix", "update"}
			}
		case "third_person":
			// Convert third person to imperative
			if strings.HasSuffix(firstWord, "s") || strings.HasSuffix(firstWord, "es") {
				base := strings.TrimSuffix(firstWord, "s")
				base = strings.TrimSuffix(base, "e")

				suggestions = []string{base}
			} else {
				suggestions = []string{"add", "fix", "update"}
			}
		default:
			suggestions = []string{"add", "fix", "update", "remove", "improve", "implement"}
		}

		// Map category to specific error code
		var errorCode appErrors.ValidationErrorCode

		switch category {
		case "past_tense":
			errorCode = appErrors.ErrPastTense
		case "gerund":
			errorCode = appErrors.ErrGerund
		case "third_person":
			errorCode = appErrors.ErrThirdPerson
		default:
			errorCode = appErrors.ErrNonImperative
		}

		// Create error with specific code based on category
		help := fmt.Sprintf("Use the imperative form of '%s'", firstWord)
		if len(suggestions) > 0 {
			help = "Try: " + strings.Join(suggestions, ", ")
		}

		return []appErrors.ValidationError{
			appErrors.NewValidationError(
				errorCode,
				"ImperativeVerb",
				fmt.Sprintf("Word '%s' is not in imperative mood", firstWord),
				help,
			).WithContextMap(map[string]string{
				"word":        firstWord,
				"suggestions": strings.Join(suggestions, ", "),
			}),
		}
	}

	return nil
}

// extractFirstWord extracts the first word from a subject.
func extractFirstWord(subject string) string {
	parts := strings.Fields(subject)
	if len(parts) == 0 {
		return ""
	}

	return parts[0]
}

// categorizeVerb determines if a verb is in a non-imperative category.
func categorizeVerb(word string, categories map[string][]string, baseFormsWithED map[string]bool) (string, bool) {
	// Check if it ends with "ed" but is a valid base form (like "need", "seed")
	if strings.HasSuffix(word, "ed") && baseFormsWithED[word] {
		return "", false
	}

	// Check all categories
	for category, words := range categories {
		for _, nonImperativeWord := range words {
			if word == strings.ToLower(nonImperativeWord) {
				return category, true
			}
		}
	}

	return "", false
}
