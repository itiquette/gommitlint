// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package config

import (
	"strings"
)

// DefaultDisabledRules is a list of rule names that are disabled by default.
var DefaultDisabledRules = map[string]bool{
	"JiraReference": true,
	"CommitBody":    true,
}

func IsRuleEnabled(ruleName string, enabled, disabled []string) bool {
	// Clean rule name for comparison
	ruleName = strings.TrimSpace(strings.Trim(ruleName, "\"'"))

	// First check if the rule is explicitly enabled
	for _, rule := range enabled {
		cleanRule := strings.TrimSpace(strings.Trim(rule, "\"'"))
		if cleanRule == ruleName {
			// Rule is explicitly enabled - this overrides disabled status
			return true
		}
	}

	// If not explicitly enabled, check if it's disabled
	for _, rule := range disabled {
		cleanRule := strings.TrimSpace(strings.Trim(rule, "\"'"))
		if cleanRule == ruleName {
			// Rule is explicitly disabled
			return false
		}
	}

	// Check if the rule is in the list of rules disabled by default
	if DefaultDisabledRules[ruleName] {
		return false
	}

	// If not in either list and not a special case, default to enabled
	return true
}
