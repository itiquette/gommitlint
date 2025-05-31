// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package domain

// RuleResult represents the result of validating a single rule against a commit.
// Simplified to focus only on essential information.
type RuleResult struct {
	// RuleName is the human-readable name of the rule.
	RuleName string

	// Status indicates whether the rule passed, failed, or was skipped.
	Status ValidationStatus

	// Errors contains any validation errors detected by the rule.
	Errors []ValidationError
}

// CommitResult represents the results of validating all rules against a single commit.
// Simplified to contain only necessary fields.
type CommitResult struct {
	// CommitInfo contains information about the validated commit.
	CommitInfo CommitInfo

	// RuleResults contains the results of each rule applied to the commit.
	RuleResults []RuleResult

	// Passed indicates whether all rules passed validation.
	Passed bool
}

// ValidationResults contains the validation results for multiple commits.
// Simplified to focus on core functionality.
type ValidationResults struct {
	// Results contains individual commit validation results
	Results []CommitResult

	// TotalCommits is the total number of commits validated.
	TotalCommits int

	// PassedCommits is the number of commits that passed all validations.
	PassedCommits int
}

// NewCommitResult creates a new commit result.
func NewCommitResult(commit CommitInfo) CommitResult {
	return CommitResult{
		CommitInfo:  commit,
		RuleResults: []RuleResult{},
		Passed:      true,
	}
}

// AddRuleResult adds a rule result to the commit result.
// This replaces the complex WithRuleResult and WithFormattedRuleResult methods.
func (r CommitResult) AddRuleResult(ruleName string, errs []ValidationError) CommitResult {
	// Create new result with copied data
	newResult := CommitResult{
		CommitInfo: r.CommitInfo,
		Passed:     r.Passed,
	}

	// Copy existing rule results
	newResult.RuleResults = make([]RuleResult, len(r.RuleResults), len(r.RuleResults)+1)
	copy(newResult.RuleResults, r.RuleResults)

	// Determine status
	status := StatusPassed
	if len(errs) > 0 {
		status = StatusFailed
		newResult.Passed = false
	}

	// Add new rule result
	ruleResult := RuleResult{
		RuleName: ruleName,
		Status:   status,
		Errors:   errs,
	}
	newResult.RuleResults = append(newResult.RuleResults, ruleResult)

	return newResult
}

// NewValidationResults creates a new ValidationResults instance.
func NewValidationResults() ValidationResults {
	return ValidationResults{
		Results: []CommitResult{},
	}
}

// AddResult adds a commit result to the validation results.
// This replaces the WithResult method with a simpler name.
func (r ValidationResults) AddResult(result CommitResult) ValidationResults {
	newResults := ValidationResults{
		TotalCommits:  r.TotalCommits + 1,
		PassedCommits: r.PassedCommits,
	}

	if result.Passed {
		newResults.PassedCommits++
	}

	// Copy existing results and add new one
	newResults.Results = make([]CommitResult, len(r.Results), len(r.Results)+1)
	copy(newResults.Results, r.Results)
	newResults.Results = append(newResults.Results, result)

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

// GetFailedRules returns a map of rule names to the number of commits that failed that rule.
// This provides the essential functionality without the complex RuleSummary map.
func (r ValidationResults) GetFailedRules() map[string]int {
	summary := make(map[string]int)

	for _, result := range r.Results {
		for _, ruleResult := range result.RuleResults {
			if ruleResult.Status == StatusFailed {
				summary[ruleResult.RuleName]++
			}
		}
	}

	return summary
}
