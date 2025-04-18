// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package rules_test

import (
	"strings"
	"testing"

	"github.com/itiquette/gommitlint/internal/core/rules"
	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestSubjectSuffixRule(t *testing.T) {
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
			expectedCode:    "missing_subject",
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

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			// Create the rule
			var rule *rules.SubjectSuffixRule

			if testCase.invalidSuffixes != "" {
				rule = rules.NewSubjectSuffixRule(rules.WithInvalidSuffixes(testCase.invalidSuffixes))
			} else {
				rule = rules.NewSubjectSuffixRule() // Use default suffixes
			}

			// Create commit with the test subject
			commit := &domain.CommitInfo{
				Subject: testCase.subject,
			}

			// Validate
			result := rule.Validate(commit)

			// Check results
			if testCase.expectedValid {
				assert.Empty(t, result, "Expected no validation errors")
				assert.Equal(t, testCase.expectedMessage, rule.Result(), "Expected success message doesn't match")
				assert.Contains(t, rule.VerboseResult(), "Subject ends with valid character",
					"VerboseResult should mention valid character")
			} else {
				assert.NotEmpty(t, result, "Expected validation errors")
				assert.Equal(t, testCase.expectedMessage, rule.Result(), "Expected error message doesn't match")

				// Access validation error directly
				validationErr := result[0]
				assert.Equal(t, testCase.expectedCode, validationErr.Code, "Error code should match expected")

				// Check that context contains required fields
				if testCase.expectedCode == "invalid_suffix" {
					assert.Contains(t, validationErr.Context, "last_char", "Context should contain last_char")
					assert.Contains(t, validationErr.Context, "invalid_suffixes", "Context should contain invalid_suffixes")

					// For special characters like Unicode emoji or tab, don't check the exact character
					if testCase.subject != "" &&
						!strings.Contains(testCase.subject, "😊") &&
						!strings.Contains(testCase.subject, "\t") {
						lastChar := string(testCase.subject[len(testCase.subject)-1])
						assert.Contains(t, rule.VerboseResult(), lastChar,
							"VerboseResult should contain the last character")
					}
				}

				// Check context contains subject
				assert.Contains(t, validationErr.Context, "subject", "Context should contain subject")
			}

			// Check name
			assert.Equal(t, "SubjectSuffix", rule.Name(), "Name should be 'SubjectSuffix'")

			// Check help text
			helpText := rule.Help()
			assert.NotEmpty(t, helpText, "Help text should not be empty")

			// Verify help text is appropriate for the error
			if !testCase.expectedValid {
				switch testCase.expectedCode {
				case "missing_subject":
					assert.Contains(t, helpText, "Provide a non-empty subject")
				case "invalid_suffix":
					assert.Contains(t, helpText, "Remove the punctuation")
				case "invalid_utf8":
					assert.Contains(t, helpText, "valid UTF-8")
				}
			} else {
				assert.Contains(t, helpText, "No errors to fix")
			}
		})
	}
}

func TestSubjectSuffixOptions(t *testing.T) {
	t.Run("Default configuration", func(t *testing.T) {
		// No options provided, should use default invalid suffixes
		rule := rules.NewSubjectSuffixRule()

		// Create commit with valid subject
		commit := &domain.CommitInfo{
			Subject: "This is valid",
		}
		errors := rule.Validate(commit)
		assert.Empty(t, errors, "Default config should accept valid subject")

		// Create commit with invalid subject (period at end)
		invalidCommit := &domain.CommitInfo{
			Subject: "This ends with period.",
		}
		errors = rule.Validate(invalidCommit)
		assert.NotEmpty(t, errors, "Default config should reject subject ending with period")

		validationErr := errors[0]
		assert.Equal(t, "invalid_suffix", validationErr.Code, "Error code should be 'invalid_suffix'")
	})

	t.Run("With custom invalid suffixes", func(t *testing.T) {
		// Custom invalid suffixes
		rule := rules.NewSubjectSuffixRule(rules.WithInvalidSuffixes("!@#"))

		// Create commit with invalid suffix (ends with !)
		commit := &domain.CommitInfo{
			Subject: "This ends with exclamation!",
		}
		errors := rule.Validate(commit)
		assert.NotEmpty(t, errors, "Should reject subject with configured invalid suffix")

		validationErr := errors[0]
		assert.Equal(t, "invalid_suffix", validationErr.Code, "Error code should be 'invalid_suffix'")

		// Verify context contains the correct invalid suffixes
		assert.Equal(t, "!@#", validationErr.Context["invalid_suffixes"],
			"Context should contain the custom invalid suffixes")

		// Create commit with period (should be allowed with custom config)
		validCommit := &domain.CommitInfo{
			Subject: "This ends with period.",
		}
		errors = rule.Validate(validCommit)
		assert.Empty(t, errors, "Should accept subject ending with period when not in invalid set")
	})

	t.Run("Empty invalid suffixes", func(t *testing.T) {
		// Empty invalid suffixes should fall back to defaults
		rule := rules.NewSubjectSuffixRule(rules.WithInvalidSuffixes(""))

		// Create commit with question mark (in default invalid suffixes)
		commit := &domain.CommitInfo{
			Subject: "This ends with question mark?",
		}
		errors := rule.Validate(commit)
		assert.NotEmpty(t, errors, "Should reject subject with default invalid suffix")

		validationErr := errors[0]
		assert.Equal(t, rules.DefaultInvalidSuffixes, validationErr.Context["invalid_suffixes"],
			"Should fall back to default invalid suffixes")
	})
}
