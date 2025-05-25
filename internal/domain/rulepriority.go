// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package domain

import (
	"context"
	"strings"
)

// RulePriorityService implements the business logic for rule priority management.
// This service provides the canonical implementation of rule priority logic in gommitlint.
//
// The rule priority system follows this order:
// 1. Explicitly enabled rules are always enabled (highest priority)
// 2. Explicitly disabled rules are disabled (unless explicitly enabled)
// 3. Default-disabled rules are disabled (unless explicitly enabled)
// 4. All other rules are enabled by default (lowest priority)
//
// This service provides both instance methods and package-level functions for convenience.
type RulePriorityService struct {
	// DefaultDisabledRules is a map of rules that are disabled by default.
	DefaultDisabledRules map[string]bool
}

// NewRulePriorityService creates a new rule priority service.
func NewRulePriorityService(defaultDisabledRules map[string]bool) RulePriorityService {
	// Make a copy to avoid shared state issues
	disabled := make(map[string]bool, len(defaultDisabledRules))
	for k, v := range defaultDisabledRules {
		disabled[k] = v
	}

	return RulePriorityService{
		DefaultDisabledRules: disabled,
	}
}

// CleanRuleName normalizes a rule name for consistent handling.
// This ensures case-insensitive rule processing.
func (s RulePriorityService) CleanRuleName(ruleName string) string {
	// Convert to lowercase
	name := strings.ToLower(ruleName)

	// Trim whitespace
	name = strings.TrimSpace(name)

	// Remove quotes
	name = strings.Trim(name, "\"'")

	return name
}

// CleanRuleName is a package-level function that provides direct access to the
// rule name normalization functionality. This makes it easier for other packages
// to use the functionality without creating a service instance.
func CleanRuleName(ruleName string) string {
	// Implement the cleaning directly to avoid unnecessary service creation
	name := strings.ToLower(ruleName)
	name = strings.TrimSpace(name)
	name = strings.Trim(name, "\"'")

	return name
}

// IsRuleEnabled determines if a rule should be enabled based on the provided configuration.
// It follows a priority system where explicitly disabled rules are handled first,
// followed by explicitly enabled rules, with default rules considered last.
func (s RulePriorityService) IsRuleEnabled(
	_ context.Context,
	ruleName string,
	enabledRules, disabledRules []string,
) bool {
	cleanRuleName := s.CleanRuleName(ruleName)

	// Create lookup maps for faster filtering
	enabledMap := s.MakeRuleMap(enabledRules)
	disabledMap := s.MakeRuleMap(disabledRules)

	// Core rule priority logic: enabled rules take highest precedence,
	// followed by disabled rules, with default rules having lowest priority

	// Apply priority filtering logic with enabled_rules having highest priority
	if enabledMap[cleanRuleName] {
		// Rule is explicitly enabled - include it regardless of other settings
		return true
	}

	if disabledMap[cleanRuleName] {
		// Rule is explicitly disabled and not explicitly enabled - exclude it
		return false
	}

	if s.DefaultDisabledRules[cleanRuleName] {
		// Rule is disabled by default and not explicitly enabled - exclude it
		return false
	}

	// Rule is not explicitly disabled, not disabled by default,
	// so it's enabled by default - include it
	return true
}

// MakeRuleMap creates a map of rule names for fast lookup.
// Normalizes rule names for consistent handling.
func (s RulePriorityService) MakeRuleMap(rules []string) map[string]bool {
	ruleMap := make(map[string]bool, len(rules))

	for _, rule := range rules {
		// Skip commented out rules (those starting with #)
		if strings.HasPrefix(strings.TrimSpace(rule), "#") {
			continue
		}

		// Normalize each rule name for consistent lookup
		cleanRule := s.CleanRuleName(rule)
		ruleMap[cleanRule] = true
	}

	return ruleMap
}

// GetDefaultDisabledRules returns a copy of the default disabled rules.
func (s RulePriorityService) GetDefaultDisabledRules() map[string]bool {
	// Make a copy to avoid shared state issues
	disabled := make(map[string]bool, len(s.DefaultDisabledRules))
	for k, v := range s.DefaultDisabledRules {
		disabled[k] = v
	}

	return disabled
}

// FilterRuleNames returns a filtered list of rule names based on enabled/disabled status.
func (s RulePriorityService) FilterRuleNames(
	ctx context.Context,
	allRules []string,
	enabledRules, disabledRules []string,
) []string {
	// Map all rules to lowercase for consistent comparison
	normalizedRules := make([]string, len(allRules))
	for i, r := range allRules {
		normalizedRules[i] = s.CleanRuleName(r)
	}

	// Keep only rules that should be enabled
	result := make([]string, 0, len(normalizedRules))

	for _, rule := range normalizedRules {
		if s.IsRuleEnabled(ctx, rule, enabledRules, disabledRules) {
			result = append(result, rule)
		}
	}

	return result
}

// FilterRules returns a filtered list of rules based on enabled/disabled status.
func (s RulePriorityService) FilterRules(
	ctx context.Context,
	allRules []Rule,
	enabledRules, disabledRules []string,
) []Rule {
	// Keep only rules that should be enabled
	result := make([]Rule, 0, len(allRules))

	for _, rule := range allRules {
		if s.IsRuleEnabled(ctx, rule.Name(), enabledRules, disabledRules) {
			result = append(result, rule)
		}
	}

	return result
}

// FilterRuleResults returns a filtered list of rule results based on enabled/disabled status.
func (s RulePriorityService) FilterRuleResults(
	ctx context.Context,
	allRules []RuleResult,
	enabledRules, disabledRules []string,
) []RuleResult {
	// Keep only rules that should be enabled
	result := make([]RuleResult, 0, len(allRules))

	for _, rule := range allRules {
		if s.IsRuleEnabled(ctx, rule.RuleID, enabledRules, disabledRules) {
			result = append(result, rule)
		}
	}

	return result
}

// IsRuleEnabled determines if a rule should be enabled based on configuration settings.
// The function applies the standard rule priority logic: enabled rules have highest
// priority, followed by disabled rules, with default rules having lowest priority.
func IsRuleEnabled(ruleName string, enabledRules, disabledRules []string) bool {
	// Create a service with default settings
	service := NewRulePriorityService(GetDefaultDisabledRules())

	// Use a background context
	ctx := context.Background()

	// Delegate to the service implementation
	return service.IsRuleEnabled(ctx, ruleName, enabledRules, disabledRules)
}

// GetDefaultDisabledRules returns a map of rules that are disabled by default.
// This function provides the canonical list of rules that should be off by default
// and converts them to a map for efficient lookups.
func GetDefaultDisabledRules() map[string]bool {
	// List of rules that are disabled by default to provide a reasonable
	// out-of-the-box experience while allowing users to explicitly enable
	// additional validation when desired.
	defaultDisabledRules := []string{
		"jirareference", // JIRAReference rule is disabled by default as it's organization-specific
		"commitbody",    // CommitBody rule is disabled by default as not all projects require detailed bodies
	}

	// Create a map for O(1) lookups
	disabledMap := make(map[string]bool, len(defaultDisabledRules))

	// Convert all names to lowercase for case-insensitive matching
	for _, rule := range defaultDisabledRules {
		disabledMap[strings.ToLower(rule)] = true
	}

	return disabledMap
}

// CreateRulePriorityService creates a new RulePriorityService with the default configuration.
func CreateRulePriorityService(_ context.Context) RulePriorityService {
	// Create service with default disabled rules - domain layer should not access context utilities
	return NewRulePriorityService(GetDefaultDisabledRules())
}

// WithDefaultDisabledRule adds a rule to the default disabled rules.
// This method is pure and doesn't modify the existing service.
func (s RulePriorityService) WithDefaultDisabledRule(ruleName string, disabled bool) RulePriorityService {
	// Create a copy of the current service
	newService := NewRulePriorityService(s.DefaultDisabledRules)

	// Set the rule in the new service
	cleanName := s.CleanRuleName(ruleName)
	newService.DefaultDisabledRules[cleanName] = disabled

	return newService
}
