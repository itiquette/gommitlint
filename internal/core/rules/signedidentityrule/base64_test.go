// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package signedidentityrule

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsBase64(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "valid standard base64",
			input:    "SGVsbG8gd29ybGQ=", // "Hello world" in base64
			expected: true,
		},
		{
			name:     "valid url-safe base64",
			input:    "SGVsbG9fV29ybGQ-", // Using URL-safe characters
			expected: true,
		},
		{
			name:     "valid raw standard base64",
			input:    "SGVsbG8gd29ybGQ", // No padding
			expected: true,
		},
		{
			name:     "invalid base64 (special chars)",
			input:    "SGVsb#G8gd29ybGQ=",
			expected: false,
		},
		{
			name:     "empty string",
			input:    "",
			expected: false, // Empty string is not valid base64
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := isBase64(test.input)
			require.Equal(t, test.expected, result)
		})
	}
}
