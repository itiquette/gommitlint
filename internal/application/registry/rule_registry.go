// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package registry provides services for rule registration and filtering.
package registry

import (
	"context"

	"github.com/itiquette/gommitlint/internal/common/contextx"
	"github.com/itiquette/gommitlint/internal/domain"
)

// RuleRegistry manages the creation and configuration of validation rules.
// It centralizes rule filtering based on configuration, eliminating the need
// for individual rules to handle their own configuration and filtering.
type RuleRegistry struct {
	allRules []domain.Rule
}

// NewRuleRegistry creates a registry with the given rules.
func NewRuleRegistry(rules ...domain.Rule) *RuleRegistry {
	return &RuleRegistry{
		allRules: rules,
	}
}

// GetActiveRules returns configured rules that are active based on context configuration.
// This is the main function to use when performing validation, as it handles
// both rule filtering and configuration based on the context.
func (r *RuleRegistry) GetActiveRules(ctx context.Context) []domain.Rule {
	if ctx == nil {
		return r.allRules // Return all rules with defaults if no context
	}

	// Get configuration from context
	cfg := contextx.GetConfig(ctx)
	if cfg == nil {
		return r.allRules // Return all rules with defaults if no config
	}

	enabled := cfg.GetStringSlice("rules.enabled")
	disabled := cfg.GetStringSlice("rules.disabled")

	// Create priority service for rule filtering
	priorityService := domain.NewRulePriorityService(domain.GetDefaultDisabledRules())

	// Apply configuration to each applicable rule
	var activeRules []domain.Rule

	for _, rule := range r.allRules {
		// Skip rules that are disabled by configuration
		if !priorityService.IsRuleEnabled(ctx, rule.Name(), enabled, disabled) {
			continue
		}

		// Configure the rule if it supports configuration
		if configurableRule, ok := rule.(domain.ConfigurableRule); ok {
			activeRules = append(activeRules, configurableRule.WithContext(ctx))
		} else {
			activeRules = append(activeRules, rule)
		}
	}

	return activeRules
}
