// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package rules

import (
	"strings"

	"github.com/itiquette/gommitlint/internal/adapters/spell"
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
		"spell": func(c config.Config) domain.CommitRule {
			checker := spell.NewMisspellAdapter(c.Spell.Locale)

			return NewSpellRule(checker, c)
		},
	}

	// Default enabled rules - explicit list, no magic strings scattered
	defaultEnabled := []string{"subject", "conventional", "signoff", "signature", "spell"}

	var rules []domain.CommitRule

	// Determine which rules to create
	enabledRules := determineEnabledRules(defaultEnabled, cfg.Rules)

	// Create only enabled rules
	for _, ruleName := range enabledRules {
		if constructor, exists := ruleConstructors[ruleName]; exists {
			rules = append(rules, constructor(cfg))
		}
	}

	return rules
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

// buildRepositoryRules creates repository rules based on constructor map and configuration.
func buildRepositoryRules(constructors map[string]func(config.Config) domain.RepositoryRule, defaultEnabled []string, cfg config.Config) []domain.RepositoryRule {
	var rules []domain.RepositoryRule

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

// determineEnabledRules applies priority logic to determine which rules should be enabled.
func determineEnabledRules(defaultEnabled []string, rulesConfig config.RulesConfig) []string {
	// Start with explicitly enabled rules (highest priority)
	enabledSet := make(map[string]bool)

	for _, name := range rulesConfig.Enabled {
		rule := strings.ToLower(strings.TrimSpace(name))
		enabledSet[rule] = true
	}

	// Remove explicitly disabled rules
	disabledSet := make(map[string]bool)

	for _, name := range rulesConfig.Disabled {
		rule := strings.ToLower(strings.TrimSpace(name))
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
