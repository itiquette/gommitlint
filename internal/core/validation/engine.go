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
	"github.com/itiquette/gommitlint/internal/infrastructure/log"
)

// Engine is responsible for running validation rules against commits.
// It does not use backward compatibility adapters anymore.
type Engine struct {
	ruleProvider domain.RuleProvider
	// No need to store configuration directly - it's in the context
}

// CreateEngine creates a validation engine using the configuration and context.
func CreateEngine(_ context.Context, _ config.Config, _ domain.CommitAnalyzer) Engine {
	// Now creates the engine with context directly
	// Parameters are unused as we've moved to context-based configuration
	return Engine{
		// Rule provider is now implemented directly in the application layer
		// using context-based configuration
		ruleProvider: nil,
	}
}

func (e Engine) GetRuleProvider() domain.RuleProvider {
	return e.ruleProvider
}
func (e Engine) ValidateCommit(ctx context.Context, commit domain.CommitInfo) domain.CommitResult {
	logger := log.Logger(ctx)
	logger.Trace().
		Str("commit_hash", commit.Hash).
		Str("subject", commit.Subject).
		Bool("is_merge_commit", commit.IsMergeCommit).
		Msg("Entering Engine.ValidateCommit")

	activeRules := e.ruleProvider.GetActiveRules(ctx)

	// Get the active rules - we no longer need special filtering here
	// as this is now handled by the rule provider's GetActiveRules method

	// Log active rules for debugging
	activeRuleNames := make([]string, 0, len(activeRules))
	for _, rule := range activeRules {
		activeRuleNames = append(activeRuleNames, rule.Name())
	}

	// Print debug info about rules
	logger.Debug().Strs("active_rules", activeRuleNames).Msg("Active rules for validation")

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
	logger := log.Logger(ctx)
	logger.Trace().
		Int("commit_count", len(commits)).
		Msg("Entering Engine.ValidateCommits")

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
