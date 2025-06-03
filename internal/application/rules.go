// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package application

import (
	"github.com/itiquette/gommitlint/internal/domain"
	"github.com/itiquette/gommitlint/internal/domain/config"
	"github.com/itiquette/gommitlint/internal/domain/rules"
)

// CreateEnabledRules creates all enabled rules based on configuration.
// This is a pure function that returns a new slice of rules.
func CreateEnabledRules(cfg *config.Config) []domain.Rule {
	var enabledRules []domain.Rule

	// Check each rule and create if enabled
	if domain.ShouldRunRule("subjectlength", cfg.Rules.Enabled, cfg.Rules.Disabled) {
		enabledRules = append(enabledRules, rules.NewSubjectLengthRule(*cfg))
	}

	if domain.ShouldRunRule("subjectcase", cfg.Rules.Enabled, cfg.Rules.Disabled) {
		enabledRules = append(enabledRules, rules.NewSubjectCaseRule(*cfg))
	}

	if domain.ShouldRunRule("subjectsuffix", cfg.Rules.Enabled, cfg.Rules.Disabled) {
		enabledRules = append(enabledRules, rules.NewSubjectSuffixRule(*cfg))
	}

	if domain.ShouldRunRule("imperative", cfg.Rules.Enabled, cfg.Rules.Disabled) {
		enabledRules = append(enabledRules, rules.NewImperativeVerbRule(*cfg))
	}

	if domain.ShouldRunRule("conventional", cfg.Rules.Enabled, cfg.Rules.Disabled) {
		enabledRules = append(enabledRules, rules.NewConventionalCommitRule(*cfg))
	}

	if domain.ShouldRunRule("commitbody", cfg.Rules.Enabled, cfg.Rules.Disabled) {
		enabledRules = append(enabledRules, rules.NewCommitBodyRule(*cfg))
	}

	if domain.ShouldRunRule("jirareference", cfg.Rules.Enabled, cfg.Rules.Disabled) {
		enabledRules = append(enabledRules, rules.NewJiraReferenceRule(*cfg))
	}

	if domain.ShouldRunRule("signoff", cfg.Rules.Enabled, cfg.Rules.Disabled) {
		enabledRules = append(enabledRules, rules.NewSignOffRule(*cfg))
	}

	if domain.ShouldRunRule("signature", cfg.Rules.Enabled, cfg.Rules.Disabled) {
		enabledRules = append(enabledRules, rules.NewSignatureRule(*cfg))
	}

	if domain.ShouldRunRule("identity", cfg.Rules.Enabled, cfg.Rules.Disabled) {
		enabledRules = append(enabledRules, rules.NewIdentityRule(*cfg))
	}

	if domain.ShouldRunRule("spell", cfg.Rules.Enabled, cfg.Rules.Disabled) {
		enabledRules = append(enabledRules, rules.NewSpellRule(*cfg))
	}

	if domain.ShouldRunRule("branchahead", cfg.Rules.Enabled, cfg.Rules.Disabled) {
		enabledRules = append(enabledRules, rules.NewBranchAheadRule(*cfg))
	}

	return enabledRules
}

// CreateAllRules creates all available rules regardless of configuration.
// Useful for listing available rules or testing.
func CreateAllRules(cfg *config.Config) []domain.Rule {
	return []domain.Rule{
		rules.NewSubjectLengthRule(*cfg),
		rules.NewSubjectCaseRule(*cfg),
		rules.NewSubjectSuffixRule(*cfg),
		rules.NewImperativeVerbRule(*cfg),
		rules.NewConventionalCommitRule(*cfg),
		rules.NewCommitBodyRule(*cfg),
		rules.NewJiraReferenceRule(*cfg),
		rules.NewSignOffRule(*cfg),
		rules.NewSignatureRule(*cfg),
		rules.NewIdentityRule(*cfg),
		rules.NewSpellRule(*cfg),
		rules.NewBranchAheadRule(*cfg),
	}
}
