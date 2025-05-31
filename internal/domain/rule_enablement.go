// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package domain

import (
	"strings"
)

// Rule configuration and utility functions.
// These functions help with determining rule activation status based on configuration.

// IsRuleEnabled determines if a rule should be active based on configuration.
// This delegates to the core implementation following hexagonal architecture.
func IsRuleEnabled(ruleName string, enabled, disabled []string) bool {
	return ShouldRunRule(ruleName, enabled, disabled)
}

// CleanRuleName standardizes rule name by removing quotes and whitespace
// and converting to lowercase for case-insensitive matching.
func CleanRuleName(rule string) string {
	return strings.ToLower(strings.TrimSpace(strings.Trim(rule, "\"'")))
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
// any rules that are explicitly enabled.
func RemoveExplicitlyEnabledFromDisabled(enabledRules, disabledRules []string) []string {
	if len(enabledRules) == 0 || len(disabledRules) == 0 {
		return disabledRules
	}

	// Create a map of enabled rules for quick lookup
	enabledMap := make(map[string]bool)
	for _, rule := range enabledRules {
		enabledMap[CleanRuleName(rule)] = true
	}

	// Create a new disabled list using functional principles - filter out enabled rules
	// Use the core.Filter function from our utility package
	return FilterSliceCompat(disabledRules, func(rule string) bool {
		return !enabledMap[CleanRuleName(rule)]
	})
}

// MergeEnabledRules merges configuration-provided enabled rules with default enabled rules.
func MergeEnabledRules(defaultRules, configRules []string) []string {
	if len(configRules) == 0 {
		return defaultRules
	}

	// Create a map for faster lookups
	ruleSet := make(map[string]bool)

	// Add all rules using a functional approach
	addToSet := func(rules []string) {
		for _, rule := range rules {
			ruleSet[CleanRuleName(rule)] = true
		}
	}

	addToSet(defaultRules)
	addToSet(configRules)

	// Convert the map back to a slice using the slices utility package
	return MapKeys(ruleSet)
}
