// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package validation provides core validation components.
package validation

import (
	"context"
	"fmt"

	"github.com/itiquette/gommitlint/internal/config"
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
		RuleResults: make([]domain.RuleResult, 0, len(activeRules)),
		Passed:      true,
	}

	// Run each rule
	for _, rule := range activeRules {
		// Validate commit against rule
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
			Message:        rule.Result(ruleErrors),
			VerboseMessage: rule.VerboseResult(ruleErrors),
			HelpMessage:    rule.Help(ruleErrors),
			Errors:         ruleErrors,
		}

		fmt.Printf("%+v\n", ruleResult)
		// Add rule result to the results collection
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
