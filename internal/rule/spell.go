// SPDX-FileCopyrightText: 2024 Sidero Labs, Inc.
// SPDX-FileCopyrightText: 2025 Itiquette/Gommitlint
//
// SPDX-License-Identifier: MPL-2.0

package rule

import (
	"fmt"
	"strings"

	"github.com/golangci/misspell"
	"github.com/itiquette/gommitlint/internal/interfaces"
)

// Spell enforces correct spelling.
type Spell struct {
	RuleErrors []error
}

// Name returns the name of the rule.
func (sc Spell) Name() string {
	return "Spellcheck"
}

// Message returns the rule message.
func (sc Spell) Message() string {
	if len(sc.RuleErrors) == 0 {
		return ""
	}

	return fmt.Sprintf("Commit contains %d misspelling(s)", len(sc.RuleErrors))
}

// Errors returns any violations of the rule.
func (sc Spell) Errors() []error {
	return sc.RuleErrors
}

// ValidateSpelling checks the spelling.
func ValidateSpelling(message string, locale string) interfaces.Rule {
	rule := &Spell{}
	replacer := misspell.New()

	switch strings.ToUpper(locale) {
	case "", "US":
		// Use default American English
	case "UK", "GB":
		// For British English, we'll use the British dictionary
		replacer.AddRuleList(misspell.DictBritish)
	case "NZ", "AU", "CA":
		rule.RuleErrors = append(rule.RuleErrors, fmt.Errorf("unsupported locale: %q", locale))

		return rule
	default:
		rule.RuleErrors = append(rule.RuleErrors, fmt.Errorf("unknown locale: %q", locale))

		return rule
	}

	replacer.Compile()
	corrected, diffs := replacer.Replace(message)

	if corrected != message {
		for _, diff := range diffs {
			rule.RuleErrors = append(rule.RuleErrors, fmt.Errorf("`%s` is a misspelling of `%s`", diff.Original, diff.Corrected))
		}
	}

	return rule
}
