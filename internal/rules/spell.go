// SPDX-FileCopyrightText: 2024 Sidero Labs, Inc.
// SPDX-FileCopyrightText: 2025 Itiquette/Gommitlint
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
func (sc SpellingCheck) Status() string {
	return "Spellcheck"
}

// Message returns the check message.
func (sc SpellingCheck) Message() string {
	if len(sc.errors) == 0 {
		return ""
	}

	return fmt.Sprintf("Commit contains %d misspelling(s)", len(sc.errors))
}

// Errors returns any violations of the check.
func (sc SpellingCheck) Errors() []error {
	return sc.errors
}

// ValidateSpelling checks the spelling.
func ValidateSpelling(message string, locale string) interfaces.Check {
	check := &SpellingCheck{}
	replacer := misspell.New()

	switch strings.ToUpper(locale) {
	case "", "US":
		// Use default American English
	case "UK", "GB":
		// For British English, we'll use the British dictionary
		replacer.AddRuleList(misspell.DictBritish)
	case "NZ", "AU", "CA":
		check.errors = append(check.errors, fmt.Errorf("unsupported locale: %q", locale))

		return check
	default:
		check.errors = append(check.errors, fmt.Errorf("unknown locale: %q", locale))

		return check
	}

	replacer.Compile()
	corrected, diffs := replacer.Replace(message)

	if corrected != message {
		for _, diff := range diffs {
			check.errors = append(check.errors, fmt.Errorf("`%s` is a misspelling of `%s`", diff.Original, diff.Corrected))
		}
	}

	return check
}
