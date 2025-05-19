// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package domain

import (
	"github.com/itiquette/gommitlint/internal/errors"
)

// RuleResult represents the result of validating a single rule against a commit.
type RuleResult struct {
	// RuleID is the unique identifier of the rule.
	RuleID string

	// RuleName is the human-readable name of the rule.
	RuleName string

	// Status indicates whether the rule passed, failed, or was skipped.
	Status ValidationStatus

	// Message is a human-readable result message.
	Message string

	// VerboseMessage is a detailed result message with more information.
	VerboseMessage string

	// HelpMessage provides guidance on how to fix rule violations.
	HelpMessage string

	// Errors contains any validation errors detected by the rule.
	Errors []errors.ValidationError
}

// CommitResult represents the results of validating all rules against a single commit.
type CommitResult struct {
	// CommitInfo contains information about the validated commit.
	CommitInfo CommitInfo

	// RuleResults contains the results of each rule applied to the commit.
	RuleResults []RuleResult

	// RuleErrorMap contains validation errors for each rule by name
	RuleErrorMap map[string][]errors.ValidationError

	// Metadata contains additional information about the validation
	Metadata map[string]string

	// Passed indicates whether all rules passed validation.
	Passed bool
}

// ValidationResults contains the validation results for multiple commits.
type ValidationResults struct {
	// Results contains individual commit validation results
	Results []CommitResult

	// RuleSummary summarizes rule failures across all commits.
	RuleSummary map[string]int

	// TotalCommits is the total number of commits validated.
	TotalCommits int

	// PassedCommits is the number of commits that passed all validations.
	PassedCommits int

	// Metadata contains additional information
	Metadata map[string]string
}

// NewCommitResult creates a new commit result.
func NewCommitResult(commit CommitInfo) CommitResult {
	return CommitResult{
		CommitInfo:   commit,
		RuleResults:  []RuleResult{},
		RuleErrorMap: make(map[string][]errors.ValidationError),
		Metadata:     make(map[string]string),
		Passed:       true,
	}
}

// WithRuleResult returns a new CommitResult with added rule errors (functional).
func (r CommitResult) WithRuleResult(ruleName string, errs []errors.ValidationError) CommitResult {
	// Create new result
	newResult := CommitResult{
		CommitInfo: r.CommitInfo,
		Passed:     r.Passed,
	}

	// Deep copy rule results
	newResult.RuleResults = make([]RuleResult, len(r.RuleResults), len(r.RuleResults)+1)
	copy(newResult.RuleResults, r.RuleResults)

	// Deep copy error map
	newResult.RuleErrorMap = make(map[string][]errors.ValidationError)

	for k, v := range r.RuleErrorMap {
		errCopy := make([]errors.ValidationError, len(v))
		copy(errCopy, v)
		newResult.RuleErrorMap[k] = errCopy
	}

	newResult.RuleErrorMap[ruleName] = errs

	// Deep copy metadata
	newResult.Metadata = make(map[string]string)
	for k, v := range r.Metadata {
		newResult.Metadata[k] = v
	}

	// Create and add rule result
	ruleResult := RuleResult{
		RuleID:   ruleName,
		RuleName: ruleName,
		Errors:   errs,
		Status:   StatusPassed,
	}

	if len(errs) > 0 {
		ruleResult.Status = StatusFailed
		newResult.Passed = false
	}

	newResult.RuleResults = append(newResult.RuleResults, ruleResult)

	return newResult
}

// NewValidationResults creates a new ValidationResults instance.
func NewValidationResults() ValidationResults {
	return ValidationResults{
		Results:     []CommitResult{},
		RuleSummary: make(map[string]int),
		Metadata:    make(map[string]string),
	}
}

// WithResult returns a new ValidationResults with an added result (functional).
func (r ValidationResults) WithResult(result CommitResult) ValidationResults {
	newResults := ValidationResults{
		TotalCommits:  r.TotalCommits + 1,
		PassedCommits: r.PassedCommits,
	}

	if result.Passed {
		newResults.PassedCommits++
	}

	// Deep copy results
	newResults.Results = make([]CommitResult, len(r.Results), len(r.Results)+1)
	copy(newResults.Results, r.Results)
	newResults.Results = append(newResults.Results, result)

	// Deep copy rule summary
	newResults.RuleSummary = make(map[string]int)
	for k, v := range r.RuleSummary {
		newResults.RuleSummary[k] = v
	}

	// Update summary for failed rules
	for _, ruleResult := range result.RuleResults {
		if ruleResult.Status == StatusFailed {
			newResults.RuleSummary[ruleResult.RuleID]++
		}
	}

	// Deep copy metadata
	newResults.Metadata = make(map[string]string)
	for k, v := range r.Metadata {
		newResults.Metadata[k] = v
	}

	return newResults
}

// AllPassed returns true if all commits passed validation.
func (r ValidationResults) AllPassed() bool {
	return r.TotalCommits == r.PassedCommits
}

// Count returns the total number of commits in the validation results.
func (r ValidationResults) Count() int {
	return r.TotalCommits
}
