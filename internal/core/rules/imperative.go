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
	appErrors "github.com/itiquette/gommitlint/internal/errors"
	"github.com/itiquette/gommitlint/internal/infrastructure/log"
)

// ImperativeVerbRule validates that commit messages use imperative mood.
type ImperativeVerbRule struct {
	baseRule                    BaseRule
	checkConventionalCommits    bool
	firstWord                   string
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
		baseRule:                 NewBaseRule("ImperativeVerb"),
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
	return r.baseRule.Name()
}

// extractFirstWord extracts the first word from a string.
func extractFirstWord(text string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return ""
	}

	// Split by whitespace and get the first part
	parts := strings.Fields(text)
	if len(parts) == 0 {
		return ""
	}

	// Remove any non-alphanumeric prefix
	firstWord := parts[0]
	firstWord = strings.TrimLeft(firstWord, "`~!@#$%^&*()-_=+[{]}\\|;:'\",<.>/?")

	return firstWord
}

// SetErrors sets the errors for this rule and returns an updated rule.
func (r ImperativeVerbRule) SetErrors(errors []appErrors.ValidationError) ImperativeVerbRule {
	result := r
	result.baseRule = r.baseRule.WithClearedErrors()

	for _, err := range errors {
		result.baseRule = result.baseRule.WithError(err)
	}

	return result
}

// Errors returns all validation errors found by this rule.
func (r ImperativeVerbRule) Errors() []appErrors.ValidationError {
	return r.baseRule.Errors()
}

// HasErrors returns true if the rule has found any errors.
func (r ImperativeVerbRule) HasErrors() bool {
	return r.baseRule.HasErrors()
}

// Result returns a concise validation result.
func (r ImperativeVerbRule) Result(errors []appErrors.ValidationError) string {
	if len(errors) > 0 {
		return fmt.Sprintf("❌ Non-imperative: '%s'", r.firstWord)
	}

	return fmt.Sprintf("✓ Imperative: '%s'", r.firstWord)
}

// VerboseResult returns a more detailed explanation for verbose mode.
func (r ImperativeVerbRule) VerboseResult(errors []appErrors.ValidationError) string {
	if len(errors) > 0 {
		for _, err := range errors {
			switch err.Code {
			case string(appErrors.ErrPastTense):
				return fmt.Sprintf("❌ Past tense found instead of imperative: '%s' (use present tense)", r.firstWord)
			case string(appErrors.ErrGerund):
				return fmt.Sprintf("❌ Gerund form found instead of imperative: '%s' (remove '-ing')", r.firstWord)
			case string(appErrors.ErrThirdPerson):
				return fmt.Sprintf("❌ Third person singular found instead of imperative: '%s' (remove 's')", r.firstWord)
			}
		}

		return fmt.Sprintf("❌ Non-imperative verb form: '%s'", r.firstWord)
	}

	return fmt.Sprintf("✓ Commit uses imperative mood: '%s'", r.firstWord)
}

// Help returns guidance for fixing rule violations.
func (r ImperativeVerbRule) Help(errors []appErrors.ValidationError) string {
	if len(errors) == 0 {
		return ""
	}

	return `Use imperative mood in commit messages. 
The imperative is a command form that expresses what the commit does when applied.

Examples:
- "Fix bug" instead of "Fixed bug" (past tense)
- "Add feature" instead of "Adding feature" (gerund)
- "Update docs" instead of "Updates docs" (third person singular)

This follows Git's own built-in conventions, as if completing this sentence:
"If applied, this commit will..." (e.g., "...fix the bug")

The first word in your commit message should be an imperative verb.`
}

// Validate checks if the commit message uses imperative mood.
// This implementation uses context to retrieve configuration.
func (r ImperativeVerbRule) Validate(ctx context.Context, commit domain.CommitInfo) []appErrors.ValidationError {
	logger := log.Logger(ctx)
	logger.Trace().
		Str("rule", r.Name()).
		Str("commit_hash", commit.Hash).
		Msg("Validating imperative verb using context configuration")

	// Create a new rule with context configuration
	rule := r.withContextConfig(ctx)

	// Validate using pure function that returns errors and updated rule
	errors, _ := validateImperativeWithState(rule, commit)

	// Return errors only
	return errors
}

// withContextConfig creates a new rule with configuration from context.
func (r ImperativeVerbRule) withContextConfig(ctx context.Context) ImperativeVerbRule {
	// Get configuration from context
	cfg := config.GetConfig(ctx)

	// Extract configuration values
	isImperativeRequired := cfg.Subject.RequireImperative
	isConventional := cfg.Conventional.Required

	// Log configuration at debug level
	logger := log.Logger(ctx)
	logger.Debug().
		Bool("require_imperative", isImperativeRequired).
		Bool("is_conventional", isConventional).
		Msg("ImperativeVerb rule configuration from context")

	// Create a copy of the rule
	result := r

	// Set conventional flag if needed
	if isConventional {
		result.checkConventionalCommits = true
		result.conventionalDescriptionOnly = true
	}

	return result
}

// validateImperativeWithState performs validation and returns both errors and updated rule state.
func validateImperativeWithState(rule ImperativeVerbRule, commit domain.CommitInfo) ([]appErrors.ValidationError, ImperativeVerbRule) {
	// Create a copy and ensure base rule is properly initialized
	result := rule
	result.baseRule = rule.baseRule.WithClearedErrors().WithRun()

	// Regular implementation
	var textToValidate string

	// Determine what text to validate based on conventional commits setting
	if result.checkConventionalCommits {
		// For conventional commits, extract and validate only the description part
		conventionalRegex := regexp.MustCompile(`^(?:\w+)(?:\([^)]*\))?(?:!)?:\s*(.+)$`)
		matches := conventionalRegex.FindStringSubmatch(commit.Subject)

		if len(matches) > 1 {
			// Found conventional format, validate the description part
			textToValidate = matches[1]
		} else if result.conventionalDescriptionOnly {
			// If not a conventional commit format and we're only checking conventional descriptions,
			// skip validation
			return result.Errors(), result
		} else {
			// Fall back to validating the whole subject
			textToValidate = commit.Subject
		}
	} else {
		// Not checking conventional commits, validate the whole subject
		textToValidate = commit.Subject
	}

	// Extract the first word for validation
	firstWord := extractFirstWord(textToValidate)

	// Store the first word in rule state for result formatting
	result.firstWord = firstWord

	if firstWord == "" {
		validationErr := appErrors.CreateBasicError(
			result.baseRule.Name(),
			appErrors.ErrNoFirstWord,
			"could not find the first word in commit message",
		).WithContext("subject", textToValidate)

		result.baseRule = result.baseRule.WithError(validationErr)

		return result.Errors(), result
	}

	// Check for non-imperative verb forms
	lowerWord := strings.ToLower(firstWord)

	// Check for past tense (often ending with 'ed')
	if strings.HasSuffix(lowerWord, "ed") {
		// Skip base forms that end with 'ed' (like 'embed', 'seed', etc.)
		if _, isBaseForm := result.baseFormsEndingWithED[lowerWord]; !isBaseForm {
			for _, pastTenseWord := range result.verbCategories["past_tense"] {
				if strings.EqualFold(pastTenseWord, firstWord) {
					validationErr := appErrors.CreateBasicError(
						result.baseRule.Name(),
						appErrors.ErrPastTense,
						"use imperative mood in the commit message (e.g., 'fix', not 'fixed')",
					).WithContext("non_imperative_word", firstWord)

					result.baseRule = result.baseRule.WithError(validationErr)

					return result.Errors(), result
				}
			}
		}
	}

	// Check for gerund form (ending with 'ing')
	if strings.HasSuffix(lowerWord, "ing") {
		for _, gerundWord := range result.verbCategories["gerund"] {
			if strings.EqualFold(gerundWord, firstWord) {
				validationErr := appErrors.CreateBasicError(
					result.baseRule.Name(),
					appErrors.ErrGerund,
					"use imperative mood in the commit message (e.g., 'add', not 'adding')",
				).WithContext("non_imperative_word", firstWord)

				result.baseRule = result.baseRule.WithError(validationErr)

				return result.Errors(), result
			}
		}
	}

	// Check for third person singular (often ending with 's')
	if strings.HasSuffix(lowerWord, "s") {
		for _, thirdPersonWord := range result.verbCategories["third_person"] {
			if strings.EqualFold(thirdPersonWord, firstWord) {
				validationErr := appErrors.CreateBasicError(
					result.baseRule.Name(),
					appErrors.ErrThirdPerson,
					"use imperative mood in the commit message (e.g., 'update', not 'updates')",
				).WithContext("non_imperative_word", firstWord)

				result.baseRule = result.baseRule.WithError(validationErr)

				return result.Errors(), result
			}
		}
	}

	// No errors detected
	return result.Errors(), result
}
