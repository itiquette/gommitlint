// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package output

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/itiquette/gommitlint/internal/domain"
)

func TestGitHub_ValidReport(t *testing.T) {
	// Create test report with validation results
	commit := domain.Commit{
		Hash:        "abc1234",
		Subject:     "feat: add new feature",
		Message:     "feat: add new feature\n\nThis adds a great new feature",
		Author:      "Test User",
		AuthorEmail: "test@example.com",
	}

	validationErrors := []domain.ValidationError{
		{
			Rule:    "TestRule",
			Code:    "test_error",
			Message: "Test validation error",
			Help:    "Fix this error",
			Context: map[string]string{
				"expected": "something",
				"found":    "something else",
			},
		},
		{
			Rule:    "TestRule",
			Code:    "another_error",
			Message: "Another test error",
			Help:    "Fix this too",
			Context: map[string]string{
				"line": "1",
			},
		},
	}

	ruleResults := []domain.RuleReport{
		{
			Name:    "TestRule",
			Status:  domain.StatusFailed,
			Message: "Rule failed",
			Errors:  validationErrors,
		},
		{
			Name:    "PassingRule",
			Status:  domain.StatusPassed,
			Message: "Rule passed",
			Errors:  nil,
		},
	}

	commitReport := domain.CommitReport{
		Commit:      commit,
		Passed:      false,
		RuleResults: ruleResults,
	}

	report := domain.Report{
		Metadata: domain.ReportMetadata{
			Timestamp: time.Date(2025, 6, 14, 10, 0, 0, 0, time.UTC),
		},
		Summary: domain.ReportSummary{
			AllPassed:     false,
			TotalCommits:  1,
			PassedCommits: 0,
			FailedCommits: 1,
			FailedRules:   map[string]int{"TestRule": 2},
		},
		Commits: []domain.CommitReport{commitReport},
		Repository: domain.RepositoryReport{
			RuleResults: nil,
		},
	}

	// Test GitHub formatting
	result := GitHub(report)

	// Check required GitHub Actions annotations
	require.Contains(t, result, "::group::Summary", "should contain summary group")
	require.Contains(t, result, "::endgroup::", "should close groups")
	require.Contains(t, result, "Validated 1 commits", "should show commit count")
	require.Contains(t, result, "Passed: 0, Failed: 1", "should show pass/fail counts")

	// Check commit group
	require.Contains(t, result, "::group::Commit #1: abc1234", "should contain commit group")
	require.Contains(t, result, "Subject: feat: add new feature", "should show commit subject")

	// Check error annotations
	require.Contains(t, result, "::error file=abc1234,line=1,title=TestRule::Test validation error",
		"should contain error annotation for first error")
	require.Contains(t, result, "::error file=abc1234,line=1,title=TestRule::Another test error",
		"should contain error annotation for second error")

	// Check failure indication
	require.Contains(t, result, "❌ 1 rules failed", "should show failed rule count")

	// Check output flag
	require.Contains(t, result, "::set-output name=passed::false", "should set passed output to false")
}

func TestGitHub_PassingReport(t *testing.T) {
	// Create test report with no failures
	commit := domain.Commit{
		Hash:    "def5678",
		Subject: "docs: update readme",
		Message: "docs: update readme",
	}

	ruleResults := []domain.RuleReport{
		{
			Name:    "TestRule",
			Status:  domain.StatusPassed,
			Message: "Rule passed",
			Errors:  nil,
		},
		{
			Name:    "AnotherRule",
			Status:  domain.StatusPassed,
			Message: "Rule passed",
			Errors:  nil,
		},
	}

	commitReport := domain.CommitReport{
		Commit:      commit,
		Passed:      true,
		RuleResults: ruleResults,
	}

	report := domain.Report{
		Metadata: domain.ReportMetadata{
			Timestamp: time.Date(2025, 6, 14, 10, 0, 0, 0, time.UTC),
		},
		Summary: domain.ReportSummary{
			AllPassed:     true,
			TotalCommits:  1,
			PassedCommits: 1,
			FailedCommits: 0,
			FailedRules:   map[string]int{},
		},
		Commits: []domain.CommitReport{commitReport},
	}

	result := GitHub(report)

	// Check summary shows success
	require.Contains(t, result, "Passed: 1, Failed: 0", "should show success counts")

	// Check commit shows success
	require.Contains(t, result, "✅ All rules passed", "should show success indicator")

	// Should not contain error annotations
	require.NotContains(t, result, "::error", "should not contain error annotations")

	// Check output flag shows success
	require.Contains(t, result, "::set-output name=passed::true", "should set passed output to true")
}

func TestGitHub_EmptyReport(t *testing.T) {
	// Test with empty report
	report := domain.Report{
		Metadata: domain.ReportMetadata{
			Timestamp: time.Date(2025, 6, 14, 10, 0, 0, 0, time.UTC),
		},
		Summary: domain.ReportSummary{
			AllPassed:     true,
			TotalCommits:  0,
			PassedCommits: 0,
			FailedCommits: 0,
			FailedRules:   map[string]int{},
		},
		Commits: nil,
	}

	result := GitHub(report)

	// Check basic structure is present
	require.Contains(t, result, "::group::Summary", "should contain summary even for empty report")
	require.Contains(t, result, "Validated 0 commits", "should show zero commits")
	require.Contains(t, result, "::set-output name=passed::true", "should pass with no commits")
}

func TestGitHub_WithRepositoryResults(t *testing.T) {
	// Test with repository-level validation results
	repoResults := []domain.RuleReport{
		{
			Name:    "BranchRule",
			Status:  domain.StatusFailed,
			Message: "Branch rule failed",
			Errors: []domain.ValidationError{
				{
					Rule:    "BranchRule",
					Code:    "branch_error",
					Message: "Branch validation failed",
					Context: map[string]string{
						"branch": "main",
					},
				},
			},
		},
		{
			Name:    "RepoRule",
			Status:  domain.StatusFailed,
			Message: "Repository rule failed",
			Errors: []domain.ValidationError{
				{
					Rule:    "RepoRule",
					Code:    "repo_error",
					Message: "Repository check failed",
				},
			},
		},
	}

	report := domain.Report{
		Metadata: domain.ReportMetadata{
			Timestamp: time.Date(2025, 6, 14, 10, 0, 0, 0, time.UTC),
		},
		Summary: domain.ReportSummary{
			AllPassed:     false,
			TotalCommits:  0,
			PassedCommits: 0,
			FailedCommits: 0,
			FailedRules:   map[string]int{"BranchRule": 1, "RepoRule": 1},
		},
		Commits: nil,
		Repository: domain.RepositoryReport{
			RuleResults: repoResults,
		},
	}

	result := GitHub(report)

	// Check repository validation group
	require.Contains(t, result, "::group::Repository Validation", "should contain repository group")

	// Check repository error annotations
	require.Contains(t, result, "::error title=BranchRule::Branch validation failed",
		"should contain repository error annotation")
	require.Contains(t, result, "::error title=RepoRule::Repository check failed",
		"should contain second repository error annotation")

	// Should fail overall
	require.Contains(t, result, "::set-output name=passed::false", "should fail with repository errors")
}

func TestGitHub_MultipleCommits(t *testing.T) {
	// Test with multiple commits
	commits := []domain.CommitReport{
		{
			Commit: domain.Commit{
				Hash:    "abc1234",
				Subject: "feat: add feature",
			},
			Passed: true,
			RuleResults: []domain.RuleReport{
				{
					Name:   "TestRule",
					Status: domain.StatusPassed,
					Errors: nil,
				},
			},
		},
		{
			Commit: domain.Commit{
				Hash:    "def5678",
				Subject: "fix: broken test",
			},
			Passed: false,
			RuleResults: []domain.RuleReport{
				{
					Name:   "TestRule",
					Status: domain.StatusFailed,
					Errors: []domain.ValidationError{
						{
							Rule:    "TestRule",
							Message: "Test failed",
						},
					},
				},
			},
		},
	}

	report := domain.Report{
		Summary: domain.ReportSummary{
			TotalCommits:  2,
			PassedCommits: 1,
			FailedCommits: 1,
			AllPassed:     false,
		},
		Commits: commits,
	}

	result := GitHub(report)

	// Check both commits are present
	require.Contains(t, result, "::group::Commit #1: abc1234", "should contain first commit")
	require.Contains(t, result, "::group::Commit #2: def5678", "should contain second commit")

	// Check individual commit results
	require.Contains(t, result, "✅ All rules passed", "should show success for passing commit")
	require.Contains(t, result, "❌ 1 rules failed", "should show failure for failing commit")

	// Check summary
	require.Contains(t, result, "Passed: 1, Failed: 1", "should show mixed results")
}

func TestGitHub_EmptyCommitHash(t *testing.T) {
	// Test that commits with empty hash are filtered out
	commits := []domain.CommitReport{
		{
			Commit: domain.Commit{
				Hash:    "", // Empty hash should be filtered out
				Subject: "test",
			},
			Passed: true,
		},
		{
			Commit: domain.Commit{
				Hash:    "abc1234", // Valid hash should be included
				Subject: "valid commit",
			},
			Passed: true,
		},
	}

	report := domain.Report{
		Summary: domain.ReportSummary{
			TotalCommits:  2,
			PassedCommits: 2,
			AllPassed:     true,
		},
		Commits: commits,
	}

	result := GitHub(report)

	// Should only show the valid commit (numbered as #2 since it's at index 1)
	require.Contains(t, result, "::group::Commit #2: abc1234", "should show valid commit")
	require.NotContains(t, result, "::group::Commit #1:", "should not show first commit group with empty hash")

	// Should not contain commits with empty hash
	commitGroups := 0

	lines := strings.Split(result, "\n")
	for _, line := range lines {
		if strings.Contains(line, "::group::Commit #") {
			commitGroups++
		}
	}

	require.Equal(t, 1, commitGroups, "should only have one commit group")
}

func TestGitHub_GroupStructure(t *testing.T) {
	// Test that all groups are properly opened and closed
	report := domain.Report{
		Summary: domain.ReportSummary{
			TotalCommits:  1,
			PassedCommits: 1,
			AllPassed:     true,
		},
		Commits: []domain.CommitReport{
			{
				Commit: domain.Commit{
					Hash:    "abc1234",
					Subject: "test commit",
				},
				Passed: true,
				RuleResults: []domain.RuleReport{
					{
						Name:   "TestRule",
						Status: domain.StatusPassed,
					},
				},
			},
		},
		Repository: domain.RepositoryReport{
			RuleResults: []domain.RuleReport{
				{
					Name:   "RepoRule",
					Status: domain.StatusPassed,
				},
			},
		},
	}

	result := GitHub(report)

	// Count groups and endgroups
	groupCount := strings.Count(result, "::group::")
	endGroupCount := strings.Count(result, "::endgroup::")

	require.Equal(t, groupCount, endGroupCount, "all groups should be properly closed")
	require.GreaterOrEqual(t, groupCount, 2, "should have at least summary and commit groups")
}

func TestGitHub_IsPureFunction(t *testing.T) {
	// Test that GitHub formatter is a pure function
	report := domain.Report{
		Summary: domain.ReportSummary{
			AllPassed:     true,
			TotalCommits:  0,
			PassedCommits: 0,
		},
	}

	// Call multiple times and ensure consistent results
	result1 := GitHub(report)
	result2 := GitHub(report)
	result3 := GitHub(report)

	require.Equal(t, result1, result2, "function should be deterministic")
	require.Equal(t, result2, result3, "function should be deterministic")
}

func TestGitHub_SpecialCharacters(t *testing.T) {
	// Test handling of special characters in commit messages and errors
	commit := domain.Commit{
		Hash:    "abc1234",
		Subject: "feat: add \"quotes\" and <brackets>",
		Message: "feat: add \"quotes\" and <brackets>\n\nWith special chars: & % $ #",
	}

	validationError := domain.ValidationError{
		Rule:    "TestRule",
		Code:    "special_chars",
		Message: "Error with \"quotes\" and <brackets> & symbols",
	}

	commitReport := domain.CommitReport{
		Commit: commit,
		Passed: false,
		RuleResults: []domain.RuleReport{
			{
				Name:   "TestRule",
				Status: domain.StatusFailed,
				Errors: []domain.ValidationError{validationError},
			},
		},
	}

	report := domain.Report{
		Summary: domain.ReportSummary{
			AllPassed:    false,
			TotalCommits: 1,
		},
		Commits: []domain.CommitReport{commitReport},
	}

	result := GitHub(report)

	// Should handle special characters without breaking format
	require.Contains(t, result, "Subject: feat: add \"quotes\" and <brackets>",
		"should preserve special characters in subject")
	require.Contains(t, result, "::error file=abc1234,line=1,title=TestRule::Error with \"quotes\" and <brackets> & symbols",
		"should preserve special characters in error messages")

	// Should be valid GitHub Actions format
	require.Contains(t, result, "::group::", "should maintain valid GitHub Actions format")
	require.Contains(t, result, "::endgroup::", "should maintain valid GitHub Actions format")
}
