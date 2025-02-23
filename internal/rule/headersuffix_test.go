// SPDX-FileCopyrightText: 2025 Itiquette/Gommitlint
//
// SPDX-License-Identifier: MPL-2.0
package rule_test

import (
	"testing"

	"github.com/itiquette/gommitlint/internal/rule"
	"github.com/stretchr/testify/require"
)

func TestValidateHeaderSuffix(t *testing.T) {
	testCases := []struct {
		name            string
		header          string
		invalidSuffixes string
		expectedValid   bool
		expectedMessage string
	}{
		{
			name:            "Valid Header Without Invalid Suffix",
			header:          "Add new feature",
			invalidSuffixes: ".:;",
			expectedValid:   true,
			expectedMessage: "Header last character is valid",
		},
		{
			name:            "Header Ending with Invalid Suffix Period",
			header:          "Update documentation.",
			invalidSuffixes: ".:;",
			expectedValid:   false,
			expectedMessage: "commit header ends with invalid character '.'",
		},
		{
			name:            "Header Ending with Invalid Suffix Colon",
			header:          "Fix bug:",
			invalidSuffixes: ".:;",
			expectedValid:   false,
			expectedMessage: "commit header ends with invalid character ':'",
		},
		{
			name:            "Empty Header",
			header:          "",
			invalidSuffixes: ".:;",
			expectedValid:   true,
			expectedMessage: "Header last character is valid",
		},
		{
			name:            "Unicode Header with Invalid Suffix",
			header:          "Fix élément.",
			invalidSuffixes: ".:;",
			expectedValid:   false,
			expectedMessage: "commit header ends with invalid character '.'",
		},
		{
			name:            "Unicode Character as Invalid Suffix",
			header:          "Update description;",
			invalidSuffixes: ";",
			expectedValid:   false,
			expectedMessage: "commit header ends with invalid character ';'",
		},
	}

	for _, tabletest := range testCases {
		t.Run(tabletest.name, func(t *testing.T) {
			// Perform the check
			check := rule.ValidateHeaderSuffix(tabletest.header, tabletest.invalidSuffixes)

			// Check errors
			if tabletest.expectedValid {
				require.Empty(t, check.Errors(), "Did not expect errors")
				require.Equal(t,
					"Header last character is valid",
					check.Message(),
					"Message should be valid",
				)
			} else {
				require.NotEmpty(t, check.Errors(), "Expected errors")
				require.Equal(t,
					tabletest.expectedMessage,
					check.Message(),
					"Error message should match expected",
				)
			}

			// Check status method
			require.Equal(t, "Header Last Character", check.Name(),
				"Status should always be 'Header Last Character'")
		})
	}
}
