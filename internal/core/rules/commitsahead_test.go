// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package rules_test

import (
	"context"
	"errors"
	"testing"

	"github.com/itiquette/gommitlint/internal/config"
	"github.com/itiquette/gommitlint/internal/core/rules"
	"github.com/itiquette/gommitlint/internal/domain"
	appErrors "github.com/itiquette/gommitlint/internal/errors"
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

func TestCommitsAheadRuleNilRepo(t *testing.T) {
	// Create rule with no repository getter
	rule := rules.NewCommitsAheadRule(
		rules.WithReference("main"),
	)

	// Create a dummy commit info
	commit := domain.CommitInfo{}

	// Validate with context using the value-based approach
	ctx := context.Background()
	errors, _ := rules.ValidateCommitsAheadWithState(ctx, rule, commit)

	// This should not generate any errors because the repositoryGetter is nil
	require.Empty(t, errors, "No errors expected when repositoryGetter is nil")
}

func TestCommitsAheadRuleNilRepoInsideGetter(t *testing.T) {
	// Create rule with repository getter that returns nil
	rule := rules.NewCommitsAheadRule(
		rules.WithReference("main"),
		rules.WithRepositoryGetter(func() domain.CommitAnalyzer {
			return nil
		}),
	)

	// Create a dummy commit info
	commit := domain.CommitInfo{}

	// Validate with context using the value-based approach
	ctx := context.Background()
	errors, updatedRule := rules.ValidateCommitsAheadWithState(ctx, rule, commit)

	// Verify the error
	require.NotEmpty(t, errors)
	validationErr := errors[0]
	require.Equal(t, string(appErrors.ErrInvalidRepo), validationErr.Code)
	require.Contains(t, validationErr.Message, "Repository object is nil")
	require.Contains(t, updatedRule.Result(errors), "Repository object is nil")
	require.Contains(t, updatedRule.Help(errors), "Repository Access Error")
}

func TestCommitsAheadRuleOptions(t *testing.T) {
	// Create rule with value semantics
	rule := rules.NewCommitsAheadRule(
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

	// Add configuration to context
	ctx := context.Background()
	ctx = config.WithConfig(ctx, config.DefaultConfig().
		WithRepository(config.RepositoryConfig{
			MaxCommitsAhead: 10,
			ReferenceBranch: "main",
		}))

	// Validate with context using value semantics
	errors, updatedRule := rules.ValidateCommitsAheadWithState(ctx, rule, commit)

	// No validation errors should be present since commits ahead is within limit
	// We used to check for status_info errors, but those were removed in the updated implementation
	require.Empty(t, errors, "No validation errors should occur when commits ahead is within limit")

	require.Equal(t, "CommitsAhead", updatedRule.Name())
}

func TestCommitsAheadRuleTooManyCommits(t *testing.T) {
	// Create rule with value semantics
	rule := rules.NewCommitsAheadRule(
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

	// Add configuration to context
	ctx := context.Background()
	ctx = config.WithConfig(ctx, config.DefaultConfig().
		WithRepository(config.RepositoryConfig{
			MaxCommitsAhead: 4, // Match the rule's limit
			ReferenceBranch: "main",
		}))

	// Validate with context using value semantics
	errors, _ := rules.ValidateCommitsAheadWithState(ctx, rule, commit)

	// Check for errors - we should have exactly one error with the ErrTooManyCommits code
	require.NotEmpty(t, errors)
	require.Len(t, errors, 1, "Should have exactly one error")

	validationErr := errors[0]
	require.Equal(t, string(appErrors.ErrTooManyCommits), validationErr.Code)
	// Error message format is "Branch is X commits ahead of reference branch 'Y'"
	require.Contains(t, validationErr.Message, "Branch is 5 commits ahead")
	require.Contains(t, validationErr.Message, "of reference branch 'main'")

	// Create a rule with errors for testing help text
	ruleWithErrors := rules.NewCommitsAheadRule(
		rules.WithReference("main"),
		rules.WithMaxCommitsAhead(4),
	)
	ruleWithErrors = ruleWithErrors.SetErrors(errors)

	// Check that the help message contains useful information
	helpText := ruleWithErrors.Help(errors)
	require.Contains(t, helpText, "Commits Ahead Error")
	require.Contains(t, helpText, "Your branch has too many commits")
	require.Contains(t, helpText, "Merge")
	require.Contains(t, helpText, "Rebase")
}

func TestCommitsAheadHelpMessage(t *testing.T) {
	t.Run("help message is appropriate for state", func(t *testing.T) {
		// Test the success case first
		rule := rules.NewCommitsAheadRule(
			rules.WithReference("main"),
			rules.WithMaxCommitsAhead(5),
			rules.WithRepositoryGetter(func() domain.CommitAnalyzer {
				return &mockCommitAnalyzer{
					commitsAhead: 3, // Within limit
				}
			}),
		)

		// Add configuration to context
		ctx := context.Background()
		ctx = config.WithConfig(ctx, config.DefaultConfig().
			WithRepository(config.RepositoryConfig{
				MaxCommitsAhead: 5,
				ReferenceBranch: "main",
			}))

		// Validate to update the rule state
		validationErrors, rule := rules.ValidateCommitsAheadWithState(ctx, rule, domain.CommitInfo{})

		// With the new implementation, a commit within limits should have no errors
		require.Empty(t, validationErrors, "No validation errors should occur when commits ahead is within limit")

		// Check the help message - for no errors, we should get general help about the rule
		helpMsg := rule.Help(validationErrors)
		require.Contains(t, helpMsg, "This rule compares your current branch with a reference branch")

		// Now test the error case
		ctx = context.Background()
		ctx = config.WithConfig(ctx, config.DefaultConfig().
			WithRepository(config.RepositoryConfig{
				MaxCommitsAhead: 5,
				ReferenceBranch: "main",
			}))

		errorRule := rules.NewCommitsAheadRule(
			rules.WithReference("main"),
			rules.WithMaxCommitsAhead(5),
			rules.WithRepositoryGetter(func() domain.CommitAnalyzer {
				return &mockCommitAnalyzer{
					commitsAhead: 10, // Exceeds limit
				}
			}),
		)

		// Get validation errors
		errorsWithTooMany, _ := rules.ValidateCommitsAheadWithState(ctx, errorRule, domain.CommitInfo{})
		require.NotEmpty(t, errorsWithTooMany, "Should have validation errors")

		// Check the help message for error state
		errorHelpMsg := errorRule.Help(errorsWithTooMany)
		require.Contains(t, errorHelpMsg, "Commits Ahead Error")
		require.Contains(t, errorHelpMsg, "Your branch has too many commits")
		require.Contains(t, errorHelpMsg, "Merge")
		require.Contains(t, errorHelpMsg, "Rebase")
	})
}

func TestCommitsAheadResultMessage(t *testing.T) {
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
				expectedInMessage: "Repository object is nil",
				hasErrors:         true,
			},
		}

		for _, testCase := range tests {
			t.Run(testCase.name, func(t *testing.T) {
				ctx := context.Background()
				// Add configuration to context
				ctx = config.WithConfig(ctx, config.DefaultConfig().
					WithRepository(config.RepositoryConfig{
						MaxCommitsAhead: testCase.maxCommitsAhead,
						ReferenceBranch: "main",
					}))

				var rule rules.CommitsAheadRule

				var errors []appErrors.ValidationError

				if testCase.hasErrors {
					if testCase.errorCode == appErrors.ErrInvalidRepo {
						// For repository error case
						rule = rules.NewCommitsAheadRule(
							rules.WithReference("main"),
							rules.WithMaxCommitsAhead(testCase.maxCommitsAhead),
							rules.WithRepositoryGetter(func() domain.CommitAnalyzer {
								return nil // This will cause a repository error
							}),
						)
					} else if testCase.errorCode == appErrors.ErrTooManyCommits {
						// For too many commits case
						rule = rules.NewCommitsAheadRule(
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
					rule = rules.NewCommitsAheadRule(
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
				errors, rule = rules.ValidateCommitsAheadWithState(ctx, rule, domain.CommitInfo{})

				// Check presence of errors matches expectation
				if testCase.hasErrors {
					require.NotEmpty(t, errors, "Expected errors but got none")
				}

				// Check the result message
				result := rule.Result(errors)
				require.Contains(t, result, testCase.expectedInMessage, "Result message should contain expected text")

				// For the repository error case, we have a specific check
				if testCase.errorCode == appErrors.ErrInvalidRepo {
					require.Contains(t, result, "Repository object is nil")
				}
			})
		}
	})
}

func TestCommitsAheadRuleWithConfig(t *testing.T) {
	tests := []struct {
		name        string
		configSetup func() config.Config
		analyzer    domain.CommitAnalyzer
		wantErrors  bool
		description string
	}{
		{
			name: "commits ahead check enabled, within limit",
			configSetup: func() config.Config {
				return config.DefaultConfig().
					WithRepository(config.RepositoryConfig{
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
			configSetup: func() config.Config {
				return config.DefaultConfig().
					WithRepository(config.RepositoryConfig{
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
			configSetup: func() config.Config {
				return config.DefaultConfig().
					WithRepository(config.RepositoryConfig{
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
			configSetup: func() config.Config {
				return config.DefaultConfig().
					WithRepository(config.RepositoryConfig{
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
			configSetup: func() config.Config {
				return config.DefaultConfig().
					WithRepository(config.RepositoryConfig{
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

			// Add the config to context
			ctx := context.Background()
			ctx = config.WithConfig(ctx, cfg)

			// Create rule (no need to pass config as it will get it from context)
			rule := rules.NewCommitsAheadRule(rules.WithRepositoryGetter(func() domain.CommitAnalyzer {
				return testCase.analyzer
			}))

			// For config checks, use the config from the context
			config := config.GetConfig(ctx)
			if config.Repository.MaxCommitsAhead <= 0 {
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
