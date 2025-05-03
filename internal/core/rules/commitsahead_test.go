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
	errors, updatedRule := rules.ValidateCommitsAheadWithState(ctx, rule, commit)

	// Verify the error
	require.NotEmpty(t, errors)
	validationErr := errors[0]
	require.Equal(t, string(appErrors.ErrInvalidRepo), validationErr.Code)
	require.Contains(t, updatedRule.Result(errors), "Repository access error")
	require.Contains(t, updatedRule.VerboseResult(errors), "Repository access error")
	require.Contains(t, updatedRule.Help(errors), "Repository access error")
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
	require.Contains(t, updatedRule.Result(errors), "Repository access error")
	require.Contains(t, updatedRule.VerboseResult(errors), "Repository access error")
	require.Contains(t, updatedRule.Help(errors), "Repository access error")
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

	// Validate with context using value semantics
	ctx := context.Background()
	errors, updatedRule := rules.ValidateCommitsAheadWithState(ctx, rule, commit)

	// Check that we have the status info error, but no real validation errors
	require.NotEmpty(t, errors)
	// Check that only status info errors are present
	for _, err := range errors {
		require.Equal(t, "status_info", err.Code)
	}

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

	// Validate with context using value semantics
	ctx := context.Background()
	errors, _ := rules.ValidateCommitsAheadWithState(ctx, rule, commit)

	// Check for errors - we need at least one error with the "too_many_commits" code
	require.NotEmpty(t, errors)

	// Find the real error, not the status info one
	var validationErr appErrors.ValidationError

	for _, err := range errors {
		if err.Code == string(appErrors.ErrTooManyCommits) {
			validationErr = err

			break
		}
	}

	// Now check the real error
	require.Equal(t, string(appErrors.ErrTooManyCommits), validationErr.Code)
	require.Contains(t, validationErr.Message, "HEAD is 5 commits ahead")
	require.Contains(t, validationErr.Message, "maximum allowed: 4")

	// Create a rule with errors for testing help text
	ruleWithErrors := rules.NewCommitsAheadRule(
		rules.WithReference("main"),
		rules.WithMaxCommitsAhead(4),
	)
	ruleWithErrors = ruleWithErrors.SetErrors(errors)

	require.Contains(t, ruleWithErrors.Help(errors), "to reduce the total count")
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

		// Validate to update the rule state
		ctx := context.Background()
		validationErrors, rule := rules.ValidateCommitsAheadWithState(ctx, rule, domain.CommitInfo{})

		// Check the help message - since these are all status_info errors, should get "Your branch is in sync" message
		helpMsg := rule.Help(validationErrors)
		require.Contains(t, helpMsg, "Your branch is in sync with main")

		// Special case for empty errors array
		emptyErrors := []appErrors.ValidationError{}
		emptyHelpMsg := rule.Help(emptyErrors)
		require.Equal(t, "No errors to fix", emptyHelpMsg)

		// Now test the error case
		errorRule := rules.NewCommitsAheadRule(
			rules.WithReference("main"),
			rules.WithMaxCommitsAhead(5),
		)

		// Create a custom error
		err := appErrors.CreateBasicError(
			"CommitsAhead",
			appErrors.ErrTooManyCommits,
			"HEAD is 10 commits ahead of main (maximum allowed: 5)",
		).WithContext("commits_ahead", "10")

		// Make a slice with the error
		errorsWithTooMany := []appErrors.ValidationError{err}

		// Set the errors on the rule
		errorRule = errorRule.SetErrors(errorsWithTooMany)

		// Check the help message for error state
		errorHelpMsg := errorRule.Help(errorsWithTooMany)
		require.NotContains(t, errorHelpMsg, "No errors to fix")
		require.Contains(t, errorHelpMsg, "Your branch is too far ahead")
		require.Contains(t, errorHelpMsg, "merge")
		require.Contains(t, errorHelpMsg, "rebase")
	})
}

func TestCommitsAheadResultMessage(t *testing.T) {
	t.Run("result message matches error state", func(t *testing.T) {
		tests := []struct {
			name            string
			commitsAhead    int
			maxCommitsAhead int
			errorMessage    string
			errorCode       appErrors.ValidationErrorCode
			expectedMessage string
			hasErrors       bool
		}{
			{
				name:            "within limit",
				commitsAhead:    3,
				maxCommitsAhead: 5,
				expectedMessage: "HEAD is 3 commit(s) ahead of main",
				hasErrors:       false,
			},
			{
				name:            "exceeds limit",
				commitsAhead:    10,
				maxCommitsAhead: 5,
				errorCode:       appErrors.ErrTooManyCommits,
				errorMessage:    "HEAD is 10 commits ahead of main (maximum allowed: 5)",
				expectedMessage: "HEAD is 10 commits ahead of main",
				hasErrors:       true,
			},
			{
				name:            "repository error",
				commitsAhead:    0,
				maxCommitsAhead: 5,
				errorCode:       appErrors.ErrInvalidRepo,
				errorMessage:    "Repository object is nil",
				expectedMessage: "Repository access error",
				hasErrors:       true,
			},
		}

		for _, testCase := range tests {
			t.Run(testCase.name, func(t *testing.T) {
				// Create a rule
				rule := rules.NewCommitsAheadRule(
					rules.WithReference("main"),
					rules.WithMaxCommitsAhead(testCase.maxCommitsAhead),
				)

				var errors []appErrors.ValidationError
				// Set up errors if needed
				if testCase.hasErrors {
					err := appErrors.CreateBasicError(
						"CommitsAhead",
						testCase.errorCode,
						testCase.errorMessage,
					)

					// Add context for commits ahead error
					if testCase.errorCode == appErrors.ErrTooManyCommits {
						err = err.WithContext("commits_ahead", "10")
					}

					errors = []appErrors.ValidationError{err}
					rule = rule.SetErrors(errors)
				} else {
					// For the success case, we need to simulate a successful validation
					// which would set the ahead count in the rule
					analyzer := &mockCommitAnalyzer{commitsAhead: testCase.commitsAhead}
					rule = rules.NewCommitsAheadRule(
						rules.WithReference("main"),
						rules.WithMaxCommitsAhead(testCase.maxCommitsAhead),
						rules.WithRepositoryGetter(func() domain.CommitAnalyzer {
							return analyzer
						}),
					)

					// Validate to update the rule state
					ctx := context.Background()
					errors, rule = rules.ValidateCommitsAheadWithState(ctx, rule, domain.CommitInfo{})
				}

				// Check the result message
				result := rule.Result(errors)
				require.Contains(t, result, testCase.expectedMessage, "Result message should match expected")

				// For error cases, the message should clearly indicate a problem
				if testCase.hasErrors && testCase.errorCode == appErrors.ErrInvalidRepo {
					require.NotContains(t, result, "HEAD is 0 commit", "Error result should not use 'HEAD is X commit' format")
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
				return config.NewConfig().
					WithCheckCommitsAhead(true).
					WithMaxCommitsAhead(5).
					WithReferenceBranch("main")
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
				return config.NewConfig().
					WithCheckCommitsAhead(true).
					WithMaxCommitsAhead(5).
					WithReferenceBranch("main")
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
				return config.NewConfig().
					WithCheckCommitsAhead(false).
					WithMaxCommitsAhead(5).
					WithReferenceBranch("main")
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
				return config.NewConfig().
					WithCheckCommitsAhead(true).
					WithMaxCommitsAhead(5).
					WithReferenceBranch("main")
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
				return config.NewConfig().
					WithCheckCommitsAhead(true).
					WithMaxCommitsAhead(5).
					WithReferenceBranch("develop")
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
			unifiedConfig := testCase.configSetup()

			// Create rule with unified config
			rule := rules.NewCommitsAheadRuleWithConfig(unifiedConfig, testCase.analyzer)

			// Skip validation on disabled check case
			if !unifiedConfig.CheckCommitsAhead() {
				return
			}

			// Create empty commit info as this rule doesn't validate the commit itself
			commit := domain.CommitInfo{
				Subject: "Test commit",
				Body:    "",
			}

			ctx := context.Background()
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
