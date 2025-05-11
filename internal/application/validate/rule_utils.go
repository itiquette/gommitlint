// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package validate

import (
	"fmt"
	"os"
	"strings"

	"github.com/itiquette/gommitlint/internal/config"
	"github.com/itiquette/gommitlint/internal/core/rules"
	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/rs/zerolog"
)

func FilterRules(allRules []domain.Rule, enabledRules, disabledRules []string) []domain.Rule {
	// First, let's check if we're getting any rules at all
	if len(allRules) == 0 {
		// Create all available rules
		allAvailableRules := []domain.Rule{
			rules.NewSubjectLengthRule(),
			rules.NewConventionalCommitRule(),
			rules.NewImperativeVerbRule(),
			rules.NewSubjectCaseRule(),
			rules.NewSubjectSuffixRule(),
			rules.NewSignOffRule(),
			rules.NewSignatureRule(),
			rules.NewSpellRule(),
			rules.NewJiraReferenceRule(),
			rules.NewCommitBodyRule(),
		}

		// Filter the rules based on the enabled/disabled lists
		// This reuses our existing logic for rule filtering
		return FilterRules(allAvailableRules, enabledRules, disabledRules)
	}

	// Debug output to file for better analysis
	f, err := os.OpenFile("debug.txt", os.O_APPEND|os.O_WRONLY, 0644)
	if err == nil {
		defer f.Close()
		fmt.Fprintf(f, "FilterRules called with %d rules, enabled_rules: %v, disabled_rules: %v\n",
			len(allRules), enabledRules, disabledRules)
		
		// Debug all available rule names
		ruleNamesFound := make([]string, 0, len(allRules))
		for _, rule := range allRules {
			ruleNamesFound = append(ruleNamesFound, rule.Name())
		}
		fmt.Fprintf(f, "All rule names found: %v\n", ruleNamesFound)
	}

	// Clean rule names in lists for consistency and handle comments
	cleanEnabledRules := make([]string, 0, len(enabledRules))
	for _, name := range enabledRules {
		// Skip commented lines (lines starting with # in YAML)
		cleanName := strings.TrimSpace(strings.Trim(name, "\"'"))
		if strings.HasPrefix(cleanName, "#") {
			continue
		}
		cleanEnabledRules = append(cleanEnabledRules, cleanName)
	}

	cleanDisabledRules := make([]string, 0, len(disabledRules))
	for _, name := range disabledRules {
		// Skip commented lines (lines starting with # in YAML)
		cleanName := strings.TrimSpace(strings.Trim(name, "\"'"))
		if strings.HasPrefix(cleanName, "#") {
			continue
		}
		cleanDisabledRules = append(cleanDisabledRules, cleanName)
	}

	// Create the enabled rules map - anything in this map will be enabled
	// regardless of whether it's disabled by default
	enabledMap := make(map[string]bool)
	for _, name := range cleanEnabledRules {
		enabledMap[name] = true
		if f != nil {
			fmt.Fprintf(f, "Explicitly enabling rule: %s\n", name)
		}
	}

	// Create the disabled rules map - anything in this map will be disabled
	// unless it's explicitly enabled
	disabledMap := make(map[string]bool)
	for _, name := range cleanDisabledRules {
		disabledMap[name] = true
		if f != nil {
			fmt.Fprintf(f, "Explicitly disabling rule: %s\n", name)
		}
	}

	// Add the DefaultDisabledRules to the disabled map
	for ruleName := range config.DefaultDisabledRules {
		disabledMap[ruleName] = true
		if f != nil {
			fmt.Fprintf(f, "Adding default-disabled rule to disabledMap: %s\n", ruleName)
		}
	}

	// Filter rules based on the maps
	result := make([]domain.Rule, 0, len(allRules))
	
	for _, rule := range allRules {
		ruleName := rule.Name()
		
		// First check if explicitly enabled - this overrides any disabling
		if enabledMap[ruleName] {
			if f != nil {
				fmt.Fprintf(f, "Including explicitly enabled rule: %s\n", ruleName)
			}
			result = append(result, rule)
			continue
		}
		
		// Then check if disabled (either explicitly or by default)
		if disabledMap[ruleName] {
			if f != nil {
				fmt.Fprintf(f, "Skipping disabled rule: %s\n", ruleName)
			}
			continue
		}
		
		// If we get here, the rule is not explicitly enabled or disabled
		// so include it (default behavior for rules not in DefaultDisabledRules)
		if f != nil {
			fmt.Fprintf(f, "Including default-enabled rule: %s\n", ruleName)
		}
		result = append(result, rule)
	}

	// Log the final result
	if f != nil {
		activeRuleNames := make([]string, 0, len(result))
		for _, rule := range result {
			activeRuleNames = append(activeRuleNames, rule.Name())
		}
		fmt.Fprintf(f, "FINAL active rules: %v\n", activeRuleNames)
	}

	return result
}

// logActiveRules logs information about the active rules.
func logActiveRules(logger *zerolog.Logger, rules []domain.Rule) {
	// Extract rule names for logging
	ruleNames := make([]string, len(rules))
	for i, rule := range rules {
		ruleNames[i] = rule.Name()
	}

	// Log the active rules
	logger.Debug().
		Strs("active_rules", ruleNames).
		Int("rule_count", len(rules)).
		Msg("Active validation rules")
}
