// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package rule

import (
	"fmt"
	"strings"

	"github.com/golangci/misspell"
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
	errors []error
}

// Name returns the name of the rule.
func (Spell) Name() string {
	return "Spell"
}

// Result returns validation results as a human-readable string.
// If no spelling errors are found, it returns a success message.
// Otherwise, it returns the count of misspellings found.
func (rule Spell) Result() string {
	if len(rule.errors) == 0 {
		return "No common misspellings found"
	}

	return fmt.Sprintf("Commit contains %d misspelling(s)", len(rule.errors))
}

// Errors returns any violations of the rule.
func (rule Spell) Errors() []error {
	return rule.errors
}

// Help returns a description of how to fix the rule violation.
func (rule Spell) Help() string {
	const noErrMsg = "No errors to fix"
	if len(rule.errors) == 0 {
		return noErrMsg
	}

	if len(rule.errors) == 1 && strings.Contains(rule.errors[0].Error(), "unknown locale") {
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
	rule := Spell{}

	// Handle empty message
	if strings.TrimSpace(message) == "" {
		return rule // Nothing to check in an empty message
	}

	// Initialize the spell checker
	replacer := misspell.New()

	// Configure locale
	switch strings.ToUpper(locale) {
	case "", "US":
		// Default American English
	case "UK", "GB":
		replacer.AddRuleList(misspell.DictBritish)
	default:
		rule.addErrorf("unknown locale: %q", locale)

		return rule
	}

	// Compile the rules
	replacer.Compile()

	// Check for misspellings
	corrected, diffs := replacer.Replace(message)
	if corrected != message {
		for _, diff := range diffs {
			rule.addErrorf("`%s` is a misspelling of `%s`", diff.Original, diff.Corrected)
		}
	}

	return rule
}

// addErrorf adds an error to the rule's errors slice.
func (rule *Spell) addErrorf(format string, args ...interface{}) {
	rule.errors = append(rule.errors, fmt.Errorf(format, args...))
}
