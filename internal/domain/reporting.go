// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package domain

import (
	"time"
)

// Report represents a validation report as a pure value type.
// This moves report structure from adapters to domain following hexagonal architecture.
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

// BuildReport creates a report from validation results using pure functions.
// This is the main domain function for report creation.
func BuildReport(commitResults []ValidationResult, repoErrors []ValidationError, options ReportOptions) Report {
	return Report{
		Summary:    buildSummary(commitResults, repoErrors),
		Commits:    buildCommitReports(commitResults),
		Repository: buildRepositoryReport(repoErrors),
		Metadata:   buildMetadata(options),
	}
}

// BuildReportWithRules creates a report including all executed rules (passed and failed).
// This version shows complete rule execution results.
func BuildReportWithRules(commitResults []ValidationResult, repoErrors []ValidationError, allRules []Rule, options ReportOptions) Report {
	return Report{
		Summary:    buildSummary(commitResults, repoErrors),
		Commits:    buildCommitReportsWithRules(commitResults, allRules),
		Repository: buildRepositoryReport(repoErrors),
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

// buildCommitReports creates commit reports from validation results.
func buildCommitReports(commitResults []ValidationResult) []CommitReport {
	reports := make([]CommitReport, len(commitResults))

	for i, result := range commitResults {
		reports[i] = CommitReport{
			Commit:      result.Commit,
			RuleResults: buildRuleReportsFromValidationResult(result),
			Passed:      !result.HasFailures(),
		}
	}

	return reports
}

// buildCommitReportsWithRules creates commit reports showing all rules (passed and failed).
func buildCommitReportsWithRules(commitResults []ValidationResult, allRules []Rule) []CommitReport {
	reports := make([]CommitReport, len(commitResults))

	for i, result := range commitResults {
		reports[i] = CommitReport{
			Commit:      result.Commit,
			RuleResults: buildRuleReportsWithAllRules(result, allRules),
			Passed:      !result.HasFailures(),
		}
	}

	return reports
}

// buildRepositoryReport creates repository report from repository errors.
func buildRepositoryReport(repoErrors []ValidationError) RepositoryReport {
	return RepositoryReport{
		RuleResults: buildRuleReportsFromErrors(repoErrors),
	}
}

// buildRuleReportsFromErrors creates rule reports from validation errors.
func buildRuleReportsFromErrors(errors []ValidationError) []RuleReport {
	// Group errors by rule
	errorsByRule := make(map[string][]ValidationError)
	for _, err := range errors {
		errorsByRule[err.Rule] = append(errorsByRule[err.Rule], err)
	}

	// Create rule reports
	reports := make([]RuleReport, 0, len(errorsByRule))

	for rule, errs := range errorsByRule {
		status := StatusFailed

		message := "Failed"
		if len(errs) > 0 {
			message = errs[0].Message // Use first error message
		}

		reports = append(reports, RuleReport{
			Name:    rule,
			Status:  status,
			Errors:  errs,
			Message: message,
		})
	}

	return reports
}

// buildRuleReportsWithAllRules creates rule reports for all executed rules (passed and failed).
func buildRuleReportsWithAllRules(result ValidationResult, allRules []Rule) []RuleReport {
	// Create error map for quick lookup
	errorsByRule := make(map[string][]ValidationError)
	for _, err := range result.Errors {
		errorsByRule[err.Rule] = append(errorsByRule[err.Rule], err)
	}

	// Filter to only commit-level rules (exclude repository rules for commit reports)
	commitRules := FilterCommitRules(allRules)

	// Create reports for all commit rules
	reports := make([]RuleReport, len(commitRules))

	for index, rule := range commitRules {
		ruleName := rule.Name()
		if errors, hasErrors := errorsByRule[ruleName]; hasErrors {
			// Rule failed
			reports[index] = RuleReport{
				Name:    ruleName,
				Status:  StatusFailed,
				Errors:  errors,
				Message: errors[0].Message, // Use first error message
			}
		} else {
			// Rule passed
			reports[index] = RuleReport{
				Name:    ruleName,
				Status:  StatusPassed,
				Errors:  nil,
				Message: "No errors",
			}
		}
	}

	return reports
}

// buildRuleReportsFromValidationResult creates rule reports from a validation result.
// This is used for the legacy BuildReport function.
func buildRuleReportsFromValidationResult(result ValidationResult) []RuleReport {
	// Group errors by rule
	errorsByRule := make(map[string][]ValidationError)
	for _, err := range result.Errors {
		errorsByRule[err.Rule] = append(errorsByRule[err.Rule], err)
	}

	// Create rule reports only for failed rules (legacy behavior)
	reports := make([]RuleReport, 0, len(errorsByRule))
	for rule, errs := range errorsByRule {
		reports = append(reports, RuleReport{
			Name:    rule,
			Status:  StatusFailed,
			Errors:  errs,
			Message: errs[0].Message,
		})
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
