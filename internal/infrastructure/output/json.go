// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package output

import (
	"encoding/json"
	"time"

	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/itiquette/gommitlint/internal/errors"
)

// JSONFormatter formats validation results as JSON.
// It implements the domain.ResultFormatter interface.
// This implementation uses value semantics throughout.
type JSONFormatter struct{}

// NewJSONFormatter creates a new JSON formatter.
// This implementation returns a value rather than a pointer.
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
// This implementation follows functional programming patterns.
func (f JSONFormatter) Format(results domain.ValidationResults) string {
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

	result := make(map[string]int, len(summary))
	for k, v := range summary {
		result[k] = v
	}

	return result
}

// transformCommitResults transforms commit results to output format.
// This pure function handles the transformation without modifying state.
func transformCommitResults(commits []domain.CommitResult) []CommitResultOutput {
	if len(commits) == 0 {
		return []CommitResultOutput{}
	}

	result := make([]CommitResultOutput, 0, len(commits))

	for _, commitResult := range commits {
		if commitResult.CommitInfo.Hash == "" {
			continue
		}

		// Create a new commit output for each commit
		commit := CommitResultOutput{
			Hash:         commitResult.CommitInfo.Hash,
			Subject:      commitResult.CommitInfo.Subject,
			IsMerge:      commitResult.CommitInfo.IsMergeCommit,
			Passed:       commitResult.Passed,
			RuleResults:  transformRuleResults(commitResult.RuleResults),
			ErrorCount:   countErrors(commitResult.RuleResults),
			WarningCount: 0, // Currently not tracking warnings
		}

		// In the new version, CommitInfo has CommitDate field
		if commitResult.CommitInfo.CommitDate != "" {
			commit.CommitDate = commitResult.CommitInfo.CommitDate
		} else {
			commit.CommitDate = time.Now().Format(time.RFC3339) // Fallback
		}

		// Set author from available info
		if commitResult.CommitInfo.AuthorName != "" {
			authorInfo := commitResult.CommitInfo.AuthorName
			if commitResult.CommitInfo.AuthorEmail != "" {
				authorInfo += " <" + commitResult.CommitInfo.AuthorEmail + ">"
			}

			commit.Author = authorInfo
		} else {
			commit.Author = "Unknown" // Fallback
		}

		result = append(result, commit)
	}

	return result
}

// transformRuleResults transforms rule results to output format.
// This pure function handles the transformation without modifying state.
func transformRuleResults(rules []domain.RuleResult) []RuleResultOutput {
	if len(rules) == 0 {
		return []RuleResultOutput{}
	}

	result := make([]RuleResultOutput, 0, len(rules))

	for _, ruleResult := range rules {
		// Create a new rule output for each rule
		rule := RuleResultOutput{
			ID:             ruleResult.RuleID,
			Name:           ruleResult.RuleName,
			Status:         string(ruleResult.Status),
			Message:        ruleResult.Message,
			VerboseMessage: ruleResult.VerboseMessage,
			Help:           ruleResult.HelpMessage,
			Errors:         transformValidationErrors(ruleResult.Errors),
		}

		result = append(result, rule)
	}

	return result
}

// transformValidationErrors transforms validation errors to output format.
// This pure function handles the transformation without modifying state.
func transformValidationErrors(errors []errors.ValidationError) []ValidationErrorOutput {
	if len(errors) == 0 {
		return nil
	}

	result := make([]ValidationErrorOutput, 0, len(errors))

	for _, err := range errors {
		// Create a new error output for each error
		errorOutput := ValidationErrorOutput{
			Rule:    err.Rule,
			Code:    err.Code,
			Message: err.Message,
			Context: copyContextMap(err.Context),
		}

		result = append(result, errorOutput)
	}

	return result
}

// copyContextMap creates a deep copy of the context map.
// This ensures immutability of the original data.
func copyContextMap(context map[string]string) map[string]string {
	if context == nil {
		return nil
	}

	result := make(map[string]string, len(context))
	for k, v := range context {
		result[k] = v
	}

	return result
}

// countErrors counts the number of errors in rule results.
// This pure function aggregates data without modifying state.
func countErrors(rules []domain.RuleResult) int {
	count := 0

	for _, rule := range rules {
		if rule.Status == domain.StatusFailed {
			count += len(rule.Errors)
		}
	}

	return count
}
