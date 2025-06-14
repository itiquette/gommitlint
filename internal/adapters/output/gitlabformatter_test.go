// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package output

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/itiquette/gommitlint/internal/domain"
)

func TestGitLab_ValidReport(t *testing.T) {
	// Create test report with validation results
	commit := domain.Commit{
		Hash:        "abc1234567890",
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

	// Test GitLab formatting
	result := GitLab(report)

	// Check required GitLab CI sections
	require.Contains(t, result, "section_start:$(date +%s):summary[collapsed=true]", "should contain summary section start")
	require.Contains(t, result, "section_end:$(date +%s):summary", "should contain summary section end")
	require.Contains(t, result, "Validated 1 commits", "should show commit count")
	require.Contains(t, result, "Passed: 0, Failed: 1", "should show pass/fail counts")

	// Check commit section
	require.Contains(t, result, "section_start:$(date +%s):commit_1[collapsed=true]", "should contain commit section start")
	require.Contains(t, result, "section_end:$(date +%s):commit_1", "should contain commit section end")
	require.Contains(t, result, "Commit #1: abc1234567890", "should show commit hash")
	require.Contains(t, result, "Subject: feat: add new feature", "should show commit subject")

	// Check error messages (with short hash)
	require.Contains(t, result, "ERROR: abc1234 - TestRule: Test validation error",
		"should contain error with short hash for first error")
	require.Contains(t, result, "ERROR: abc1234 - TestRule: Another test error",
		"should contain error with short hash for second error")

	// Check failure indication
	require.Contains(t, result, "❌ 1 rules failed", "should show failed rule count")
}

func TestGitLab_PassingReport(t *testing.T) {
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

	result := GitLab(report)

	// Check summary shows success
	require.Contains(t, result, "Passed: 1, Failed: 0", "should show success counts")

	// Check commit shows success
	require.Contains(t, result, "✅ All rules passed", "should show success indicator")

	// Should not contain error messages
	require.NotContains(t, result, "ERROR:", "should not contain error messages")
}

func TestGitLab_EmptyReport(t *testing.T) {
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

	result := GitLab(report)

	// Check basic structure is present
	require.Contains(t, result, "section_start:$(date +%s):summary[collapsed=true]",
		"should contain summary section even for empty report")
	require.Contains(t, result, "Validated 0 commits", "should show zero commits")
	require.Contains(t, result, "section_end:$(date +%s):summary", "should close summary section")
}

func TestGitLab_WithRepositoryResults(t *testing.T) {
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
			Name:    "PassingRepoRule",
			Status:  domain.StatusPassed,
			Message: "Repository rule passed",
			Errors:  nil,
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
			FailedRules:   map[string]int{"BranchRule": 1},
		},
		Commits: nil,
		Repository: domain.RepositoryReport{
			RuleResults: repoResults,
		},
	}

	result := GitLab(report)

	// Check repository validation section
	require.Contains(t, result, "section_start:$(date +%s):repository[collapsed=true]",
		"should contain repository section start")
	require.Contains(t, result, "section_end:$(date +%s):repository",
		"should contain repository section end")
	require.Contains(t, result, "Repository Validation", "should contain repository section title")

	// Check repository error messages
	require.Contains(t, result, "ERROR: BranchRule - Branch validation failed",
		"should contain repository error")

	// Check passing repository rule
	require.Contains(t, result, "✅ PassingRepoRule: passed",
		"should show passing repository rule")
}

func TestGitLab_MultipleCommits(t *testing.T) {
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
				Hash:    "def567890123456",
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

	result := GitLab(report)

	// Check both commits are present
	require.Contains(t, result, "section_start:$(date +%s):commit_1[collapsed=true]",
		"should contain first commit section")
	require.Contains(t, result, "section_start:$(date +%s):commit_2[collapsed=true]",
		"should contain second commit section")
	require.Contains(t, result, "Commit #1: abc1234", "should contain first commit")
	require.Contains(t, result, "Commit #2: def567890123456", "should contain second commit")

	// Check individual commit results
	require.Contains(t, result, "✅ All rules passed", "should show success for passing commit")
	require.Contains(t, result, "❌ 1 rules failed", "should show failure for failing commit")

	// Check short hash in error (def5678 for second commit)
	require.Contains(t, result, "ERROR: def5678 - TestRule: Test failed",
		"should show error with short hash")

	// Check summary
	require.Contains(t, result, "Passed: 1, Failed: 1", "should show mixed results")
}

func TestGitLab_EmptyCommitHash(t *testing.T) {
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

	result := GitLab(report)

	// Should only show the valid commit (numbered as #2 since it's at index 1)
	require.Contains(t, result, "section_start:$(date +%s):commit_2[collapsed=true]",
		"should show valid commit section")
	require.Contains(t, result, "Commit #2: abc1234", "should show valid commit")
	require.NotContains(t, result, "commit_1[collapsed=true]",
		"should not show first commit section with empty hash")

	// Should not contain commits with empty hash
	commitSections := 0

	lines := strings.Split(result, "\n")
	for _, line := range lines {
		if strings.Contains(line, "section_start:$(date +%s):commit_") {
			commitSections++
		}
	}

	require.Equal(t, 1, commitSections, "should only have one commit section")
}

func TestGitLab_SectionStructure(t *testing.T) {
	// Test that all sections are properly opened and closed
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

	result := GitLab(report)

	// Count sections starts and ends
	sectionStartCount := strings.Count(result, "section_start:")
	sectionEndCount := strings.Count(result, "section_end:")

	require.Equal(t, sectionStartCount, sectionEndCount, "all sections should be properly closed")
	require.GreaterOrEqual(t, sectionStartCount, 2, "should have at least summary and commit sections")
}

func TestGitLab_ShortHashGeneration(t *testing.T) {
	// Test that long hashes are properly shortened to 7 characters
	testCases := []struct {
		name         string
		inputHash    string
		expectedHash string
	}{
		{
			name:         "long hash",
			inputHash:    "abcdef1234567890",
			expectedHash: "abcdef1",
		},
		{
			name:         "exactly 7 characters",
			inputHash:    "abcdef1",
			expectedHash: "abcdef1",
		},
		{
			name:         "short hash",
			inputHash:    "abc123",
			expectedHash: "abc123",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			commit := domain.Commit{
				Hash:    testCase.inputHash,
				Subject: "test commit",
			}

			commitReport := domain.CommitReport{
				Commit: commit,
				Passed: false,
				RuleResults: []domain.RuleReport{
					{
						Name:   "TestRule",
						Status: domain.StatusFailed,
						Errors: []domain.ValidationError{
							{
								Rule:    "TestRule",
								Message: "Test error",
							},
						},
					},
				},
			}

			report := domain.Report{
				Summary: domain.ReportSummary{
					TotalCommits: 1,
					AllPassed:    false,
				},
				Commits: []domain.CommitReport{commitReport},
			}

			result := GitLab(report)
			expectedError := fmt.Sprintf("ERROR: %s - TestRule: Test error", testCase.expectedHash)
			require.Contains(t, result, expectedError,
				"should show error with correct hash length")
		})
	}
}

func TestGitLab_IsPureFunction(t *testing.T) {
	// Test that GitLab formatter is a pure function
	report := domain.Report{
		Summary: domain.ReportSummary{
			AllPassed:     true,
			TotalCommits:  0,
			PassedCommits: 0,
		},
	}

	// Call multiple times and ensure consistent results
	result1 := GitLab(report)
	result2 := GitLab(report)
	result3 := GitLab(report)

	require.Equal(t, result1, result2, "function should be deterministic")
	require.Equal(t, result2, result3, "function should be deterministic")
}

func TestGitLab_SpecialCharacters(t *testing.T) {
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

	result := GitLab(report)

	// Should handle special characters without breaking format
	require.Contains(t, result, "Subject: feat: add \"quotes\" and <brackets>",
		"should preserve special characters in subject")
	require.Contains(t, result, "ERROR: abc1234 - TestRule: Error with \"quotes\" and <brackets> & symbols",
		"should preserve special characters in error messages")

	// Should be valid GitLab CI format
	require.Contains(t, result, "section_start:", "should maintain valid GitLab CI format")
	require.Contains(t, result, "section_end:", "should maintain valid GitLab CI format")
}
