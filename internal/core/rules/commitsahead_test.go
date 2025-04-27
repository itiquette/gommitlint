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
	require.Contains(t, updatedRule.Result(), "Repository object is nil - Git repository not accessible")
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
	require.Contains(t, updatedRule.Result(), "Repository object is nil - Git repository not accessible")
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
