// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package output

import (
	"context"
	"encoding/json"
	"time"

	"github.com/itiquette/gommitlint/internal/contextx"
	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/itiquette/gommitlint/internal/errors"
	"github.com/itiquette/gommitlint/internal/infrastructure/log"
)

// JSONFormatter formats validation results as JSON.
// It implements the domain.ResultFormatter interface.
type JSONFormatter struct{}

// NewJSONFormatter creates a new JSON formatter.
func NewJSONFormatter() JSONFormatter {
	return JSONFormatter{}
}

// Ensure JSONFormatter implements domain.ResultFormatter.
var _ domain.ResultFormatter = JSONFormatter{}

// ValidationErrorOutput represents a validation error in JSON format.
type ValidationErrorOutput struct {
	Rule    string            `json:"rule"`
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Context map[string]string `json:"context,omitempty"`
}

// RuleResultOutput represents the result of a single rule validation in JSON format.
type RuleResultOutput struct {
	ID             string                  `json:"id"`
	Name           string                  `json:"name"`
	Status         string                  `json:"status"`
	Message        string                  `json:"message"`
	VerboseMessage string                  `json:"verboseMessage"`
	Help           string                  `json:"help,omitempty"`
	Errors         []ValidationErrorOutput `json:"errors,omitempty"`
}

// CommitResultOutput represents the result of a commit validation in JSON format.
type CommitResultOutput struct {
	Hash         string             `json:"hash"`
	Subject      string             `json:"subject"`
	CommitDate   string             `json:"commitDate,omitempty"`
	Author       string             `json:"author,omitempty"`
	IsMerge      bool               `json:"isMerge,omitempty"`
	Passed       bool               `json:"passed"`
	RuleResults  []RuleResultOutput `json:"ruleResults,omitempty"`
	ErrorCount   int                `json:"errorCount"`
	WarningCount int                `json:"warningCount"`
}

// ValidationResultsOutput represents the overall validation results in JSON format.
type ValidationResultsOutput struct {
	Timestamp     string               `json:"timestamp"`
	AllPassed     bool                 `json:"allPassed"`
	TotalCommits  int                  `json:"totalCommits"`
	PassedCommits int                  `json:"passedCommits"`
	RuleSummary   map[string]int       `json:"ruleSummary,omitempty"`
	CommitResults []CommitResultOutput `json:"commitResults,omitempty"`
}

// Format formats validation results as JSON.
func (JSONFormatter) Format(ctx context.Context, results domain.ValidationResults) string {
	logger := log.Logger(ctx)
	logger.Trace().Int("total_commits", results.TotalCommits).Int("passed_commits", results.PassedCommits).Msg("Entering JSONFormatter.Format")
	// Create the initial report structure with proper initialization
	report := ValidationResultsOutput{
		Timestamp:     time.Now().Format(time.RFC3339),
		AllPassed:     results.AllPassed(),
		TotalCommits:  results.TotalCommits,
		PassedCommits: results.PassedCommits,
		RuleSummary:   copyRuleSummary(results.RuleSummary),
		CommitResults: make([]CommitResultOutput, 0, len(results.CommitResults)),
	}

	// Use a functional approach to transform commit results
	if len(results.CommitResults) > 0 {
		// Transform each commit result without mutating state
		commitOutputs := transformCommitResults(results.CommitResults)
		report.CommitResults = commitOutputs
	}

	// Marshal to JSON
	jsonBytes, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return `{"error": "Failed to marshal report to JSON"}`
	}

	return string(jsonBytes)
}

// copyRuleSummary creates a deep copy of the rule summary map.
// This ensures immutability of the original data.
func copyRuleSummary(summary map[string]int) map[string]int {
	if summary == nil {
		return nil
	}

	// Use contextx.DeepCopyMap for consistency with other functions
	return contextx.DeepCopyMap(summary)
}

// transformCommitResults transforms commit results to output format.
// This pure function handles the transformation without modifying state.
func transformCommitResults(commits []domain.CommitResult) []CommitResultOutput {
	if len(commits) == 0 {
		return []CommitResultOutput{}
	}

	// Filter and map commits in a single pass using FilterMap
	return contextx.FilterMap(commits,
		// Predicate to filter out commits with empty hash
		func(commit domain.CommitResult) bool {
			return commit.CommitInfo.Hash != ""
		},
		// Map function to transform valid commits to CommitResultOutput
		func(commitResult domain.CommitResult) CommitResultOutput {
			// Create a new commit output
			commit := CommitResultOutput{
				Hash:         commitResult.CommitInfo.Hash,
				Subject:      commitResult.CommitInfo.Subject,
				IsMerge:      commitResult.CommitInfo.IsMergeCommit,
				Passed:       commitResult.Passed,
				RuleResults:  transformRuleResults(commitResult.RuleResults),
				ErrorCount:   countErrors(commitResult.RuleResults),
				WarningCount: 0, // Currently not tracking warnings
			}

			// Set commit date with fallback
			if commitResult.CommitInfo.CommitDate != "" {
				commit.CommitDate = commitResult.CommitInfo.CommitDate
			} else {
				commit.CommitDate = time.Now().Format(time.RFC3339)
			}

			// Set author with fallback
			if commitResult.CommitInfo.AuthorName != "" {
				authorInfo := commitResult.CommitInfo.AuthorName
				if commitResult.CommitInfo.AuthorEmail != "" {
					authorInfo += " <" + commitResult.CommitInfo.AuthorEmail + ">"
				}

				commit.Author = authorInfo
			} else {
				commit.Author = "Unknown"
			}

			return commit
		})
}

// transformRuleResults transforms rule results to output format.
// This pure function handles the transformation without modifying state.
func transformRuleResults(rules []domain.RuleResult) []RuleResultOutput {
	if len(rules) == 0 {
		return []RuleResultOutput{}
	}

	// Map rule results to rule outputs
	return contextx.Map(rules, func(ruleResult domain.RuleResult) RuleResultOutput {
		return RuleResultOutput{
			ID:             ruleResult.RuleID,
			Name:           ruleResult.RuleName,
			Status:         string(ruleResult.Status),
			Message:        ruleResult.Message,
			VerboseMessage: ruleResult.VerboseMessage,
			Help:           ruleResult.HelpMessage,
			Errors:         transformValidationErrors(ruleResult.Errors),
		}
	})
}

// transformValidationErrors transforms validation errors to output format.
// This pure function handles the transformation without modifying state.
func transformValidationErrors(validationErrors []errors.ValidationError) []ValidationErrorOutput {
	if len(validationErrors) == 0 {
		return nil
	}

	// Map validation errors to error outputs
	return contextx.Map(validationErrors, func(err errors.ValidationError) ValidationErrorOutput {
		return ValidationErrorOutput{
			Rule:    err.Rule,
			Code:    err.Code,
			Message: err.Message,
			Context: copyContextMap(err.Context),
		}
	})
}

// copyContextMap creates a deep copy of the context map.
// This ensures immutability of the original data.
func copyContextMap(context map[string]string) map[string]string {
	if context == nil {
		return nil
	}

	// Use contextx.DeepCopyMap for consistency with other functions
	return contextx.DeepCopyMap(context)
}

// countErrors counts the number of errors in rule results.
// This pure function aggregates data without modifying state.
func countErrors(rules []domain.RuleResult) int {
	// Use FilterMap to both filter failed rules and extract error counts in one pass
	errorCounts := contextx.FilterMap(rules,
		// Filter predicate: only include failed rules
		func(rule domain.RuleResult) bool {
			return rule.Status == domain.StatusFailed
		},
		// Map function: extract the error count from each rule
		func(rule domain.RuleResult) int {
			return len(rule.Errors)
		})

	// Sum up all error counts
	return contextx.Reduce(errorCounts, 0, func(sum, count int) int {
		return sum + count
	})
}
