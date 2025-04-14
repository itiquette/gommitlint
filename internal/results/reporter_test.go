// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package results_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/itiquette/gommitlint/internal/results"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReporter(t *testing.T) {
	// Set up test environment
	cleanup := setupTestPrintReport()
	defer cleanup()

	t.Run("text mode basic output", func(t *testing.T) {
		// Create aggregator with a passing commit
		aggregator := results.NewAggregator()
		mockCommit := createMockCommit("1111111111111111111111111111111111111111", "Test commit", "Good body")
		mockRules := createMockRules(2, 0) // 2 passing rules
		aggregator.AddCommitResult(mockCommit, mockRules)

		// Create buffer to capture output
		var buf bytes.Buffer
		testWriter = &buf

		// Create reporter options
		options := results.ReporterOptions{
			Format:   results.FormatText,
			Verbose:  false,
			ShowHelp: false,
			Writer:   &buf,
		}

		// Create reporter and generate report
		reporter := results.NewReporter(aggregator, options)
		err := reporter.GenerateReport()
		require.NoError(t, err, "generating report should not error")

		// Verify output
		output := buf.String()
		assert.Contains(t, output, "PassingRuleA", "output should mention the rule name")
		assert.Contains(t, output, "Rule passed", "output should contain success message")
		assert.Contains(t, output, "SUCCESS", "output should contain SUCCESS indicator")
		assert.NotContains(t, output, "with flying colors", "non-verbose output should not contain verbose details")
	})

	t.Run("text mode verbose output", func(t *testing.T) {
		// Create aggregator with a failing commit
		aggregator := results.NewAggregator()
		mockCommit := createMockCommit("2222222222222222222222222222222222222222", "Test commit", "Bad body")
		mockRules := createMockRules(1, 1) // 1 passing, 1 failing rule
		aggregator.AddCommitResult(mockCommit, mockRules)

		// Create buffer to capture output
		var buf bytes.Buffer
		testWriter = &buf

		// Create reporter options with verbose mode
		options := results.ReporterOptions{
			Format:   results.FormatText,
			Verbose:  true,
			ShowHelp: false,
			Writer:   &buf,
		}

		// Create reporter and generate report
		reporter := results.NewReporter(aggregator, options)
		err := reporter.GenerateReport()
		require.NoError(t, err, "generating report should not error")

		// Verify verbose output
		output := buf.String()
		assert.Contains(t, output, "PassingRuleA", "output should mention the passing rule")
		assert.Contains(t, output, "FailingRuleA", "output should mention the failing rule")
		assert.Contains(t, output, "Rule failed because of reasons", "verbose output should contain detailed explanation")
		assert.Contains(t, output, "with flying colors", "verbose output should contain detailed success explanation")
		assert.Contains(t, output, "TIP:", "verbose output should contain tip for failing rules")
		assert.Contains(t, output, "FAIL:", "output should contain FAIL indicator")
	})

	t.Run("text mode help output", func(t *testing.T) {
		// Create aggregator with a failing commit
		aggregator := results.NewAggregator()
		mockCommit := createMockCommit("3333333333333333333333333333333333333333", "Test commit", "Bad body")
		mockRules := createMockRules(0, 1) // 0 passing, 1 failing rule
		aggregator.AddCommitResult(mockCommit, mockRules)

		// Create buffer to capture output
		var buf bytes.Buffer
		testWriter = &buf

		// Create reporter options with help mode
		options := results.ReporterOptions{
			Format:   results.FormatText,
			Verbose:  false,
			ShowHelp: true,
			Writer:   &buf,
		}

		// Create reporter and generate report
		reporter := results.NewReporter(aggregator, options)
		err := reporter.GenerateReport()
		require.NoError(t, err, "generating report should not error")

		// Verify help output
		output := buf.String()
		assert.Contains(t, output, "Fix this error by doing something", "help output should contain help message")
	})

	t.Run("text mode specific rule help", func(t *testing.T) {
		// Create aggregator with multiple rules
		aggregator := results.NewAggregator()
		mockCommit := createMockCommit("4444444444444444444444444444444444444444", "Test commit", "Bad body")
		mockRules := createMockRules(1, 2) // 1 passing, 2 failing rules
		aggregator.AddCommitResult(mockCommit, mockRules)

		// Create buffer to capture output
		var buf bytes.Buffer
		testWriter = &buf

		// Create reporter options with specific rule help
		options := results.ReporterOptions{
			Format:         results.FormatText,
			Verbose:        false,
			ShowHelp:       true,
			RuleToShowHelp: "FailingRuleB",
			Writer:         &buf,
		}

		// Create reporter and generate report
		reporter := results.NewReporter(aggregator, options)
		err := reporter.GenerateReport()
		require.NoError(t, err, "generating report should not error")

		// Verify specific rule help output
		output := buf.String()
		assert.Contains(t, output, "FailingRuleB", "output should focus on specific rule")
		assert.Contains(t, output, "Fix this error by doing something", "output should contain help for the rule")
	})

	t.Run("text mode summary for multiple commits", func(t *testing.T) {
		// Create aggregator with multiple commits
		aggregator := results.NewAggregator()

		// Add first commit - passing
		commit1 := createMockCommit("1111111111111111111111111111111111111111", "First commit", "Good body")
		rules1 := createMockRules(2, 0) // All passing
		aggregator.AddCommitResult(commit1, rules1)

		// Add second commit - failing
		commit2 := createMockCommit("2222222222222222222222222222222222222222", "Second commit", "Bad body")
		rules2 := createMockRules(1, 1) // One passing, one failing
		aggregator.AddCommitResult(commit2, rules2)

		// Create buffer to capture output
		var buf bytes.Buffer
		testWriter = &buf

		// Create reporter options
		options := results.ReporterOptions{
			Format:  results.FormatText,
			Verbose: false,
			Writer:  &buf,
		}

		// Create reporter and generate report
		reporter := results.NewReporter(aggregator, options)
		err := reporter.GenerateReport()
		require.NoError(t, err, "generating report should not error")

		// Verify multi-commit summary
		output := buf.String()
		assert.Contains(t, output, "OVERALL SUMMARY", "output should contain summary section")
		assert.Contains(t, output, "Validated 2 commits", "summary should mention total commits")
		assert.Contains(t, output, "Passed: 1", "summary should mention passed commits")
		assert.Contains(t, output, "Failed: 1", "summary should mention failed commits")
	})

	t.Run("json mode output", func(t *testing.T) {
		// Create aggregator with mixed results
		aggregator := results.NewAggregator()

		// Add a passing and a failing commit
		commit1 := createMockCommit("1111111111111111111111111111111111111111", "First commit", "Good body")
		rules1 := createMockRules(2, 0) // All passing
		aggregator.AddCommitResult(commit1, rules1)

		commit2 := createMockCommit("2222222222222222222222222222222222222222", "Second commit", "Bad body")
		rules2 := createMockRules(0, 1) // All failing
		aggregator.AddCommitResult(commit2, rules2)

		// Create buffer to capture output
		var buf bytes.Buffer
		testWriter = &buf

		// Create reporter options with JSON format
		options := results.ReporterOptions{
			Format: results.FormatJSON,
			Writer: &buf,
		}

		// Create reporter and generate report
		reporter := results.NewReporter(aggregator, options)
		err := reporter.GenerateReport()
		require.NoError(t, err, "generating report should not error")

		// Verify JSON output
		output := buf.String()
		assert.True(t, json.Valid([]byte(output)), "output should be valid JSON")

		// Unmarshal and verify structure
		var result map[string]interface{}
		err = json.Unmarshal([]byte(output), &result)
		require.NoError(t, err, "should be able to unmarshal JSON output")

		// Get and verify totalCommits
		totalCommits, isOk := result["totalCommits"].(float64)
		require.True(t, isOk, "totalCommits should be a float64")
		assert.InDelta(t, 2.0, totalCommits, 0.0001, "JSON should have correct totalCommits")

		// Get and verify passedCommits
		passedCommits, isOk := result["passedCommits"].(float64)
		require.True(t, isOk, "passedCommits should be a float64")
		assert.InDelta(t, 1.0, passedCommits, 0.0001, "JSON should have correct passedCommits")

		// Get and verify failedCommits
		failedCommits, isOk := result["failedCommits"].(float64)
		require.True(t, isOk, "failedCommits should be a float64")
		assert.InDelta(t, 1.0, failedCommits, 0.0001, "JSON should have correct failedCommits")

		// Get and verify passRate
		passRate, isOk := result["passRate"].(float64)
		require.True(t, isOk, "passRate should be a float64")
		assert.InDelta(t, 0.5, passRate, 0.0001, "JSON should have correct passRate")

		assert.Equal(t, "failure", result["status"], "JSON should have correct status")

		commits, isCommitsArray := result["commits"].([]interface{})
		require.True(t, isCommitsArray, "JSON should have commits array")
		require.Len(t, commits, 2, "JSON should have 2 commits")

		// Check first commit
		commit, isCommitMap := commits[0].(map[string]interface{})
		require.True(t, isCommitMap, "commit should be an object")
		assert.Equal(t, "1111111111111111111111111111111111111111", commit["sha"], "commit should have correct SHA")
		assert.Equal(t, "First commit", commit["subject"], "commit should have correct subject")
		assert.Equal(t, true, commit["passed"], "first commit should have passed=true")
	})

	t.Run("empty results", func(t *testing.T) {
		// Create empty aggregator
		aggregator := results.NewAggregator()

		// Create buffer to capture output
		var buf bytes.Buffer
		testWriter = &buf

		// Create reporter options
		options := results.ReporterOptions{
			Format: results.FormatText,
			Writer: &buf,
		}

		// Create reporter and generate report
		reporter := results.NewReporter(aggregator, options)
		err := reporter.GenerateReport()
		require.NoError(t, err, "generating report should not error")

		// Verify empty output
		output := buf.String()
		assert.Contains(t, output, "No commits were validated", "output should indicate no commits validated")
	})

	t.Run("invalid format", func(t *testing.T) {
		// Create aggregator with a commit
		aggregator := results.NewAggregator()
		mockCommit := createMockCommit("1111111111111111111111111111111111111111", "Test commit", "Good body")
		mockRules := createMockRules(1, 0)
		aggregator.AddCommitResult(mockCommit, mockRules)

		// Create buffer to capture output
		var buf bytes.Buffer
		testWriter = &buf

		// Create reporter options with invalid format
		options := results.ReporterOptions{
			Format: "invalid",
			Writer: &buf,
		}

		// Create reporter and generate report
		reporter := results.NewReporter(aggregator, options)
		err := reporter.GenerateReport()

		// Verify error for invalid format
		require.Error(t, err, "should error with invalid format")
		assert.Contains(t, err.Error(), "unsupported output format",
			"error should mention unsupported format")
	})

	t.Run("rule help not found", func(t *testing.T) {
		// Create aggregator with a commit
		aggregator := results.NewAggregator()
		mockCommit := createMockCommit("1111111111111111111111111111111111111111", "Test commit", "Good body")
		mockRules := createMockRules(1, 0)
		aggregator.AddCommitResult(mockCommit, mockRules)

		// Create buffer to capture output
		var buf bytes.Buffer
		testWriter = &buf

		// Create reporter options with nonexistent rule help
		options := results.ReporterOptions{
			Format:         results.FormatText,
			ShowHelp:       true,
			RuleToShowHelp: "NonexistentRule",
			Writer:         &buf,
		}

		// Create reporter and generate report
		reporter := results.NewReporter(aggregator, options)
		err := reporter.GenerateReport()
		require.NoError(t, err, "generating report should not error")

		// Verify output for nonexistent rule
		output := buf.String()
		assert.Contains(t, output, "No help found for rule: NonexistentRule",
			"output should indicate rule not found")
		assert.Contains(t, output, "Available rules:",
			"output should list available rules")
	})
}
