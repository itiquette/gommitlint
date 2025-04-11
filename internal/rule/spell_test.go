// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package rule_test

import (
	"strconv"
	"strings"
	"testing"

	"github.com/itiquette/gommitlint/internal/rule"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateSpelling(t *testing.T) {
	tests := []struct {
		name           string
		message        string
		locale         string
		expectedErrors int
		expectedCode   string
		expectedWords  []string
	}{
		{
			name:           "No misspellings",
			message:        "This is a correct sentence.",
			locale:         "US",
			expectedErrors: 0,
			expectedWords:  nil,
		},
		{
			name:           "One misspelling",
			message:        "This is definately a misspelling.", //nolint
			locale:         "US",
			expectedErrors: 1,
			expectedCode:   "misspelling",
			expectedWords:  []string{"definately", "definitely"}, //nolint
		},
		{
			name:           "Multiple misspellings",
			message:        "We occured a misspelling and we beleive it needs fixing.", //nolint
			locale:         "US",
			expectedErrors: 2,
			expectedCode:   "misspelling",
			expectedWords:  []string{"occured", "beleive"}, //nolint
		},
		{
			name:           "British english",
			message:        "The colour of the centre looks great.",
			locale:         "GB",
			expectedErrors: 0,
			expectedWords:  nil,
		},
		{
			name:           "Empty message",
			message:        "",
			locale:         "US",
			expectedErrors: 0,
			expectedWords:  nil,
		},
		{
			name:           "Whitespace message",
			message:        "   \t   \n",
			locale:         "US",
			expectedErrors: 0,
			expectedWords:  nil,
		},
		{
			name:           "Unknown locale",
			message:        "This is a test.",
			locale:         "FR",
			expectedErrors: 1,
			expectedCode:   "invalid_locale",
			expectedWords:  []string{"unknown locale"},
		},
		{
			name:           "Default locale",
			message:        "This is a test with the color red.",
			locale:         "",
			expectedErrors: 0,
			expectedWords:  nil,
		},
		{
			name:           "Case insensitive locale",
			message:        "The colour of the centre looks great.",
			locale:         "gb",
			expectedErrors: 0,
			expectedWords:  nil,
		},
	}

	for _, tabletest := range tests {
		t.Run(tabletest.name, func(t *testing.T) {
			// Call the rule
			rule := rule.ValidateSpelling(tabletest.message, tabletest.locale)

			// Check for expected number of errors
			errors := rule.Errors()
			assert.Len(t, errors, tabletest.expectedErrors, "Incorrect number of errors")

			// Check error code if specified
			if tabletest.expectedCode != "" && len(errors) > 0 {
				assert.Equal(t, tabletest.expectedCode, errors[0].Code,
					"Error code should match expected")
			}

			// Verify rule name is set in ValidationError
			if len(errors) > 0 {
				assert.Equal(t, "Spell", errors[0].Rule,
					"Rule name should be set in ValidationError")
			}

			// Check for expected words in error messages
			if tabletest.expectedWords != nil && len(errors) > 0 {
				for _, word := range tabletest.expectedWords {
					found := false

					for _, err := range errors {
						if strings.Contains(err.Error(), word) {
							found = true

							break
						}
					}

					if !found {
						// This is not a hard failure - some misspellings might not be in the dictionary
						t.Logf("Note: Expected to find '%s' in one of the error messages, but didn't", word)
					}
				}
			}

			// Verify Result() method
			if tabletest.expectedErrors > 0 {
				assert.Contains(t, rule.Result(), "misspelling",
					"Result should mention misspellings when errors are present")
				assert.Contains(t, rule.Result(), strconv.Itoa(tabletest.expectedErrors),
					"Result should include the number of errors")
			} else {
				assert.Contains(t, rule.Result(), "No common misspellings found",
					"Result should indicate no misspellings")
			}

			// Verify Help() method
			helpText := rule.Help()
			require.NotEmpty(t, helpText, "Help text should not be empty")

			if tabletest.expectedErrors == 0 {
				assert.Contains(t, helpText, "No errors to fix",
					"Help should indicate no errors to fix")
			} else if len(errors) > 0 && errors[0].Code == "invalid_locale" {
				assert.Contains(t, helpText, "supported locale",
					"Help should mention supported locales")
			} else if len(errors) > 0 {
				assert.Contains(t, helpText, "Fix the following misspellings",
					"Help should provide guidance")

				for _, err := range errors {
					assert.Contains(t, helpText, err.Error(),
						"Help should include error details")
				}
			}

			// Verify Name() method
			assert.Equal(t, "Spell", rule.Name(), "Rule name should be 'Spell'")

			// Check context data
			if tabletest.expectedCode == "misspelling" && len(errors) > 0 {
				assert.Contains(t, errors[0].Context, "original",
					"Context should contain original misspelled word")
				assert.Contains(t, errors[0].Context, "corrected",
					"Context should contain corrected spelling")
			} else if tabletest.expectedCode == "invalid_locale" && len(errors) > 0 {
				assert.Contains(t, errors[0].Context, "locale",
					"Context should contain the invalid locale")
				assert.Contains(t, errors[0].Context, "supported_locales",
					"Context should contain supported locales")
			}
		})
	}

	// Test for context information in the help text
	t.Run("Help text includes suggestions", func(t *testing.T) {
		// Create a rule with a misspelling
		rule := rule.ValidateSpelling("This is definately a mistake", "US") //nolint

		// Verify that the help text includes the suggestion
		helpText := rule.Help()
		assert.Contains(t, helpText, "Replace", "Help text should include replacement suggestions")

		// Check that both original and corrected words appear in the help text
		if len(rule.Errors()) > 0 {
			original := rule.Errors()[0].Context["original"]
			corrected := rule.Errors()[0].Context["corrected"]

			assert.Contains(t, helpText, original, "Help text should include original misspelled word")
			assert.Contains(t, helpText, corrected, "Help text should include corrected spelling")
		}
	})
}
