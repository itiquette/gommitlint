// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package results_test

import (
	"testing"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/itiquette/gommitlint/internal/model"
	"github.com/itiquette/gommitlint/internal/results"
	"github.com/itiquette/gommitlint/internal/rule"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockCommitRule creates a simple rule for testing.
type mockCommitRule struct {
	name    string
	result  string
	verbose string
	errors  []*model.ValidationError
	help    string
}

func (m mockCommitRule) Name() string {
	return m.name
}

func (m mockCommitRule) Result() string {
	return m.result
}

func (m mockCommitRule) VerboseResult() string {
	return m.verbose
}

func (m mockCommitRule) Errors() []*model.ValidationError {
	return m.errors
}

func (m mockCommitRule) Help() string {
	return m.help
}

// createMockRules creates a set of mock rules for testing.
func createMockRules(passing, failing int) []model.CommitRule {
	rules := make([]model.CommitRule, 0, passing+failing)

	// Add passing rules
	for passIndex := 0; passIndex < passing; passIndex++ {
		rules = append(rules, mockCommitRule{
			name:    "PassingRule" + string(rune('A'+passIndex)),
			result:  "Rule passed",
			verbose: "Rule passed with flying colors",
			errors:  []*model.ValidationError{},
			help:    "No errors to fix",
		})
	}

	// Add failing rules
	for failIndex := 0; failIndex < failing; failIndex++ {
		rules = append(rules, mockCommitRule{
			name:    "FailingRule" + string(rune('A'+failIndex)),
			result:  "Rule failed",
			verbose: "Rule failed because of reasons",
			errors: []*model.ValidationError{
				model.NewValidationError("FailingRule"+string(rune('A'+failIndex)), "test_error", "Test error message"),
			},
			help: "Fix this error by doing something",
		})
	}

	return rules
}

// createMockCommit creates a mock commit for testing.
func createMockCommit(hash string, subject string, body string) model.CommitInfo {
	gitHash := plumbing.NewHash(hash)
	gitCommit := &object.Commit{
		Hash:    gitHash,
		Message: subject + "\n\n" + body,
	}

	return model.CommitInfo{
		Subject:   subject,
		Body:      body,
		RawCommit: gitCommit,
	}
}

func TestAggregator(t *testing.T) {
	t.Run("new aggregator starts empty", func(t *testing.T) {
		aggregator := results.NewAggregator()
		assert.Equal(t, 0, aggregator.GetSummary().TotalCommits, "new aggregator should start with 0 commits")
		assert.Equal(t, 0, aggregator.GetSummary().PassedCommits, "new aggregator should start with 0 passed commits")
		assert.Empty(t, aggregator.GetSummary().CommitResults, "new aggregator should start with empty commit results")
		assert.Empty(t, aggregator.GetSummary().FailedRuleTypes, "new aggregator should start with empty failed rule types")
		assert.False(t, aggregator.HasAnyResults(), "new aggregator should report having no results")
	})

	t.Run("add passing commit", func(t *testing.T) {
		aggregator := results.NewAggregator()
		mockCommit := createMockCommit("abcdef1234567890abcdef1234567890abcdef12", "Test commit", "Test body")
		mockRules := createMockRules(3, 0) // 3 passing rules, 0 failing

		aggregator.AddCommitResult(mockCommit, mockRules)

		assert.Equal(t, 1, aggregator.GetSummary().TotalCommits, "should have 1 total commit")
		assert.Equal(t, 1, aggregator.GetSummary().PassedCommits, "should have 1 passed commit")
		assert.Len(t, aggregator.GetSummary().CommitResults, 1, "should have 1 commit result")
		assert.True(t, aggregator.GetSummary().CommitResults[0].Passed, "commit should be marked as passed")
		assert.Empty(t, aggregator.GetSummary().FailedRuleTypes, "should have no failed rule types")
		assert.True(t, aggregator.AllRulesPassed(), "all rules should be marked as passed")
		assert.True(t, aggregator.DidAnyCommitPass(), "at least one commit should pass")
		assert.True(t, aggregator.HasAnyResults(), "aggregator should have results")
	})

	t.Run("add failing commit", func(t *testing.T) {
		aggregator := results.NewAggregator()
		mockCommit := createMockCommit("abcdef1234567890abcdef1234567890abcdef12", "Test commit", "Test body")
		mockRules := createMockRules(1, 2) // 1 passing rule, 2 failing

		aggregator.AddCommitResult(mockCommit, mockRules)

		assert.Equal(t, 1, aggregator.GetSummary().TotalCommits, "should have 1 total commit")
		assert.Equal(t, 0, aggregator.GetSummary().PassedCommits, "should have 0 passed commits")
		assert.Len(t, aggregator.GetSummary().CommitResults, 1, "should have 1 commit result")
		assert.False(t, aggregator.GetSummary().CommitResults[0].Passed, "commit should be marked as failed")
		assert.Len(t, aggregator.GetSummary().FailedRuleTypes, 2, "should have 2 failed rule types")
		assert.False(t, aggregator.AllRulesPassed(), "not all rules should pass")
		assert.False(t, aggregator.DidAnyCommitPass(), "no commits should pass")
		assert.True(t, aggregator.HasAnyResults(), "aggregator should have results")
	})

	t.Run("add multiple commits", func(t *testing.T) {
		aggregator := results.NewAggregator()

		// Add first commit - passing
		commit1 := createMockCommit("1111111111111111111111111111111111111111", "First commit", "Good body")
		rules1 := createMockRules(2, 0) // All passing
		aggregator.AddCommitResult(commit1, rules1)

		// Add second commit - failing
		commit2 := createMockCommit("2222222222222222222222222222222222222222", "Second commit", "Bad body")
		rules2 := createMockRules(1, 1) // One passing, one failing
		aggregator.AddCommitResult(commit2, rules2)

		// Add third commit - failing with same rule as second
		commit3 := createMockCommit("3333333333333333333333333333333333333333", "Third commit", "Another bad body")
		rules3 := createMockRules(0, 1) // All failing
		aggregator.AddCommitResult(commit3, rules3)

		assert.Equal(t, 3, aggregator.GetSummary().TotalCommits, "should have 3 total commits")
		assert.Equal(t, 1, aggregator.GetSummary().PassedCommits, "should have 1 passed commit")

		// Verify failing rules tracking
		assert.Len(t, aggregator.GetSummary().FailedRuleTypes, 1, "should have 1 type of failed rule")
		assert.Equal(t, 2, aggregator.GetSummary().FailedRuleTypes["FailingRuleA"], "FailingRuleA should have failed 2 times")

		// Verify commit filtering functions
		passingCommits := aggregator.GetPassingCommits()
		failingCommits := aggregator.GetFailedCommits()

		assert.Len(t, passingCommits, 1, "should have 1 passing commit")
		assert.Len(t, failingCommits, 2, "should have 2 failing commits")

		assert.False(t, aggregator.AllRulesPassed(), "not all rules should pass")
		assert.True(t, aggregator.DidAnyCommitPass(), "at least one commit should pass")
	})

	t.Run("get most frequent failures", func(t *testing.T) {
		aggregator := results.NewAggregator()

		// Create different failing rules patterns
		// First commit - fails rules A and B
		commit1 := createMockCommit("1111111111111111111111111111111111111111", "First commit", "Bad body")
		rules1 := []model.CommitRule{
			mockCommitRule{
				name: "RuleA",
				errors: []*model.ValidationError{
					model.NewValidationError("RuleA", "error", "Error"),
				},
			},
			mockCommitRule{
				name: "RuleB",
				errors: []*model.ValidationError{
					model.NewValidationError("RuleB", "error", "Error"),
				},
			},
		}
		aggregator.AddCommitResult(commit1, rules1)

		// Second commit - fails rule A only
		commit2 := createMockCommit("2222222222222222222222222222222222222222", "Second commit", "Another bad body")
		rules2 := []model.CommitRule{
			mockCommitRule{
				name: "RuleA",
				errors: []*model.ValidationError{
					model.NewValidationError("RuleA", "error", "Error"),
				},
			},
			mockCommitRule{
				name:   "RuleB",
				errors: []*model.ValidationError{}, // Passing
			},
		}
		aggregator.AddCommitResult(commit2, rules2)

		// Third commit - fails rule C only
		commit3 := createMockCommit("3333333333333333333333333333333333333333", "Third commit", "Yet another bad body")
		rules3 := []model.CommitRule{
			mockCommitRule{
				name: "RuleC",
				errors: []*model.ValidationError{
					model.NewValidationError("RuleC", "error", "Error"),
				},
			},
		}
		aggregator.AddCommitResult(commit3, rules3)

		// Get the most frequent failures
		mostFrequent := aggregator.GetMostFrequentFailures(2)

		// RuleA should be first (2 occurrences), followed by RuleB or RuleC (1 occurrence each)
		require.Len(t, mostFrequent, 2, "should return 2 most frequent failures")
		assert.Equal(t, "RuleA", mostFrequent[0], "RuleA should be the most frequent failure")
		assert.Contains(t, []string{"RuleB", "RuleC"}, mostFrequent[1], "second most frequent should be either RuleB or RuleC")
	})

	t.Run("summary text generation", func(t *testing.T) {
		aggregator := results.NewAggregator()

		// Add a mix of passing and failing commits with different rules
		// First commit - passes
		commit1 := createMockCommit("1111111111111111111111111111111111111111", "First commit", "Good body")
		rules1 := createMockRules(2, 0) // All passing
		aggregator.AddCommitResult(commit1, rules1)

		// Second commit - fails with SubjectLength
		commit2 := createMockCommit("2222222222222222222222222222222222222222", "Second commit", "Bad body")
		rules2 := []model.CommitRule{
			rule.ValidateSubjectLength("Very long subject line that exceeds the maximum allowed length by a lot", 50),
		}
		aggregator.AddCommitResult(commit2, rules2)

		// Generate summary text
		summaryText := aggregator.GenerateSummaryText()

		// Verify the summary contains expected information
		assert.Contains(t, summaryText, "Validated 2 commits", "summary should mention total number of commits")
		assert.Contains(t, summaryText, "Passed: 1", "summary should mention number of passing commits")
		assert.Contains(t, summaryText, "Failed: 1", "summary should mention number of failing commits")
		assert.Contains(t, summaryText, "Most common rule failures", "summary should list common failures")
		assert.Contains(t, summaryText, "SubjectLength", "summary should mention failing rule name")
	})
}

func TestCommitResult(t *testing.T) {
	t.Run("commit result properly captures info", func(t *testing.T) {
		// Create a mock commit
		mockCommit := createMockCommit("abcdef1234567890abcdef1234567890abcdef12", "Test subject", "Test body")

		// Create a mix of passing and failing rules
		mockRules := []model.CommitRule{
			mockCommitRule{
				name:   "PassingRule",
				result: "Rule passed",
				errors: []*model.ValidationError{},
			},
			mockCommitRule{
				name:   "FailingRule",
				result: "Rule failed",
				errors: []*model.ValidationError{
					model.NewValidationError("FailingRule", "test_error", "Test error message"),
				},
			},
		}

		// Create aggregator and add the result
		aggregator := results.NewAggregator()
		aggregator.AddCommitResult(mockCommit, mockRules)

		// Get the commit results and verify
		commitResults := aggregator.GetCommitResults()
		require.Len(t, commitResults, 1, "should have 1 commit result")

		commitResult := commitResults[0]
		assert.Equal(t, "Test subject", commitResult.CommitInfo.Subject, "subject should match")
		assert.Equal(t, "Test body", commitResult.CommitInfo.Body, "body should match")
		assert.Equal(t, "abcdef1234567890abcdef1234567890abcdef12", commitResult.CommitInfo.RawCommit.Hash.String(), "hash should match")
		assert.False(t, commitResult.Passed, "commit should be marked as failed due to failing rule")
		assert.Len(t, commitResult.Rules, 2, "should have 2 rules")
	})
}
