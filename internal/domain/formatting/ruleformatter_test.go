// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package formatting_test

import (
	"testing"

	"github.com/itiquette/gommitlint/internal/domain/formatting"
	"github.com/itiquette/gommitlint/internal/errors"
	"github.com/stretchr/testify/require"
)

func TestFormatResult(t *testing.T) {
	tests := []struct {
		name     string
		ruleName string
		errors   []errors.ValidationError
		expected string
	}{
		{
			name:     "no errors",
			ruleName: "SubjectLength",
			errors:   nil,
			expected: "SubjectLength: Passed",
		},
		{
			name:     "single error",
			ruleName: "SubjectLength",
			errors: []errors.ValidationError{
				{Rule: "SubjectLength", Code: "too_long", Message: "subject too long"},
			},
			expected: "SubjectLength: Failed with 1 error(s)",
		},
		{
			name:     "multiple errors",
			ruleName: "CommitBody",
			errors: []errors.ValidationError{
				{Rule: "CommitBody", Code: "missing", Message: "body missing"},
				{Rule: "CommitBody", Code: "too_short", Message: "body too short"},
			},
			expected: "CommitBody: Failed with 2 error(s)",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := formatting.FormatResult(testCase.ruleName, testCase.errors)
			require.Equal(t, testCase.expected, result)
		})
	}
}

func TestFormatVerboseResult(t *testing.T) {
	tests := []struct {
		name     string
		ruleName string
		errors   []errors.ValidationError
		expected string
	}{
		{
			name:     "no errors",
			ruleName: "SubjectLength",
			errors:   nil,
			expected: "SubjectLength: All checks passed",
		},
		{
			name:     "with errors",
			ruleName: "SubjectLength",
			errors: []errors.ValidationError{
				{Rule: "SubjectLength", Code: "too_long", Message: "subject exceeds 50 characters"},
			},
			expected: "SubjectLength: Found 1 error(s):\n  1. subject exceeds 50 characters\n",
		},
		{
			name:     "with errors and help",
			ruleName: "ImperativeVerb",
			errors: []errors.ValidationError{
				errors.WithHelp(
					errors.ValidationError{
						Rule:    "ImperativeVerb",
						Code:    "non_imperative",
						Message: "subject should use imperative mood",
					},
					"Use imperative verbs like 'Add' instead of 'Added'",
				),
			},
			expected: "ImperativeVerb: Found 1 error(s):\n  1. subject should use imperative mood\n     Help: Use imperative verbs like 'Add' instead of 'Added'\n",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := formatting.FormatVerboseResult(testCase.ruleName, testCase.errors)
			require.Equal(t, testCase.expected, result)
		})
	}
}

func TestFormatHelp(t *testing.T) {
	tests := []struct {
		name     string
		ruleName string
		errors   []errors.ValidationError
		expected string
	}{
		{
			name:     "no errors - default help",
			ruleName: "SubjectLength",
			errors:   nil,
			expected: "Keep commit subjects under the configured maximum length (default: 50 characters)",
		},
		{
			name:     "errors with help",
			ruleName: "ImperativeVerb",
			errors: []errors.ValidationError{
				errors.WithHelp(
					errors.ValidationError{
						Rule:    "ImperativeVerb",
						Code:    "non_imperative",
						Message: "subject should use imperative mood",
					},
					"Use imperative verbs like 'Add' instead of 'Added'",
				),
			},
			expected: "Use imperative verbs like 'Add' instead of 'Added'",
		},
		{
			name:     "unknown rule - generic help",
			ruleName: "UnknownRule",
			errors:   nil,
			expected: "Follow the UnknownRule rule guidelines",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := formatting.FormatHelp(testCase.ruleName, testCase.errors)
			require.Equal(t, testCase.expected, result)
		})
	}
}

func TestPureFunctionality(t *testing.T) {
	// Test that functions are pure - same input always produces same output
	ruleName := "TestRule"
	errors := []errors.ValidationError{
		{Rule: "TestRule", Code: "test", Message: "test error"},
	}

	// Call multiple times to ensure no side effects
	result1 := formatting.FormatResult(ruleName, errors)
	result2 := formatting.FormatResult(ruleName, errors)
	result3 := formatting.FormatResult(ruleName, errors)

	require.Equal(t, result1, result2)
	require.Equal(t, result2, result3)

	// Test with nil input
	nilResult := formatting.FormatResult(ruleName, nil)
	require.Equal(t, "TestRule: Passed", nilResult)
}
