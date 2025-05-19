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

// RuleProvider interface has been removed in favor of RuleRegistry.
// Use RuleRegistry and the registry-based validation engine directly.

// RulePriorityManager defines an interface for determining rule activation status.
// This is primarily used by output formatters to filter results.
type RulePriorityManager interface {
	// IsRuleEnabled determines if a rule should be active based on configuration.
	IsRuleEnabled(ctx context.Context, ruleName string, enabledRules, disabledRules []string) bool

	// FilterRuleResults filters rule results based on configuration.
	FilterRuleResults(ctx context.Context, results []RuleResult, enabledRules, disabledRules []string) []RuleResult
}
