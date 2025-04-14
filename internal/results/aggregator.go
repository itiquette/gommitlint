// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package results

import (
	"fmt"
	"sort"
	"strings"

	"github.com/itiquette/gommitlint/internal/model"
)

// ValidationSummary represents the aggregated validation results.
type ValidationSummary struct {
	// TotalCommits is the number of commits validated
	TotalCommits int

	// PassedCommits is the number of commits that passed all rules
	PassedCommits int

	// CommitResults holds detailed per-commit validation results
	CommitResults []CommitResult

	// FailedRuleTypes tracks the types of rules that failed across all commits
	FailedRuleTypes map[string]int
}

// CommitResult holds the validation results for a single commit.
type CommitResult struct {
	// CommitInfo contains metadata about the commit
	CommitInfo model.CommitInfo

	// Rules contains validation rules applied to this commit
	Rules []model.CommitRule

	// Passed indicates whether the commit passed all validation rules
	Passed bool
}

// Aggregator collects and processes validation results.
type Aggregator struct {
	summary ValidationSummary
}

// NewAggregator creates a new result aggregator instance.
func NewAggregator() *Aggregator {
	return &Aggregator{
		summary: ValidationSummary{
			TotalCommits:    0,
			PassedCommits:   0,
			CommitResults:   make([]CommitResult, 0),
			FailedRuleTypes: make(map[string]int),
		},
	}
}

// AddCommitResult adds the results for a single commit to the aggregator.
func (a *Aggregator) AddCommitResult(commitInfo model.CommitInfo, rules []model.CommitRule) {
	commitPassed := true

	// Check if all rules passed for this commit
	for _, rule := range rules {
		if len(rule.Errors()) > 0 {
			commitPassed = false
			// Track failed rule types for analytics
			a.summary.FailedRuleTypes[rule.Name()]++
		}
	}

	// Create and add commit result
	result := CommitResult{
		CommitInfo: commitInfo,
		Rules:      rules,
		Passed:     commitPassed,
	}

	a.summary.CommitResults = append(a.summary.CommitResults, result)
	a.summary.TotalCommits++

	if commitPassed {
		a.summary.PassedCommits++
	}
}

// GetSummary returns the complete validation summary.
func (a *Aggregator) GetSummary() ValidationSummary {
	return a.summary
}

// GetCommitResults returns only the commit results.
func (a *Aggregator) GetCommitResults() []CommitResult {
	return a.summary.CommitResults
}

// GetFailedCommits returns only the failing commit results.
func (a *Aggregator) GetFailedCommits() []CommitResult {
	var failedCommits []CommitResult

	for _, result := range a.summary.CommitResults {
		if !result.Passed {
			failedCommits = append(failedCommits, result)
		}
	}

	return failedCommits
}

// GetPassingCommits returns only the passing commit results.
func (a *Aggregator) GetPassingCommits() []CommitResult {
	var passingCommits []CommitResult

	for _, result := range a.summary.CommitResults {
		if result.Passed {
			passingCommits = append(passingCommits, result)
		}
	}

	return passingCommits
}

// GetMostFrequentFailures returns the most commonly failed rule types.
func (a *Aggregator) GetMostFrequentFailures(limit int) []string {
	// Convert map to slice of key-value pairs
	type failureCount struct {
		ruleName string
		count    int
	}

	failureCounts := make([]failureCount, 0, len(a.summary.FailedRuleTypes))
	for ruleName, count := range a.summary.FailedRuleTypes {
		failureCounts = append(failureCounts, failureCount{ruleName, count})
	}

	// Sort by count (descending)
	sort.Slice(failureCounts, func(i, j int) bool {
		return failureCounts[i].count > failureCounts[j].count
	})

	// Build result with limit
	resultSize := limit
	if len(failureCounts) < limit {
		resultSize = len(failureCounts)
	}

	result := make([]string, resultSize)
	for ruleIndex := 0; ruleIndex < resultSize; ruleIndex++ {
		result[ruleIndex] = failureCounts[ruleIndex].ruleName
	}

	return result
}

// AllRulesPassed returns whether all commits passed all rules.
func (a *Aggregator) AllRulesPassed() bool {
	return a.summary.TotalCommits == a.summary.PassedCommits
}

// DidAnyCommitPass returns whether at least one commit passed all rules.
func (a *Aggregator) DidAnyCommitPass() bool {
	return a.summary.PassedCommits > 0
}

// HasAnyResults returns whether the aggregator has collected any results.
func (a *Aggregator) HasAnyResults() bool {
	return a.summary.TotalCommits > 0
}

// GenerateSummaryText produces a text summary of validation results.
func (a *Aggregator) GenerateSummaryText() string {
	var stringBuilder strings.Builder

	stringBuilder.WriteString(fmt.Sprintf("Validated %d commits\n", a.summary.TotalCommits))
	stringBuilder.WriteString(fmt.Sprintf("  Passed: %d commits\n", a.summary.PassedCommits))
	stringBuilder.WriteString(fmt.Sprintf("  Failed: %d commits\n", a.summary.TotalCommits-a.summary.PassedCommits))

	if len(a.summary.FailedRuleTypes) > 0 {
		stringBuilder.WriteString("\nMost common rule failures:\n")

		// Get top failures (max 5)
		topFailures := a.GetMostFrequentFailures(5)
		for ruleIndex, ruleName := range topFailures {
			count := a.summary.FailedRuleTypes[ruleName]
			stringBuilder.WriteString(fmt.Sprintf("  %d. %s: %d failures\n", ruleIndex+1, ruleName, count))
		}
	}

	return stringBuilder.String()
}
