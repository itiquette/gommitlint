// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package rule

import (
	"fmt"
	"strings"

	"github.com/golangci/misspell"
	"github.com/itiquette/gommitlint/internal/model"
)

// Spell enforces correct spelling in commit messages by identifying common misspellings.
//
// This rule helps maintain consistent, professional-looking commit messages by catching
// frequently misspelled words and offering corrections. It supports different English
// language variants (locales) to accommodate regional spelling differences.
//
// The rule uses a comprehensive dictionary of common misspellings to identify likely
// errors without requiring a full dictionary. This makes it efficient while still
// catching the most frequent spelling errors in commit messages.
//
// Examples:
//
//   - With locale="US" (American English):
//     "Fix authorization bug" would pass
//     "Fix authorisation bug" would fail (should be "authorization")
//
//   - With locale="GB" (British English):
//     "Update colour scheme" would pass
//     "Update color scheme" would fail (should be "colour")
//
//   - Common misspellings caught in any locale:
//     "Fix defenite issue" would fail (should be "definite")
//     "Accidentally deleted file" would fail (should be "accidentally")
type Spell struct {
	errors []*model.ValidationError
	locale string          // Store the locale for verbose output
	diffs  []misspell.Diff // Store misspelling details for verbose output
}

// Name returns the name of the rule.
func (Spell) Name() string {
	return "Spell"
}

// Result returns a concise validation result as a human-readable string.
func (rule Spell) Result() string {
	if len(rule.errors) == 0 {
		return "No spelling errors"
	}

	return fmt.Sprintf("%d misspelling(s)", len(rule.errors))
}

// VerboseResult returns a more detailed explanation for verbose mode.
func (rule Spell) VerboseResult() string {
	if len(rule.errors) == 0 {
		localeDesc := "US (American English)"
		if rule.locale == "UK" || rule.locale == "GB" {
			localeDesc = "UK/GB (British English)"
		}

		return fmt.Sprintf("No common misspellings found using %s dictionary", localeDesc)
	}

	if len(rule.errors) == 1 && rule.errors[0].Code == "invalid_locale" {
		locale := ""

		for k, v := range rule.errors[0].Context {
			if k == "locale" {
				locale = v

				break
			}
		}

		return fmt.Sprintf("Invalid locale: '%s'. Supported locales are: US, UK, GB", locale)
	}

	// List the specific misspellings
	var stringBuilder strings.Builder

	stringBuilder.WriteString(fmt.Sprintf("Found %d misspelling(s):", len(rule.errors)))

	// Limit the number of misspellings to display in verbose mode
	limit := 5
	if len(rule.diffs) > limit {
		for i := 0; i < limit; i++ {
			diff := rule.diffs[i]
			stringBuilder.WriteString(fmt.Sprintf("\n- '%s' should be '%s'", diff.Original, diff.Corrected))
		}

		remaining := len(rule.diffs) - limit
		stringBuilder.WriteString(fmt.Sprintf("\n- and %d more...", remaining))
	} else {
		for _, diff := range rule.diffs {
			stringBuilder.WriteString(fmt.Sprintf("\n- '%s' should be '%s'", diff.Original, diff.Corrected))
		}
	}

	return stringBuilder.String()
}

// Errors returns any violations of the rule.
func (rule Spell) Errors() []*model.ValidationError {
	return rule.errors
}

// addError adds a structured validation error.
func (rule *Spell) addError(code, message string, context map[string]string) {
	err := model.NewValidationError("Spell", code, message)

	// Add any context values
	for key, value := range context {
		_ = err.WithContext(key, value)
	}

	rule.errors = append(rule.errors, err)
}

// Help returns a description of how to fix the rule violation.
func (rule Spell) Help() string {
	const noErrMsg = "No errors to fix"
	if len(rule.errors) == 0 {
		return noErrMsg
	}

	// Check for specific error codes
	if len(rule.errors) == 1 && rule.errors[0].Code == "invalid_locale" {
		return "Use a supported locale for spell checking: 'US' (default), 'UK', or 'GB'."
	}

	var corrections strings.Builder

	corrections.WriteString("Fix the following misspellings in your commit message:\n")

	for i, err := range rule.errors {
		if i > 0 {
			corrections.WriteString("\n")
		}

		corrections.WriteString("- ")
		corrections.WriteString(err.Error())

		// Add suggestion if available in context
		if original, ok := err.Context["original"]; ok {
			if corrected, ok := err.Context["corrected"]; ok {
				corrections.WriteString(fmt.Sprintf(" (Replace '%s' with '%s')", original, corrected))
			}
		}
	}

	corrections.WriteString("\n\nIf any of these are intentional or domain-specific terms, consider rewording.")

	return corrections.String()
}

// ValidateSpelling checks the spelling in a commit message based on the specified locale.
// It identifies common misspellings and suggests corrections.
//
// Parameters:
//   - message: The commit message to check for spelling errors
//   - locale: The language variant to use for spell checking
//
// The locale parameter accepts:
//   - "US" (default): American English spelling (e.g., "color", "organize")
//   - "UK" or "GB": British English spelling (e.g., "colour", "organise")
//
// If an empty string is provided for locale, it defaults to "US".
// If an unsupported locale is provided, it returns an error.
//
// The spell checker uses a curated list of common misspellings rather than a
// full dictionary, so it won't catch all possible spelling errors but focuses
// on the most frequent mistakes.
//
// Returns:
//   - A Spell instance with validation results including any detected misspellings
func ValidateSpelling(message string, locale string) Spell {
	rule := Spell{
		locale: locale,
	}

	// Handle empty message
	if strings.TrimSpace(message) == "" {
		return rule // Nothing to check in an empty message
	}

	// Initialize the spell checker
	replacer := misspell.New()

	// Configure locale
	switch strings.ToUpper(locale) {
	case "", "US":
		rule.locale = "US" // Ensure locale is set for verbose output
	case "UK", "GB":
		replacer.AddRuleList(misspell.DictBritish)

		rule.locale = strings.ToUpper(locale) // Store normalized locale
	default:
		rule.addError(
			"invalid_locale",
			fmt.Sprintf("unknown locale: %q", locale),
			map[string]string{
				"locale":            locale,
				"supported_locales": "US,UK,GB",
			},
		)

		return rule
	}

	// Compile the rules
	replacer.Compile()

	// Check for misspellings
	corrected, diffs := replacer.Replace(message)
	if corrected != message {
		// Store diffs for verbose output
		rule.diffs = diffs

		for _, diff := range diffs {
			rule.addError(
				"misspelling",
				fmt.Sprintf("`%s` is a misspelling of `%s`", diff.Original, diff.Corrected),
				map[string]string{
					"original":  diff.Original,
					"corrected": diff.Corrected,
				},
			)
		}
	}

	return rule
}
