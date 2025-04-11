// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package rule_test

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/itiquette/gommitlint/internal/rule"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateSubjectLength(t *testing.T) {
	testCases := []struct {
		name           string
		subject        string
		maxLength      int
		expectedValid  bool
		expectedLength int
		expectedError  string
		expectedCode   string
	}{
		{
			name:           "Within default length",
			subject:        "Fix authentication service",
			maxLength:      0,
			expectedValid:  true,
			expectedLength: 26,
			expectedError:  "",
		},
		{
			name:           "Exactly default max length",
			subject:        "A message that is exactly the default maximum length allowed",
			maxLength:      0,
			expectedValid:  true,
			expectedLength: 60,
			expectedError:  "",
		},
		{
			name:           "Exceeds default length",
			subject:        "A very long message that definitely exceeds the default maximum length allowed for cddddddddddddddddd",
			maxLength:      0,
			expectedValid:  false,
			expectedLength: rule.DefaultMaxCommitSubjectLength + 1,
			expectedCode:   "subject_too_long",
			expectedError:  "subject too long: " + strconv.Itoa(rule.DefaultMaxCommitSubjectLength+1) + " characters (maximum allowed: " + strconv.Itoa(rule.DefaultMaxCommitSubjectLength) + ")",
		},
		{
			name:           "Custom max length exceeded",
			subject:        "A message that exceeds a custom maximum length",
			maxLength:      20,
			expectedValid:  false,
			expectedLength: 46,
			expectedCode:   "subject_too_long",
			expectedError:  "subject too long: 46 characters (maximum allowed: 20)",
		},
		{
			name:           "Unicode characters",
			subject:        "Fix élément with special characters éçà",
			maxLength:      0,
			expectedValid:  true,
			expectedLength: 39,
			expectedError:  "",
		},
		{
			name:           "Unicode characters exceeding length",
			subject:        "A very long message with unicode characters like élément that makes it exceed the length límit",
			maxLength:      89,
			expectedValid:  false,
			expectedLength: 94,
			expectedCode:   "subject_too_long",
			expectedError:  "subject too long: 94 characters (maximum allowed: 89)",
		},
		{
			name:           "Empty message",
			subject:        "",
			maxLength:      0,
			expectedValid:  true,
			expectedLength: 0,
			expectedError:  "",
		},
		{
			name:          "Unusual unicode characters",
			subject:       "Fix bug with 𝓤𝓷𝓲𝓬𝓸𝓭𝓮 𝕗𝕒𝕟𝕔𝕪 𝖙𝖊𝖝𝖙 red at for 👨‍👩‍👧‍👦🇺🇦",
			maxLength:     50,
			expectedValid: false,
			// Each code point counts as one rune in Go, including fancy Unicode and emoji components
			expectedLength: 52,
			expectedCode:   "subject_too_long",
			expectedError:  "subject too long: 52 characters (maximum allowed: 50)",
		},
		{
			name:           "Control characters",
			subject:        "Message with control\u0000characters\u0007and\tnull\u0000bytes",
			maxLength:      50,
			expectedValid:  true,
			expectedLength: 46,
			expectedError:  "",
		},
		{
			name:           "Very long string",
			subject:        strings.Repeat("abcdefghij", 100), // 1000 printable characters
			maxLength:      0,
			expectedValid:  false,
			expectedLength: 1000,
			expectedCode:   "subject_too_long",
			expectedError:  fmt.Sprintf("subject too long: %d characters (maximum allowed: %d)", 1000, rule.DefaultMaxCommitSubjectLength),
		},
	}

	for _, tabletest := range testCases {
		t.Run(tabletest.name, func(t *testing.T) {
			// Compute UTF-8 aware length
			actualLength := utf8.RuneCountInString(tabletest.subject)
			require.Equal(t, tabletest.expectedLength, actualLength, "UTF-8 length calculation should be correct")

			// Perform the rule
			rule := rule.ValidateSubjectLength(tabletest.subject, tabletest.maxLength)

			// Check errors
			if tabletest.expectedValid {
				require.Empty(t, rule.Errors(), "Did not expect errors")
				require.Equal(t,
					fmt.Sprintf("Subject is %d characters", actualLength),
					rule.Result(),
					"Message should report correct length",
				)
			} else {
				require.NotEmpty(t, rule.Errors(), "Expected errors")
				require.Equal(t, tabletest.expectedError, rule.Result(),
					"Error message should match expected")

				// Verify error code if specified
				if tabletest.expectedCode != "" {
					assert.Equal(t, tabletest.expectedCode, rule.Errors()[0].Code,
						"Error code should match expected")
				}

				// Verify rule name is set in ValidationError
				assert.Equal(t, "SubjectLength", rule.Errors()[0].Rule,
					"Rule name should be set in ValidationError")

				// Verify context information is present
				assert.Contains(t, rule.Errors()[0].Context, "actual_length",
					"Context should contain actual length")
				assert.Contains(t, rule.Errors()[0].Context, "max_length",
					"Context should contain maximum length")

				// Check if context values are correct
				assert.Equal(t, strconv.Itoa(tabletest.expectedLength), rule.Errors()[0].Context["actual_length"],
					"Actual length in context should match expected length")

				expectedMaxLength := tabletest.maxLength
				if expectedMaxLength == 0 {
					expectedMaxLength = 100
				}

				assert.Equal(t, strconv.Itoa(expectedMaxLength), rule.Errors()[0].Context["max_length"],
					"Max length in context should match expected max length")
			}

			// Check name method
			require.Equal(t, "SubjectLength", rule.Name(),
				"Name should always be 'SubjectLength'")

			// Check Help method is available
			if len(rule.Errors()) > 0 {
				helpText := rule.Help()
				require.NotEmpty(t, helpText, "Help text should not be empty for errors")
				require.Contains(t, helpText, "Shorten your commit message", "Help should provide guidance on fixing long subjects")

				// Check if help text includes the max length from context
				if maxLength, ok := rule.Errors()[0].Context["max_length"]; ok {
					assert.Contains(t, helpText, maxLength, "Help text should include the maximum length")
				}
			}
		})
	}

	// Test for case with explicit context information in the help text
	t.Run("Help message includes max length", func(t *testing.T) {
		customMaxLength := 42
		longSubject := strings.Repeat("a", customMaxLength+1)
		rule := rule.ValidateSubjectLength(longSubject, customMaxLength)

		helpText := rule.Help()
		assert.Contains(t, helpText, strconv.Itoa(customMaxLength),
			"Help text should include the custom maximum length")
	})
}
