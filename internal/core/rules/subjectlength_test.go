// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package rules_test

import (
	"context"
	"strconv"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/itiquette/gommitlint/internal/core/rules"
	"github.com/itiquette/gommitlint/internal/domain"
	appErrors "github.com/itiquette/gommitlint/internal/errors"
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

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Verify test assumptions
			actualLength := utf8.RuneCountInString(testCase.subject)

			// Create commit info
			commit := domain.CommitInfo{
				Subject: testCase.subject,
			}

			// Create rule using functional options pattern
			var rule rules.SubjectLengthRule
			if testCase.maxLength > 0 {
				rule = rules.NewSubjectLengthRule(rules.WithMaxLength(testCase.maxLength))
			} else {
				rule = rules.NewSubjectLengthRule() // Use default
			}

			ctx := context.Background()
			// Validate using the stateless method
			errors := rule.Validate(ctx, commit)

			// Check result
			if testCase.expectError {
				require.NotEmpty(t, errors, "Expected errors but got none")

				// Check error details
				err := errors[0]
				require.Equal(t, "SubjectLength", err.Rule, "Rule name should be set in ValidationError")
				require.Equal(t, string(appErrors.ErrMaxLengthExceeded), err.Code, "Error code should match expected")

				// Check context - we now use actual_length instead of subject_length
				require.Contains(t, err.Context, "actual_length", "Context should contain subject length")
				require.Contains(t, err.Context, "max_length", "Context should contain maximum length")
				require.Equal(t, strconv.Itoa(actualLength), err.Context["actual_length"],
					"Subject length in context should match expected length")

				// Test pure function implementation explicitly
				_, updatedRule := rules.ValidateSubjectLengthWithState(rule, commit)
				require.Equal(t, "Subject too long", updatedRule.Result(), "Result message should indicate subject is too long")
				require.True(t, updatedRule.HasErrors(), "HasErrors should return true for invalid subjects")
			} else {
				require.Empty(t, errors, "Expected no errors but got: %v", errors)

				// Test pure function implementation explicitly
				_, updatedRule := rules.ValidateSubjectLengthWithState(rule, commit)
				require.Equal(t, "Subject length OK", updatedRule.Result(), "Result message should indicate length is OK")
				require.False(t, updatedRule.HasErrors(), "HasErrors should return false for valid subjects")
			}

			// Check name
			require.Equal(t, "SubjectLength", rule.Name(), "Name should always be 'SubjectLength'")

			// For verbose result and help message, we need a rule with state
			_, ruleWithState := rules.ValidateSubjectLengthWithState(rule, commit)

			// Check verbose result
			require.NotEmpty(t, ruleWithState.VerboseResult(), "VerboseResult should not be empty")

			// Check help message
			require.NotEmpty(t, ruleWithState.Help(), "Help should not be empty")
		})
	}
}

func TestSubjectLengthRuleWithConfig(t *testing.T) {
	// Create a rule with options
	rule := rules.NewSubjectLengthRule(
		rules.WithMaxLength(50),
	)

	// Verify the rule uses the provided value
	commit := domain.CommitInfo{
		Subject: strings.Repeat("a", 51), // One character over the limit
	}

	ctx := context.Background()
	// Validate and check for error
	errors := rule.Validate(ctx, commit)
	require.Len(t, errors, 1, "Should have exactly one error")
	require.Equal(t, string(appErrors.ErrMaxLengthExceeded), errors[0].Code)

	// Check context values
	require.Equal(t, "51", errors[0].Context["actual_length"])
	require.Equal(t, "50", errors[0].Context["max_length"])
}

// Note: Mock provider implementation has been removed as it's not used in the tests
// The tests use functional options pattern (rules.WithMaxLength) instead of configuration providers
