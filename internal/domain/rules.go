// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package domain

import (
	"slices"
	"strings"

	"github.com/itiquette/gommitlint/internal/domain/config"
)

// Rule defines the interface for all validation rules.
// Rules are pure validators that receive all dependencies and use only what they need.
type Rule interface {
	// Name returns the rule's name.
	Name() string

	// Validate performs validation with explicit dependencies.
	// Rules check for nil repository/config if they need them.
	Validate(commit Commit, repo Repository, config *config.Config) []ValidationError
}

// IsRepositoryLevelRule checks if a rule operates at repository level.
// Currently only BranchAhead is a repository-level rule.
func IsRepositoryLevelRule(rule Rule) bool {
	return rule.Name() == "BranchAhead"
}

// SeparateRules separates rules into commit-level and repository-level rules.
func SeparateRules(rules []Rule) ([]Rule, []Rule) {
	var commitRules []Rule

	var repositoryRules []Rule

	for _, rule := range rules {
		if IsRepositoryLevelRule(rule) {
			repositoryRules = append(repositoryRules, rule)
		} else {
			commitRules = append(commitRules, rule)
		}
	}

	return commitRules, repositoryRules
}

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
