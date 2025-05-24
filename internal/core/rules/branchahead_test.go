// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package rules

import (
	"context"
	"errors"
	"testing"

	"github.com/itiquette/gommitlint/internal/common/contextx"
	"github.com/itiquette/gommitlint/internal/config"
	"github.com/itiquette/gommitlint/internal/config/types"
	"github.com/itiquette/gommitlint/internal/domain"
	appErrors "github.com/itiquette/gommitlint/internal/errors"
	testconfig "github.com/itiquette/gommitlint/internal/testutils/config"
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

// TestBranchAheadRule_WithContext tests configuration via context.
func TestBranchAheadRule_WithContext(t *testing.T) {
	// We'll create test contexts with different configurations
	testCases := []struct {
		name           string
		configModifier func(types.Config) types.Config
		commitsAhead   int
		wantErrors     bool
	}{
		{
			name: "context: default limits",
			configModifier: func(cfg types.Config) types.Config {
				return cfg // No changes
			},
			commitsAhead: 5,
			wantErrors:   false,
		},
		{
			name: "context: custom max ahead within limit",
			configModifier: func(cfg types.Config) types.Config {
				cfg.Repo.MaxCommitsAhead = 10

				return cfg
			},
			commitsAhead: 8,
			wantErrors:   false,
		},
		{
			name: "context: custom max ahead exceeded",
			configModifier: func(cfg types.Config) types.Config {
				cfg.Repo.MaxCommitsAhead = 4

				return cfg
			},
			commitsAhead: 5,
			wantErrors:   true,
		},
		{
			name: "context: custom reference branch",
			configModifier: func(cfg types.Config) types.Config {
				cfg.Repo.Branch = "develop"

				return cfg
			},
			commitsAhead: 3,
			wantErrors:   false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			// Apply custom configuration
			cfg := testconfig.NewBuilder().Build()
			cfg = testCase.configModifier(cfg)
			ctx := context.Background()

			// Create rule with configuration
			configuredRule := NewBranchAheadRule(
				WithRepositoryGetter(func() domain.CommitAnalyzer {
					return &mockCommitAnalyzer{
						commitsAhead: testCase.commitsAhead,
					}
				}),
				WithMaxCommitsAhead(cfg.Repo.MaxCommitsAhead),
				WithReference(cfg.Repo.Branch),
			)

			// Check that rule has correct values
			require.Equal(t, cfg.Repo.MaxCommitsAhead, configuredRule.maxCommitsAhead,
				"Rule should have max commits ahead from config")
			require.Equal(t, cfg.Repo.Branch, configuredRule.reference,
				"Rule should have reference branch from config")

			// Create an empty commit to validate
			commit := domain.CommitInfo{
				Hash:    "test-commit",
				Subject: "test subject",
			}

			// Validate the commit
			errors := configuredRule.Validate(ctx, commit)

			// Check for expected errors
			if testCase.wantErrors {
				require.NotEmpty(t, errors, "Expected validation errors")
			} else {
				require.Empty(t, errors, "Expected no validation errors")
			}
		})
	}
}

func TestBranchAheadRuleWithConfig(t *testing.T) {
	// More direct tests for configuration handling
	tests := []struct {
		name        string
		configSetup func() types.Config
		analyzer    domain.CommitAnalyzer
		wantErrors  bool
		description string
	}{
		{
			name: "commits ahead check enabled, within limit",
			configSetup: func() types.Config {
				cfg := config.NewDefaultConfig()
				cfg.Repo.MaxCommitsAhead = 5
				cfg.Repo.Branch = "main"

				return cfg
			},
			analyzer: &mockCommitAnalyzer{
				commitsAhead: 3,
				err:          nil,
			},
			wantErrors:  false,
			description: "Should pass when number of commits ahead is within the limit",
		},
		{
			name: "commits ahead check enabled, exceeds limit",
			configSetup: func() types.Config {
				cfg := config.NewDefaultConfig()
				cfg.Repo.MaxCommitsAhead = 5
				cfg.Repo.Branch = "main"

				return cfg
			},
			analyzer: &mockCommitAnalyzer{
				commitsAhead: 8,
				err:          nil,
			},
			wantErrors:  true,
			description: "Should fail when number of commits ahead exceeds the limit",
		},
		{
			name: "commits ahead check disabled",
			configSetup: func() types.Config {
				cfg := config.NewDefaultConfig()
				cfg.Repo.MaxCommitsAhead = 0 // Disabled
				cfg.Repo.Branch = "main"

				return cfg
			},
			analyzer: &mockCommitAnalyzer{
				commitsAhead: 10,
				err:          nil,
			},
			wantErrors:  false,
			description: "Should pass when the check is disabled",
		},
		{
			name: "commits ahead check enabled, analyzer error",
			configSetup: func() types.Config {
				cfg := config.NewDefaultConfig()
				cfg.Repo.MaxCommitsAhead = 5
				cfg.Repo.Branch = "main"

				return cfg
			},
			analyzer: &mockCommitAnalyzer{
				commitsAhead: 0,
				err:          errors.New("git error"),
			},
			wantErrors:  true,
			description: "Should fail when the analyzer returns an error",
		},
		{
			name: "custom reference branch",
			configSetup: func() types.Config {
				cfg := config.NewDefaultConfig()
				cfg.Repo.MaxCommitsAhead = 5
				cfg.Repo.Branch = "develop"

				return cfg
			},
			analyzer: &mockCommitAnalyzer{
				commitsAhead: 3,
				err:          nil,
			},
			wantErrors:  false,
			description: "Should use the custom reference branch from config",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Create the config using the setup function
			cfg := testCase.configSetup()

			// Create context with test config
			ctx := context.Background()
			adapter := testconfig.NewAdapter(cfg)
			ctx = contextx.WithConfig(ctx, adapter.Adapter)

			// Log test parameters
			t.Logf("Test '%s': maxCommitsAhead=%d", testCase.name, cfg.Repo.MaxCommitsAhead)

			// Create rule with the analyzer and configuration
			configuredRule := NewBranchAheadRule(
				WithRepositoryGetter(func() domain.CommitAnalyzer {
					return testCase.analyzer
				}),
				WithMaxCommitsAhead(cfg.Repo.MaxCommitsAhead),
				WithReference(cfg.Repo.Branch),
			)

			// Log the configured rule parameters
			t.Logf("Configured rule: maxCommitsAhead=%d, reference=%s",
				configuredRule.maxCommitsAhead, configuredRule.reference)

			// Create empty commit info for validation
			commit := domain.CommitInfo{
				Subject: "Test commit",
				Body:    "",
			}

			// Validate commit
			errors := configuredRule.Validate(ctx, commit)

			// Log validation results
			t.Logf("Validation errors (%d): %v", len(errors), errors)

			// Verify results
			if testCase.wantErrors {
				// For special case with analyzer error
				mockAnalyzer, isAnalyzerMock := testCase.analyzer.(*mockCommitAnalyzer)
				if testCase.analyzer != nil && isAnalyzerMock && mockAnalyzer.err != nil {
					// We expect a git operation error
					hasGitError := false

					for _, err := range errors {
						if err.Code == string(appErrors.ErrGitOperationFailed) {
							hasGitError = true

							break
						}
					}

					require.True(t, hasGitError, "Expected git operation error but got: %v", errors)
				} else {
					// For limit exceeded case
					hasLimitError := false

					for _, err := range errors {
						if err.Code == string(appErrors.ErrTooManyCommits) {
							hasLimitError = true

							break
						}
					}

					require.True(t, hasLimitError, "Expected too many commits error but got: %v", errors)
				}
			} else {
				require.Empty(t, errors, "Expected no errors but got: %v", errors)
			}
		})
	}
}
