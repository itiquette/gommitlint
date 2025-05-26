// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package domain provides test helpers for domain components.
package domain

import (
	"github.com/itiquette/gommitlint/internal/domain"
)

// NewTestRuleRegistry creates a new rule registry with custom default disabled rules for testing.
func NewTestRuleRegistry(defaultDisabledRules map[string]bool) domain.RuleRegistry {
	priorityService := domain.NewRulePriorityService(defaultDisabledRules)

	return domain.NewRuleRegistry(
		domain.WithPriorityService(priorityService),
	)
}

// WithDefaultDisabledRules creates a test rule registry with specific rules disabled by default.
func WithDefaultDisabledRules(rules ...string) domain.RuleRegistry {
	disabledMap := make(map[string]bool, len(rules))
	for _, rule := range rules {
		disabledMap[rule] = true
	}

	return NewTestRuleRegistry(disabledMap)
}
