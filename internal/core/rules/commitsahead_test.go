// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package rules_test

import (
	"context"
	"testing"

	"github.com/itiquette/gommitlint/internal/core/rules"
	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/stretchr/testify/assert"
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
	assert.Equal(t, "nil_repo", errors[0].Code)
	assert.Equal(t, "CommitsAhead", rule.Name())
	assert.Contains(t, rule.Result(), "Too many commits ahead of main")
	assert.Contains(t, rule.VerboseResult(), "Repository object is nil")
	assert.Contains(t, rule.Help(), "valid git repository")
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
	assert.Equal(t, "nil_repo", errors[0].Code)
	assert.Contains(t, rule.Result(), "Too many commits ahead of main")
	assert.Contains(t, rule.VerboseResult(), "Repository object is nil")
	assert.Contains(t, rule.Help(), "valid git repository")
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
	assert.Equal(t, "empty_ref", errors[0].Code)
	assert.Contains(t, rule.Result(), "Too many commits ahead of ")
	assert.Contains(t, rule.VerboseResult(), "Reference branch name is empty")
	assert.Contains(t, rule.Help(), "reference branch name")
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
	assert.Empty(t, errors)
	assert.Equal(t, "CommitsAhead", rule.Name())
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
	assert.Equal(t, "too_many_commits", errors[0].Code)
	assert.Contains(t, rule.Result(), "Too many commits ahead of main")
	assert.Contains(t, rule.VerboseResult(), "10 commit(s) ahead of main (maximum allowed: 5)")
	assert.Contains(t, rule.Help(), "too many commits ahead of main")
}
