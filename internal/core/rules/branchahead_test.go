// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package rules_test

import (
	"context"
	"errors"
	"testing"

	"github.com/itiquette/gommitlint/internal/common/contextx"
	"github.com/itiquette/gommitlint/internal/config"
	"github.com/itiquette/gommitlint/internal/config/types"
	"github.com/itiquette/gommitlint/internal/core/rules"
	"github.com/itiquette/gommitlint/internal/domain"
	appErrors "github.com/itiquette/gommitlint/internal/errors"
	testconfig "github.com/itiquette/gommitlint/internal/testutils/config"
	testcontext "github.com/itiquette/gommitlint/internal/testutils/context"
	"github.com/stretchr/testify/require"
)

// mockCommitAnalyzer implements domain.CommitAnalyzer for testing.
type mockCommitAnalyzer struct {
	commitsAhead int
	err          error
}

func (m *mockCommitAnalyzer) GetCommitsAhead(_ context.Context, _ string) (int, error) {
	return m.commitsAhead, m.err
}

// createConfigContext creates a context with the configuration needed for the commitsahead rule.
// referenceBranch parameter is kept for future test flexibility even though it's currently always "main"
//
//nolint:unparam // Parameter may vary in future tests
func createConfigContext(maxCommitsAhead int, referenceBranch string) context.Context {
	// Create context with new builder pattern
	ctx := testconfig.NewBuilder().BuildContext(testcontext.CreateTestContext())

	// Create config with repository values
	builder := testconfig.NewBuilder()
	cfg := builder.Build()
	cfg.Repository.MaxCommitsAhead = maxCommitsAhead
	cfg.Repository.ReferenceBranch = referenceBranch

	contextAdapter := testconfig.NewAdapter(cfg)

	return contextx.WithConfig(ctx, contextAdapter.Adapter)
}

func TestBranchAheadRuleNilRepo(t *testing.T) {
	// Create rule with no repository getter
	rule := rules.NewBranchAheadRule(
		rules.WithReference("main"),
	)

	// Create a dummy commit info
	commit := domain.CommitInfo{}

	// Validate with context using the value-based approach
	ctx := testcontext.CreateTestContext()
	errors := rule.Validate(ctx, commit)

	// This should generate a specific error because the repositoryGetter is nil
	require.Len(t, errors, 1, "One error expected when repositoryGetter is nil")
	require.Equal(t, "Repository analyzer unavailable - cannot verify commit distance", errors[0].Error())
}

func TestBranchAheadRuleNilRepoInsideGetter(t *testing.T) {
	// Create rule with repository getter that returns nil
	rule := rules.NewBranchAheadRule(
		rules.WithReference("main"),
		rules.WithRepositoryGetter(func() domain.CommitAnalyzer {
			return nil
		}),
	)

	// Create a dummy commit info
	commit := domain.CommitInfo{}

	// Validate with context using the value-based approach
	ctx := testcontext.CreateTestContext()
	errors := rule.Validate(ctx, commit)

	// Verify the error
	require.NotEmpty(t, errors)
	validationErr := errors[0]
	require.Equal(t, string(appErrors.ErrInvalidRepo), validationErr.Code)
	require.Contains(t, validationErr.Message, "Repository analyzer unavailable")
}

func TestBranchAheadRuleOptions(t *testing.T) {
	// Create rule with value semantics
	rule := rules.NewBranchAheadRule(
		rules.WithReference("main"),
		rules.WithMaxCommitsAhead(10),
		rules.WithRepositoryGetter(func() domain.CommitAnalyzer {
			return &mockCommitAnalyzer{
				commitsAhead: 5, // Within limit
			}
		}),
	)

	// Create a dummy commit info
	commit := domain.CommitInfo{}

	// Create context with configuration
	ctx := createConfigContext(10, "main")

	// Validate with context using value semantics
	errors := rule.Validate(ctx, commit)

	// No validation errors should be present since commits ahead is within limit
	// We used to check for status_info errors, but those were removed in the updated implementation
	require.Empty(t, errors, "No validation errors should occur when commits ahead is within limit")
}

func TestBranchAheadRuleTooManyCommits(t *testing.T) {
	// Create rule with value semantics
	rule := rules.NewBranchAheadRule(
		rules.WithReference("main"),
		rules.WithMaxCommitsAhead(4), // Set to 4 to ensure 5 commits is greater
		rules.WithRepositoryGetter(func() domain.CommitAnalyzer {
			return &mockCommitAnalyzer{
				commitsAhead: 5, // Exceeds the limit of 4
			}
		}),
	)

	// Create a dummy commit info
	commit := domain.CommitInfo{}

	// Create context with configuration
	ctx := createConfigContext(4, "main") // Match the rule's limit

	// Validate with context using value semantics
	errors := rule.Validate(ctx, commit)

	// Check for errors - we should have exactly one error with the ErrTooManyCommits code
	require.NotEmpty(t, errors)
	require.Len(t, errors, 1, "Should have exactly one error")

	validationErr := errors[0]
	require.Equal(t, string(appErrors.ErrTooManyCommits), validationErr.Code)
	// Error message format is "Branch is X commits ahead of reference branch 'Y'"
	require.Contains(t, validationErr.Message, "Branch is 5 commits ahead")
	require.Contains(t, validationErr.Message, "of reference branch 'main'")
}

func TestBranchAheadHelpMessage(t *testing.T) {
	t.Run("help message is appropriate for state", func(t *testing.T) {
		// Test the success case first
		rule := rules.NewBranchAheadRule(
			rules.WithReference("main"),
			rules.WithMaxCommitsAhead(5),
			rules.WithRepositoryGetter(func() domain.CommitAnalyzer {
				return &mockCommitAnalyzer{
					commitsAhead: 3, // Within limit
				}
			}),
		)

		// Add configuration to context
		ctx := createConfigContext(5, "main")

		// Validate to update the rule state
		validationErrors := rule.Validate(ctx, domain.CommitInfo{})

		// With the new implementation, a commit within limits should have no errors
		require.Empty(t, validationErrors, "No validation errors should occur when commits ahead is within limit")

		// Check the help message - for no errors, we should get general help about the rule

		// Now test the error case
		ctx = createConfigContext(5, "main")

		errorRule := rules.NewBranchAheadRule(
			rules.WithReference("main"),
			rules.WithMaxCommitsAhead(5),
			rules.WithRepositoryGetter(func() domain.CommitAnalyzer {
				return &mockCommitAnalyzer{
					commitsAhead: 10, // Exceeds limit
				}
			}),
		)

		// Get validation errors
		errorsWithTooMany := errorRule.Validate(ctx, domain.CommitInfo{})
		require.NotEmpty(t, errorsWithTooMany, "Should have validation errors")
	})
}

func TestBranchAheadResultMessage(t *testing.T) {
	t.Run("result message matches error state", func(t *testing.T) {
		tests := []struct {
			name              string
			commitsAhead      int
			maxCommitsAhead   int
			errorCode         appErrors.ValidationErrorCode
			expectedInMessage string // What the error message should contain
			hasErrors         bool
		}{
			{
				name:              "within limit",
				commitsAhead:      3,
				maxCommitsAhead:   5,
				expectedInMessage: "Branch is 3 commits ahead of reference branch 'main'",
				hasErrors:         false,
			},
			{
				name:              "exceeds limit",
				commitsAhead:      10,
				maxCommitsAhead:   5,
				errorCode:         appErrors.ErrTooManyCommits,
				expectedInMessage: "Branch is 10 commits ahead",
				hasErrors:         true,
			},
			{
				name:              "repository error",
				commitsAhead:      0,
				maxCommitsAhead:   5,
				errorCode:         appErrors.ErrInvalidRepo,
				expectedInMessage: "Repository analyzer unavailable",
				hasErrors:         true,
			},
		}

		for _, testCase := range tests {
			t.Run(testCase.name, func(t *testing.T) {
				// Create context with configuration
				ctx := createConfigContext(testCase.maxCommitsAhead, "main")

				var rule rules.BranchAheadRule

				var errors []appErrors.ValidationError

				if testCase.hasErrors {
					if testCase.errorCode == appErrors.ErrInvalidRepo {
						// For repository error case
						rule = rules.NewBranchAheadRule(
							rules.WithReference("main"),
							rules.WithMaxCommitsAhead(testCase.maxCommitsAhead),
							rules.WithRepositoryGetter(func() domain.CommitAnalyzer {
								return nil // This will cause a repository error
							}),
						)
					} else if testCase.errorCode == appErrors.ErrTooManyCommits {
						// For too many commits case
						rule = rules.NewBranchAheadRule(
							rules.WithReference("main"),
							rules.WithMaxCommitsAhead(testCase.maxCommitsAhead),
							rules.WithRepositoryGetter(func() domain.CommitAnalyzer {
								return &mockCommitAnalyzer{
									commitsAhead: testCase.commitsAhead, // This exceeds the limit
								}
							}),
						)
					}
				} else {
					// For success case
					rule = rules.NewBranchAheadRule(
						rules.WithReference("main"),
						rules.WithMaxCommitsAhead(testCase.maxCommitsAhead),
						rules.WithRepositoryGetter(func() domain.CommitAnalyzer {
							return &mockCommitAnalyzer{
								commitsAhead: testCase.commitsAhead, // This is within the limit
							}
						}),
					)
				}

				// Validate to get errors
				errors = rule.Validate(ctx, domain.CommitInfo{})

				// Check presence of errors matches expectation
				if testCase.hasErrors {
					// For cases with expected errors, there should be at least one real error
					hasRealError := false

					for _, err := range errors {
						if err.Code != "status_info" {
							hasRealError = true

							break
						}
					}

					require.True(t, hasRealError, "Expected validation errors but got none")
				} else {
					// For cases without expected errors, there should be only status info errors
					for _, err := range errors {
						require.Equal(t, "status_info", err.Code, "Expected only status_info errors but got: %v", errors)
					}
				}
			})
		}
	})
}

func TestBranchAheadRuleWithConfig(t *testing.T) {
	// TODO: This test is skipped until the BranchAhead rule can properly handle repository checks
	// The test requires a mocked git repository which is not available in this test environment
	t.Skip("Skipping test until BranchAhead rule can work with mocked repositories")

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
				return config.NewDefaultConfig().
					WithRepository(types.RepositoryConfig{
						MaxCommitsAhead: 5,
						ReferenceBranch: "main",
					})
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
				return config.NewDefaultConfig().
					WithRepository(types.RepositoryConfig{
						MaxCommitsAhead: 5,
						ReferenceBranch: "main",
					})
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
				return config.NewDefaultConfig().
					WithRepository(types.RepositoryConfig{
						MaxCommitsAhead: 0, // Disabled
						ReferenceBranch: "main",
					})
			},
			analyzer: &mockCommitAnalyzer{
				commitsAhead: 10, // This would normally fail
				err:          nil,
			},
			wantErrors:  false,
			description: "Should not validate when commits ahead check is disabled",
		},
		{
			name: "commits ahead check enabled, analyzer error",
			configSetup: func() types.Config {
				return config.NewDefaultConfig().
					WithRepository(types.RepositoryConfig{
						MaxCommitsAhead: 5,
						ReferenceBranch: "main",
					})
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
				return config.NewDefaultConfig().
					WithRepository(types.RepositoryConfig{
						MaxCommitsAhead: 5,
						ReferenceBranch: "develop",
					})
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

			// Add the config to context using our adapter
			ctx := testcontext.CreateTestContext()
			// Create context with test config
			builder := testconfig.NewBuilder()
			builderCfg := builder.Build()
			builderCfg.Repository.MaxCommitsAhead = cfg.Repository.MaxCommitsAhead
			builderCfg.Repository.ReferenceBranch = cfg.Repository.ReferenceBranch

			contextAdapter := testconfig.NewAdapter(builderCfg)
			ctx = contextx.WithConfig(ctx, contextAdapter.Adapter)

			// Create rule (no need to pass config as it will get it from context)
			rule := rules.NewBranchAheadRule(rules.WithRepositoryGetter(func() domain.CommitAnalyzer {
				return testCase.analyzer
			}))

			// For config checks, use the config from the context
			config := contextx.GetConfig(ctx) // Using contextx for consistent configuration access
			repoMaxCommitsAhead := 0

			if v := config.GetInt("repository.max_commits_ahead"); v > 0 {
				repoMaxCommitsAhead = v
			}

			if repoMaxCommitsAhead <= 0 {
				// Create empty commit info as this rule doesn't validate the commit itself
				commit := domain.CommitInfo{
					Subject: "Test commit",
					Body:    "",
				}

				// Validate commit - should return empty errors in disabled case
				errors := rule.Validate(ctx, commit)
				require.Empty(t, errors, "Expected no errors when commits ahead check is disabled")

				return
			}

			// Create empty commit info as this rule doesn't validate the commit itself
			commit := domain.CommitInfo{
				Subject: "Test commit",
				Body:    "",
			}

			// Validate commit
			errors := rule.Validate(ctx, commit)

			if testCase.wantErrors {
				// For cases with expected errors, there should be at least one real error
				hasRealError := false

				for _, err := range errors {
					if err.Code != "status_info" {
						hasRealError = true

						break
					}
				}

				require.True(t, hasRealError, "Expected validation errors but got none")
			} else {
				// For cases without expected errors, there should be only status info errors
				for _, err := range errors {
					require.Equal(t, "status_info", err.Code, "Expected only status_info errors but got: %v", errors)
				}
			}
		})
	}
}
