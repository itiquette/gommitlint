// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package rules

import (
	"context"
	"errors"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/itiquette/gommitlint/internal/domain/config"
)

// mockRepository is a test implementation of the Repository interface.
// Only implements GetCommitsAhead for testing the branchahead rule.
type mockRepository struct {
	commitsAhead     int
	err              error
	refBranchName    string
	currentBranch    string
	currentBranchErr error
	shouldError      bool
	errorMsg         string
}

// GetCommitsAhead returns the number of commits ahead stored in the mock.
func (m *mockRepository) GetCommitsAhead(_ context.Context, refBranch string) (int, error) {
	// Save the reference branch name for verification
	m.refBranchName = refBranch

	return m.commitsAhead, m.err
}

// GetCommitsAheadCount returns the number of commits ahead (alias for GetCommitsAhead).
func (m *mockRepository) GetCommitsAheadCount(_ context.Context, refBranch string) (int, error) {
	// Save the reference branch name for verification
	m.refBranchName = refBranch

	// Check if we should return an error
	if m.shouldError {
		return 0, errors.New(m.errorMsg)
	}

	if m.err != nil {
		return 0, m.err
	}

	return m.commitsAhead, nil
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
func (m *mockRepository) GetCurrentBranch(_ context.Context) (string, error) {
	if m.currentBranchErr != nil {
		return "", m.currentBranchErr
	}

	if m.currentBranch != "" {
		return m.currentBranch, nil
	}

	return "feature/test", nil
}
func (m *mockRepository) GetRepositoryName(_ context.Context) string { return "" }
func (m *mockRepository) IsValid(_ context.Context) (bool, error)    { return true, nil }

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
			expectedErrors:    true, // Now provides enhanced error handling
			expectedErrorCode: string(domain.ErrGitOperationFailed),
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
				require.Equal(t, testCase.expectedErrorCode, failures[0].Code)
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

	// Create config with explicit MaxCommitsAhead to ensure validation runs
	cfg := config.Config{
		Repo: config.RepoConfig{
			MaxCommitsAhead: 5, // Set explicit value so validation runs
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
			wantError:       false, // 0 means disabled, so no validation
			wantErrorCode:   "",
		},
		{
			name:            "git error",
			maxCommitsAhead: 5,
			reference:       "main",
			analyzerErr:     errors.New("git error"),
			wantError:       true, // Now provides enhanced git error handling
			wantErrorCode:   string(domain.ErrGitOperationFailed),
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
				require.Equal(t, testCase.wantErrorCode, failures[0].Code)
			} else {
				require.Empty(t, failures, "Expected no validation failures")
			}

			// Verify correct reference was used (only if validation actually ran)
			if testCase.analyzerErr == nil && testCase.maxCommitsAhead > 0 {
				require.Equal(t, testCase.reference, analyzer.refBranchName)
			}
		})
	}
}

// TestBranchAheadRule_EnhancedErrorHandling tests the enhanced error handling functionality.
func TestBranchAheadRule_EnhancedErrorHandling(t *testing.T) {
	tests := []struct {
		name            string
		repoError       error
		currentBranch   string
		wantErrCode     string
		wantErrContains string
		expectNoError   bool
	}{
		{
			name:          "reference not found error",
			repoError:     errors.New("reference not found"),
			currentBranch: "feature/auth",
			expectNoError: true, // Reference not found is treated as 0 commits ahead (original behavior)
		},
		{
			name:            "repository access error",
			repoError:       errors.New("not a git repository"),
			currentBranch:   "feature/auth",
			wantErrCode:     string(domain.ErrInvalidRepo),
			wantErrContains: "Cannot access repository",
		},
		{
			name:            "general git operation error",
			repoError:       errors.New("network timeout"),
			currentBranch:   "feature/auth",
			wantErrCode:     string(domain.ErrGitOperationFailed),
			wantErrContains: "Failed to check commits ahead",
		},
		{
			name:          "unknown revision error",
			repoError:     errors.New("unknown revision or path"),
			currentBranch: "feature/auth",
			expectNoError: true, // Unknown revision is treated as 0 commits ahead (original behavior)
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			mock := &mockRepository{
				err:           testCase.repoError,
				currentBranch: testCase.currentBranch,
			}

			cfg := config.Config{
				Repo: config.RepoConfig{
					MaxCommitsAhead: 5,
					ReferenceBranch: "main",
				},
			}

			rule := NewBranchAheadRule(cfg)
			commit := domain.Commit{Hash: "test", Subject: "test commit"}
			failures := rule.Validate(commit, mock, cfg)

			if testCase.expectNoError {
				require.Empty(t, failures, "Expected no validation error")
			} else {
				require.NotEmpty(t, failures, "Expected validation error")
				require.Equal(t, testCase.wantErrCode, failures[0].Code)
				require.Contains(t, failures[0].Message, testCase.wantErrContains)
			}
		})
	}
}

// TestBranchAheadRule_ConfigurationValidation tests configuration validation.
func TestBranchAheadRule_ConfigurationValidation(t *testing.T) {
	tests := []struct {
		name            string
		maxCommits      int
		referenceBranch string
		wantErrCode     string
		wantErrMessage  string
	}{
		{
			name:            "negative max commits",
			maxCommits:      -5,
			referenceBranch: "main",
			wantErrCode:     string(domain.ErrInvalidConfig),
			wantErrMessage:  "maxCommitsAhead cannot be negative",
		},
		{
			name:            "empty reference branch",
			maxCommits:      10,
			referenceBranch: "",
			wantErrCode:     string(domain.ErrMissingReference),
			wantErrMessage:  "Reference branch cannot be empty",
		},
		{
			name:            "valid configuration",
			maxCommits:      5,
			referenceBranch: "main",
			wantErrCode:     "",
			wantErrMessage:  "",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Create rule with invalid config by directly setting fields
			rule := BranchAheadRule{
				maxCommitsAhead: testCase.maxCommits,
				reference:       testCase.referenceBranch,
			}

			mock := &mockRepository{
				commitsAhead:  1,
				currentBranch: "feature/test",
			}

			commit := domain.Commit{Hash: "test", Subject: "test commit"}
			failures := rule.Validate(commit, mock, config.Config{})

			if testCase.wantErrCode != "" {
				require.NotEmpty(t, failures, "Expected configuration validation error")
				require.Equal(t, testCase.wantErrCode, failures[0].Code)
				require.Contains(t, failures[0].Message, testCase.wantErrMessage)
			} else {
				// Should pass validation or have no config errors
				if len(failures) > 0 {
					require.NotEqual(t, string(domain.ErrInvalidConfig), failures[0].Code)
					require.NotEqual(t, string(domain.ErrMissingReference), failures[0].Code)
				}
			}
		})
	}
}

// TestBranchAheadRule_EnhancedErrorMessages tests the enhanced error messages with branch context.
func TestBranchAheadRule_EnhancedErrorMessages(t *testing.T) {
	tests := []struct {
		name            string
		commitsAhead    int
		maxCommits      int
		currentBranch   string
		referenceBranch string
		wantMessage     string
		wantHelpText    string
	}{
		{
			name:            "small excess - squashing suggestion",
			commitsAhead:    8,
			maxCommits:      5,
			currentBranch:   "feature/auth",
			referenceBranch: "main",
			wantMessage:     "Current branch is 8 commits ahead of 'main' (maximum allowed: 5)",
			wantHelpText:    "Consider squashing the 3 excess commits",
		},
		{
			name:            "medium excess - rebasing suggestion",
			commitsAhead:    15,
			maxCommits:      5,
			currentBranch:   "feature/ui-overhaul",
			referenceBranch: "develop",
			wantMessage:     "Current branch is 15 commits ahead of 'develop' (maximum allowed: 5)",
			wantHelpText:    "You have 10 excess commits. Consider rebasing onto 'develop'",
		},
		{
			name:            "large excess - break up suggestion",
			commitsAhead:    25,
			maxCommits:      5,
			currentBranch:   "feature/major-refactor",
			referenceBranch: "main",
			wantMessage:     "Current branch is 25 commits ahead of 'main' (maximum allowed: 5)",
			wantHelpText:    "You have 20 excess commits. Consider breaking this into smaller pull requests",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			mock := &mockRepository{
				commitsAhead:  testCase.commitsAhead,
				currentBranch: testCase.currentBranch,
			}

			cfg := config.Config{
				Repo: config.RepoConfig{
					MaxCommitsAhead: testCase.maxCommits,
					ReferenceBranch: testCase.referenceBranch,
				},
			}

			rule := NewBranchAheadRule(cfg)
			commit := domain.Commit{Hash: "test", Subject: "test commit"}
			failures := rule.Validate(commit, mock, cfg)

			require.NotEmpty(t, failures, "Expected validation error")
			require.Equal(t, string(domain.ErrTooManyCommits), failures[0].Code)
			require.Equal(t, testCase.wantMessage, failures[0].Message)
			require.Contains(t, failures[0].Help, testCase.wantHelpText)

			// Verify ultra-minimal context
			require.Equal(t, strconv.Itoa(testCase.commitsAhead), failures[0].Context["actual"])
			require.Equal(t, "max "+strconv.Itoa(testCase.maxCommits), failures[0].Context["expected"])
		})
	}
}

// TestBranchAheadRule_DisabledCheck tests that validation is skipped when maxCommitsAhead is 0.
func TestBranchAheadRule_DisabledCheck(t *testing.T) {
	mock := &mockRepository{
		commitsAhead:  100, // Way over any reasonable limit
		currentBranch: "feature/massive",
	}

	cfg := config.Config{
		Repo: config.RepoConfig{
			MaxCommitsAhead: 0, // Disabled
			ReferenceBranch: "main",
		},
	}

	rule := NewBranchAheadRule(cfg)
	commit := domain.Commit{Hash: "test", Subject: "test commit"}
	failures := rule.Validate(commit, mock, cfg)

	require.Empty(t, failures, "Expected no validation errors when check is disabled")
}

// TestBranchAheadRule_BackwardCompatibility tests that existing functionality is preserved.
func TestBranchAheadRule_BackwardCompatibility(t *testing.T) {
	tests := []struct {
		name         string
		commitsAhead int
		maxCommits   int
		wantError    bool
	}{
		{
			name:         "within limit - no error",
			commitsAhead: 3,
			maxCommits:   5,
			wantError:    false,
		},
		{
			name:         "exceeds limit - error",
			commitsAhead: 8,
			maxCommits:   5,
			wantError:    true,
		},
		{
			name:         "exactly at limit - no error",
			commitsAhead: 5,
			maxCommits:   5,
			wantError:    false,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			mock := &mockRepository{
				commitsAhead:  testCase.commitsAhead,
				currentBranch: "feature/test",
			}

			cfg := config.Config{
				Repo: config.RepoConfig{
					MaxCommitsAhead: testCase.maxCommits,
					ReferenceBranch: "main",
				},
			}

			rule := NewBranchAheadRule(cfg)
			commit := domain.Commit{Hash: "test", Subject: "test commit"}
			failures := rule.Validate(commit, mock, cfg)

			if testCase.wantError {
				require.NotEmpty(t, failures, "Expected validation error")
				require.Equal(t, string(domain.ErrTooManyCommits), failures[0].Code)
			} else {
				require.Empty(t, failures, "Expected no validation errors")
			}
		})
	}
}

// TestEnsureFullReference tests the reference format conversion functionality.
func TestEnsureFullReference(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		scenario string
	}{
		// Local branch references
		{
			name:     "short local branch",
			input:    "main",
			expected: "refs/heads/main",
			scenario: "Most common case - short branch name",
		},
		{
			name:     "develop branch",
			input:    "develop",
			expected: "refs/heads/develop",
			scenario: "Alternative main branch",
		},
		{
			name:     "feature branch",
			input:    "feature-abc",
			expected: "refs/heads/feature-abc",
			scenario: "Feature branch with dash",
		},

		// Remote branch references (most important for CI/CD)
		{
			name:     "origin main",
			input:    "origin/main",
			expected: "refs/remotes/origin/main",
			scenario: "Standard remote branch reference",
		},
		{
			name:     "origin develop",
			input:    "origin/develop",
			expected: "refs/remotes/origin/develop",
			scenario: "Remote develop branch",
		},
		{
			name:     "upstream main",
			input:    "upstream/main",
			expected: "refs/remotes/upstream/main",
			scenario: "Fork workflow - upstream remote",
		},
		{
			name:     "complex remote path",
			input:    "my-remote/feature/complex",
			expected: "refs/remotes/my-remote/feature/complex",
			scenario: "Remote with complex branch name",
		},

		// Already formatted references (should pass through)
		{
			name:     "already full local ref",
			input:    "refs/heads/main",
			expected: "refs/heads/main",
			scenario: "Full reference should pass through unchanged",
		},
		{
			name:     "already full remote ref",
			input:    "refs/remotes/origin/main",
			expected: "refs/remotes/origin/main",
			scenario: "Full remote reference should pass through",
		},
		{
			name:     "tag reference",
			input:    "refs/tags/v1.0.0",
			expected: "refs/tags/v1.0.0",
			scenario: "Tag reference should pass through",
		},

		// Edge cases
		{
			name:     "branch with underscores",
			input:    "feature_branch_name",
			expected: "refs/heads/feature_branch_name",
			scenario: "Branch name with underscores",
		},
		{
			name:     "branch with numbers",
			input:    "release-1.2.3",
			expected: "refs/heads/release-1.2.3",
			scenario: "Branch name with version numbers",
		},
		{
			name:     "remote with numbers",
			input:    "origin/release-1.2.3",
			expected: "refs/remotes/origin/release-1.2.3",
			scenario: "Remote branch with version numbers",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := ensureFullReference(testCase.input)
			require.Equal(t, testCase.expected, result,
				"Scenario: %s\\nInput: %s\\nExpected: %s\\nGot: %s",
				testCase.scenario, testCase.input, testCase.expected, result)
		})
	}
}

// TestBranchAheadRule_ReferenceNotFoundBehavior tests the original behavior of treating
// reference not found as 0 commits ahead (no validation error).
func TestBranchAheadRule_ReferenceNotFoundBehavior(t *testing.T) {
	tests := []struct {
		name        string
		error       string
		expectError bool
		description string
	}{
		{
			name:        "reference not found - no error",
			error:       "reference not found",
			expectError: false,
			description: "Original behavior: reference not found means 0 commits ahead",
		},
		{
			name:        "unknown revision - no error",
			error:       "unknown revision or path not in working tree",
			expectError: false,
			description: "Git error for unknown revision should be treated as 0 commits",
		},
		{
			name:        "bad revision - no error",
			error:       "bad revision 'nonexistent-branch'",
			expectError: false,
			description: "Bad revision should be treated as 0 commits ahead",
		},
		{
			name:        "does not exist - no error",
			error:       "branch does not exist",
			expectError: false,
			description: "Branch existence error should not fail validation",
		},
		{
			name:        "permission denied - should error",
			error:       "permission denied",
			expectError: true,
			description: "Permission errors are real problems, not missing references",
		},
		{
			name:        "not a git repository - should error",
			error:       "not a git repository",
			expectError: true,
			description: "Repository access errors should fail validation",
		},
		{
			name:        "network error - should error",
			error:       "network timeout",
			expectError: true,
			description: "Network errors should fail validation",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			mock := &mockRepository{
				shouldError: true,
				errorMsg:    testCase.error,
			}

			cfg := config.Config{
				Repo: config.RepoConfig{
					MaxCommitsAhead: 10,
					ReferenceBranch: "main",
				},
			}

			rule := NewBranchAheadRule(cfg)
			commit := domain.Commit{Hash: "test", Subject: "test commit"}
			failures := rule.Validate(commit, mock, cfg)

			if testCase.expectError {
				require.NotEmpty(t, failures, "Expected validation error for: %s", testCase.description)
			} else {
				require.Empty(t, failures, "Expected no validation error for: %s", testCase.description)
			}
		})
	}
}
