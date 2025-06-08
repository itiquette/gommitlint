// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package output

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/itiquette/gommitlint/internal/domain"
)

func TestJSON_ValidReport(t *testing.T) {
	// Create test report with validation results
	commit := domain.Commit{
		Hash:        "abc1234",
		Subject:     "feat: add new feature",
		Message:     "feat: add new feature\n\nThis adds a great new feature",
		Author:      "Test User",
		AuthorEmail: "test@example.com",
		CommitDate:  "2025-06-14T10:00:00Z",
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
			FailedRules:   map[string]int{"TestRule": 1},
		},
		Commits: []domain.CommitReport{commitReport},
		Repository: domain.RepositoryReport{
			RuleResults: nil,
		},
	}

	// Test JSON formatting
	result := JSON(report)

	// Validate it's valid JSON
	var jsonData map[string]interface{}
	err := json.Unmarshal([]byte(result), &jsonData)
	require.NoError(t, err, "should produce valid JSON")

	// Check required fields
	require.Equal(t, "2025-06-14T10:00:00Z", jsonData["timestamp"])
	require.Equal(t, false, jsonData["allPassed"])
	require.InDelta(t, 1, jsonData["totalCommits"], 0.01)
	require.InDelta(t, 0, jsonData["passedCommits"], 0.01)

	// Check rule summary
	ruleSummary, hasRuleSummary := jsonData["ruleSummary"].(map[string]interface{})
	require.True(t, hasRuleSummary)
	require.Len(t, ruleSummary, 1)
	require.InDelta(t, 1, ruleSummary["TestRule"], 0.01)

	// Check commit results
	commitResults, hasCommitResults := jsonData["commitResults"].([]interface{})
	require.True(t, hasCommitResults)
	require.Len(t, commitResults, 1)

	commitData, isCommitMap := commitResults[0].(map[string]interface{})
	require.True(t, isCommitMap)
	require.Equal(t, "abc1234", commitData["hash"])
	require.Equal(t, "feat: add new feature", commitData["subject"])
	require.Equal(t, false, commitData["passed"])
	require.Equal(t, "Test User <test@example.com>", commitData["author"])
	require.Equal(t, "2025-06-14T10:00:00Z", commitData["commitDate"])
	require.InDelta(t, 1, commitData["errorCount"], 0.01)

	// Check rule results in commit
	jsonRuleResults, hasRuleResults := commitData["ruleResults"].([]interface{})
	require.True(t, hasRuleResults)
	require.Len(t, jsonRuleResults, 2)

	// Check failed rule
	failedRule, isRuleMap := jsonRuleResults[0].(map[string]interface{})
	require.True(t, isRuleMap)
	require.Equal(t, "TestRule", failedRule["name"])
	require.Equal(t, "failed", failedRule["status"])
	require.Equal(t, "Rule failed", failedRule["message"])

	// Check validation errors
	errors, hasErrors := failedRule["errors"].([]interface{})
	require.True(t, hasErrors)
	require.Len(t, errors, 1)

	errorData, isErrorMap := errors[0].(map[string]interface{})
	require.True(t, isErrorMap)
	require.Equal(t, "TestRule", errorData["rule"])
	require.Equal(t, "test_error", errorData["code"])
	require.Equal(t, "Test validation error", errorData["message"])
	require.Equal(t, "Fix this error", errorData["help"])

	// Check context
	context, ok := errorData["context"].(map[string]interface{})
	require.True(t, ok)
	require.Equal(t, "something", context["expected"])
	require.Equal(t, "something else", context["found"])
}

func TestJSON_EmptyReport(t *testing.T) {
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
			FailedRules:   nil,
		},
		Commits: nil,
		Repository: domain.RepositoryReport{
			RuleResults: nil,
		},
	}

	result := JSON(report)

	// Validate it's valid JSON
	var jsonData map[string]interface{}
	err := json.Unmarshal([]byte(result), &jsonData)
	require.NoError(t, err, "should produce valid JSON for empty report")

	require.Equal(t, true, jsonData["allPassed"])
	require.InDelta(t, 0, jsonData["totalCommits"], 0.01)
	require.InDelta(t, 0, jsonData["passedCommits"], 0.01)
}

func TestJSON_WithRepositoryResults(t *testing.T) {
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

	result := JSON(report)

	// Validate it's valid JSON
	var jsonData map[string]interface{}
	err := json.Unmarshal([]byte(result), &jsonData)
	require.NoError(t, err, "should produce valid JSON")

	// Check repository results are included
	repoData, exists := jsonData["repositoryResults"]
	require.True(t, exists, "should include repository results")

	repoArray, isRepoArray := repoData.([]interface{})
	require.True(t, isRepoArray)
	require.Len(t, repoArray, 1)

	repoRule, isRepoRuleMap := repoArray[0].(map[string]interface{})
	require.True(t, isRepoRuleMap)
	require.Equal(t, "BranchRule", repoRule["name"])
	require.Equal(t, "failed", repoRule["status"])
}

func TestJSON_CommitWithMissingFields(t *testing.T) {
	// Test with commit that has missing optional fields
	commit := domain.Commit{
		Hash:    "abc1234",
		Subject: "test commit",
		Message: "test commit",
		// Missing Author, AuthorEmail, CommitDate
	}

	commitReport := domain.CommitReport{
		Commit:      commit,
		Passed:      true,
		RuleResults: nil,
	}

	report := domain.Report{
		Metadata: domain.ReportMetadata{
			Timestamp: time.Date(2025, 6, 14, 10, 0, 0, 0, time.UTC),
		},
		Summary: domain.ReportSummary{
			AllPassed:     true,
			TotalCommits:  1,
			PassedCommits: 1,
		},
		Commits: []domain.CommitReport{commitReport},
	}

	result := JSON(report)

	// Validate it's valid JSON
	var jsonData map[string]interface{}
	err := json.Unmarshal([]byte(result), &jsonData)
	require.NoError(t, err, "should produce valid JSON")

	commitResults, hasCommitResults := jsonData["commitResults"].([]interface{})
	commitData, isCommitData := commitResults[0].(map[string]interface{})

	require.True(t, hasCommitResults)
	require.True(t, isCommitData)

	// Check defaults are applied
	require.Equal(t, "Unknown", commitData["author"])
	require.NotEmpty(t, commitData["commitDate"], "should have a default commit date")
}

func TestJSON_EmptyCommitHash(t *testing.T) {
	// Test that commits with empty hash are filtered out
	commits := []domain.CommitReport{
		{
			Commit: domain.Commit{
				Hash:    "", // Empty hash should be filtered out
				Subject: "test",
				Message: "test",
			},
			Passed: true,
		},
		{
			Commit: domain.Commit{
				Hash:    "abc1234", // Valid hash should be included
				Subject: "valid commit",
				Message: "valid commit",
			},
			Passed: true,
		},
	}

	report := domain.Report{
		Metadata: domain.ReportMetadata{
			Timestamp: time.Date(2025, 6, 14, 10, 0, 0, 0, time.UTC),
		},
		Summary: domain.ReportSummary{
			AllPassed:     true,
			TotalCommits:  2,
			PassedCommits: 2,
		},
		Commits: commits,
	}

	result := JSON(report)

	var jsonData map[string]interface{}
	err := json.Unmarshal([]byte(result), &jsonData)
	require.NoError(t, err)

	commitResults, hasCommitResults := jsonData["commitResults"].([]interface{})
	commitData, isCommitData := commitResults[0].(map[string]interface{})

	require.True(t, hasCommitResults)
	require.True(t, isCommitData)
	require.Len(t, commitResults, 1, "should filter out commits with empty hash")
	require.Equal(t, "abc1234", commitData["hash"])
}

func TestJSON_ErrorHandling(t *testing.T) {
	// Test error handling when JSON marshaling might fail
	// Note: In practice, json.Marshal rarely fails for simple data structures,
	// but we test that error handling returns valid JSON
	//
	// This test verifies the error handling code path exists and returns valid JSON
	// Even if we can't easily trigger a marshal error, the code should be robust
	report := domain.Report{
		Metadata: domain.ReportMetadata{
			Timestamp: time.Date(2025, 6, 14, 10, 0, 0, 0, time.UTC),
		},
		Summary: domain.ReportSummary{
			AllPassed:     true,
			TotalCommits:  0,
			PassedCommits: 0,
		},
	}

	result := JSON(report)

	// Should always return valid JSON, even in error cases
	var jsonData map[string]interface{}
	err := json.Unmarshal([]byte(result), &jsonData)
	require.NoError(t, err, "should always return valid JSON")

	// Should contain expected fields for successful case
	require.Contains(t, jsonData, "timestamp", "should contain timestamp")
	require.Contains(t, jsonData, "allPassed", "should contain allPassed")
}

func TestJSON_IsPureFunction(t *testing.T) {
	// Test that JSON formatter is a pure function
	report := domain.Report{
		Metadata: domain.ReportMetadata{
			Timestamp: time.Date(2025, 6, 14, 10, 0, 0, 0, time.UTC),
		},
		Summary: domain.ReportSummary{
			AllPassed:     true,
			TotalCommits:  0,
			PassedCommits: 0,
		},
	}

	// Call multiple times and ensure consistent results
	result1 := JSON(report)
	result2 := JSON(report)
	result3 := JSON(report)

	require.Equal(t, result1, result2, "function should be deterministic")
	require.Equal(t, result2, result3, "function should be deterministic")
}
