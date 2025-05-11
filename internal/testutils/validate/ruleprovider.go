// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package validate contains test utilities for the validate package.
// This package is intended for testing purposes only.
package validate

import (
	"context"

	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/itiquette/gommitlint/internal/errors"
)

// ProviderFactory is a function that creates a rule provider for testing.
// This can be set in tests to provide custom rule implementations.
type ProviderFactory func(ctx context.Context, analyzer domain.CommitAnalyzer) domain.RuleProvider

// MockRule is a test implementation of the Rule interface.
type MockRule struct {
	name         string
	shouldPass   bool
	errors       []errors.ValidationError
	validationFn func(domain.CommitInfo) []errors.ValidationError
}

// NewMockRule creates a simple pass/fail mock rule for testing.
func NewMockRule(name string, shouldPass bool) *MockRule {
	return &MockRule{
		name:       name,
		shouldPass: shouldPass,
	}
}

// NewMockRuleWithErrors creates a mock rule that returns the specified errors.
func NewMockRuleWithErrors(name string, errs []errors.ValidationError) *MockRule {
	return &MockRule{
		name:       name,
		shouldPass: false,
		errors:     errs,
	}
}

// NewMockRuleWithCustomValidation creates a mock rule with a custom validation function.
func NewMockRuleWithCustomValidation(name string, validationFn func(domain.CommitInfo) []errors.ValidationError) *MockRule {
	return &MockRule{
		name:         name,
		validationFn: validationFn,
	}
}

// Name returns the rule name.
func (r *MockRule) Name() string {
	return r.name
}

// Validate performs the validation.
func (r *MockRule) Validate(_ context.Context, commit domain.CommitInfo) []errors.ValidationError {
	// If we have a custom validation function, use that
	if r.validationFn != nil {
		return r.validationFn(commit)
	}

	// Otherwise, use the simple pass/fail behavior
	if r.shouldPass {
		return nil
	}

	// If we have specific errors, return those
	if len(r.errors) > 0 {
		return r.errors
	}

	// Default error
	return []errors.ValidationError{
		{
			Rule:    r.name,
			Code:    "error",
			Message: "This is a mock error for testing",
		},
	}
}

// MockRuleProvider is a test implementation of the RuleProvider interface.
type MockRuleProvider struct {
	rules []domain.Rule
}

// NewMockRuleProvider creates a new mock rule provider with the given rules.
func NewMockRuleProvider(rules []domain.Rule) *MockRuleProvider {
	return &MockRuleProvider{
		rules: rules,
	}
}

// GetRules returns all rules regardless of active status.
func (p *MockRuleProvider) GetRules(_ context.Context) []domain.Rule {
	return p.rules
}

// GetActiveRules returns active rules.
func (p *MockRuleProvider) GetActiveRules(_ context.Context) []domain.Rule {
	return p.rules // For mock, return all rules as active
}

// WithActiveRules returns a new provider with only the specified rules active.
func (p *MockRuleProvider) WithActiveRules(ruleNames []string) domain.RuleProvider {
	// Create a map for O(1) lookup
	enabled := make(map[string]bool)
	for _, name := range ruleNames {
		enabled[name] = true
	}

	// Create a new slice with only enabled rules
	filteredRules := make([]domain.Rule, 0)

	for _, rule := range p.rules {
		if enabled[rule.Name()] {
			filteredRules = append(filteredRules, rule)
		}
	}

	return NewMockRuleProvider(filteredRules)
}

// WithDisabledRules returns a new provider with the specified rules disabled.
func (p *MockRuleProvider) WithDisabledRules(ruleNames []string) domain.RuleProvider {
	// Create a map for O(1) lookup
	disabled := make(map[string]bool)
	for _, name := range ruleNames {
		disabled[name] = true
	}

	// Create a new slice excluding disabled rules
	filteredRules := make([]domain.Rule, 0)

	for _, rule := range p.rules {
		if !disabled[rule.Name()] {
			filteredRules = append(filteredRules, rule)
		}
	}

	return NewMockRuleProvider(filteredRules)
}

// WithCustomRule returns a new provider with the custom rule added.
func (p *MockRuleProvider) WithCustomRule(rule domain.Rule) domain.RuleProvider {
	newRules := make([]domain.Rule, len(p.rules), len(p.rules)+1)
	copy(newRules, p.rules)
	newRules = append(newRules, rule)

	return NewMockRuleProvider(newRules)
}

// GetAvailableRuleNames returns the names of all available rules.
func (p *MockRuleProvider) GetAvailableRuleNames(_ context.Context) []string {
	names := make([]string, 0, len(p.rules))
	for _, rule := range p.rules {
		names = append(names, rule.Name())
	}

	return names
}
