// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package cli

import (
	"strings"
)

// containsRule checks if a rule name is in the enabled_rules list
// It handles the cleanup of rule names for proper comparison
func containsRule(rules []string, ruleName string) bool {
	for _, rule := range rules {
		// Clean the rule name by removing quotes and whitespace
		cleanRule := strings.TrimSpace(strings.Trim(rule, "\"'"))
		if cleanRule == ruleName {
			return true
		}
	}

	return false
}
