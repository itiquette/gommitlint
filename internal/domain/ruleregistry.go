// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package domain

import (
	"context"
	"fmt"
	"sort"

	"github.com/itiquette/gommitlint/internal/common/contextx"
)

// RuleRegistry provides a simplified rule management system.
type RuleRegistry struct {
	factories       map[string]func(context.Context) Rule
	defaultDisabled map[string]bool
}

// NewRuleRegistry creates a new rule registry.
func NewRuleRegistry() *RuleRegistry {
	return &RuleRegistry{
		factories:       make(map[string]func(context.Context) Rule),
		defaultDisabled: GetDefaultDisabledRules(),
	}
}

// Register adds a rule to the registry.
func (r *RuleRegistry) Register(name string, factory func(context.Context) Rule) {
	r.RegisterWithContext(context.Background(), name, factory)
}

// RegisterWithContext adds a rule to the registry with a context.
func (r *RuleRegistry) RegisterWithContext(ctx context.Context, name string, factory func(context.Context) Rule) {
	r.factories[name] = factory
	logger := contextx.GetLogger(ctx)
	logger.Debug("Registered rule", "rule", name)
}

// Create instantiates a rule by name.
func (r *RuleRegistry) Create(ctx context.Context, name string) (Rule, error) {
	factory, exists := r.factories[name]
	if !exists {
		return nil, fmt.Errorf("rule not found: %s", name)
	}

	rule := factory(ctx)
	if rule == nil {
		return nil, fmt.Errorf("failed to create rule: %s", name)
	}

	return rule, nil
}

// GetEnabledRules returns a list of enabled rules based on configuration.
func (r *RuleRegistry) GetEnabledRules(ctx context.Context) []Rule {
	var rules []Rule

	// Get enabled/disabled from config
	config := contextx.GetConfig(ctx)
	enabledMap := makeRuleMap(config.GetStringSlice("rules.enabled_rules"))
	disabledMap := makeRuleMap(config.GetStringSlice("rules.disabled_rules"))

	// Create rules that are enabled
	for name, factory := range r.factories {
		if r.shouldEnableRule(name, enabledMap, disabledMap) {
			if rule := factory(ctx); rule != nil {
				rules = append(rules, rule)
			}
		}
	}

	return rules
}

// shouldEnableRule determines if a rule should be enabled.
func (r *RuleRegistry) shouldEnableRule(name string, enabled, disabled map[string]bool) bool {
	// Explicitly disabled takes precedence
	if disabled[name] {
		return false
	}

	// Explicitly enabled
	if enabled[name] {
		return true
	}

	// Check default disabled list
	return !r.defaultDisabled[name]
}

// makeRuleMap converts a slice of rule names to a map for quick lookup.
func makeRuleMap(rules []string) map[string]bool {
	m := make(map[string]bool)
	for _, rule := range rules {
		m[rule] = true
	}

	return m
}

// CreateRule creates a rule instance with the specified name.
func (r *RuleRegistry) CreateRule(ctx context.Context, name string) Rule {
	factory, exists := r.factories[name]
	if !exists {
		logger := contextx.GetLogger(ctx)
		logger.Warn("Rule factory not found", "rule", name)

		return nil
	}

	return factory(ctx)
}

// IsRuleRegistered checks if a rule with the specified name is registered.
func (r *RuleRegistry) IsRuleRegistered(name string) bool {
	_, exists := r.factories[name]

	return exists
}

// GetRuleNames returns a sorted list of all registered rule names.
func (r *RuleRegistry) GetRuleNames() []string {
	names := make([]string, 0, len(r.factories))
	for name := range r.factories {
		names = append(names, name)
	}
	// Sort for consistent output
	sort.Strings(names)

	return names
}

// CreateAllRules creates all registered rules.
func (r *RuleRegistry) CreateAllRules(ctx context.Context) []Rule {
	names := r.GetRuleNames()
	rules := make([]Rule, 0, len(names))

	for _, name := range names {
		rule := r.CreateRule(ctx, name)
		if rule != nil {
			rules = append(rules, rule)
		}
	}

	return rules
}

// CreateActiveRules creates a list of active rules based on configuration.
func (r *RuleRegistry) CreateActiveRules(ctx context.Context, enabledRules, disabledRules []string) []Rule {
	names := r.GetRuleNames()
	active := make([]Rule, 0, len(names))

	enabledMap := makeRuleMap(enabledRules)
	disabledMap := makeRuleMap(disabledRules)

	for _, name := range names {
		if r.shouldEnableRule(name, enabledMap, disabledMap) {
			rule := r.CreateRule(ctx, name)
			if rule != nil {
				active = append(active, rule)
			}
		}
	}

	return active
}

// ValidateRule validates that a rule with the specified name exists.
func (r *RuleRegistry) ValidateRule(name string) error {
	if !r.IsRuleRegistered(name) {
		return fmt.Errorf("rule %q not found", name)
	}

	return nil
}

// IsDefaultDisabled checks if a rule is disabled by default.
func (r *RuleRegistry) IsDefaultDisabled(rule string) bool {
	return r.defaultDisabled[rule]
}

// SetDefaultDisabled sets whether a rule is disabled by default.
func (r *RuleRegistry) SetDefaultDisabled(rule string, disabled bool) {
	r.defaultDisabled[rule] = disabled
}

// GetDefaultDisabledRules returns a copy of the default disabled rules map.
func (r *RuleRegistry) GetDefaultDisabledRules() map[string]bool {
	result := make(map[string]bool, len(r.defaultDisabled))
	for k, v := range r.defaultDisabled {
		result[k] = v
	}

	return result
}
