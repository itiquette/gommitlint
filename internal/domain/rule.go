// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package domain

import (
	"context"

	"github.com/itiquette/gommitlint/internal/errors"
)

// Rule defines the interface for all validation rules.
// Rules are pure validators that check commits against specific criteria.
type Rule interface {
	// Name returns the rule's name.
	Name() string

	// Validate performs validation against a commit.
	// This should be a pure function that returns errors without side effects.
	Validate(ctx context.Context, commit CommitInfo) []errors.ValidationError
}

// ConfigurableRule extends Rule with configuration capabilities.
// Implementing this interface allows rules to be configured by the RuleRegistry.
type ConfigurableRule interface {
	Rule

	// WithConfig returns a new rule instance with applied configuration from context.
	// Following functional programming principles, it doesn't modify the original rule.
	WithContext(ctx context.Context) Rule
}

// RulePriorityManager defines an interface for determining rule activation status.
// This is primarily used by output formatters to filter results.
type RulePriorityManager interface {
	// IsRuleEnabled determines if a rule should be active based on configuration.
	IsRuleEnabled(ctx context.Context, ruleName string, enabledRules, disabledRules []string) bool

	// FilterRuleResults filters rule results based on configuration.
	FilterRuleResults(ctx context.Context, results []RuleResult, enabledRules, disabledRules []string) []RuleResult
}
