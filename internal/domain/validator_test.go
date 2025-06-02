// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package domain_test

import (
	"testing"

	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/stretchr/testify/require"
)

// MockRule for testing - simple and direct.
type MockRule struct {
	name     string
	failures []domain.RuleFailure
}

func (m MockRule) Name() string {
	return m.name
}

func (m MockRule) Validate(_ domain.ValidationContext) []domain.RuleFailure {
	return m.failures
}

func TestValidator_ValidateCommit(t *testing.T) {
	tests := []struct {
		name     string
		rules    []domain.Rule
		commit   domain.Commit
		expected domain.ValidationResult
	}{
		{
			name:  "no rules - always passes",
			rules: []domain.Rule{},
			commit: domain.Commit{
				Hash:    "abc123",
				Subject: "test commit",
			},
			expected: domain.ValidationResult{
				Commit: domain.Commit{
					Hash:    "abc123",
					Subject: "test commit",
				},
				Failures: nil,
			},
		},
		{
			name: "single rule - passes",
			rules: []domain.Rule{
				MockRule{name: "TestRule", failures: nil},
			},
			commit: domain.Commit{
				Hash:    "abc123",
				Subject: "test commit",
			},
			expected: domain.ValidationResult{
				Commit: domain.Commit{
					Hash:    "abc123",
					Subject: "test commit",
				},
				Failures: nil,
			},
		},
		{
			name: "single rule - fails",
			rules: []domain.Rule{
				MockRule{
					name: "TestRule",
					failures: []domain.RuleFailure{
						{Rule: "TestRule", Message: "test failure"},
					},
				},
			},
			commit: domain.Commit{
				Hash:    "abc123",
				Subject: "test commit",
			},
			expected: domain.ValidationResult{
				Commit: domain.Commit{
					Hash:    "abc123",
					Subject: "test commit",
				},
				Failures: []domain.RuleFailure{
					{Rule: "TestRule", Message: "test failure"},
				},
			},
		},
		{
			name: "multiple rules - mixed results",
			rules: []domain.Rule{
				MockRule{name: "Rule1", failures: nil},
				MockRule{
					name: "Rule2",
					failures: []domain.RuleFailure{
						{Rule: "Rule2", Message: "failure 1"},
					},
				},
				MockRule{
					name: "Rule3",
					failures: []domain.RuleFailure{
						{Rule: "Rule3", Message: "failure 2"},
						{Rule: "Rule3", Message: "failure 3"},
					},
				},
			},
			commit: domain.Commit{
				Hash:    "abc123",
				Subject: "test commit",
			},
			expected: domain.ValidationResult{
				Commit: domain.Commit{
					Hash:    "abc123",
					Subject: "test commit",
				},
				Failures: []domain.RuleFailure{
					{Rule: "Rule2", Message: "failure 1"},
					{Rule: "Rule3", Message: "failure 2"},
					{Rule: "Rule3", Message: "failure 3"},
				},
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Create validator
			validator := domain.NewValidator(testCase.rules)

			// Validate
			result := validator.ValidateCommit(testCase.commit, nil, nil)

			// Assert
			require.Equal(t, testCase.expected.Commit, result.Commit)
			require.Equal(t, testCase.expected.Failures, result.Failures)
			require.Equal(t, len(testCase.expected.Failures) == 0, result.Passed())
		})
	}
}
