// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package output

import (
	"encoding/json"
	"time"

	"github.com/itiquette/gommitlint/internal/domain"
)

// JSONFormatter formats validation results as JSON.
type JSONFormatter struct{}

// NewJSONFormatter creates a new JSON formatter.
func NewJSONFormatter() *JSONFormatter {
	return &JSONFormatter{}
}

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
	VerboseMessage string                  `json:"verbose_message"`
	Help           string                  `json:"help,omitempty"`
	Errors         []ValidationErrorOutput `json:"errors,omitempty"`
}

// CommitResultOutput represents the result of a commit validation in JSON format.
type CommitResultOutput struct {
	Hash         string             `json:"hash"`
	Subject      string             `json:"subject"`
	CommitDate   string             `json:"commit_date,omitempty"`
	Author       string             `json:"author,omitempty"`
	IsMerge      bool               `json:"is_merge,omitempty"`
	Passed       bool               `json:"passed"`
	RuleResults  []RuleResultOutput `json:"rule_results,omitempty"`
	ErrorCount   int                `json:"error_count"`
	WarningCount int                `json:"warning_count"`
}

// ValidationResultsOutput represents the overall validation results in JSON format.
type ValidationResultsOutput struct {
	Timestamp     string               `json:"timestamp"`
	AllPassed     bool                 `json:"all_passed"`
	TotalCommits  int                  `json:"total_commits"`
	PassedCommits int                  `json:"passed_commits"`
	RuleSummary   map[string]int       `json:"rule_summary,omitempty"`
	CommitResults []CommitResultOutput `json:"commit_results,omitempty"`
}

// Format formats validation results as JSON.
func (f *JSONFormatter) Format(results *domain.ValidationResults) string {
	report := ValidationResultsOutput{
		Timestamp:     time.Now().Format(time.RFC3339),
		AllPassed:     results.AllPassed(),
		TotalCommits:  results.TotalCommits,
		PassedCommits: results.PassedCommits,
		RuleSummary:   results.RuleSummary,
		CommitResults: make([]CommitResultOutput, 0, len(results.CommitResults)),
	}

	// Format each commit result
	if len(results.CommitResults) > 0 {
		for _, commitResult := range results.CommitResults {
			if commitResult.CommitInfo != nil {
				commit := CommitResultOutput{
					Hash:        commitResult.CommitInfo.Hash,
					Subject:     commitResult.CommitInfo.Subject,
					IsMerge:     commitResult.CommitInfo.IsMergeCommit,
					Passed:      commitResult.Passed,
					RuleResults: make([]RuleResultOutput, 0, len(commitResult.RuleResults)),
				}

				// In the new version, CommitInfo doesn't have CommitDate or Author fields
				commit.CommitDate = time.Now().Format(time.RFC3339) // Use current time as placeholder
				commit.Author = "Unknown"                           // Use placeholder

				// Format rule results
				errorCount := 0
				warningCount := 0

				if len(commitResult.RuleResults) > 0 {
					for _, ruleResult := range commitResult.RuleResults {
						// Create rule result
						rule := RuleResultOutput{
							ID:             ruleResult.RuleID,
							Name:           ruleResult.RuleName,
							Status:         string(ruleResult.Status),
							Message:        ruleResult.Message,
							VerboseMessage: ruleResult.VerboseMessage,
							Help:           ruleResult.HelpMessage,
						}

						// Convert errors
						if len(ruleResult.Errors) > 0 {
							rule.Errors = make([]ValidationErrorOutput, 0, len(ruleResult.Errors))

							for _, err := range ruleResult.Errors {
								// ValidationError is already the correct type
								validationErr := err
								rule.Errors = append(rule.Errors, ValidationErrorOutput{
									Rule:    validationErr.Rule,
									Code:    validationErr.Code,
									Message: validationErr.Message,
									Context: validationErr.Context,
								})
							}
						}

						commit.RuleResults = append(commit.RuleResults, rule)
					}
				}

				commit.ErrorCount = errorCount
				commit.WarningCount = warningCount

				report.CommitResults = append(report.CommitResults, commit)
			}
		}
	}

	// Marshal to JSON
	jsonBytes, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return `{"error": "Failed to marshal report to JSON"}`
	}

	return string(jsonBytes)
}
