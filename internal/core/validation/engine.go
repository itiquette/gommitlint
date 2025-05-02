// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package validation provides core validation components.
package validation

import (
	"context"

	"github.com/itiquette/gommitlint/internal/config"
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
		// Check for context cancellation
		if ctx.Err() != nil {
			break
		}

		ruleErrors := rule.Validate(ctx, commit)

		// Determine status first based on errors
		status := domain.StatusPassed
		if len(ruleErrors) > 0 {
			status = domain.StatusFailed
			result.Passed = false
		}
		// For rules that need errors to properly return status messages,
		// try to set errors if possible
		var ruleWithErrors = rule
		// Try to set errors on the rule if it supports it
		// This is needed for rules like CommitsAhead that need to know about errors to produce correct messages
		if setter, ok := rule.(interface {
			SetErrors([]errors.ValidationError) domain.Rule //nolint
		}); ok {
			// Filter out any special internal state transfer errors before using the errors for status
			// This ensures they don't affect the validation result incorrectly
			displayErrors := []errors.ValidationError{}

			for _, err := range ruleErrors {
				if code := errors.ValidationErrorCode(err.Code); code != "internal_state_transfer" {
					displayErrors = append(displayErrors, err)
				}
			}

			// Always set the original errors to ensure we transfer state correctly
			ruleWithErrors = setter.SetErrors(ruleErrors)

			// But if there were only internal state transfer errors, update status to passed
			if len(displayErrors) == 0 && len(ruleErrors) > 0 {
				status = domain.StatusPassed
				result.Passed = true
			}
		}

		// Create rule result with status already set
		// Prepare the errors list - filter out any internal state transfer errors
		displayErrors := ruleErrors
		if _, ok := rule.(interface {
			SetErrors([]errors.ValidationError) domain.Rule //nolint
		}); ok {
			// We only need to filter if we are using setter rules
			displayErrors = []errors.ValidationError{}

			for _, err := range ruleErrors {
				if code := errors.ValidationErrorCode(err.Code); code != "internal_state_transfer" {
					displayErrors = append(displayErrors, err)
				}
			}
		}

		ruleResult := domain.RuleResult{
			RuleID:         rule.Name(),
			RuleName:       rule.Name(),
			Status:         status,
			Message:        ruleWithErrors.Result(),
			VerboseMessage: ruleWithErrors.VerboseResult(),
			HelpMessage:    ruleWithErrors.Help(), // Use the rule with errors set
			Errors:         displayErrors,         // Only include real validation errors, not internal state transfer errors
		}

		// Add rule result to the results collection
		// All rules in the activeRules list should always be included in results
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
