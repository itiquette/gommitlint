// SPDX-FileCopyrightText: 2025 Itiquette/Gommitlint
//
// SPDX-License-Identifier: MPL-2.0

package rules

import (
	"fmt"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/require"
)

func TestValidateHeaderLength(t *testing.T) {
	testCases := []struct {
		name           string
		message        string
		maxLength      int
		expectedValid  bool
		expectedLength int
		expectedError  string
	}{
		{
			name:           "Within Default Length",
			message:        "Fix authentication service",
			maxLength:      0,
			expectedValid:  true,
			expectedLength: 26,
			expectedError:  "",
		},
		{
			name:           "Exactly Default Max Length",
			message:        "A message that is exactly the default maximum length allowed",
			maxLength:      0,
			expectedValid:  true,
			expectedLength: 60,
			expectedError:  "",
		},
		{
			name:           "Exceeds Default Length",
			message:        "A very long message that definitely exceeds the default maximum length allowed for commit messages..",
			maxLength:      0,
			expectedValid:  false,
			expectedLength: 100,
			expectedError:  "commit header is too long: 100 characters (maximum allowed: 89)",
		},
		{
			name:           "Custom Max Length Exceeded",
			message:        "A message that exceeds a custom maximum length",
			maxLength:      20,
			expectedValid:  false,
			expectedLength: 46,
			expectedError:  "commit header is too long: 46 characters (maximum allowed: 20)",
		},
		{
			name:           "Unicode Characters",
			message:        "Fix élément with special characters éçà",
			maxLength:      0,
			expectedValid:  true,
			expectedLength: 39,
			expectedError:  "",
		},
		{
			name:           "Unicode Characters Exceeding Length",
			message:        "A very long message with unicode characters like élément that makes it exceed the length límit",
			maxLength:      0,
			expectedValid:  false,
			expectedLength: 94,
			expectedError:  "commit header is too long: 94 characters (maximum allowed: 89)",
		},
		{
			name:           "Empty Message",
			message:        "",
			maxLength:      0,
			expectedValid:  true,
			expectedLength: 0,
			expectedError:  "",
		},
	}

	for _, tabletest := range testCases {
		t.Run(tabletest.name, func(t *testing.T) {
			// Compute UTF-8 aware length
			actualLength := utf8.RuneCountInString(tabletest.message)
			require.Equal(t, tabletest.expectedLength, actualLength, "UTF-8 length calculation should be correct")

			// Perform the check
			check := ValidateHeaderLength(tabletest.message, tabletest.maxLength)

			// Check errors
			if tabletest.expectedValid {
				require.Empty(t, check.Errors(), "Did not expect errors")
				require.Equal(t,
					fmt.Sprintf("Header is %d characters", actualLength),
					check.Message(),
					"Message should report correct length",
				)
			} else {
				require.NotEmpty(t, check.Errors(), "Expected errors")
				require.Equal(t, tabletest.expectedError, check.Message(),
					"Error message should match expected")
			}

			// Check status method
			require.Equal(t, "Header Length", check.Status(),
				"Status should always be 'Header Length'")
		})
	}
}
