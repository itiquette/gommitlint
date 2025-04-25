// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package domain

import (
	"github.com/itiquette/gommitlint/internal/errors"
)

// ValidationStatus represents the result status of a validation rule.
type ValidationStatus string

const (
	// StatusPassed indicates the rule passed validation.
	StatusPassed ValidationStatus = "passed"

	// StatusFailed indicates the rule failed validation.
	StatusFailed ValidationStatus = "failed"

	// StatusSkipped indicates the rule was skipped for some reason.
	StatusSkipped ValidationStatus = "skipped"
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

	// Passed indicates whether all rules passed validation.
	Passed bool
}

// ValidationResults contains the validation results for multiple commits.
type ValidationResults struct {
	// CommitResults contains the results for each validated commit.
	CommitResults []CommitResult

	// RuleSummary summarizes rule failures across all commits.
	RuleSummary map[string]int

	// TotalCommits is the total number of commits validated.
	TotalCommits int

	// PassedCommits is the number of commits that passed all validations.
	PassedCommits int
}

// AddCommitResult adds a commit result to the validation results.
func (r *ValidationResults) AddCommitResult(result CommitResult) {
	r.CommitResults = append(r.CommitResults, result)
	r.TotalCommits++

	if result.Passed {
		r.PassedCommits++
	}

	// Update rule summary
	for _, ruleResult := range result.RuleResults {
		if ruleResult.Status == StatusFailed {
			r.RuleSummary[ruleResult.RuleID]++
		}
	}
}

// AllPassed returns true if all commits passed validation.
func (r *ValidationResults) AllPassed() bool {
	return r.TotalCommits == r.PassedCommits
}

// Count returns the total number of commits in the validation results.
func (r *ValidationResults) Count() int {
	return r.TotalCommits
}

// NewValidationResults creates a new ValidationResults instance.
func NewValidationResults() ValidationResults {
	return ValidationResults{
		CommitResults: make([]CommitResult, 0),
		RuleSummary:   make(map[string]int),
	}
}
