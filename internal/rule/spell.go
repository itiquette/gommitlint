// SPDX-FileCopyrightText: 2025 itiquette/gommitlint
//
// SPDX-License-Identifier: EUPL-1.2

package rule

import (
	"fmt"
	"strings"

	"github.com/golangci/misspell"
)

// Spell enforces correct spelling in commit messages by identifying common misspellings.
// This rule helps maintain consistent, professional-looking commit messages by catching
// frequently misspelled words based on the specified locale.
//
// For example, with locale="US", a commit message containing "occurred" would be flagged
// because it should be spelled "occurred", while "colour" would be acceptable with
// locale="GB" but flagged with locale="US".
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
// It returns a Spell rule with any detected misspellings.
//
// The locale parameter accepts:
// - "US" (default): American English spelling
// - "UK" or "GB": British English spelling
// If an empty string is provided, it defaults to "US".
// If an unsupported locale is provided, it returns an error.
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
