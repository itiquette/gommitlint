// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package validation provides core validation components.
package validation

import (
	"context"

	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/itiquette/gommitlint/internal/domain/formatting"
)

// Config holds configuration for validation engine.
type Config struct {
	EnabledRules  []string
	DisabledRules []string
}

// RegistryEngine is responsible for running validation rules against commits.
// This implementation uses RuleRegistry directly and follows dependency injection principles.
type RegistryEngine struct {
	registry      *domain.RuleRegistry
	enabledRules  []string
	disabledRules []string
}

// CreateEngine creates a validation engine with explicit dependencies.
func CreateEngine(config Config, _ domain.CommitAnalyzer, ruleRegistry *domain.RuleRegistry) domain.ValidationEngine {
	return &RegistryEngine{
		registry:      ruleRegistry,
		enabledRules:  config.EnabledRules,
		disabledRules: config.DisabledRules,
	}
}

// ValidateCommit validates a single commit against all active rules.
func (e RegistryEngine) ValidateCommit(ctx context.Context, commit domain.CommitInfo) domain.CommitResult {
	// Use injected configuration instead of service locator pattern
	activeRules := e.registry.GetActiveRules(ctx, e.enabledRules, e.disabledRules)

	// Use helper method to validate with the provided rules
	return e.validateWithRules(ctx, commit, activeRules)
}

// validateWithRules is a helper method that validates a commit against a set of rules.
func (e RegistryEngine) validateWithRules(ctx context.Context, commit domain.CommitInfo, rules []domain.Rule) domain.CommitResult {
	// Handle no active rules case
	if len(rules) == 0 {
		return domain.NewCommitResult(commit)
	}

	// Initialize result using functional constructor
	result := domain.NewCommitResult(commit)

	// Validate against each rule using functional methods
	for _, rule := range rules {
		ruleErrors := rule.Validate(ctx, commit)

		// Determine status
		status := domain.StatusPassed
		if len(ruleErrors) > 0 {
			status = domain.StatusFailed
		}

		// Create formatted rule result
		ruleResult := domain.RuleResult{
			RuleID:         rule.Name(),
			RuleName:       rule.Name(),
			Status:         status,
			Message:        formatting.FormatResult(rule.Name(), ruleErrors),
			VerboseMessage: formatting.FormatVerboseResult(rule.Name(), ruleErrors),
			HelpMessage:    formatting.FormatHelp(rule.Name(), ruleErrors),
			Errors:         ruleErrors,
		}

		// Use WithFormattedRuleResult to maintain immutability
		result = result.WithFormattedRuleResult(ruleResult)
	}

	return result
}

// ValidateCommits validates multiple commits against all active rules.
func (e RegistryEngine) ValidateCommits(ctx context.Context, commits []domain.CommitInfo) domain.ValidationResults {
	// Use injected configuration instead of service locator pattern
	activeRules := e.registry.GetActiveRules(ctx, e.enabledRules, e.disabledRules)

	// Create results using functional approach
	results := domain.NewValidationResults()

	for _, commit := range commits {
		commitResult := e.validateWithRules(ctx, commit, activeRules)
		results = results.WithResult(commitResult)
	}

	return results
}
