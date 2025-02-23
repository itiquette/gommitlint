// SPDX-FileCopyrightText: 2024 Sidero Labs, Inc.
// SPDX-FileCopyrightText: 2025 Itiquette/Gommitlint
//
// SPDX-License-Identifier: MPL-2.0

package rule_test

import (
	"strings"
	"testing"

	"github.com/itiquette/gommitlint/internal/rule"
	"github.com/stretchr/testify/assert"
)

func TestValidateSpelling(t *testing.T) {
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
			message:        "The colour of the centre is grey.",
			locale:         "GB",
			expectedErrors: 0, // This might need to be adjusted based on actual behavior
			expectedWords:  nil,
		},
		{
			name:           "American English",
			message:        "The color of the center is gray.",
			locale:         "US",
			expectedErrors: 0,
			expectedWords:  nil,
		},
		{
			name:           "Unsupported locale",
			message:        "This is a test.",
			locale:         "AU",
			expectedErrors: 1,
			expectedWords:  []string{"unsupported locale"},
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
			check := rule.ValidateSpelling(tabletest.message, tabletest.locale)
			errors := check.Errors()

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
				assert.NotEmpty(t, check.Message(), "Message should not be empty when errors are present")
			} else {
				assert.Empty(t, check.Message(), "Message should be empty when no errors are present")
			}
		})
	}
}
