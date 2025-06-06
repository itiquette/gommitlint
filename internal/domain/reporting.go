// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package domain

import (
	"strings"
	"time"
)

// Report represents a validation report.
type Report struct {
	Summary    ReportSummary
	Commits    []CommitReport
	Repository RepositoryReport
	Metadata   ReportMetadata
}

// ReportSummary contains high-level validation statistics.
type ReportSummary struct {
	TotalCommits  int
	PassedCommits int
	FailedCommits int
	AllPassed     bool
	FailedRules   map[string]int // Rule name -> failure count
}

// CommitReport contains formatted information about a single commit validation.
type CommitReport struct {
	Commit      Commit
	RuleResults []RuleReport
	Passed      bool
}

// RuleReport contains formatted rule validation information.
type RuleReport struct {
	Name    string
	Status  ValidationStatus
	Errors  []ValidationError
	Message string // Formatted message for display
}

// RepositoryReport contains repository-level validation results.
type RepositoryReport struct {
	RuleResults []RuleReport
}

// ReportMetadata contains report generation context.
type ReportMetadata struct {
	Timestamp time.Time
	Format    string
	Options   ReportOptions
}

// BuildReport creates a report showing all executed rules (both passed and failed).
func BuildReport(commitResults []ValidationResult, repoErrors []ValidationError,
	commitRules []CommitRule, repoRules []RepositoryRule, options ReportOptions) Report {
	return Report{
		Summary:    buildSummary(commitResults, repoErrors),
		Commits:    buildCommitReports(commitResults, commitRules),
		Repository: buildRepositoryReport(repoErrors, repoRules),
		Metadata:   buildMetadata(options),
	}
}

// buildSummary creates report summary from validation results.
func buildSummary(commitResults []ValidationResult, repoErrors []ValidationError) ReportSummary {
	totalCommits := len(commitResults)
	passedCommits := 0
	failedRules := make(map[string]int)

	// Count passed commits and collect failed rules
	for _, result := range commitResults {
		if !result.HasFailures() {
			passedCommits++
		}

		// Count rule failures
		for _, err := range result.Errors {
			failedRules[err.Rule]++
		}
	}

	// Count repository rule failures
	for _, err := range repoErrors {
		failedRules[err.Rule]++
	}

	failedCommits := totalCommits - passedCommits
	allPassed := failedCommits == 0 && len(repoErrors) == 0

	return ReportSummary{
		TotalCommits:  totalCommits,
		PassedCommits: passedCommits,
		FailedCommits: failedCommits,
		AllPassed:     allPassed,
		FailedRules:   failedRules,
	}
}

// buildCommitReports creates commit reports showing all executed rules.
func buildCommitReports(commitResults []ValidationResult, commitRules []CommitRule) []CommitReport {
	reports := make([]CommitReport, len(commitResults))

	for i, result := range commitResults {
		reports[i] = CommitReport{
			Commit:      result.Commit,
			RuleResults: buildRuleReports(result, commitRules),
			Passed:      !result.HasFailures(),
		}
	}

	return reports
}

// buildRepositoryReport creates repository report showing all executed rules.
func buildRepositoryReport(repoErrors []ValidationError, repoRules []RepositoryRule) RepositoryReport {
	return RepositoryReport{
		RuleResults: buildRepositoryRuleReports(repoErrors, repoRules),
	}
}

// buildRuleReports creates rule reports showing all executed commit rules.
func buildRuleReports(result ValidationResult, commitRules []CommitRule) []RuleReport {
	// Group errors by rule
	errorsByRule := make(map[string][]ValidationError)
	for _, err := range result.Errors {
		errorsByRule[err.Rule] = append(errorsByRule[err.Rule], err)
	}

	// Create reports for all executed rules
	reports := make([]RuleReport, 0, len(commitRules))

	for _, rule := range commitRules {
		ruleName := rule.Name()
		errs, hasFailed := errorsByRule[ruleName]

		if hasFailed {
			// Failed rule
			var messageBuilder strings.Builder

			for i, err := range errs {
				if i > 0 {
					messageBuilder.WriteString("; ")
				}

				messageBuilder.WriteString(err.Message)
			}

			reports = append(reports, RuleReport{
				Name:    ruleName,
				Status:  StatusFailed,
				Errors:  errs,
				Message: messageBuilder.String(),
			})
		} else {
			// Passed rule
			reports = append(reports, RuleReport{
				Name:    ruleName,
				Status:  StatusPassed,
				Errors:  nil,
				Message: "Passed",
			})
		}
	}

	return reports
}

// buildRepositoryRuleReports creates rule reports showing all executed repository rules.
func buildRepositoryRuleReports(repoErrors []ValidationError, repoRules []RepositoryRule) []RuleReport {
	// Group errors by rule
	errorsByRule := make(map[string][]ValidationError)
	for _, err := range repoErrors {
		errorsByRule[err.Rule] = append(errorsByRule[err.Rule], err)
	}

	// Create reports for all executed rules
	reports := make([]RuleReport, 0, len(repoRules))

	for _, rule := range repoRules {
		ruleName := rule.Name()
		errs, hasFailed := errorsByRule[ruleName]

		if hasFailed {
			// Failed rule
			var messageBuilder strings.Builder

			for i, err := range errs {
				if i > 0 {
					messageBuilder.WriteString("; ")
				}

				messageBuilder.WriteString(err.Message)
			}

			reports = append(reports, RuleReport{
				Name:    ruleName,
				Status:  StatusFailed,
				Errors:  errs,
				Message: messageBuilder.String(),
			})
		} else {
			// Passed rule
			reports = append(reports, RuleReport{
				Name:    ruleName,
				Status:  StatusPassed,
				Errors:  nil,
				Message: "Passed",
			})
		}
	}

	return reports
}

// buildMetadata creates report metadata.
func buildMetadata(options ReportOptions) ReportMetadata {
	return ReportMetadata{
		Timestamp: time.Now(),
		Format:    options.Format,
		Options:   options,
	}
}
