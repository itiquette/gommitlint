// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package rules

import (
	"strings"
)

// CleanRuleName standardizes rule name by removing quotes and whitespace.
func CleanRuleName(rule string) string {
	// Remove leading/trailing whitespace
	rule = strings.TrimSpace(rule)

	// Remove both double and single quotes if they're at the start and end
	if (strings.HasPrefix(rule, "\"") && strings.HasSuffix(rule, "\"")) ||
		(strings.HasPrefix(rule, "'") && strings.HasSuffix(rule, "'")) {
		rule = rule[1 : len(rule)-1]
	}

	// Final trim to remove any remaining whitespace
	rule = strings.TrimSpace(rule)

	return rule
}

// IsRuleEnabled determines if a rule should be active based on configuration.
// This function implements the rule priority system by delegating to the centralized RulePriority function.
func IsRuleEnabled(ruleName string, enabled, disabled []string) bool {
	return RulePriority(ruleName, enabled, disabled, nil)
}

// IsExplicitlyEnabled checks if a rule is explicitly enabled in the configuration.
func IsExplicitlyEnabled(ruleName string, enabled []string) bool {
	cleanName := CleanRuleName(ruleName)
	for _, rule := range enabled {
		if CleanRuleName(rule) == cleanName {
			return true
		}
	}

	return false
}

// RemoveExplicitlyEnabledFromDisabled creates a new disabled rules list without
// any rules that are explicitly enabled. This function is generic and works with any rule.
// It implements the rule priority principle where explicitly enabled rules always win.
func RemoveExplicitlyEnabledFromDisabled(enabledRules, disabledRules []string) []string {
	if len(enabledRules) == 0 || len(disabledRules) == 0 {
		return disabledRules
	}

	// Create a map of enabled rules for quick lookup
	enabledMap := make(map[string]bool)
	for _, rule := range enabledRules {
		enabledMap[CleanRuleName(rule)] = true
	}

	// Create a new disabled list, excluding any rules that are explicitly enabled
	newDisabled := make([]string, 0, len(disabledRules))

	for _, rule := range disabledRules {
		cleanRule := CleanRuleName(rule)
		if !enabledMap[cleanRule] {
			// Only include rules that are NOT explicitly enabled
			newDisabled = append(newDisabled, rule)
		}
	}

	return newDisabled
}

// MergeEnabledRules merges configuration-provided enabled rules with default enabled rules.
// This ensures that rules explicitly enabled in config are added to defaults rather than replacing them.
func MergeEnabledRules(defaultRules, configRules []string) []string {
	if len(configRules) == 0 {
		return defaultRules
	}

	// Create a map for faster lookups
	ruleSet := make(map[string]bool)

	// Add all default rules
	for _, rule := range defaultRules {
		ruleSet[CleanRuleName(rule)] = true
	}

	// Add all config rules
	for _, rule := range configRules {
		ruleSet[CleanRuleName(rule)] = true
	}

	// Convert the map back to a slice
	result := make([]string, 0, len(ruleSet))
	for rule := range ruleSet {
		result = append(result, rule)
	}

	return result
}
