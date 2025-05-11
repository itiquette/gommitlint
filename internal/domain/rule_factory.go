// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package domain

import (
	"context"
)

// RuleFactory defines a factory type for rule creation.
type RuleFactory interface {
	CreateRule(ctx context.Context, analyzer CommitAnalyzer) Rule
}

// DefaultRuleProvider is a default implementation of the RuleProvider interface.
type DefaultRuleProvider struct {
	ctxValue context.Context //nolint:containedctx // This is a special case for testing
	analyzer CommitAnalyzer
	rules    []Rule
}

// NewDefaultRuleProvider creates a new DefaultRuleProvider with the given context and analyzer.
// This is temporary and should be replaced by a real implementation that creates actual rule instances.
// It's kept here to maintain backward compatibility until all usages are updated.
func NewDefaultRuleProvider(ctx context.Context, analyzer CommitAnalyzer) RuleProvider {
	return &DefaultRuleProvider{
		ctxValue: ctx,
		analyzer: analyzer,
	}
}

// GetRules returns all rules in the provider.
// This is a stub implementation that will be replaced by a real implementation
// that creates actual rule instances.
func (p *DefaultRuleProvider) GetRules(_ context.Context) []Rule {
	// This is a stub that should be replaced in production by
	// implementations that return real rule instances
	return []Rule{}
}

// GetActiveRules returns the active rules based on configuration.
func (p *DefaultRuleProvider) GetActiveRules(ctx context.Context) []Rule {
	// In a real implementation, this would filter rules based on configuration
	return p.GetRules(ctx)
}

// WithActiveRules returns a new provider with the specified active rules.
func (p *DefaultRuleProvider) WithActiveRules(_ []string) RuleProvider {
	// Create a copy of the provider
	newProvider := &DefaultRuleProvider{
		ctxValue: p.ctxValue,
		analyzer: p.analyzer,
		rules:    make([]Rule, len(p.rules)),
	}
	copy(newProvider.rules, p.rules)

	return newProvider
}

// WithDisabledRules returns a new provider with the specified rules disabled.
func (p *DefaultRuleProvider) WithDisabledRules(_ []string) RuleProvider {
	// Create a copy of the provider
	newProvider := &DefaultRuleProvider{
		ctxValue: p.ctxValue,
		analyzer: p.analyzer,
		rules:    make([]Rule, len(p.rules)),
	}
	copy(newProvider.rules, p.rules)

	return newProvider
}

// WithCustomRule returns a new provider with the custom rule added.
func (p *DefaultRuleProvider) WithCustomRule(rule Rule) RuleProvider {
	// Create a copy of the provider
	newProvider := &DefaultRuleProvider{
		ctxValue: p.ctxValue,
		analyzer: p.analyzer,
		rules:    make([]Rule, len(p.rules), len(p.rules)+1),
	}
	copy(newProvider.rules, p.rules)

	// Add the new rule
	newProvider.rules = append(newProvider.rules, rule)

	return newProvider
}
