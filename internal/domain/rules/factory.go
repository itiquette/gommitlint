// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package rules

import (
	"strings"

	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/itiquette/gommitlint/internal/domain/config"
)

// CreateCommitRules creates commit rules based on configuration.
func CreateCommitRules(cfg config.Config) []domain.CommitRule {
	// Map of rule constructors - explicit, type-safe, no string magic
	ruleConstructors := map[string]func(config.Config) domain.CommitRule{
		"subject":       func(c config.Config) domain.CommitRule { return NewSubjectRule(c) },
		"conventional":  func(c config.Config) domain.CommitRule { return NewConventionalCommitRule(c) },
		"commitbody":    func(c config.Config) domain.CommitRule { return NewCommitBodyRule(c) },
		"jirareference": func(c config.Config) domain.CommitRule { return NewJiraReferenceRule(c) },
		"signoff":       func(c config.Config) domain.CommitRule { return NewSignOffRule(c) },
		"signature":     func(c config.Config) domain.CommitRule { return NewSignatureRule(c) },
		"identity":      func(c config.Config) domain.CommitRule { return NewIdentityRule(c) },
		"spell":         func(c config.Config) domain.CommitRule { return NewSpellRule(c) },
	}

	// Default enabled rules - explicit list, no magic strings scattered
	defaultEnabled := []string{"subject", "conventional", "signoff", "signature", "identity"}

	return buildRules(ruleConstructors, defaultEnabled, cfg)
}

// CreateRepositoryRules creates repository rules based on configuration.
func CreateRepositoryRules(cfg config.Config) []domain.RepositoryRule {
	// Map of rule constructors - type-safe
	ruleConstructors := map[string]func(config.Config) domain.RepositoryRule{
		"branchahead": func(c config.Config) domain.RepositoryRule { return NewBranchAheadRule(c) },
	}

	// Default enabled rules
	defaultEnabled := []string{"branchahead"}

	return buildRepositoryRules(ruleConstructors, defaultEnabled, cfg)
}

// buildRules creates rules based on constructor map and configuration.
func buildRules[T any](constructors map[string]func(config.Config) T, defaultEnabled []string, cfg config.Config) []T {
	var rules []T

	// Determine which rules to create
	enabledRules := determineEnabledRules(defaultEnabled, cfg.Rules)

	// Create only enabled rules
	for _, ruleName := range enabledRules {
		if constructor, exists := constructors[ruleName]; exists {
			rules = append(rules, constructor(cfg))
		}
	}

	return rules
}

// buildRepositoryRules specializes buildRules for repository rules.
func buildRepositoryRules(constructors map[string]func(config.Config) domain.RepositoryRule, defaultEnabled []string, cfg config.Config) []domain.RepositoryRule {
	return buildRules(constructors, defaultEnabled, cfg)
}

// determineEnabledRules applies priority logic to determine which rules should be enabled.
func determineEnabledRules(defaultEnabled []string, rulesConfig config.RulesConfig) []string {
	// Start with explicitly enabled rules (highest priority)
	enabledSet := make(map[string]bool)
	for _, rule := range normalizeRuleNames(rulesConfig.Enabled) {
		enabledSet[rule] = true
	}

	// Remove explicitly disabled rules
	disabledSet := make(map[string]bool)
	for _, rule := range normalizeRuleNames(rulesConfig.Disabled) {
		disabledSet[rule] = true
	}

	// Add default enabled rules if not explicitly disabled
	for _, rule := range defaultEnabled {
		if !disabledSet[rule] {
			enabledSet[rule] = true
		}
	}

	// Convert to slice - pre-allocate for performance
	result := make([]string, 0, len(enabledSet))
	for rule := range enabledSet {
		result = append(result, rule)
	}

	return result
}

// normalizeRuleNames normalizes rule names.
func normalizeRuleNames(names []string) []string {
	result := make([]string, len(names))
	for i, name := range names {
		result[i] = strings.ToLower(strings.TrimSpace(name))
	}

	return result
}
