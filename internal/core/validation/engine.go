// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package validation provides core validation components.
package validation

import (
	"context"

	"github.com/itiquette/gommitlint/internal/config"
	"github.com/itiquette/gommitlint/internal/contextx"
	"github.com/itiquette/gommitlint/internal/domain"
)

// Engine is responsible for running validation rules against commits.
type Engine struct {
	ruleProvider domain.RuleProvider
}

// NewEngine creates a new validation engine.
func NewEngine(provider domain.RuleProvider) Engine {
	return Engine{
		ruleProvider: provider,
	}
}

// NewFakeConfigForTesting creates a simple config for testing purposes.
// This is only used in tests and should not be used in production code.
func NewFakeConfigForTesting() config.Config {
	return config.NewConfig()
}

func (e Engine) GetRuleProvider() domain.RuleProvider {
	return e.ruleProvider
}
func (e Engine) ValidateCommit(ctx context.Context, commit domain.CommitInfo) domain.CommitResult {
	activeRules := e.ruleProvider.GetActiveRules()
	// Initialize result
	result := domain.CommitResult{
		CommitInfo:  commit,
		RuleResults: nil, // Will be populated by Map function
		Passed:      true,
	}

	// Transform rules into rule results using Map
	ruleResults := contextx.Map(activeRules, func(rule domain.Rule) domain.RuleResult {
		// Validate commit against rule
		ruleErrors := rule.Validate(ctx, commit)

		// Determine status based on errors
		status := domain.StatusPassed
		if len(ruleErrors) > 0 {
			status = domain.StatusFailed
			result.Passed = false // Side effect: update the overall pass/fail status
		}

		// Create rule result
		ruleResult := domain.RuleResult{
			RuleID:         rule.Name(),
			RuleName:       rule.Name(),
			Status:         status,
			Message:        rule.Result(ruleErrors),
			VerboseMessage: rule.VerboseResult(ruleErrors),
			HelpMessage:    rule.Help(ruleErrors),
			Errors:         ruleErrors,
		}

		return ruleResult
	})

	// Add all rule results to the commit result
	result.RuleResults = ruleResults

	return result
}

// ValidateCommits validates multiple commits against all active rules.
func (e Engine) ValidateCommits(ctx context.Context, commits []domain.CommitInfo) domain.ValidationResults {
	// Create a new ValidationResults
	results := domain.NewValidationResults()

	// Define a validator function
	validateCommit := func(acc struct {
		results domain.ValidationResults
		done    bool
	}, commit domain.CommitInfo) struct {
		results domain.ValidationResults
		done    bool
	} {
		// Check if we're already done
		if acc.done {
			return struct {
				results domain.ValidationResults
				done    bool
			}{
				results: acc.results,
				done:    true,
			}
		}

		// Create a new results object to maintain immutability
		newResults := domain.NewValidationResults()

		// Copy existing results
		newResults.CommitResults = contextx.DeepCopy(acc.results.CommitResults)

		// Validate the commit and append to results
		commitResult := e.ValidateCommit(ctx, commit)
		newResults.CommitResults = append(newResults.CommitResults, commitResult)

		// Update total and passed commits counts
		newResults.TotalCommits = acc.results.TotalCommits + 1
		newResults.PassedCommits = acc.results.PassedCommits

		if commitResult.Passed {
			newResults.PassedCommits++
		}

		// Copy and update rule summary
		newResults.RuleSummary = contextx.DeepCopyMap(acc.results.RuleSummary)

		for _, ruleResult := range commitResult.RuleResults {
			if ruleResult.Status == domain.StatusFailed {
				newResults.RuleSummary[ruleResult.RuleID]++
			}
		}

		return struct {
			results domain.ValidationResults
			done    bool
		}{
			results: newResults,
			done:    false,
		}
	}

	// Use reduce to process commits in a functional way
	accumulator := contextx.Reduce(
		commits,
		struct {
			results domain.ValidationResults
			done    bool
		}{
			results: results,
			done:    false,
		},
		func(acc struct {
			results domain.ValidationResults
			done    bool
		}, commit domain.CommitInfo) struct {
			results domain.ValidationResults
			done    bool
		} {
			// Skip processing if we're already done
			if acc.done {
				return acc
			}

			return validateCommit(acc, commit)
		},
	)

	return accumulator.results
}
