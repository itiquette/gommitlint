// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package rules_test

import (
	"strconv"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/itiquette/gommitlint/internal/core/rules"
	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/stretchr/testify/require"
)

func TestSubjectLengthRule(t *testing.T) {
	tests := []struct {
		name        string
		subject     string
		maxLength   int
		expectError bool
	}{
		{
			name:        "Within default length",
			subject:     "Fix authentication service",
			maxLength:   0, // Use default
			expectError: false,
		},
		{
			name:        "Exactly max length",
			subject:     strings.Repeat("a", 80),
			maxLength:   80,
			expectError: false,
		},
		{
			name:        "Exceeds max length",
			subject:     strings.Repeat("a", 81),
			maxLength:   80,
			expectError: true,
		},
		{
			name:        "Empty message",
			subject:     "",
			maxLength:   80,
			expectError: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Verify test assumptions
			actualLength := utf8.RuneCountInString(test.subject)

			// Create commit info
			commit := &domain.CommitInfo{
				Subject: test.subject,
			}

			// Create rule
			rule := rules.NewSubjectLengthRule(test.maxLength)

			// Validate
			errors := rule.Validate(commit)

			// Check result
			if test.expectError {
				require.NotEmpty(t, errors, "Expected errors but got none")

				// Check error details
				err := errors[0]
				require.Equal(t, "SubjectLength", err.Rule, "Rule name should be set in ValidationError")
				require.Equal(t, string(domain.ValidationErrorTooLong), err.Code, "Error code should match expected")

				// Check context
				require.Contains(t, err.Context, "actual_length", "Context should contain actual length")
				require.Contains(t, err.Context, "max_length", "Context should contain maximum length")
				require.Equal(t, strconv.Itoa(actualLength), err.Context["actual_length"],
					"Actual length in context should match expected length")

				// Check result message
				require.Equal(t, "Subject too long", rule.Result(), "Result message should indicate subject is too long")
			} else {
				require.Empty(t, errors, "Expected no errors but got: %v", errors)
				require.Equal(t, "Subject length OK", rule.Result(), "Result message should indicate length is OK")
			}

			// Check name
			require.Equal(t, "SubjectLength", rule.Name(), "Name should always be 'SubjectLength'")

			// Check verbose result
			require.NotEmpty(t, rule.VerboseResult(), "VerboseResult should not be empty")

			// Check help message
			require.NotEmpty(t, rule.Help(), "Help should not be empty")
		})
	}
}
