// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package rules

import (
	"context"

	"github.com/itiquette/gommitlint/internal/common/contextx"
	configRules "github.com/itiquette/gommitlint/internal/config/rules"
)

// IsRuleEnabled checks if a specific rule is enabled based on the current configuration.
// It uses the standard rule priority logic to determine if a rule should be active.
func IsRuleEnabled(ctx context.Context, ruleName string) bool {
	cfg := contextx.GetConfig(ctx)
	if cfg == nil {
		return false
	}

	// Get enabled and disabled rules from configuration
	enabledRules := cfg.GetStringSlice("rules.enabled_rules")
	disabledRules := cfg.GetStringSlice("rules.disabled_rules")

	// Use the standard rule priority function
	return configRules.RulePriority(ruleName, enabledRules, disabledRules, nil)
}
