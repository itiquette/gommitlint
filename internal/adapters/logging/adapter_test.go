// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package logging

import (
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    zerolog.Level
		description string
	}{
		{
			name:        "error level",
			input:       "error",
			expected:    zerolog.ErrorLevel,
			description: "should map error to ErrorLevel (only errors)",
		},
		{
			name:        "warn level",
			input:       "warn",
			expected:    zerolog.WarnLevel,
			description: "should map warn to WarnLevel (warnings and above)",
		},
		{
			name:        "info level",
			input:       "info",
			expected:    zerolog.InfoLevel,
			description: "should map info to InfoLevel (default)",
		},
		{
			name:        "debug level",
			input:       "debug",
			expected:    zerolog.DebugLevel,
			description: "should map debug to DebugLevel (debug and above)",
		},
		{
			name:        "trace level",
			input:       "trace",
			expected:    zerolog.TraceLevel,
			description: "should map trace to TraceLevel (all levels)",
		},
		{
			name:        "invalid level",
			input:       "invalid",
			expected:    zerolog.InfoLevel,
			description: "should fallback to InfoLevel for invalid input",
		},
		{
			name:        "empty string",
			input:       "",
			expected:    zerolog.InfoLevel,
			description: "should fallback to InfoLevel for empty string",
		},
		{
			name:        "case sensitivity",
			input:       "ERROR",
			expected:    zerolog.InfoLevel,
			description: "should fallback to InfoLevel for wrong case",
		},
		{
			name:        "whitespace",
			input:       " info ",
			expected:    zerolog.InfoLevel,
			description: "should fallback to InfoLevel for whitespace input",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := parseLogLevel(testCase.input)

			require.Equal(t, testCase.expected, result, testCase.description)
		})
	}
}

func TestParseLogLevel_IsPureFunction(t *testing.T) {
	// Test that the function is pure (same input always gives same output)
	input := "error"

	// Call multiple times and ensure consistent results
	result1 := parseLogLevel(input)
	result2 := parseLogLevel(input)
	result3 := parseLogLevel(input)

	require.Equal(t, result1, result2, "function should be deterministic")
	require.Equal(t, result2, result3, "function should be deterministic")
	require.Equal(t, zerolog.ErrorLevel, result1, "should consistently return ErrorLevel for error")
}

func TestParseLogLevel_AllAdvertisedLevels(t *testing.T) {
	// Test all levels mentioned in the CLI flag usage
	advertisedLevels := []struct {
		level    string
		expected zerolog.Level
	}{
		{"error", zerolog.ErrorLevel},
		{"warn", zerolog.WarnLevel},
		{"info", zerolog.InfoLevel},
		{"debug", zerolog.DebugLevel},
		{"trace", zerolog.TraceLevel},
	}

	for _, testCase := range advertisedLevels {
		t.Run(testCase.level, func(t *testing.T) {
			result := parseLogLevel(testCase.level)
			require.Equal(t, testCase.expected, result,
				"advertised level %s should map correctly", testCase.level)
		})
	}
}
