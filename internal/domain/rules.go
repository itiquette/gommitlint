// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package domain

import (
	"strings"

	"github.com/itiquette/gommitlint/internal/domain/config"
)

// CommitRule defines interface for rules that only need commit data.
type CommitRule interface {
	// Name returns the rule's name.
	Name() string

	// Validate performs commit validation.
	Validate(commit Commit, config config.Config) []ValidationError
}

// RepositoryRule defines interface for rules that need repository access.
type RepositoryRule interface {
	// Name returns the rule's name.
	Name() string

	// Validate performs validation with repository access.
	Validate(commit Commit, repo Repository, config config.Config) []ValidationError
}

// ValidateCommitRules validates commit using CommitRule implementations.
func ValidateCommitRules(commit Commit, rules []CommitRule, cfg config.Config) []ValidationError {
	var errors []ValidationError
	for _, rule := range rules {
		errors = append(errors, rule.Validate(commit, cfg)...)
	}

	return errors
}

// ValidateRepositoryRules validates commit using RepositoryRule implementations.
func ValidateRepositoryRules(commit Commit, rules []RepositoryRule, repo Repository, cfg config.Config) []ValidationError {
	var errors []ValidationError
	for _, rule := range rules {
		errors = append(errors, rule.Validate(commit, repo, cfg)...)
	}

	return errors
}

// DefaultDisabledRulesList contains rules that are disabled by default.
// Only rules that require explicit opt-in should be listed here.
var DefaultDisabledRulesList = []string{
	"jirareference", // Organization-specific, requires JIRA setup
	"commitbody",    // Not all projects require detailed commit bodies
	"spell",         // Spell checking requires dictionary setup
}

// IsRuleActive determines if a rule should run based on configuration.
// Priority: 1) Explicit enable wins, 2) Explicit disable, 3) Default disabled, 4) Enabled by default.
func IsRuleActive(ruleName string, enabled, disabled []string) bool {
	cleanName := CleanRuleName(ruleName)

	// Priority 1: Explicitly enabled always runs
	if contains(enabled, cleanName) {
		return true
	}

	// Priority 2: Explicitly disabled never runs
	if contains(disabled, cleanName) {
		return false
	}

	// Priority 3: Default behavior
	return !contains(DefaultDisabledRulesList, cleanName)
}

// contains is a helper to check if a rule is in a list.
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if CleanRuleName(s) == item {
			return true
		}
	}

	return false
}

// CleanRuleName standardizes rule name by removing quotes and whitespace and converting to lowercase for case-insensitive matching.
func CleanRuleName(rule string) string {
	return strings.ToLower(strings.TrimSpace(strings.Trim(rule, "\"'")))
}

// RemoveExplicitlyEnabledFromDisabled creates a new disabled rules list without any rules that are explicitly enabled.
func RemoveExplicitlyEnabledFromDisabled(enabledRules, disabledRules []string) []string {
	if len(enabledRules) == 0 || len(disabledRules) == 0 {
		return disabledRules
	}

	// Create a map of enabled rules for quick lookup
	enabledMap := make(map[string]bool)
	for _, rule := range enabledRules {
		enabledMap[CleanRuleName(rule)] = true
	}

	// Filter out enabled rules from disabled list
	filtered := make([]string, 0, len(disabledRules))

	for _, rule := range disabledRules {
		if !enabledMap[CleanRuleName(rule)] {
			filtered = append(filtered, rule)
		}
	}

	return filtered
}

// MergeEnabledRules merges configuration-provided enabled rules with default enabled rules.
func MergeEnabledRules(defaultRules, configRules []string) []string {
	if len(configRules) == 0 {
		return defaultRules
	}

	// Create a map for faster lookups
	ruleSet := make(map[string]bool)

	// Add all rules to the set
	addToSet := func(rules []string) {
		for _, rule := range rules {
			ruleSet[CleanRuleName(rule)] = true
		}
	}

	addToSet(defaultRules)
	addToSet(configRules)

	// Extract keys from map to slice
	keys := make([]string, 0, len(ruleSet))
	for key := range ruleSet {
		keys = append(keys, key)
	}

	return keys
}
