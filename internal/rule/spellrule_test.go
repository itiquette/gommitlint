// SPDX-FileCopyrightText: 2025 itiquette/gommitlint
//
// SPDX-License-Identifier: EUPL-1.2

package rule_test

import (
	"strings"
	"testing"

	"github.com/itiquette/gommitlint/internal/rule"
	"github.com/stretchr/testify/assert"
)

func TestValidateSpellingRule(t *testing.T) {
	tests := []struct {
		name           string
		message        string
		locale         string
		expectedErrors int
		expectedWords  []string
	}{
		{
			name:           "No misspellings",
			message:        "This is a correct sentence.",
			locale:         "US",
			expectedErrors: 0,
			expectedWords:  nil,
		},
		// {
		// 	name:           "One misspelling",
		// 	message:        "This langage has a misspelling.",
		// 	locale:         "US",
		// 	expectedErrors: 1,
		// 	expectedWords:  []string{"langiage"},
		// },
		// {
		// 	name:           "Multiple misspellings",
		// 	message:        "This langiage has multiple misspellings.",
		// 	locale:         "US",
		// 	expectedErrors: 2,
		// 	expectedWords:  []string{"language", "multiple"},
		// },
		{
			name:           "British English",
			message:        "langauge", //nolint
			locale:         "GB",
			expectedErrors: 1,
			expectedWords:  nil,
		},
		{
			name:           "American English",
			message:        "langauge", //nolint
			locale:         "US",
			expectedErrors: 1,
			expectedWords:  nil,
		},
		{
			name:           "Unknown locale",
			message:        "This is a test.",
			locale:         "FR",
			expectedErrors: 1,
			expectedWords:  []string{"unknown locale"},
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			rule := rule.ValidateSpellingRule(tabletest.message, tabletest.locale)
			errors := rule.Errors()

			assert.Len(t, errors, tabletest.expectedErrors, "Incorrect number of errors")

			for _, word := range tabletest.expectedWords {
				found := false

				for _, err := range errors {
					if strings.Contains(err.Error(), word) {
						found = true

						break
					}
				}

				assert.True(t, found, "Expected to find error for word '%s'", word)
			}

			if tabletest.expectedErrors > 0 {
				assert.NotEmpty(t, rule.Result(), "Message should not be empty when errors are present")
			} else {
				assert.Contains(t, rule.Result(), "No common misspellings found")
			}
		})
	}
}
