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
	commit := &domain.CommitInfo{}

	// Validate
	errors := rule.Validate(commit)

	// Verify the error
	require.NotEmpty(t, errors)
	validationErr := errors[0]
	require.Equal(t, string(appErrors.ErrInvalidRepo), validationErr.Code)
	require.Equal(t, "CommitsAhead", rule.Name())
	require.Contains(t, rule.Result(), "Repository object is nil - Git repository not accessible")
	require.Contains(t, rule.VerboseResult(), "Repository object is nil")
	require.Contains(t, rule.Help(), "Git repository is not accessible")
}

func TestCommitsAheadRuleNilRepoInsideGetter(t *testing.T) {
	// Create rule with nil repo
	rule := rules.NewCommitsAheadRule(
		rules.WithReference("main"),
		rules.WithRepositoryGetter(func() domain.CommitAnalyzer {
			return nil
		}),
	)

	// Create a dummy commit info
	commit := &domain.CommitInfo{}

	// Validate
	errors := rule.Validate(commit)

	// Verify the error
	require.NotEmpty(t, errors)
	validationErr := errors[0]
	require.Equal(t, string(appErrors.ErrInvalidRepo), validationErr.Code)
	require.Contains(t, rule.Result(), "Repository object is nil - Git repository not accessible")
	require.Contains(t, rule.VerboseResult(), "Repository object is nil")
	require.Contains(t, rule.Help(), "Git repository is not accessible")
}

func TestCommitsAheadRuleEmptyReference(t *testing.T) {
	// Create rule with empty reference
	rule := rules.NewCommitsAheadRule(
		rules.WithReference(""),
		rules.WithRepositoryGetter(func() domain.CommitAnalyzer {
			return &mockCommitAnalyzer{}
		}),
	)

	// Create a dummy commit info
	commit := &domain.CommitInfo{}

	// Validate
	errors := rule.Validate(commit)

	// Verify the error
	require.NotEmpty(t, errors)
	validationErr := errors[0]
	require.Equal(t, string(appErrors.ErrInvalidConfig), validationErr.Code)
	require.Contains(t, rule.Result(), "Too many commits ahead of reference branch")
	require.Contains(t, rule.VerboseResult(), "Reference branch name is empty")
	require.Contains(t, rule.Help(), "reference branch name")
}

func TestCommitsAheadRuleOptions(t *testing.T) {
	// Test setting max commits ahead
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
	commit := &domain.CommitInfo{}

	// Just ensure the rule was created successfully
	errors := rule.Validate(commit)
	require.Empty(t, errors)
	require.Equal(t, "CommitsAhead", rule.Name())
}

func TestCommitsAheadRuleTooManyCommits(t *testing.T) {
	// Test exceeding max commits ahead
	rule := rules.NewCommitsAheadRule(
		rules.WithReference("main"),
		rules.WithMaxCommitsAhead(5),
		rules.WithRepositoryGetter(func() domain.CommitAnalyzer {
			return &mockCommitAnalyzer{
				commitsAhead: 10, // Exceeds limit
			}
		}),
	)

	// Create a dummy commit info
	commit := &domain.CommitInfo{}

	// Validate
	errors := rule.Validate(commit)

	// Should have an error
	require.NotEmpty(t, errors)
	validationErr := errors[0]
	require.Equal(t, string(appErrors.ErrTooManyCommits), validationErr.Code)
	require.Contains(t, rule.Result(), "Too many commits ahead of main (10 > 5)")
	require.Contains(t, rule.VerboseResult(), "HEAD is 10 commit(s) ahead of main (maximum allowed: 5)")
	require.Contains(t, rule.Help(), "Your branch is too far ahead of main")
}
