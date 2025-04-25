// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package validation provides core validation components.
package validation

import (
	"github.com/itiquette/gommitlint/internal/core/rules"
	"github.com/itiquette/gommitlint/internal/domain"
)

// RuleRegistry provides a registry for validation rules.
// This registry is a central repository of all available rules and their state.
type RuleRegistry struct {
	rules         map[string]domain.Rule // All registered rules
	activeRules   map[string]bool        // Which rules are active
	configuration Config                 // Configuration for rules
	analyzer      domain.CommitAnalyzer  // Commit analyzer for rules that need it
}

// NewRuleRegistry creates a new RuleRegistry with the given configuration and analyzer.
func NewRuleRegistry(config Config, analyzer domain.CommitAnalyzer) *RuleRegistry {
	registry := &RuleRegistry{
		rules:         make(map[string]domain.Rule),
		activeRules:   make(map[string]bool),
		configuration: config,
		analyzer:      analyzer,
	}
	// Initialize the registry with rules
	registry.registerRules()

	return registry
}

// registerRules registers all available validation rules.
func (r *RuleRegistry) registerRules() {
	// Register Subject Length rule
	r.registerRule(rules.NewSubjectLengthRuleWithConfig(r.configuration))

	// Register Conventional Commit rule
	r.registerRule(rules.NewConventionalCommitRuleWithConfig(r.configuration))

	// Register Imperative Verb rule
	r.registerRule(rules.NewImperativeVerbRuleWithConfig(r.configuration, r.configuration))

	// Register Jira Reference rule only if explicitly enabled
	if r.configuration.Subject.Jira.Required {
		// Create and register the Jira rule using domain-based interface
		r.registerRule(rules.NewJiraReferenceRuleWithConfig(r.configuration, r.configuration))
	}

	// Register Signature rule
	r.registerRule(rules.NewSignatureRuleWithConfig(r.configuration))

	// Register SignOff rule
	r.registerRule(rules.NewSignOffRuleWithConfig(r.configuration))

	// Register Spell rule
	r.registerRule(rules.NewSpellRuleWithConfig(r.configuration))

	// Register Subject Case rule
	r.registerRule(rules.NewSubjectCaseRuleWithConfig(r.configuration, r.configuration))

	// Register Subject Suffix rule
	r.registerRule(rules.NewSubjectSuffixRuleWithConfig(r.configuration))

	// Register Commit Body rule
	r.registerRule(rules.NewCommitBodyRuleWithConfig(r.configuration))

	// Register Commits Ahead rule
	r.registerRule(rules.NewCommitsAheadRuleWithConfig(r.configuration, r.analyzer))

	// Apply enabled/disabled rules configuration
	r.applyRuleConfiguration()
}

// Helper functions to create rules with options.
// createJiraRule has been replaced by NewJiraReferenceRuleWithConfig

// createSignatureRule has been replaced by NewSignatureRuleWithConfig

// createSignOffRule has been replaced by NewSignOffRuleWithConfig

// createSpellRule has been replaced by NewSpellRuleWithConfig

// createSubjectCaseRule has been replaced by NewSubjectCaseRuleWithConfig

// createSubjectSuffixRule has been replaced by NewSubjectSuffixRuleWithConfig

// No longer needed - using the domain-oriented constructor directly

// createCommitsAheadRule has been replaced by NewCommitsAheadRuleWithConfig

// registerRule adds a rule to the registry and marks it as active by default.
func (r *RuleRegistry) registerRule(rule domain.Rule) {
	name := rule.Name()
	r.rules[name] = rule
	r.activeRules[name] = true
}

// applyRuleConfiguration applies the enabled/disabled rules settings.
func (r *RuleRegistry) applyRuleConfiguration() {
	// If specific rules are enabled, disable all others
	if len(r.configuration.Rules.EnabledRules) > 0 {
		// First, set all rules as inactive
		for name := range r.activeRules {
			r.activeRules[name] = false
		}

		// Then enable only the specified rules
		for _, name := range r.configuration.Rules.EnabledRules {
			if _, exists := r.rules[name]; exists {
				r.activeRules[name] = true
			}
		}
	} else if len(r.configuration.Rules.DisabledRules) > 0 {
		// Just disable the specified rules
		for _, name := range r.configuration.Rules.DisabledRules {
			r.activeRules[name] = false
		}
	}
}

// GetRules returns all available validation rules.
func (r *RuleRegistry) GetRules() []domain.Rule {
	rules := make([]domain.Rule, 0, len(r.rules))
	for _, rule := range r.rules {
		rules = append(rules, rule)
	}

	return rules
}

// GetActiveRules returns all active validation rules.
func (r *RuleRegistry) GetActiveRules() []domain.Rule {
	activeRules := make([]domain.Rule, 0)
	for name, active := range r.activeRules {
		if active {
			activeRules = append(activeRules, r.rules[name])
		}
	}

	return activeRules
}

// ReplaceRule replaces an existing rule with a new implementation.
func (r *RuleRegistry) ReplaceRule(name string, rule domain.Rule) {
	// Check if the rule exists and has the expected name
	if _, exists := r.rules[name]; exists {
		// Store the current active state
		wasActive := r.activeRules[name]

		// Replace the rule
		r.rules[name] = rule

		// Maintain the active state
		r.activeRules[name] = wasActive
	}
}

// SetActiveRules sets which rules are active based on a list of rule names.
func (r *RuleRegistry) SetActiveRules(ruleNames []string) {
	// If the list is empty, enable all rules
	if len(ruleNames) == 0 {
		for name := range r.rules {
			r.activeRules[name] = true
		}

		return
	}

	// First, disable all rules
	for name := range r.activeRules {
		r.activeRules[name] = false
	}

	// Then enable only the specified rules
	for _, name := range ruleNames {
		if _, exists := r.rules[name]; exists {
			r.activeRules[name] = true
		}
	}
}

// DisableRules disables specific rules by name.
func (r *RuleRegistry) DisableRules(ruleNames []string) {
	// If no rules to disable, do nothing
	if len(ruleNames) == 0 {
		return
	}

	// Disable the specified rules
	for _, name := range ruleNames {
		if _, exists := r.rules[name]; exists {
			r.activeRules[name] = false
		}
	}
}

// GetRuleByName returns a rule by its name.
func (r *RuleRegistry) GetRuleByName(name string) domain.Rule {
	return r.rules[name]
}

// RuleProvider implements the domain.RuleProvider interface.
type RuleProvider struct {
	registry *RuleRegistry
}

// NewRuleProvider creates a RuleProvider with the given configuration and analyzer.
// This is the main provider for rule creation.
func NewRuleProvider(config Config, analyzer domain.CommitAnalyzer) *RuleProvider {
	return &RuleProvider{
		registry: NewRuleRegistry(config, analyzer),
	}
}

// GetRules returns all available validation rules.
func (p *RuleProvider) GetRules() []domain.Rule {
	return p.registry.GetRules()
}

// GetActiveRules returns all active validation rules based on configuration.
func (p *RuleProvider) GetActiveRules() []domain.Rule {
	return p.registry.GetActiveRules()
}

// ReplaceRule replaces a rule with a new implementation.
func (p *RuleProvider) ReplaceRule(name string, rule domain.Rule) {
	p.registry.ReplaceRule(name, rule)
}

// SetActiveRules sets which rules are active based on a list of rule names.
func (p *RuleProvider) SetActiveRules(ruleNames []string) {
	p.registry.SetActiveRules(ruleNames)
}

// DisableRules disables specific rules by name.
func (p *RuleProvider) DisableRules(ruleNames []string) {
	p.registry.DisableRules(ruleNames)
}

// GetRuleByName returns a rule by its name.
func (p RuleProvider) GetRuleByName(name string) domain.Rule {
	return p.registry.GetRuleByName(name)
}
