// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package domain_test

import (
	"testing"

	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/itiquette/gommitlint/internal/domain/config"
	"github.com/stretchr/testify/require"
)

// MockRule for testing - simple and direct.
type MockRule struct {
	name     string
	failures []domain.ValidationError
}

func (m MockRule) Name() string {
	return m.name
}

func (m MockRule) Validate(_ domain.Commit, _ domain.Repository, _ *config.Config) []domain.ValidationError {
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
				Errors: nil,
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
				Errors: nil,
			},
		},
		{
			name: "single rule - fails",
			rules: []domain.Rule{
				MockRule{
					name: "TestRule",
					failures: []domain.ValidationError{
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
				Errors: []domain.ValidationError{
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
					failures: []domain.ValidationError{
						{Rule: "Rule2", Message: "failure 1"},
					},
				},
				MockRule{
					name: "Rule3",
					failures: []domain.ValidationError{
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
				Errors: []domain.ValidationError{
					{Rule: "Rule2", Message: "failure 1"},
					{Rule: "Rule3", Message: "failure 2"},
					{Rule: "Rule3", Message: "failure 3"},
				},
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Validate using pure function
			result := domain.ValidateCommit(testCase.commit, testCase.rules, nil, nil)

			// Assert
			require.Equal(t, testCase.expected.Commit, result.Commit)
			require.Equal(t, testCase.expected.Errors, result.Errors)
			require.Equal(t, len(testCase.expected.Errors) == 0, !result.HasFailures())
		})
	}
}
