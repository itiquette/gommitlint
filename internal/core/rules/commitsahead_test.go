// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package rules_test

import (
	"context"
	"testing"

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
	require.Contains(t, updatedRule.Result(), "Git repository not accessible")
	require.Contains(t, updatedRule.VerboseResult(), "Repository object is nil")
	require.Contains(t, updatedRule.Help(), "Git repository is not accessible")
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
	require.Contains(t, updatedRule.Result(), "Git repository not accessible")
	require.Contains(t, updatedRule.VerboseResult(), "Repository object is nil")
	require.Contains(t, updatedRule.Help(), "Git repository is not accessible")
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

	// Check for errors
	require.Empty(t, errors)
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

	// Check for errors
	require.NotEmpty(t, errors)
	validationErr := errors[0]
	require.Equal(t, string(appErrors.ErrTooManyCommits), validationErr.Code)

	// Use the errors directly for our assertions, as they contain the error information
	require.Contains(t, validationErr.Message, "HEAD is 5 commits ahead")
	require.Contains(t, validationErr.Message, "maximum allowed: 4")

	// Create a rule with errors for testing help text
	ruleWithErrors := rules.NewCommitsAheadRule(
		rules.WithReference("main"),
		rules.WithMaxCommitsAhead(4),
	)
	ruleWithErrors = ruleWithErrors.SetErrors(errors)

	require.Contains(t, ruleWithErrors.Help(), "to reduce the total count")
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
		_, rule = rules.ValidateCommitsAheadWithState(ctx, rule, domain.CommitInfo{})

		// Check the help message
		helpMsg := rule.Help()
		require.Equal(t, "No errors to fix", helpMsg)

		// Now test the error case
		errorRule := rules.NewCommitsAheadRule(
			rules.WithReference("main"),
			rules.WithMaxCommitsAhead(5),
		)

		// Create a custom error
		err := appErrors.New(
			"CommitsAhead",
			appErrors.ErrTooManyCommits,
			"HEAD is 10 commits ahead of main (maximum allowed: 5)",
		).WithContext("commits_ahead", "10")

		errorRule = errorRule.SetErrors([]appErrors.ValidationError{err})

		// Check the help message for error state
		errorHelpMsg := errorRule.Help()
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
				expectedMessage: "Too many commits ahead of main (10)",
				hasErrors:       true,
			},
			{
				name:            "repository error",
				commitsAhead:    0,
				maxCommitsAhead: 5,
				errorCode:       appErrors.ErrInvalidRepo,
				errorMessage:    "Repository object is nil",
				expectedMessage: "Git repository not accessible",
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

				// Set up errors if needed
				if testCase.hasErrors {
					err := appErrors.New(
						"CommitsAhead",
						testCase.errorCode,
						testCase.errorMessage,
					)

					// Add context for commits ahead error
					if testCase.errorCode == appErrors.ErrTooManyCommits {
						err = err.WithContext("commits_ahead", "10")
					}

					rule = rule.SetErrors([]appErrors.ValidationError{err})
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
					_, rule = rules.ValidateCommitsAheadWithState(ctx, rule, domain.CommitInfo{})
				}

				// Check the result message
				result := rule.Result()
				require.Contains(t, result, testCase.expectedMessage, "Result message should match expected")

				// For error cases, the message should clearly indicate a problem
				if testCase.hasErrors {
					require.NotContains(t, result, "HEAD is 0 commit", "Error result should not use 'HEAD is X commit' format")
				}
			})
		}
	})
}
