// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package rules

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/itiquette/gommitlint/internal/config"
	"github.com/itiquette/gommitlint/internal/domain"
)

// ImperativeVerbRule validates that commit messages use imperative mood.
type ImperativeVerbRule struct {
	name                        string
	checkConventionalCommits    bool
	verbCategories              map[string][]string
	conventionalDescriptionOnly bool
	baseFormsEndingWithED       map[string]bool // Words ending with 'ed' that are already base forms
}

// NewImperativeVerbRule creates a new rule for validating imperative verbs from config.
func NewImperativeVerbRule(cfg config.Config) ImperativeVerbRule {
	// Check if conventional commit is enabled
	isConventionalEnabled := domain.ShouldRunRule("conventional", cfg.Rules.Enabled, cfg.Rules.Disabled)

	return ImperativeVerbRule{
		name:                        "ImperativeVerb",
		checkConventionalCommits:    isConventionalEnabled,
		conventionalDescriptionOnly: isConventionalEnabled,
		baseFormsEndingWithED:       make(map[string]bool),
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
}

// Name returns the rule name.
func (r ImperativeVerbRule) Name() string {
	return r.name
}

// Validate checks if the commit message uses imperative mood.
func (r ImperativeVerbRule) Validate(_ context.Context, commit domain.CommitInfo) []domain.ValidationError {
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
		return []domain.ValidationError{
			domain.New(
				"ImperativeVerb",
				domain.ErrNoFirstWord,
				"Cannot extract first word from commit message",
			).WithHelp("Ensure your commit message starts with a verb").WithContextMap(map[string]string{"subject": subject}),
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
		var errorCode domain.ValidationErrorCode

		switch category {
		case "past_tense":
			errorCode = domain.ErrPastTense
		case "gerund":
			errorCode = domain.ErrGerund
		case "third_person":
			errorCode = domain.ErrThirdPerson
		default:
			errorCode = domain.ErrNonImperative
		}

		// Create error with specific code based on category
		help := fmt.Sprintf("Use the imperative form of '%s'", firstWord)
		if len(suggestions) > 0 {
			help = "Try: " + strings.Join(suggestions, ", ")
		}

		return []domain.ValidationError{
			domain.New(
				"ImperativeVerb",
				errorCode,
				fmt.Sprintf("Word '%s' is not in imperative mood", firstWord),
			).WithHelp(help).WithContextMap(map[string]string{
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
