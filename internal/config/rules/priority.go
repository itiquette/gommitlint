// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package rules

import (
	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/rs/zerolog"
)

// RulePriority determines if a rule should be active based on configuration.
// This function implements the rule priority system:
// 1. Explicitly disabled rules are always disabled - highest priority
// 2. Explicitly enabled rules are enabled (unless explicitly disabled)
// 3. Default-disabled rules are disabled (unless explicitly enabled)
// 4. All other rules are enabled by default.
func RulePriority(ruleName string, enabledRules, disabledRules []string, logger *zerolog.Logger) bool {
	// Clean rule name for comparison
	ruleName = CleanRuleName(ruleName)

	// Create lookup maps for faster filtering
	enabledMap := makeRuleMap(enabledRules)
	disabledMap := makeRuleMap(disabledRules)

	// Core rule priority logic: disabled_rules take precedence over enabled_rules,
	// ensuring users can disable any rule regardless of default settings

	// Apply priority filtering logic with disabled_rules having highest priority
	if disabledMap[ruleName] {
		// Rule is explicitly disabled - exclude it
		if logger != nil {
			logger.Debug().Str("rule", ruleName).Msg("Excluding explicitly disabled rule")
		}

		return false
	}

	if enabledMap[ruleName] {
		// Rule is explicitly enabled - include it
		if logger != nil {
			logger.Debug().Str("rule", ruleName).Msg("Including explicitly enabled rule")
		}

		return true
	}

	domainDefaultDisabled := domain.GetDefaultDisabledRules()
	if domainDefaultDisabled[ruleName] {
		// Rule is disabled by default and not explicitly enabled - exclude it
		if logger != nil {
			logger.Debug().Str("rule", ruleName).Msg("Excluding default-disabled rule")
		}

		return false
	}

	// Rule is enabled by default and not explicitly disabled - include it
	if logger != nil {
		logger.Debug().Str("rule", ruleName).Msg("Including default-enabled rule")
	}

	return true
}

// makeRuleMap creates a map of rule names for faster lookup.
// It handles cleaning rule names by removing quotes and whitespace.
func makeRuleMap(ruleNames []string) map[string]bool {
	result := make(map[string]bool)

	for _, name := range ruleNames {
		// Clean the rule name
		cleanName := CleanRuleName(name)

		// Skip commented lines
		if len(cleanName) > 0 && cleanName[0] != '#' {
			result[cleanName] = true
		}
	}

	return result
}
