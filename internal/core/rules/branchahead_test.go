// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package rules

import (
	"context"
	"errors"
	"testing"

	"github.com/itiquette/gommitlint/internal/domain"
	appErrors "github.com/itiquette/gommitlint/internal/errors"
	"github.com/stretchr/testify/require"
)

// mockCommitAnalyzer is a test implementation of the CommitAnalyzer interface.
type mockCommitAnalyzer struct {
	commitsAhead  int
	err           error
	refBranchName string
}

// GetCommitsAhead returns the number of commits ahead stored in the mock.
func (m *mockCommitAnalyzer) GetCommitsAhead(_ context.Context, refBranch string) (int, error) {
	// Save the reference branch name for verification
	m.refBranchName = refBranch

	return m.commitsAhead, m.err
}

// TestBranchAheadRule tests the basic functionality of the BranchAheadRule.
func TestBranchAheadRule(t *testing.T) {
	tests := []struct {
		name               string
		maxCommitsAheadCfg int
		analyzer           domain.CommitAnalyzer
		expectedErrors     bool
		expectedErrorCode  string
	}{
		{
			name:               "valid - within limit",
			maxCommitsAheadCfg: 10,
			analyzer: &mockCommitAnalyzer{
				commitsAhead: 5,
				err:          nil,
			},
			expectedErrors:    false,
			expectedErrorCode: "",
		},
		{
			name:               "invalid - exceeds limit",
			maxCommitsAheadCfg: 5,
			analyzer: &mockCommitAnalyzer{
				commitsAhead: 10,
				err:          nil,
			},
			expectedErrors:    true,
			expectedErrorCode: string(appErrors.ErrTooManyCommits),
		},
		{
			name:               "analyzer error",
			maxCommitsAheadCfg: 10,
			analyzer: &mockCommitAnalyzer{
				commitsAhead: 0,
				err:          errors.New("test error"),
			},
			expectedErrors:    true,
			expectedErrorCode: string(appErrors.ErrGitOperationFailed),
		},
		{
			name:               "nil repository",
			maxCommitsAheadCfg: 10,
			analyzer:           nil,
			expectedErrors:     true,
			expectedErrorCode:  string(appErrors.ErrInvalidRepo),
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Create rule with options
			rule := NewBranchAheadRule(
				WithMaxCommitsAhead(testCase.maxCommitsAheadCfg),
				WithRepositoryGetter(func() domain.CommitAnalyzer {
					return testCase.analyzer
				}),
			)

			// Check rule name
			require.Equal(t, "BranchAhead", rule.Name(), "Rule name should be BranchAhead")

			// Create an empty commit to validate
			commit := domain.CommitInfo{
				Hash:    "test-commit",
				Subject: "test subject",
			}

			// Validate the commit
			errors := rule.Validate(context.Background(), commit)

			// Check for expected errors
			if testCase.expectedErrors {
				require.NotEmpty(t, errors, "Expected validation errors")
				require.Equal(t, testCase.expectedErrorCode, errors[0].Code, "Error code mismatch")
			} else {
				require.Empty(t, errors, "Expected no validation errors")
			}
		})
	}
}

// TestBranchAheadRule_WithOptions tests the functional options pattern.
func TestBranchAheadRule_WithOptions(t *testing.T) {
	t.Run("default options", func(t *testing.T) {
		rule := NewBranchAheadRule()

		require.Equal(t, 10, rule.maxCommitsAhead, "Default max commits ahead should be 10")
		require.Equal(t, "main", rule.reference, "Default reference branch should be 'main'")
	})

	t.Run("custom max commits ahead", func(t *testing.T) {
		rule := NewBranchAheadRule(
			WithMaxCommitsAhead(5),
		)

		require.Equal(t, 5, rule.maxCommitsAhead, "Custom max commits ahead should be 5")
		require.Equal(t, "main", rule.reference, "Reference branch should remain default")
	})

	t.Run("custom reference branch", func(t *testing.T) {
		rule := NewBranchAheadRule(
			WithReference("develop"),
		)

		require.Equal(t, 10, rule.maxCommitsAhead, "Max commits ahead should remain default")
		require.Equal(t, "develop", rule.reference, "Custom reference branch should be 'develop'")
	})

	t.Run("multiple options", func(t *testing.T) {
		rule := NewBranchAheadRule(
			WithMaxCommitsAhead(3),
			WithReference("staging"),
		)

		require.Equal(t, 3, rule.maxCommitsAhead, "Custom max commits ahead should be 3")
		require.Equal(t, "staging", rule.reference, "Custom reference branch should be 'staging'")
	})
}

// TestBranchAheadRule_ReferenceBranch tests that the correct reference branch is used.
func TestBranchAheadRule_ReferenceBranch(t *testing.T) {
	// Create a mock analyzer to capture what reference branch is passed
	mockAnalyzer := &mockCommitAnalyzer{
		commitsAhead: 3,
		err:          nil,
	}

	// Create rule with custom reference branch
	rule := NewBranchAheadRule(
		WithReference("develop"),
		WithRepositoryGetter(func() domain.CommitAnalyzer {
			return mockAnalyzer
		}),
	)

	// Validate any commit (it doesn't matter what commit for this test)
	commit := domain.CommitInfo{
		Hash:    "test-commit",
		Subject: "test subject",
	}

	// Run validation (this should use the analyzer)
	rule.Validate(context.Background(), commit)

	// Check what reference branch was passed to the analyzer
	require.Equal(t, "develop", mockAnalyzer.refBranchName,
		"Expected reference branch 'develop' to be passed to analyzer")
}

// TestBranchAheadRule_Configuration tests the rule with various configurations.
func TestBranchAheadRule_Configuration(t *testing.T) {
	tests := []struct {
		name            string
		maxCommitsAhead int
		reference       string
		commitsAhead    int
		analyzerErr     error
		wantError       bool
		wantErrorCode   string
	}{
		{
			name:            "within limit",
			maxCommitsAhead: 5,
			reference:       "main",
			commitsAhead:    3,
			wantError:       false,
		},
		{
			name:            "exceeds limit",
			maxCommitsAhead: 5,
			reference:       "main",
			commitsAhead:    8,
			wantError:       true,
			wantErrorCode:   string(appErrors.ErrTooManyCommits),
		},
		{
			name:            "disabled check (max=0)",
			maxCommitsAhead: 0,
			reference:       "main",
			commitsAhead:    100,
			wantError:       false,
		},
		{
			name:            "git error",
			maxCommitsAhead: 5,
			reference:       "main",
			analyzerErr:     errors.New("git error"),
			wantError:       true,
			wantErrorCode:   string(appErrors.ErrGitOperationFailed),
		},
		{
			name:            "custom reference branch",
			maxCommitsAhead: 5,
			reference:       "develop",
			commitsAhead:    3,
			wantError:       false,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Create mock analyzer
			analyzer := &mockCommitAnalyzer{
				commitsAhead: testCase.commitsAhead,
				err:          testCase.analyzerErr,
			}

			// Create rule with configuration
			rule := NewBranchAheadRule(
				WithMaxCommitsAhead(testCase.maxCommitsAhead),
				WithReference(testCase.reference),
				WithRepositoryGetter(func() domain.CommitAnalyzer {
					return analyzer
				}),
			)

			// Validate
			commit := domain.CommitInfo{
				Hash:    "test-commit",
				Subject: "test subject",
			}
			errors := rule.Validate(context.Background(), commit)

			// Check results
			if testCase.wantError {
				require.NotEmpty(t, errors, "Expected validation errors")
				require.Equal(t, testCase.wantErrorCode, errors[0].Code)
			} else {
				require.Empty(t, errors, "Expected no validation errors")
			}

			// Verify correct reference was used
			if testCase.analyzerErr == nil {
				require.Equal(t, testCase.reference, analyzer.refBranchName)
			}
		})
	}
}
