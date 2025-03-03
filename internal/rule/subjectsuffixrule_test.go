// SPDX-FileCopyrightText: 2025 itiquette/gommitlint
//
// SPDX-License-Identifier: EUPL-1.2
package rule_test

import (
	"testing"

	"github.com/itiquette/gommitlint/internal/rule"
	"github.com/stretchr/testify/require"
)

func TestValidateSubjectSuffixRule(t *testing.T) {
	testCases := []struct {
		name            string
		subject         string
		invalidSuffixes string
		expectedValid   bool
		expectedMessage string
	}{
		{
			name:            "Valid Subject Without Invalid Suffix",
			subject:         "Add new feature",
			invalidSuffixes: ".:;",
			expectedValid:   true,
			expectedMessage: "Subject last character is valid",
		},
		{
			name:            "Subject Ending with Invalid Suffix Period",
			subject:         "Update documentation.",
			invalidSuffixes: ".:;",
			expectedValid:   false,
			expectedMessage: "subject has invalid suffix '.' (invalid suffixes: '.:;')",
		},
		{
			name:            "Subject Ending with Invalid Suffix Colon",
			subject:         "Fix bug:",
			invalidSuffixes: ".:;",
			expectedValid:   false,
			expectedMessage: "subject has invalid suffix ':' (invalid suffixes: '.:;')",
		},
		{
			name:            "Unicode Subject with Invalid Suffix",
			subject:         "Fix élément.",
			invalidSuffixes: ".:;",
			expectedValid:   false,
			expectedMessage: "subject has invalid suffix '.' (invalid suffixes: '.:;')",
		},
		{
			name:            "Unicode Character as Invalid Suffix",
			subject:         "Update description;",
			invalidSuffixes: ";",
			expectedValid:   false,
			expectedMessage: "subject has invalid suffix ';' (invalid suffixes: ';')",
		},
	}

	for _, tabletest := range testCases {
		t.Run(tabletest.name, func(t *testing.T) {
			rule := rule.ValidateSubjectSuffix(tabletest.subject, tabletest.invalidSuffixes)

			// Check errors
			if tabletest.expectedValid {
				require.Empty(t, rule.Errors(), "Did not expect errors")
				require.Equal(t,
					"Subject last character is valid",
					rule.Result(),
					"Message should be valid",
				)
			} else {
				require.NotEmpty(t, rule.Errors(), "Expected errors")
				require.Equal(t,
					tabletest.expectedMessage,
					rule.Result(),
					"Error message should match expected",
				)
			}

			// Check status method
			require.Equal(t, "SubjectSuffixRule", rule.Name(),
				"Status should always be 'SubjectSuffixRule'")
		})
	}
}
