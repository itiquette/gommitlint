// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package validation provides core validation components.
package validation

import (
	"context"

	"github.com/itiquette/gommitlint/internal/application/factories"
	"github.com/itiquette/gommitlint/internal/common/contextx"
	"github.com/itiquette/gommitlint/internal/config/types"
	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/itiquette/gommitlint/internal/domain/formatting"
)

// Engine defines the interface for commit validation engines.
type Engine interface {
	// ValidateCommit validates a single commit against active rules.
	ValidateCommit(ctx context.Context, commit domain.CommitInfo) domain.CommitResult

	// ValidateCommits validates multiple commits against active rules.
	ValidateCommits(ctx context.Context, commits []domain.CommitInfo) domain.ValidationResults

	// GetRegistry returns the rule registry used by this engine.
	GetRegistry() *domain.RuleRegistry
}

// RegistryEngine is responsible for running validation rules against commits.
// This implementation uses RuleRegistry directly.
type RegistryEngine struct {
	registry *domain.RuleRegistry
}

// CreateEngine creates a validation engine using the context.
func CreateEngine(ctx context.Context, _ types.Config, analyzer domain.CommitAnalyzer) Engine {
	logger := contextx.GetLogger(ctx)
	logger.Debug("Entering validation.CreateEngine")

	// Create rule registry
	simpleRegistry := domain.NewRuleRegistry()

	// Create simple rule factory
	simpleFactory := factories.NewSimpleRuleFactory()

	// Create basic rules
	basicRules := simpleFactory.CreateBasicRules()

	// Create analyzer rules if we have one
	if analyzer != nil {
		analyzerRules := simpleFactory.CreateAnalyzerRules(analyzer)
		for name, factory := range analyzerRules {
			basicRules[name] = factory
		}
	}

	// Register all rules
	for name, factory := range basicRules {
		simpleRegistry.RegisterWithContext(ctx, name, factory)
	}

	// Create and return the registry engine
	return &RegistryEngine{
		registry: simpleRegistry,
	}
}

// ValidateCommit validates a single commit against all active rules.
func (e *RegistryEngine) ValidateCommit(ctx context.Context, commit domain.CommitInfo) domain.CommitResult {
	// Get logger from context
	logger := contextx.GetLogger(ctx)
	logger.Debug("Entering Engine.ValidateCommit",
		"commit_hash", commit.Hash,
		"subject", commit.Subject,
		"is_merge_commit", commit.IsMergeCommit)

	// Get enabled/disabled rules directly from context configuration
	cfg := contextx.GetConfig(ctx)
	enabledRules := cfg.GetStringSlice("rules.enabled_rules")
	disabledRules := cfg.GetStringSlice("rules.disabled_rules")

	// Log configuration values
	logger.Debug("Rule configuration from context",
		"enabled_rules", enabledRules,
		"disabled_rules", disabledRules)

	// Create active rules using the registry
	activeRules := e.registry.CreateActiveRules(ctx, enabledRules, disabledRules)

	// Log active rules for debugging
	activeRuleNames := make([]string, 0, len(activeRules))
	for _, rule := range activeRules {
		activeRuleNames = append(activeRuleNames, rule.Name())
	}

	logger.Debug("Active rules for validation", "active_rules", activeRuleNames)

	// Use helper method to validate with the provided rules
	return e.validateWithRules(ctx, commit, activeRules)
}

// validateWithRules is a helper method that validates a commit against a set of rules.
func (e *RegistryEngine) validateWithRules(ctx context.Context, commit domain.CommitInfo, rules []domain.Rule) domain.CommitResult {
	// Get logger from context
	logger := contextx.GetLogger(ctx)

	// Handle no active rules case
	if len(rules) == 0 {
		logger.Debug("No active rules, returning passing result")

		return domain.CommitResult{
			CommitInfo:  commit,
			RuleResults: []domain.RuleResult{},
			Passed:      true,
		}
	}

	// Initialize result
	result := domain.CommitResult{
		CommitInfo:  commit,
		RuleResults: make([]domain.RuleResult, 0, len(rules)),
		Passed:      true,
	}

	// Validate against each rule
	for _, rule := range rules {
		ruleErrors := rule.Validate(ctx, commit)

		status := domain.StatusPassed
		if len(ruleErrors) > 0 {
			status = domain.StatusFailed
			result.Passed = false
		}

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

// ValidateCommits validates multiple commits against all active rules.
func (e *RegistryEngine) ValidateCommits(ctx context.Context, commits []domain.CommitInfo) domain.ValidationResults {
	logger := contextx.GetLogger(ctx)
	logger.Debug("Entering Engine.ValidateCommits",
		"commit_count", len(commits))

	// Get config from context
	cfg := contextx.GetConfig(ctx)
	enabledRules := cfg.GetStringSlice("rules.enabled_rules")
	disabledRules := cfg.GetStringSlice("rules.disabled_rules")

	// Create active rules once to avoid repeated rule creation
	activeRules := e.registry.CreateActiveRules(ctx, enabledRules, disabledRules)

	// Create results
	results := domain.NewValidationResults()

	for _, commit := range commits {
		commitResult := e.validateWithRules(ctx, commit, activeRules)

		results.Results = append(results.Results, commitResult)
		results.TotalCommits++

		if commitResult.Passed {
			results.PassedCommits++
		}

		// Update rule summary
		for _, ruleResult := range commitResult.RuleResults {
			if ruleResult.Status == domain.StatusFailed {
				results.RuleSummary[ruleResult.RuleID]++
			}
		}
	}

	return results
}

// GetRegistry returns the rule registry used by this engine.
func (e *RegistryEngine) GetRegistry() *domain.RuleRegistry {
	return e.registry
}
