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
	VerboseMessage string                  `json:"verboseMessage,omitempty"`
	Help           string                  `json:"help,omitempty"`
	Errors         []ValidationErrorOutput `json:"errors,omitempty"`
}

// CommitResultOutput represents the result of a single commit validation in JSON format.
type CommitResultOutput struct {
	Hash        string             `json:"hash"`
	ShortHash   string             `json:"shortHash,omitempty"`
	Subject     string             `json:"subject"`
	Body        string             `json:"body,omitempty"`
	Passed      bool               `json:"passed"`
	RuleResults []RuleResultOutput `json:"ruleResults,omitempty"`
}

// DetailedReport is a detailed JSON report structure.
type DetailedReport struct {
	Status        string               `json:"status"`
	Total         int                  `json:"total"`
	Passed        int                  `json:"passed"`
	Failed        int                  `json:"failed"`
	CommitResults []CommitResultOutput `json:"commitResults,omitempty"`
	Timestamp     string               `json:"timestamp"`
}

// Format formats validation results as JSON.
func (f *JSONFormatter) Format(results *domain.ValidationResults) string {
	// Build detailed JSON report
	report := DetailedReport{
		Total:     results.TotalCommits,
		Passed:    results.PassedCommits,
		Failed:    results.TotalCommits - results.PassedCommits,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	// Set overall status
	if results.AllPassed() {
		report.Status = "pass"
	} else {
		report.Status = "fail"
	}

	// Convert commit results
	if len(results.CommitResults) > 0 {
		report.CommitResults = make([]CommitResultOutput, 0, len(results.CommitResults))

		for _, commitResult := range results.CommitResults {
			// Create commit result
			commit := CommitResultOutput{
				Hash:    commitResult.CommitInfo.Hash,
				Subject: commitResult.CommitInfo.Subject,
				Body:    commitResult.CommitInfo.Body,
				Passed:  commitResult.Passed,
			}

			// Add short hash if available
			if len(commitResult.CommitInfo.Hash) >= 8 {
				commit.ShortHash = commitResult.CommitInfo.Hash[:8]
			}

			// Convert rule results
			if len(commitResult.RuleResults) > 0 {
				commit.RuleResults = make([]RuleResultOutput, 0, len(commitResult.RuleResults))

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
							rule.Errors = append(rule.Errors, ValidationErrorOutput{
								Rule:    err.Rule,
								Code:    err.Code,
								Message: err.Message,
								Context: err.Context,
							})
						}
					}

					commit.RuleResults = append(commit.RuleResults, rule)
				}
			}

			report.CommitResults = append(report.CommitResults, commit)
		}
	}

	// Marshal to JSON
	jsonBytes, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return `{"error": "Failed to marshal report to JSON"}`
	}

	return string(jsonBytes)
}
