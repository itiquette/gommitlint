// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package encutils_test

import (
	"testing"

	"github.com/itiquette/gommitlint/internal/common/encutils"
	"github.com/stretchr/testify/require"
)

func TestIsBase64(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "Empty string",
			input:    "",
			expected: false,
		},
		{
			name:     "Standard Base64",
			input:    "SGVsbG8gV29ybGQ=",
			expected: true,
		},
		{
			name:     "URL-safe Base64",
			input:    "SGVsbG8_V29ybGQ=",
			expected: true,
		},
		{
			name:     "Raw Base64 (no padding)",
			input:    "SGVsbG8gV29ybGQ",
			expected: true,
		},
		{
			name:     "Raw URL-safe Base64 (no padding)",
			input:    "SGVsbG8_V29ybGQ",
			expected: true,
		},
		{
			name:     "Invalid Base64",
			input:    "Not base64!",
			expected: false,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := encutils.IsBase64(testCase.input)
			require.Equal(t, testCase.expected, result)
		})
	}
}

func TestDecodeBase64(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantOutput string
		wantOk     bool
	}{
		{
			name:       "Empty string",
			input:      "",
			wantOutput: "",
			wantOk:     false,
		},
		{
			name:       "Standard Base64",
			input:      "SGVsbG8gV29ybGQ=",
			wantOutput: "Hello World",
			wantOk:     true,
		},
		{
			name:       "URL-safe Base64",
			input:      "SGVsbG8_V29ybGQ=",
			wantOutput: "Hello?World",
			wantOk:     true,
		},
		{
			name:       "Raw Base64 (no padding)",
			input:      "SGVsbG8gV29ybGQ",
			wantOutput: "Hello World",
			wantOk:     true,
		},
		{
			name:       "Invalid Base64",
			input:      "Not base64!",
			wantOutput: "",
			wantOk:     false,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			decoded, ok := encutils.DecodeBase64(testCase.input)

			require.Equal(t, testCase.wantOk, ok)

			if ok {
				require.Equal(t, testCase.wantOutput, string(decoded))
			} else {
				require.Nil(t, decoded)
			}
		})
	}
}
