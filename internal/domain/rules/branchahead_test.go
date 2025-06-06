// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package rules

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/itiquette/gommitlint/internal/domain/config"
)

// mockRepository is a test implementation of the Repository interface.
// Only implements GetCommitsAhead for testing the branchahead rule.
type mockRepository struct {
	commitsAhead  int
	err           error
	refBranchName string
}

// GetCommitsAhead returns the number of commits ahead stored in the mock.
func (m *mockRepository) GetCommitsAhead(_ context.Context, refBranch string) (int, error) {
	// Save the reference branch name for verification
	m.refBranchName = refBranch

	return m.commitsAhead, m.err
}

// GetCommitsAheadCount returns the number of commits ahead (alias for GetCommitsAhead).
func (m *mockRepository) GetCommitsAheadCount(ctx context.Context, refBranch string) (int, error) {
	return m.GetCommitsAhead(ctx, refBranch)
}

// Stub implementations for Repository interface (not used in branchahead tests).
func (m *mockRepository) GetCommit(_ context.Context, _ string) (domain.Commit, error) {
	return domain.Commit{}, nil
}
func (m *mockRepository) GetCommits(_ context.Context, _ int) ([]domain.Commit, error) {
	return nil, nil
}
func (m *mockRepository) GetCommitRange(_ context.Context, _, _ string) ([]domain.Commit, error) {
	return nil, nil
}
func (m *mockRepository) GetHeadCommits(_ context.Context, _ int) ([]domain.Commit, error) {
	return nil, nil
}
func (m *mockRepository) GetCurrentBranch(_ context.Context) (string, error) { return "", nil }
func (m *mockRepository) GetRepositoryName(_ context.Context) string         { return "" }
func (m *mockRepository) IsValid(_ context.Context) (bool, error)            { return true, nil }

// TestBranchAheadRule tests the basic functionality of the BranchAheadRule.
func TestBranchAheadRule(t *testing.T) {
	tests := []struct {
		name               string
		maxCommitsAheadCfg int
		analyzer           domain.Repository
		expectedErrors     bool
		expectedErrorCode  string
	}{
		{
			name:               "valid - within limit",
			maxCommitsAheadCfg: 10,
			analyzer: &mockRepository{
				commitsAhead: 5,
				err:          nil,
			},
			expectedErrors:    false,
			expectedErrorCode: "",
		},
		{
			name:               "invalid - exceeds limit",
			maxCommitsAheadCfg: 5,
			analyzer: &mockRepository{
				commitsAhead: 10,
				err:          nil,
			},
			expectedErrors:    true,
			expectedErrorCode: string(domain.ErrTooManyCommits),
		},
		{
			name:               "analyzer error",
			maxCommitsAheadCfg: 10,
			analyzer: &mockRepository{
				commitsAhead: 0,
				err:          errors.New("test error"),
			},
			expectedErrors:    false, // Now gracefully skips errors
			expectedErrorCode: "",
		},
		{
			name:               "nil repository",
			maxCommitsAheadCfg: 10,
			analyzer:           nil,
			expectedErrors:     false, // Now gracefully skips when no repository
			expectedErrorCode:  "",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Create config with test values
			cfg := config.Config{
				Repo: config.RepoConfig{
					MaxCommitsAhead: testCase.maxCommitsAheadCfg,
					ReferenceBranch: "main",
				},
			}

			// Create rule with config only (simplified interface)
			rule := NewBranchAheadRule(cfg)

			// Check rule name
			require.Equal(t, "BranchAhead", rule.Name(), "Rule name should be BranchAhead")

			// Validate using new signature
			commit := domain.Commit{Hash: "test", Subject: "test commit"}
			failures := rule.Validate(commit, testCase.analyzer, cfg)

			// Check for expected errors
			if testCase.expectedErrors {
				require.NotEmpty(t, failures, "Expected validation failures")
				// Check the message contains expected content
				if testCase.analyzer == nil {
					require.Contains(t, failures[0].Message, "Repository is nil", "Error message should indicate nil repository")
				} else if mockRepo, ok := testCase.analyzer.(*mockRepository); ok && mockRepo.err != nil {
					require.Contains(t, failures[0].Message, "Failed to check", "Error message should indicate failure")
				} else {
					require.Contains(t, failures[0].Message, "commits ahead", "Error message should mention commits ahead")
				}
			} else {
				require.Empty(t, failures, "Expected no validation failures")
			}
		})
	}
}

// TestBranchAheadRule_WithConfig tests the config-based pattern.
func TestBranchAheadRule_WithConfig(t *testing.T) {
	t.Run("default options", func(t *testing.T) {
		cfg := config.Config{}
		rule := NewBranchAheadRule(cfg)

		require.Equal(t, 10, rule.maxCommitsAhead, "Default max commits ahead should be 10")
		require.Equal(t, "main", rule.reference, "Default reference branch should be 'main'")
	})

	t.Run("custom max commits ahead", func(t *testing.T) {
		cfg := config.Config{
			Repo: config.RepoConfig{
				MaxCommitsAhead: 5,
			},
		}
		rule := NewBranchAheadRule(cfg)

		require.Equal(t, 5, rule.maxCommitsAhead, "Custom max commits ahead should be 5")
		require.Equal(t, "main", rule.reference, "Reference branch should remain default")
	})

	t.Run("custom reference branch", func(t *testing.T) {
		cfg := config.Config{
			Repo: config.RepoConfig{
				ReferenceBranch: "develop",
			},
		}
		rule := NewBranchAheadRule(cfg)

		require.Equal(t, 10, rule.maxCommitsAhead, "Max commits ahead should remain default")
		require.Equal(t, "develop", rule.reference, "Custom reference branch should be 'develop'")
	})

	t.Run("multiple options", func(t *testing.T) {
		cfg := config.Config{
			Repo: config.RepoConfig{
				MaxCommitsAhead: 3,
				ReferenceBranch: "staging",
			},
		}
		rule := NewBranchAheadRule(cfg)

		require.Equal(t, 3, rule.maxCommitsAhead, "Custom max commits ahead should be 3")
		require.Equal(t, "staging", rule.reference, "Custom reference branch should be 'staging'")
	})
}

// TestBranchAheadRule_ReferenceBranch tests that the correct reference branch is used.
func TestBranchAheadRule_ReferenceBranch(t *testing.T) {
	// Create a mock analyzer to capture what reference branch is passed
	mockAnalyzer := &mockRepository{
		commitsAhead: 3,
		err:          nil,
	}

	// Create config
	cfg := config.Config{
		Repo: config.RepoConfig{
			ReferenceBranch: "develop",
		},
	}

	// Create rule with custom reference branch
	rule := NewBranchAheadRule(cfg)

	// Run validation using new signature
	commit := domain.Commit{Hash: "test", Subject: "test commit"}
	rule.Validate(commit, mockAnalyzer, cfg)

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
			wantErrorCode:   string(domain.ErrTooManyCommits),
		},
		{
			name:            "disabled check (max=0)",
			maxCommitsAhead: 0,
			reference:       "main",
			commitsAhead:    100,
			wantError:       true, // 0 is converted to default 10, so 100 > 10 = error
			wantErrorCode:   string(domain.ErrTooManyCommits),
		},
		{
			name:            "git error",
			maxCommitsAhead: 5,
			reference:       "main",
			analyzerErr:     errors.New("git error"),
			wantError:       false, // Now gracefully skips git errors
			wantErrorCode:   "",
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
			analyzer := &mockRepository{
				commitsAhead: testCase.commitsAhead,
				err:          testCase.analyzerErr,
			}

			// Create config and dependencies
			cfg := config.Config{
				Repo: config.RepoConfig{
					MaxCommitsAhead: testCase.maxCommitsAhead,
					ReferenceBranch: testCase.reference,
				},
			}

			// Create rule with configuration
			rule := NewBranchAheadRule(cfg)

			// Validate using new signature
			commit := domain.Commit{Hash: "test", Subject: "test commit"}
			failures := rule.Validate(commit, analyzer, cfg)

			// Check results
			if testCase.wantError {
				require.NotEmpty(t, failures, "Expected validation failures")
				// Check the message content since RuleFailure doesn't have Code field
				if testCase.analyzerErr != nil {
					require.Contains(t, failures[0].Message, "Failed to check", "Error message should indicate failure")
				} else {
					require.Contains(t, failures[0].Message, "commits ahead", "Error message should mention commits ahead")
				}
			} else {
				require.Empty(t, failures, "Expected no validation failures")
			}

			// Verify correct reference was used
			if testCase.analyzerErr == nil {
				require.Equal(t, testCase.reference, analyzer.refBranchName)
			}
		})
	}
}
