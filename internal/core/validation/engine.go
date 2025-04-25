// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package validation provides core validation components.
package validation

import (
	"context"

	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/itiquette/gommitlint/internal/errors"
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

// ValidateCommit validates a single commit against all active rules.
func (e Engine) ValidateCommit(ctx context.Context, commit domain.CommitInfo) domain.CommitResult {
	activeRules := e.ruleProvider.GetActiveRules()

	// Initialize result
	result := domain.CommitResult{
		CommitInfo:  commit,
		RuleResults: make([]domain.RuleResult, 0, len(activeRules)),
		Passed:      true,
	}

	// Run each rule
	for _, rule := range activeRules {
		// Check for context cancellation
		if ctx.Err() != nil {
			break
		}

		// Check if the rule supports context
		var ruleErrors []errors.ValidationError
		if contextualRule, ok := rule.(domain.ContextualRule); ok {
			// Use the context-aware validation method
			ruleErrors = contextualRule.ValidateWithContext(ctx, commit)
		} else {
			// Fall back to the regular validation method
			ruleErrors = rule.Validate(commit)
		}

		// Create rule result
		ruleResult := domain.RuleResult{
			RuleID:         rule.Name(),
			RuleName:       rule.Name(),
			Message:        rule.Result(),
			VerboseMessage: rule.VerboseResult(),
			HelpMessage:    rule.Help(),
			Errors:         ruleErrors,
		}

		// Set status based on errors
		if len(ruleErrors) > 0 {
			ruleResult.Status = domain.StatusFailed
			result.Passed = false
		} else {
			ruleResult.Status = domain.StatusPassed
		}

		// Add to results
		result.RuleResults = append(result.RuleResults, ruleResult)
	}

	return result
}

// ValidateCommits validates multiple commits against all active rules.
func (e Engine) ValidateCommits(ctx context.Context, commits []domain.CommitInfo) domain.ValidationResults {
	results := domain.NewValidationResults()

	for _, commit := range commits {
		// Check for context cancellation
		if ctx.Err() != nil {
			break
		}

		// Validate commit
		commitResult := e.ValidateCommit(ctx, commit)

		// Add to results
		results.AddCommitResult(commitResult)
	}

	return results
}
