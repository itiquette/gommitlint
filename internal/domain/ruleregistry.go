// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package domain

import (
	"context"
	"fmt"
	"sort"
	"strings"
)

// RuleRegistry provides an immutable rule management system.
// All operations return new instances rather than modifying state.
type RuleRegistry struct {
	factories       map[string]func(context.Context) Rule
	priorityService RulePriorityService
	rules           map[string]Rule
}

// RuleRegistryOption is a functional option for configuring a RuleRegistry.
type RuleRegistryOption func(RuleRegistry) RuleRegistry

// WithPriorityService sets a custom priority service for the registry.
func WithPriorityService(service RulePriorityService) RuleRegistryOption {
	return func(r RuleRegistry) RuleRegistry {
		r.priorityService = service

		return r
	}
}

// NewRuleRegistry creates a new rule registry.
func NewRuleRegistry(opts ...RuleRegistryOption) RuleRegistry {
	registry := RuleRegistry{
		factories:       make(map[string]func(context.Context) Rule),
		priorityService: NewRulePriorityService(GetDefaultDisabledRules()),
		rules:           make(map[string]Rule),
	}

	// Apply options
	for _, opt := range opts {
		registry = opt(registry)
	}

	return registry
}

// Register returns a new registry with the added rule factory.
// This maintains immutability by creating a new instance.
func (r RuleRegistry) Register(name string, factory func(context.Context) Rule) RuleRegistry {
	// Create new factories map
	newFactories := make(map[string]func(context.Context) Rule, len(r.factories)+1)
	for k, v := range r.factories {
		newFactories[k] = v
	}

	newFactories[name] = factory

	return RuleRegistry{
		factories:       newFactories,
		priorityService: r.priorityService,
		rules:           r.rules,
	}
}

// WithInitializedRules returns a new registry with all rules initialized.
// This replaces the mutable InitializeRules method.
func (r RuleRegistry) WithInitializedRules(ctx context.Context) RuleRegistry {
	// Create new rules map
	newRules := make(map[string]Rule, len(r.factories))

	// Create all rules
	for name, factory := range r.factories {
		rule := factory(ctx)
		if rule != nil {
			newRules[name] = rule
		}
	}

	return RuleRegistry{
		factories:       r.factories,
		priorityService: r.priorityService,
		rules:           newRules,
	}
}

// HasRules returns true if rules have been initialized.
func (r RuleRegistry) HasRules() bool {
	return len(r.rules) > 0
}

// Create instantiates a rule by name.
func (r RuleRegistry) Create(ctx context.Context, name string) (Rule, error) {
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
// Configuration must be provided via parameters to maintain hexagonal architecture.
func (r RuleRegistry) GetEnabledRules(ctx context.Context, enabledRules, disabledRules []string) []Rule {
	// If we have pre-created rules, use them
	if len(r.rules) > 0 {
		return r.GetActiveRules(ctx, enabledRules, disabledRules)
	}

	// Otherwise create rules that are enabled on-demand
	return r.createEnabledRules(ctx, enabledRules, disabledRules)
}

// createEnabledRules creates rules that are enabled based on configuration.
func (r RuleRegistry) createEnabledRules(ctx context.Context, enabledRules, disabledRules []string) []Rule {
	var rules []Rule

	for name, factory := range r.factories {
		if r.priorityService.IsRuleEnabled(ctx, name, enabledRules, disabledRules) {
			if rule := factory(ctx); rule != nil {
				rules = append(rules, rule)
			}
		}
	}

	return rules
}

// MakeRuleMap converts rule names to a map for fast lookups.
// This is a pure function that creates a new map.
func MakeRuleMap(rules []string) map[string]bool {
	ruleMap := make(map[string]bool, len(rules))

	for _, rule := range rules {
		// Skip commented out rules (those starting with #)
		if strings.HasPrefix(strings.TrimSpace(rule), "#") {
			continue
		}

		// Normalize each rule name for consistent lookup
		cleanRule := CleanRuleName(rule)
		ruleMap[cleanRule] = true
	}

	return ruleMap
}

// CreateRule creates a rule instance with the specified name.
func (r RuleRegistry) CreateRule(ctx context.Context, name string) Rule {
	factory, exists := r.factories[name]
	if !exists {
		return nil
	}

	return factory(ctx)
}

// IsRuleRegistered checks if a rule with the specified name is registered.
func (r RuleRegistry) IsRuleRegistered(name string) bool {
	_, exists := r.factories[name]

	return exists
}

// GetPriorityService returns the registry's priority service for consistent rule enablement logic.
func (r RuleRegistry) GetPriorityService() RulePriorityService {
	return r.priorityService
}

// GetRuleNames returns a sorted list of all registered rule names.
func (r RuleRegistry) GetRuleNames() []string {
	names := make([]string, 0, len(r.factories))
	for name := range r.factories {
		names = append(names, name)
	}
	// Sort for consistent output
	sort.Strings(names)

	return names
}

// CreateAllRules creates all registered rules.
func (r RuleRegistry) CreateAllRules(ctx context.Context) []Rule {
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

// GetActiveRules returns a list of active rules based on configuration.
// This uses pre-created rules for better performance and consistency.
func (r RuleRegistry) GetActiveRules(ctx context.Context, enabledRules, disabledRules []string) []Rule {
	// If context is nil, we still need to proceed with validation
	// But we shouldn't create a new context from scratch
	if ctx == nil {
		// We explicitly return early here to avoid nil context issues
		// Create rules without filtering
		return r.getRulesAsList()
	}

	// If rules haven't been initialized, return an empty slice
	if len(r.rules) == 0 {
		return []Rule{}
	}

	// Get rules directly using the priority service for filtering
	return r.priorityService.FilterRules(
		ctx,
		r.getRulesAsList(),
		enabledRules,
		disabledRules,
	)
}

// getRulesAsList converts the rules map to a slice.
func (r RuleRegistry) getRulesAsList() []Rule {
	result := make([]Rule, 0, len(r.rules))

	for _, rule := range r.rules {
		if rule != nil {
			result = append(result, rule)
		}
	}

	return result
}

// ValidateRule validates that a rule with the specified name exists.
func (r RuleRegistry) ValidateRule(name string) error {
	if !r.IsRuleRegistered(name) {
		return fmt.Errorf("rule %q not found", name)
	}

	return nil
}

// IsDefaultDisabled checks if a rule is disabled by default.
func (r RuleRegistry) IsDefaultDisabled(rule string) bool {
	// Use the clean rule name for consistent comparison
	cleanName := r.priorityService.CleanRuleName(rule)

	return r.priorityService.DefaultDisabledRules[cleanName]
}

// GetDefaultDisabledRules returns a copy of the default disabled rules map.
func (r RuleRegistry) GetDefaultDisabledRules() map[string]bool {
	return r.priorityService.GetDefaultDisabledRules()
}
