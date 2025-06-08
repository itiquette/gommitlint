// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package spell

import (
	"strings"

	"github.com/client9/misspell"
	"github.com/itiquette/gommitlint/internal/domain"
)

// MisspellAdapter implements the SpellChecker interface using the misspell library.
type MisspellAdapter struct {
	replacer *misspell.Replacer
}

// NewMisspellAdapter creates a new spell checker adapter configured for the specified locale.
func NewMisspellAdapter(locale string) *MisspellAdapter {
	replacer := misspell.New()

	// Configure for different English variants
	switch strings.ToLower(locale) {
	case "uk", "en-gb", "british", "gb":
		// Enable British spelling dictionary
		replacer.AddRuleList(misspell.DictBritish)
		replacer.Compile()
	case "us", "en-us", "american", "usa", "":
		// US is default - misspell.New() already includes DictMain
		replacer.Compile()
	default:
		// Default to US English for unknown locales
		replacer.Compile()
	}

	return &MisspellAdapter{
		replacer: replacer,
	}
}

// CheckText implements the SpellChecker interface by checking text for misspellings.
func (a *MisspellAdapter) CheckText(text string) []domain.Misspelling {
	_, diffs := a.replacer.Replace(text)

	misspellings := make([]domain.Misspelling, 0, len(diffs))

	for _, diff := range diffs {
		misspellings = append(misspellings, domain.Misspelling{
			Word:       diff.Original,
			Position:   diff.Column,
			Suggestion: diff.Corrected,
		})
	}

	return misspellings
}
