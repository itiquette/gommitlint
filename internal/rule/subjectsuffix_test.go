// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package rule_test

import (
	"strings"
	"testing"

	"github.com/itiquette/gommitlint/internal/rule"
	"github.com/stretchr/testify/assert"
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
			expectedMessage: "Valid subject suffix",
		},
		{
			name:            "Subject ending with invalid suffix period",
			subject:         "Update documentation.",
			invalidSuffixes: ".:;",
			expectedValid:   false,
			expectedMessage: "Invalid subject suffix",
			expectedCode:    "invalid_suffix",
		},
		{
			name:            "Subject ending with invalid suffix colon",
			subject:         "Fix bug:",
			invalidSuffixes: ".:;",
			expectedValid:   false,
			expectedMessage: "Invalid subject suffix",
			expectedCode:    "invalid_suffix",
		},
		{
			name:            "Unicode subject with invalid suffix",
			subject:         "Fix élément.",
			invalidSuffixes: ".:;",
			expectedValid:   false,
			expectedMessage: "Invalid subject suffix",
			expectedCode:    "invalid_suffix",
		},
		{
			name:            "Unicode character as invalid suffix",
			subject:         "Update description;",
			invalidSuffixes: ";",
			expectedValid:   false,
			expectedMessage: "Invalid subject suffix",
			expectedCode:    "invalid_suffix",
		},
		{
			name:            "Empty subject",
			subject:         "",
			invalidSuffixes: ".:;",
			expectedValid:   false,
			expectedMessage: "Invalid subject suffix",
			expectedCode:    "subject_empty",
		},
		{
			name:            "Subject with Unicode invalid suffix",
			subject:         "Add new emoji😊",
			invalidSuffixes: "😊😀",
			expectedValid:   false,
			expectedMessage: "Invalid subject suffix",
			expectedCode:    "invalid_suffix",
		},
		{
			name:            "Subject with space as invalid suffix",
			subject:         "Add feature ",
			invalidSuffixes: " \t\n",
			expectedValid:   false,
			expectedMessage: "Invalid subject suffix",
			expectedCode:    "invalid_suffix",
		},
		{
			name:            "Subject with tab as invalid suffix",
			subject:         "Add feature\t",
			invalidSuffixes: " \t\n",
			expectedValid:   false,
			expectedMessage: "Invalid subject suffix",
			expectedCode:    "invalid_suffix",
		},
		{
			name:            "Valid Unicode subject",
			subject:         "修复问题",
			invalidSuffixes: ".:;",
			expectedValid:   true,
			expectedMessage: "Valid subject suffix",
		},
		{
			name:            "Default invalid suffixes",
			subject:         "Update feature?",
			invalidSuffixes: "",
			expectedValid:   false,
			expectedMessage: "Invalid subject suffix",
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
					"Valid subject suffix",
					result.Result(),
					"Message should be valid",
				)
			} else {
				require.NotEmpty(t, result.Errors(), "Expected errors")
				require.Equal(t,
					"Invalid subject suffix",
					result.Result(),
					"Result should indicate invalid suffix",
				)

				// Check the detailed error message
				if tabletest.expectedCode == "invalid_suffix" {
					errorMessage := result.Errors()[0].Error()
					assert.Contains(t, errorMessage, "subject has invalid suffix",
						"Error should mention invalid suffix")

					// For special characters like Unicode emoji or tab, don't check for the exact character
					// as they might be represented differently in the error message
					if tabletest.subject != "" &&
						!strings.Contains(tabletest.subject, "😊") &&
						!strings.Contains(tabletest.subject, "\t") {
						lastChar := string(tabletest.subject[len(tabletest.subject)-1])
						assert.Contains(t, errorMessage, lastChar,
							"Error should contain the last character")
					}
				} else if tabletest.expectedCode == "subject_empty" {
					assert.Contains(t, result.Errors()[0].Error(), "subject is empty",
						"Error should mention empty subject")
				}

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

	// Test VerboseResult method if available
	t.Run("VerboseResult method", func(t *testing.T) {
		// Test if the method exists by type assertion
		validRule := rule.ValidateSubjectSuffix("Valid subject", ".:;")

		if verboseRule, ok := interface{}(validRule).(interface{ VerboseResult() string }); ok {
			// Valid case
			result := verboseRule.VerboseResult()
			assert.Contains(t, result, "Subject ends with valid character",
				"Verbose result should mention valid ending")

			// Invalid case - empty subject
			emptyRule := rule.ValidateSubjectSuffix("", ".:;")
			if verboseEmptyRule, ok := interface{}(emptyRule).(interface{ VerboseResult() string }); ok {
				result = verboseEmptyRule.VerboseResult()
				assert.Contains(t, result, "Subject is empty",
					"Verbose result should mention empty subject")
			}

			// Invalid case - invalid suffix
			invalidRule := rule.ValidateSubjectSuffix("Invalid suffix.", ".:;")
			if verboseInvalidRule, ok := interface{}(invalidRule).(interface{ VerboseResult() string }); ok {
				result = verboseInvalidRule.VerboseResult()
				assert.Contains(t, result, "Subject ends with forbidden character",
					"Verbose result should describe the forbidden character")
			}
		}
	})
}
