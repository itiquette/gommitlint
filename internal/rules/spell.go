// SPDX-FileCopyrightText: 2024 Sidero Labs, Inc.
//
// SPDX-License-Identifier: MPL-2.0

package rules

import (
	"fmt"
	"strings"

	"github.com/golangci/misspell"
	"github.com/itiquette/gommitlint/internal/interfaces"
)

// SpellingCheck enforces correct spelling.
type SpellingCheck struct {
	errors []error
}

// Status returns the name of the check.
func (h SpellingCheck) Status() string {
	return "Spellcheck"
}

// Message returns to check message.
func (h SpellingCheck) Message() string {
	return fmt.Sprintf("Commit contains %d misspellings", len(h.errors))
}

// Errors returns any violations of the check.
func (h SpellingCheck) Errors() []error {
	return h.errors
}

// ValidateSpelling checks the spelling.
func ValidateSpelling(message string, locale string) interfaces.Check { //nolint:ireturn
	check := &SpellingCheck{}

	replacer := misspell.Replacer{
		Replacements: misspell.DictMain,
	}

	switch strings.ToUpper(locale) {
	case "":
	case "US":
		replacer.AddRuleList(misspell.DictAmerican)
	case "UK", "GB":
		replacer.AddRuleList(misspell.DictBritish)
	case "NZ", "AU", "CA":
		check.errors = append(check.errors, fmt.Errorf("unknown locale: %q", locale))
	}

	replacer.Compile()

	_, diffs := replacer.Replace(message)

	for _, diff := range diffs {
		check.errors = append(check.errors, fmt.Errorf("`%s` is a misspelling of `%s`", diff.Original, diff.Corrected))
	}

	return check
}
