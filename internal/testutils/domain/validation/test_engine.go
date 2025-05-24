// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package validation

import (
	"context"

	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/itiquette/gommitlint/internal/domain/formatting"
)

// TestEngine is a thin wrapper around domain.ValidationEngine that implements
// domain.ValidationEngine for testing purposes.
type TestEngine struct {
	Registry *domain.RuleRegistry
}

// ValidateCommit implements domain.ValidationEngine.
func (e TestEngine) ValidateCommit(ctx context.Context, commit domain.CommitInfo) domain.CommitResult {
	// No need to create a domain.ValidationEngine since we're implementing the interface directly
	// Since we can't directly set the unexported field, use the validation
	// functions directly with our provider
	// Get active rules from our registry
	activeRules := e.Registry.GetActiveRules(ctx, []string{}, []string{})

	// Initialize result
	result := domain.CommitResult{
		CommitInfo:  commit,
		RuleResults: make([]domain.RuleResult, 0, len(activeRules)),
		Passed:      true,
	}

	// Validate against each rule
	for _, rule := range activeRules {
		ruleErrors := rule.Validate(ctx, commit)

		// Determine status based on errors
		status := domain.StatusPassed
		if len(ruleErrors) > 0 {
			status = domain.StatusFailed
			result.Passed = false
		}

		// Create rule result
		ruleResult := domain.RuleResult{
			RuleID:         rule.Name(),
			RuleName:       rule.Name(),
			Status:         status,
			Message:        formatting.FormatResult(rule.Name(), ruleErrors),
			VerboseMessage: formatting.FormatVerboseResult(rule.Name(), ruleErrors),
			HelpMessage:    formatting.FormatHelp(rule.Name(), ruleErrors),
			Errors:         ruleErrors,
		}

		result.RuleResults = append(result.RuleResults, ruleResult)
	}

	return result
}

// ValidateCommits implements domain.ValidationEngine.
func (e TestEngine) ValidateCommits(ctx context.Context, commits []domain.CommitInfo) domain.ValidationResults {
	// Create a new ValidationResults
	results := domain.NewValidationResults()

	// Process each commit
	for _, commit := range commits {
		// Validate the commit
		commitResult := e.ValidateCommit(ctx, commit)

		// Add to results
		results.Results = append(results.Results, commitResult)

		// Update totals
		results.TotalCommits++
		if commitResult.Passed {
			results.PassedCommits++
		}

		// Update rule summary
		for _, ruleResult := range commitResult.RuleResults {
			if ruleResult.Status == domain.StatusFailed {
				if results.RuleSummary == nil {
					results.RuleSummary = make(map[string]int)
				}

				results.RuleSummary[ruleResult.RuleID]++
			}
		}
	}

	return results
}

// GetRegistry implements domain.ValidationEngine.
func (e TestEngine) GetRegistry() *domain.RuleRegistry {
	return e.Registry
}
