// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2
package output

import (
	"context"
	"encoding/json"
	"maps"
	"time"

	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/itiquette/gommitlint/internal/errors"
	"github.com/itiquette/gommitlint/internal/ports/outgoing"
)

// JSONFormatter formats validation results as JSON.
// It implements the domain.ResultFormatter interface.
type JSONFormatter struct{}

// NewJSONFormatter creates a new JSON formatter.
func NewJSONFormatter() JSONFormatter {
	return JSONFormatter{}
}

// Ensure JSONFormatter implements outgoing.ResultFormatter.
var _ outgoing.ResultFormatter = JSONFormatter{}

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
func (JSONFormatter) Format(_ context.Context, results interface{}) string {
	validationResults, ok := results.(domain.ValidationResults)
	if !ok {
		return `{"error": "invalid results type"}`
	}

	// Create the initial report structure with proper initialization
	report := ValidationResultsOutput{
		Timestamp:     time.Now().Format(time.RFC3339),
		AllPassed:     validationResults.AllPassed(),
		TotalCommits:  validationResults.TotalCommits,
		PassedCommits: validationResults.PassedCommits,
		RuleSummary:   maps.Clone(validationResults.RuleSummary),
		CommitResults: make([]CommitResultOutput, 0, len(validationResults.Results)),
	}

	// Use a functional approach to transform commit results
	if len(validationResults.Results) > 0 {
		// Transform each commit result without mutating state
		commitOutputs := transformCommitResults(validationResults.Results)
		report.CommitResults = commitOutputs
	}

	// Marshal to JSON
	jsonBytes, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return `{"error": "Failed to marshal report to JSON"}`
	}

	return string(jsonBytes)
}

// transformCommitResults transforms commit results to output format.
// This pure function handles the transformation without modifying state.
func transformCommitResults(commits []domain.CommitResult) []CommitResultOutput {
	if len(commits) == 0 {
		return []CommitResultOutput{}
	}

	// Filter and transform commits
	results := make([]CommitResultOutput, 0, len(commits))

	for _, commitResult := range commits {
		// Skip commits with empty hash
		if commitResult.CommitInfo.Hash == "" {
			continue
		}

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

		results = append(results, commit)
	}

	return results
}

// transformRuleResults transforms rule results to output format.
// This pure function handles the transformation without modifying state.
func transformRuleResults(rules []domain.RuleResult) []RuleResultOutput {
	if len(rules) == 0 {
		return []RuleResultOutput{}
	}

	// Transform rule results to rule outputs
	results := make([]RuleResultOutput, len(rules))
	for i, ruleResult := range rules {
		results[i] = RuleResultOutput{
			ID:             ruleResult.RuleID,
			Name:           ruleResult.RuleName,
			Status:         string(ruleResult.Status),
			Message:        ruleResult.Message,
			VerboseMessage: ruleResult.VerboseMessage,
			Help:           ruleResult.HelpMessage,
			Errors:         transformValidationErrors(ruleResult.Errors),
		}
	}

	return results
}

// transformValidationErrors transforms validation errors to output format.
// This pure function handles the transformation without modifying state.
func transformValidationErrors(validationErrors []errors.ValidationError) []ValidationErrorOutput {
	if len(validationErrors) == 0 {
		return nil
	}

	// Transform validation errors to error outputs
	results := make([]ValidationErrorOutput, len(validationErrors))
	for i, err := range validationErrors {
		results[i] = ValidationErrorOutput{
			Rule:    err.Rule,
			Code:    err.Code,
			Message: err.Message,
			Context: maps.Clone(err.Context),
		}
	}

	return results
}

// countErrors counts the number of errors in rule results.
// This pure function aggregates data without modifying state.
func countErrors(rules []domain.RuleResult) int {
	total := 0

	for _, rule := range rules {
		if rule.Status == domain.StatusFailed {
			total += len(rule.Errors)
		}
	}

	return total
}
