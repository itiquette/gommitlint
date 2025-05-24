// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package encutils_test

import (
	"encoding/base64"
	"testing"

	"github.com/itiquette/gommitlint/internal/testutils/encutils"
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
			name:     "Invalid URL-safe Base64 with question mark",
			input:    "SGVsbG8_V29ybGQ=",
			expected: true, // Our implementation allows this even though it's not standard URL-safe base64
		},
		{
			name:     "URL-safe with dash and underscore",
			input:    "SGVsbG8tV29ybGRf",
			expected: true,
		},
		{
			name:     "Raw Base64 (no padding)",
			input:    "SGVsbG8gV29ybGQ",
			expected: true,
		},
		{
			name:     "Raw URL-safe Base64 (no padding)",
			input:    "SGVsbG8tV29ybGRf",
			expected: true,
		},
		{
			name:     "Invalid Base64",
			input:    "Not base64!",
			expected: false,
		},
		{
			name:     "Short base64",
			input:    "YWJj", // "abc"
			expected: true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := encutils.IsBase64(testCase.input)
			require.Equal(t, testCase.expected, result, "Input: %s", testCase.input)
		})
	}
}

func TestDecodeBase64(t *testing.T) {
	// Create test strings with different encodings
	stdString := base64.StdEncoding.EncodeToString([]byte("Standard encoding test"))
	urlString := base64.URLEncoding.EncodeToString([]byte("URL encoding test"))
	rawStdString := base64.RawStdEncoding.EncodeToString([]byte("Raw standard encoding test"))
	rawURLString := base64.RawURLEncoding.EncodeToString([]byte("Raw URL encoding test"))

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
			input:      stdString,
			wantOutput: "Standard encoding test",
			wantOk:     true,
		},
		{
			name:       "URL Base64",
			input:      urlString,
			wantOutput: "URL encoding test",
			wantOk:     true,
		},
		{
			name:       "Raw Standard Base64",
			input:      rawStdString,
			wantOutput: "Raw standard encoding test",
			wantOk:     true,
		},
		{
			name:       "Raw URL Base64",
			input:      rawURLString,
			wantOutput: "Raw URL encoding test",
			wantOk:     true,
		},
		{
			name:       "Common base64",
			input:      "SGVsbG8gV29ybGQ=",
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

			require.Equal(t, testCase.wantOk, ok, "Input: %s", testCase.input)

			if ok {
				require.Equal(t, testCase.wantOutput, string(decoded))
			} else {
				require.Nil(t, decoded)
			}
		})
	}
}
