// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package rule_test

import (
	"strings"
	"testing"

	"github.com/itiquette/gommitlint/internal/rule"
	"github.com/stretchr/testify/require"
)

func TestValidateSubjectSuffix(t *testing.T) {
	testCases := []struct {
		name            string
		subject         string
		invalidSuffixes string
		expectedValid   bool
		expectedMessage string
	}{
		{
			name:            "Valid subject without invalid suffix",
			subject:         "Add new feature",
			invalidSuffixes: ".:;",
			expectedValid:   true,
			expectedMessage: "Subject last character is valid",
		},
		{
			name:            "Subject ending with invalid suffix period",
			subject:         "Update documentation.",
			invalidSuffixes: ".:;",
			expectedValid:   false,
			expectedMessage: "subject has invalid suffix '.' (invalid suffixes: \".:;\")",
		},
		{
			name:            "Subject ending with invalid suffix colon",
			subject:         "Fix bug:",
			invalidSuffixes: ".:;",
			expectedValid:   false,
			expectedMessage: "subject has invalid suffix ':' (invalid suffixes: \".:;\")",
		},
		{
			name:            "Unicode subject with invalid suffix",
			subject:         "Fix élément.",
			invalidSuffixes: ".:;",
			expectedValid:   false,
			expectedMessage: "subject has invalid suffix '.' (invalid suffixes: \".:;\")",
		},
		{
			name:            "Unicode character as invalid suffix",
			subject:         "Update description;",
			invalidSuffixes: ";",
			expectedValid:   false,
			expectedMessage: "subject has invalid suffix ';' (invalid suffixes: \";\")",
		},
		{
			name:            "Empty subject",
			subject:         "",
			invalidSuffixes: ".:;",
			expectedValid:   false,
			expectedMessage: "subject is empty",
		},
		{
			name:            "Subject with Unicode invalid suffix",
			subject:         "Add new emoji😊",
			invalidSuffixes: "😊😀",
			expectedValid:   false,
			expectedMessage: "subject has invalid suffix '😊' (invalid suffixes: \"😊😀\")",
		},
		{
			name:            "Subject with space as invalid suffix",
			subject:         "Add feature ",
			invalidSuffixes: " \t\n",
			expectedValid:   false,
			expectedMessage: "subject has invalid suffix ' ' (invalid suffixes: \" \\t\\n\")",
		},
		{
			name:            "Subject with tab as invalid suffix",
			subject:         "Add feature\t",
			invalidSuffixes: " \t\n",
			expectedValid:   false,
			expectedMessage: "subject has invalid suffix '\\t' (invalid suffixes: \" \\t\\n\")",
		},
		{
			name:            "Valid Unicode subject",
			subject:         "修复问题",
			invalidSuffixes: ".:;",
			expectedValid:   true,
			expectedMessage: "Subject last character is valid",
		},
	}

	for _, tabletest := range testCases {
		t.Run(tabletest.name, func(t *testing.T) {
			result := rule.ValidateSubjectSuffix(tabletest.subject, tabletest.invalidSuffixes)

			// Check errors
			if tabletest.expectedValid {
				require.Empty(t, result.Errors(), "Did not expect errors")
				require.Equal(t,
					"Subject last character is valid",
					result.Result(),
					"Message should be valid",
				)
			} else {
				require.NotEmpty(t, result.Errors(), "Expected errors")
				require.Equal(t,
					tabletest.expectedMessage,
					result.Result(),
					"Error message should match expected",
				)
			}

			// Check name method
			require.Equal(t, "SubjectSuffixRule", result.Name(),
				"Name should always be 'SubjectSuffixRule'")

			// Check help method
			helpText := result.Help()
			require.NotEmpty(t, helpText, "Help text should not be empty")

			// Verify help text is appropriate for the error
			if !tabletest.expectedValid {
				errMsg := result.Result()

				if strings.Contains(errMsg, "subject is empty") {
					require.Contains(t, helpText, "Provide a non-empty subject")
				} else if strings.Contains(errMsg, "invalid suffix") {
					require.Contains(t, helpText, "Remove the punctuation")
				}
			} else {
				require.Contains(t, helpText, "No errors to fix")
			}
		})
	}
}
