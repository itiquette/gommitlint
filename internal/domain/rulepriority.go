// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package domain

import (
	"context"
	"strings"

	"github.com/itiquette/gommitlint/internal/common/contextx"
)

// DefaultDisabledRuleMap defines rules that are disabled by default unless explicitly enabled.
type DefaultDisabledRuleMap map[string]bool

// Standard default disabled rules that are used throughout the application.
var standardDefaultDisabledRules = DefaultDisabledRuleMap{
	"JiraReference":  true,
	"SignedIdentity": true,
	"CommitBody":     true,
}

// GetDefaultDisabledRules returns a copy of the standard default disabled rules map.
// This ensures consistent default rule behavior across the application.
func GetDefaultDisabledRules() DefaultDisabledRuleMap {
	result := make(DefaultDisabledRuleMap, len(standardDefaultDisabledRules))
	for k, v := range standardDefaultDisabledRules {
		result[k] = v
	}

	return result
}

// RulePriorityService provides a centralized way to determine rule activation status
// based on configuration. It implements the priority logic for rule enablement.
type RulePriorityService struct {
	// DefaultDisabledRules contains rules that are disabled by default
	DefaultDisabledRules DefaultDisabledRuleMap
}

// NewRulePriorityService creates a new centralized rule priority service.
func NewRulePriorityService(defaultDisabledRules DefaultDisabledRuleMap) *RulePriorityService {
	return &RulePriorityService{
		DefaultDisabledRules: defaultDisabledRules,
	}
}

// CleanRuleName standardizes rule name by removing quotes and whitespace.
func (s *RulePriorityService) CleanRuleName(ruleName string) string {
	// Remove leading/trailing whitespace
	ruleName = strings.TrimSpace(ruleName)

	// Remove both double and single quotes if they're at the start and end
	if (strings.HasPrefix(ruleName, "\"") && strings.HasSuffix(ruleName, "\"")) ||
		(strings.HasPrefix(ruleName, "'") && strings.HasSuffix(ruleName, "'")) {
		ruleName = ruleName[1 : len(ruleName)-1]
	}

	// Final trim to remove any remaining whitespace
	ruleName = strings.TrimSpace(ruleName)

	return ruleName
}

// MakeRuleMap creates a cleaned rule name map for faster lookups.
func (s *RulePriorityService) MakeRuleMap(ruleNames []string) map[string]bool {
	result := make(map[string]bool)

	for _, name := range ruleNames {
		// Clean the rule name
		cleanName := s.CleanRuleName(name)

		// Skip commented lines
		if len(cleanName) > 0 && cleanName[0] != '#' {
			result[cleanName] = true
		}
	}

	return result
}

// IsRuleEnabled determines if a rule should be active based on configuration.
// This is the core priority logic implementation:
// 1. Explicitly disabled rules are always disabled - highest priority
// 2. Explicitly enabled rules are enabled (unless explicitly disabled)
// 3. Default-disabled rules are disabled (unless explicitly enabled)
// 4. All other rules are enabled by default.
func (s *RulePriorityService) IsRuleEnabled(
	ctx context.Context,
	ruleName string,
	enabledRules, disabledRules []string,
) bool {
	// Log entry at trace level
	logger := contextx.GetLogger(ctx)
	logger.Debug("Entering RulePriorityService.IsRuleEnabled",
		"rule", ruleName,
		"enabled", enabledRules,
		"disabled", disabledRules)

	cleanRuleName := s.CleanRuleName(ruleName)

	// Create lookup maps for faster filtering
	enabledMap := s.MakeRuleMap(enabledRules)
	disabledMap := s.MakeRuleMap(disabledRules)

	// Core rule priority logic: disabled take precedence over enabled,
	// ensuring users can disable any rule regardless of default settings

	// Apply priority filtering logic with disabled_rules having highest priority
	if disabledMap[cleanRuleName] {
		// Rule is explicitly disabled - exclude it
		logger.Debug("Excluding explicitly disabled rule", "rule", cleanRuleName)

		return false
	}

	if enabledMap[cleanRuleName] {
		// Rule is explicitly enabled - include it
		logger.Debug("Including explicitly enabled rule", "rule", cleanRuleName)

		return true
	}

	if s.DefaultDisabledRules[cleanRuleName] {
		// Rule is disabled by default and not explicitly enabled - exclude it
		logger.Debug("Excluding default-disabled rule", "rule", cleanRuleName)

		return false
	}

	// Rule is enabled by default and not explicitly disabled - include it
	logger.Debug("Including default-enabled rule", "rule", cleanRuleName)

	return true
}

// IsExplicitlyEnabled checks if a rule is explicitly enabled in the configuration.
func (s *RulePriorityService) IsExplicitlyEnabled(ruleName string, enabledRules []string) bool {
	cleanName := s.CleanRuleName(ruleName)
	enabledMap := s.MakeRuleMap(enabledRules)

	return enabledMap[cleanName]
}

// FilterRules applies rule priority logic to filter a slice of rules based on configuration.
func (s *RulePriorityService) FilterRules(
	ctx context.Context,
	rules []Rule,
	enabledRules, disabledRules []string,
) []Rule {
	// Log entry at trace level
	logger := contextx.GetLogger(ctx)
	logger.Debug("Entering RulePriorityService.FilterRules",
		"rule_count", len(rules),
		"enabled", enabledRules,
		"disabled", disabledRules)

	logger.Debug("Filtering rules with configuration",
		"enabled", enabledRules,
		"disabled", disabledRules)

	// Filter the rules based on configuration
	filtered := make([]Rule, 0, len(rules))

	for _, rule := range rules {
		ruleName := rule.Name()

		// Use the centralized rule priority function
		if s.IsRuleEnabled(ctx, ruleName, enabledRules, disabledRules) {
			filtered = append(filtered, rule)
		}
	}

	logger.Debug("Rule filtering complete",
		"total_rules", len(rules),
		"filtered_rules", len(filtered))

	return filtered
}

// FilterRuleResults filters rule results based on configuration.
func (s *RulePriorityService) FilterRuleResults(
	ctx context.Context,
	results []RuleResult,
	enabledRules, disabledRules []string,
) []RuleResult {
	// Log entry at trace level
	logger := contextx.GetLogger(ctx)
	logger.Debug("Entering RulePriorityService.FilterRuleResults",
		"result_count", len(results),
		"enabled", enabledRules,
		"disabled", disabledRules)

	logger.Debug("Filtering rule results with configuration",
		"enabled", enabledRules,
		"disabled", disabledRules)

	// Filter the rule results
	filtered := make([]RuleResult, 0, len(results))

	for _, result := range results {
		ruleName := result.RuleName

		// Skip results that have been explicitly skipped
		if result.Status == StatusSkipped {
			logger.Debug("Skipping rule result with status=skipped", "rule", ruleName)

			continue
		}

		// Use the centralized rule priority function
		if s.IsRuleEnabled(ctx, ruleName, enabledRules, disabledRules) {
			filtered = append(filtered, result)
		}
	}

	logger.Debug("Rule result filtering complete",
		"total_results", len(results),
		"filtered_results", len(filtered))

	return filtered
}
