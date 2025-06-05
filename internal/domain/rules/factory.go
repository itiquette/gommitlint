// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package rules

import (
	"slices"
	"strings"

	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/itiquette/gommitlint/internal/domain/config"
)

// CreateEnabledRules creates a slice of enabled rules based on configuration.
// This is the main entry point for rule creation in the functional approach.
func CreateEnabledRules(cfg *config.Config) []domain.Rule {
	var rules []domain.Rule

	// Check each rule and create if enabled
	if shouldEnableRule("subject", cfg) {
		rules = append(rules, NewSubjectRule(*cfg))
	}

	if shouldEnableRule("imperative", cfg) {
		rules = append(rules, NewImperativeVerbRule(*cfg))
	}

	if shouldEnableRule("conventional", cfg) {
		rules = append(rules, NewConventionalCommitRule(*cfg))
	}

	if shouldEnableRule("commitbody", cfg) {
		rules = append(rules, NewCommitBodyRule(*cfg))
	}

	if shouldEnableRule("jirareference", cfg) {
		rules = append(rules, NewJiraReferenceRule(*cfg))
	}

	if shouldEnableRule("signoff", cfg) {
		rules = append(rules, NewSignOffRule(*cfg))
	}

	if shouldEnableRule("signature", cfg) {
		rules = append(rules, NewSignatureRule(*cfg))
	}

	if shouldEnableRule("identity", cfg) {
		rules = append(rules, NewIdentityRule(*cfg))
	}

	if shouldEnableRule("spell", cfg) {
		rules = append(rules, NewSpellRule(*cfg))
	}

	if shouldEnableRule("branchahead", cfg) {
		rules = append(rules, NewBranchAheadRule(*cfg))
	}

	return rules
}

// shouldEnableRule determines if a rule should be enabled based on configuration.
func shouldEnableRule(ruleName string, cfg *config.Config) bool {
	// Normalize rule name
	ruleName = strings.ToLower(strings.TrimSpace(ruleName))

	// Check explicit enable list first
	for _, enabled := range cfg.Rules.Enabled {
		if strings.ToLower(strings.TrimSpace(enabled)) == ruleName {
			return true
		}
	}

	// Check if explicitly disabled
	for _, disabled := range cfg.Rules.Disabled {
		if strings.ToLower(strings.TrimSpace(disabled)) == ruleName {
			return false
		}
	}

	// Check if rule is disabled by default
	defaultDisabled := []string{"jirareference", "commitbody", "spell"}
	if slices.Contains(defaultDisabled, ruleName) {
		return false
	}

	// Default is enabled
	return true
}

// CreateAllRules creates all available rules regardless of configuration.
// Useful for listing available rules or testing.
func CreateAllRules(cfg *config.Config) []domain.Rule {
	return []domain.Rule{
		NewSubjectRule(*cfg),
		NewImperativeVerbRule(*cfg),
		NewConventionalCommitRule(*cfg),
		NewCommitBodyRule(*cfg),
		NewJiraReferenceRule(*cfg),
		NewSignOffRule(*cfg),
		NewSignatureRule(*cfg),
		NewIdentityRule(*cfg),
		NewSpellRule(*cfg),
		NewBranchAheadRule(*cfg),
	}
}
