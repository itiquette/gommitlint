// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package output

import (
	"encoding/json"
	"time"

	"github.com/itiquette/gommitlint/internal/contextx"
	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/itiquette/gommitlint/internal/errors"
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
func (JSONFormatter) Format(results domain.ValidationResults) string {
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

	// Filter out commits with empty hash
	validCommits := contextx.Filter(commits, func(commit domain.CommitResult) bool {
		return commit.CommitInfo.Hash != ""
	})

	// Map valid commits to CommitResultOutput objects
	return contextx.Map(validCommits, func(commitResult domain.CommitResult) CommitResultOutput {
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
	// Filter for failed rules and then use Reduce to sum error counts
	failedRules := contextx.Filter(rules, func(rule domain.RuleResult) bool {
		return rule.Status == domain.StatusFailed
	})

	return contextx.Reduce(failedRules, 0, func(total int, rule domain.RuleResult) int {
		return total + len(rule.Errors)
	})
}
