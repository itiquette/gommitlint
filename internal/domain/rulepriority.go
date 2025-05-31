// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package domain

import (
	"slices"
	"strings"
)

// DefaultDisabledRulesList contains rules that are disabled by default.
// Only rules that require explicit opt-in should be listed here.
var DefaultDisabledRulesList = []string{
	"jirareference", // Organization-specific, requires JIRA setup
	"commitbody",    // Not all projects require detailed commit bodies
	"spell",         // Spell checking requires dictionary setup
}

// ShouldRunRule determines if a rule should be enabled based on configuration.
// Priority: 1) Explicit enable wins, 2) Explicit disable, 3) Default disabled, 4) Enabled by default.
func ShouldRunRule(ruleName string, enabledRules, disabledRules []string) bool {
	// Normalize rule name (lowercase, trimmed)
	ruleName = strings.ToLower(strings.TrimSpace(ruleName))

	// Normalize all rule lists
	normalizedEnabled := normalizeRuleList(enabledRules)
	normalizedDisabled := normalizeRuleList(disabledRules)

	// Explicit enable always wins
	if slices.Contains(normalizedEnabled, ruleName) {
		return true
	}

	// Explicit disable
	if slices.Contains(normalizedDisabled, ruleName) {
		return false
	}

	// Default disabled rules
	if slices.Contains(DefaultDisabledRulesList, ruleName) {
		return false
	}

	// Default is enabled
	return true
}

// normalizeRuleList normalizes a list of rule names for consistent comparison.
func normalizeRuleList(rules []string) []string {
	normalized := make([]string, 0, len(rules))

	for _, rule := range rules {
		// Skip empty or commented rules
		rule = strings.TrimSpace(rule)
		if rule == "" || strings.HasPrefix(rule, "#") {
			continue
		}
		// Normalize: lowercase, trim quotes
		rule = strings.ToLower(rule)
		rule = strings.Trim(rule, "\"'")
		normalized = append(normalized, rule)
	}

	return normalized
}

// FilterEnabledRules returns only the rules that should be enabled.
func FilterEnabledRules(allRules []Rule, enabledRules, disabledRules []string) []Rule {
	filtered := make([]Rule, 0, len(allRules))

	for _, rule := range allRules {
		if ShouldRunRule(rule.Name(), enabledRules, disabledRules) {
			filtered = append(filtered, rule)
		}
	}

	return filtered
}
