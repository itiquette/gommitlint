// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package rules

import (
	"github.com/itiquette/gommitlint/internal/config"
	"github.com/itiquette/gommitlint/internal/domain"
)

// RuleConstructor creates a rule from configuration and dependencies.
type RuleConstructor func(cfg *config.Config, deps domain.RuleDependencies) domain.Rule

// RuleConstructors maps rule names to their constructor functions.
// This provides the exported interface for rule creation.
var RuleConstructors = map[string]RuleConstructor{
	"subjectlength": createSubjectLengthRule,
	"subjectcase":   createSubjectCaseRule,
	"subjectsuffix": createSubjectSuffixRule,
	"imperative":    createImperativeRule,
	"conventional":  createConventionalRule,
	"commitbody":    createCommitBodyRule,
	"jirareference": createJiraReferenceRule,
	"signoff":       createSignOffRule,
	"signature":     createSignatureRule,
	"identity":      createIdentityRule,
	"spell":         createSpellRule,
	"branchahead":   createBranchAheadRule,
}

// CreateEnabledRules creates all rules that should be enabled based on configuration.
func CreateEnabledRules(cfg *config.Config, deps domain.RuleDependencies) []domain.Rule {
	var enabledRules []domain.Rule

	enabledList := cfg.Rules.Enabled
	disabledList := cfg.Rules.Disabled

	for name, constructor := range RuleConstructors {
		if domain.ShouldRunRule(name, enabledList, disabledList) {
			rule := constructor(cfg, deps)
			if rule != nil {
				enabledRules = append(enabledRules, rule)
			}
		}
	}

	return enabledRules
}

// Rule creation functions - each rule knows how to configure itself from config

func createSubjectLengthRule(cfg *config.Config, _ domain.RuleDependencies) domain.Rule {
	return NewSubjectLengthRule(*cfg)
}

func createSubjectCaseRule(cfg *config.Config, _ domain.RuleDependencies) domain.Rule {
	return NewSubjectCaseRule(*cfg)
}

func createSubjectSuffixRule(cfg *config.Config, _ domain.RuleDependencies) domain.Rule {
	return NewSubjectSuffixRule(*cfg)
}

func createImperativeRule(cfg *config.Config, _ domain.RuleDependencies) domain.Rule {
	return NewImperativeVerbRule(*cfg)
}

func createConventionalRule(cfg *config.Config, _ domain.RuleDependencies) domain.Rule {
	return NewConventionalCommitRule(*cfg)
}

func createCommitBodyRule(cfg *config.Config, _ domain.RuleDependencies) domain.Rule {
	return NewCommitBodyRule(*cfg)
}

func createJiraReferenceRule(cfg *config.Config, _ domain.RuleDependencies) domain.Rule {
	return NewJiraReferenceRule(*cfg)
}

func createSignOffRule(cfg *config.Config, _ domain.RuleDependencies) domain.Rule {
	return NewSignOffRule(*cfg)
}

func createSignatureRule(cfg *config.Config, _ domain.RuleDependencies) domain.Rule {
	return NewSignatureRule(*cfg)
}

func createIdentityRule(cfg *config.Config, deps domain.RuleDependencies) domain.Rule {
	return NewIdentityRule(*cfg, deps)
}

func createSpellRule(cfg *config.Config, _ domain.RuleDependencies) domain.Rule {
	return NewSpellRule(*cfg)
}

func createBranchAheadRule(cfg *config.Config, deps domain.RuleDependencies) domain.Rule {
	return NewBranchAheadRule(*cfg, deps)
}
