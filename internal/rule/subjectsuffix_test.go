// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package rule_test

import (
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
		expectedCode    string
	}{
		{
			name:            "Valid subject without invalid suffix",
			subject:         "Add new feature",
			invalidSuffixes: ".:;",
			expectedValid:   true,
			expectedMessage: "Subject has valid suffix",
		},
		{
			name:            "Subject ending with invalid suffix period",
			subject:         "Update documentation.",
			invalidSuffixes: ".:;",
			expectedValid:   false,
			expectedMessage: "subject has invalid suffix \".\" (invalid suffixes: \".:;\")",
			expectedCode:    "invalid_suffix",
		},
		{
			name:            "Subject ending with invalid suffix colon",
			subject:         "Fix bug:",
			invalidSuffixes: ".:;",
			expectedValid:   false,
			expectedMessage: "subject has invalid suffix \":\" (invalid suffixes: \".:;\")",
			expectedCode:    "invalid_suffix",
		},
		{
			name:            "Unicode subject with invalid suffix",
			subject:         "Fix élément.",
			invalidSuffixes: ".:;",
			expectedValid:   false,
			expectedMessage: "subject has invalid suffix \".\" (invalid suffixes: \".:;\")",
			expectedCode:    "invalid_suffix",
		},
		{
			name:            "Unicode character as invalid suffix",
			subject:         "Update description;",
			invalidSuffixes: ";",
			expectedValid:   false,
			expectedMessage: "subject has invalid suffix \";\" (invalid suffixes: \";\")",
			expectedCode:    "invalid_suffix",
		},
		{
			name:            "Empty subject",
			subject:         "",
			invalidSuffixes: ".:;",
			expectedValid:   false,
			expectedMessage: "subject is empty",
			expectedCode:    "subject_empty",
		},
		{
			name:            "Subject with Unicode invalid suffix",
			subject:         "Add new emoji😊",
			invalidSuffixes: "😊😀",
			expectedValid:   false,
			expectedMessage: "subject has invalid suffix \"😊\" (invalid suffixes: \"😊😀\")",
			expectedCode:    "invalid_suffix",
		},
		{
			name:            "Subject with space as invalid suffix",
			subject:         "Add feature ",
			invalidSuffixes: " \t\n",
			expectedValid:   false,
			expectedMessage: "subject has invalid suffix \" \" (invalid suffixes: \" \\t\\n\")",
			expectedCode:    "invalid_suffix",
		},
		{
			name:            "Subject with tab as invalid suffix",
			subject:         "Add feature\t",
			invalidSuffixes: " \t\n",
			expectedValid:   false,
			expectedMessage: "subject has invalid suffix \"\\t\" (invalid suffixes: \" \\t\\n\")",
			expectedCode:    "invalid_suffix",
		},
		{
			name:            "Valid Unicode subject",
			subject:         "修复问题",
			invalidSuffixes: ".:;",
			expectedValid:   true,
			expectedMessage: "Subject has valid suffix",
		},
		{
			name:            "Default invalid suffixes",
			subject:         "Update feature?",
			invalidSuffixes: "",
			expectedValid:   false,
			expectedMessage: "subject has invalid suffix \"?\" (invalid suffixes: \".,;:!?\")",
			expectedCode:    "invalid_suffix",
		},
	}

	for _, tabletest := range testCases {
		t.Run(tabletest.name, func(t *testing.T) {
			result := rule.ValidateSubjectSuffix(tabletest.subject, tabletest.invalidSuffixes)

			// Check errors
			if tabletest.expectedValid {
				require.Empty(t, result.Errors(), "Did not expect errors")
				require.Equal(t,
					"Subject has valid suffix",
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

				// Check error code
				require.Equal(t,
					tabletest.expectedCode,
					result.Errors()[0].Code,
					"Error code should match expected",
				)

				// Verify context is set
				require.NotEmpty(t, result.Errors()[0].Context, "Error context should not be empty")

				// Check specific context values based on error type
				if tabletest.expectedCode == "invalid_suffix" {
					require.Contains(t, result.Errors()[0].Context, "last_char", "Context should contain last_char")
					require.Contains(t, result.Errors()[0].Context, "invalid_suffixes", "Context should contain invalid_suffixes")
				}

				require.Contains(t, result.Errors()[0].Context, "subject", "Context should contain subject")
			}

			// Check name method
			require.Equal(t, "SubjectSuffix", result.Name(),
				"Name should be 'SubjectSuffix'")

			// Check help method
			helpText := result.Help()
			require.NotEmpty(t, helpText, "Help text should not be empty")

			// Verify help text is appropriate for the error
			if !tabletest.expectedValid {
				errCode := ""
				if len(result.Errors()) > 0 {
					errCode = result.Errors()[0].Code
				}

				switch errCode {
				case "subject_empty":
					require.Contains(t, helpText, "Provide a non-empty subject")
				case "invalid_suffix":
					require.Contains(t, helpText, "Remove the punctuation")
				case "invalid_utf8":
					require.Contains(t, helpText, "valid UTF-8")
				}
			} else {
				require.Contains(t, helpText, "No errors to fix")
			}
		})
	}
}
