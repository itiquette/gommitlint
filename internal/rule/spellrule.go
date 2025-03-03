// SPDX-FileCopyrightText: 2025 itiquette/gommitlint
//
// SPDX-License-Identifier: EUPL-1.2

package rule

import (
	"fmt"
	"strings"

	"github.com/golangci/misspell"
)

// Spell enforces correct spelling.
type Spell struct {
	RuleErrors []error
}

// Name returns the name of the rule.
func (sc Spell) Name() string {
	return "SpellRule"
}

// Result returns validation resulits.
func (sc Spell) Result() string {
	if len(sc.RuleErrors) == 0 {
		return "No common misspellings found"
	}

	return fmt.Sprintf("Commit contains %d misspelling(s)", len(sc.RuleErrors))
}

// Errors returns any violations of the rule.
func (sc Spell) Errors() []error {
	return sc.RuleErrors
}

// ValidateSpellingRule checks the spelling.
func ValidateSpellingRule(message string, locale string) *Spell {
	rule := &Spell{}
	replacer := misspell.New()

	switch strings.ToUpper(locale) {
	case "", "US":
	case "UK", "GB":
		replacer.AddRuleList(misspell.DictBritish)
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
