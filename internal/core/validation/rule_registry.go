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
	configuration interface{}            // Configuration for rules
	analyzer      domain.CommitAnalyzer  // Commit analyzer for rules that need it
}

// NewRuleRegistry creates a new RuleRegistry with the given configuration and analyzer.
// The config should implement all required configuration interfaces.
func NewRuleRegistry(config interface{}, analyzer domain.CommitAnalyzer) *RuleRegistry {
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

// CoreRuleProvider is a function that creates a rule using various config interfaces.
// Each rule only requires the specific config interfaces it needs.
type CoreRuleProvider func(config interface{}, analyzer domain.CommitAnalyzer) domain.Rule

// ruleFactory represents a factory for creating a rule, with a condition for when to create it.
type ruleFactory struct {
	provider         CoreRuleProvider
	requiresAnalyzer bool
	condition        func(interface{}) bool // Function that determines if the rule should be created
}

// Standard rule factories for all built-in rules.
var standardRuleFactories = map[string]ruleFactory{
	"SubjectLength": {
		provider: func(config interface{}, _ domain.CommitAnalyzer) domain.Rule {
			// Only need SubjectConfigProvider
			if subjectConfig, ok := config.(domain.SubjectConfigProvider); ok {
				return rules.NewSubjectLengthRuleWithConfig(subjectConfig)
			}

			return nil
		},
		requiresAnalyzer: false,
		condition:        func(_ interface{}) bool { return true }, // Always create
	},
	"ConventionalCommit": {
		provider: func(config interface{}, _ domain.CommitAnalyzer) domain.Rule {
			// Only need ConventionalConfigProvider
			if convConfig, ok := config.(domain.ConventionalConfigProvider); ok {
				return rules.NewConventionalCommitRuleWithConfig(convConfig)
			}

			return nil
		},
		requiresAnalyzer: false,
		condition:        func(_ interface{}) bool { return true }, // Always create
	},
	"ImperativeVerb": {
		provider: func(config interface{}, _ domain.CommitAnalyzer) domain.Rule {
			// Need SubjectConfigProvider and ConventionalConfigProvider
			if subjectConfig, ok := config.(domain.SubjectConfigProvider); ok {
				if convConfig, ok := config.(domain.ConventionalConfigProvider); ok {
					return rules.NewImperativeVerbRuleWithConfig(subjectConfig, convConfig)
				}
			}

			return nil
		},
		requiresAnalyzer: false,
		condition:        func(_ interface{}) bool { return true }, // Always create
	},
	"JiraReference": {
		provider: func(config interface{}, _ domain.CommitAnalyzer) domain.Rule {
			// Need JiraConfigProvider and ConventionalConfigProvider
			if jiraConfig, ok := config.(domain.JiraConfigProvider); ok {
				if convConfig, ok := config.(domain.ConventionalConfigProvider); ok {
					return rules.NewJiraReferenceRuleWithConfig(jiraConfig, convConfig)
				}
			}

			return nil
		},
		requiresAnalyzer: false,
		condition: func(config interface{}) bool {
			if jiraConfig, ok := config.(domain.JiraConfigProvider); ok {
				return jiraConfig.JiraRequired()
			}

			return false
		},
	},
	"Signature": {
		provider: func(config interface{}, _ domain.CommitAnalyzer) domain.Rule {
			// Need SecurityConfigProvider
			if securityConfig, ok := config.(domain.SecurityConfigProvider); ok {
				return rules.NewSignatureRuleWithConfig(securityConfig)
			}

			return nil
		},
		requiresAnalyzer: false,
		condition:        func(_ interface{}) bool { return true },
	},
	"SignOff": {
		provider: func(config interface{}, _ domain.CommitAnalyzer) domain.Rule {
			// Need SecurityConfigProvider
			if securityConfig, ok := config.(domain.SecurityConfigProvider); ok {
				return rules.NewSignOffRuleWithConfig(securityConfig)
			}

			return nil
		},
		requiresAnalyzer: false,
		condition:        func(_ interface{}) bool { return true },
	},
	"Spell": {
		provider: func(config interface{}, _ domain.CommitAnalyzer) domain.Rule {
			// Need SpellCheckConfigProvider
			if spellConfig, ok := config.(domain.SpellCheckConfigProvider); ok {
				return rules.NewSpellRuleWithConfig(spellConfig)
			}

			return nil
		},
		requiresAnalyzer: false,
		condition:        func(_ interface{}) bool { return true },
	},
	"SubjectCase": {
		provider: func(config interface{}, _ domain.CommitAnalyzer) domain.Rule {
			// Need SubjectConfigProvider and ConventionalConfigProvider
			if subjectConfig, ok := config.(domain.SubjectConfigProvider); ok {
				if convConfig, ok := config.(domain.ConventionalConfigProvider); ok {
					return rules.NewSubjectCaseRuleWithConfig(subjectConfig, convConfig)
				}
			}

			return nil
		},
		requiresAnalyzer: false,
		condition:        func(_ interface{}) bool { return true },
	},
	"SubjectSuffix": {
		provider: func(config interface{}, _ domain.CommitAnalyzer) domain.Rule {
			// Need SubjectConfigProvider
			if subjectConfig, ok := config.(domain.SubjectConfigProvider); ok {
				return rules.NewSubjectSuffixRuleWithConfig(subjectConfig)
			}

			return nil
		},
		requiresAnalyzer: false,
		condition:        func(_ interface{}) bool { return true },
	},
	"CommitBody": {
		provider: func(config interface{}, _ domain.CommitAnalyzer) domain.Rule {
			// Need BodyConfigProvider
			if bodyConfig, ok := config.(domain.BodyConfigProvider); ok {
				return rules.NewCommitBodyRuleWithConfig(bodyConfig)
			}

			return nil
		},
		requiresAnalyzer: false,
		condition:        func(_ interface{}) bool { return true },
	},
	"CommitsAhead": {
		provider: func(config interface{}, analyzer domain.CommitAnalyzer) domain.Rule {
			// Need RepositoryConfigProvider
			if repoConfig, ok := config.(domain.RepositoryConfigProvider); ok {
				return rules.NewCommitsAheadRuleWithConfig(repoConfig, analyzer)
			}

			return nil
		},
		requiresAnalyzer: true,
		condition: func(config interface{}) bool {
			if repoConfig, ok := config.(domain.RepositoryConfigProvider); ok {
				return repoConfig.CheckCommitsAhead()
			}

			return false
		},
	},
}

// registerRules registers all available validation rules.
func (r *RuleRegistry) registerRules() {
	// Register all standard rules based on their factory conditions
	for ruleName, factory := range standardRuleFactories {
		// Check if the rule should be created based on its condition
		if factory.condition(r.configuration) {
			var rule domain.Rule
			if factory.requiresAnalyzer {
				rule = factory.provider(r.configuration, r.analyzer)
			} else {
				rule = factory.provider(r.configuration, nil)
			}

			// Only register the rule if it was successfully created
			if rule != nil {
				r.registerRule(rule)
			} else {
				// Log that we couldn't create the rule due to missing interface implementation
				// In a real implementation, we would use a logger here
				// For now, just skip registering the rule
				// fmt.Printf("Warning: Could not create rule '%s' because the config doesn't implement required interfaces\n", ruleName)
				_ = ruleName // Just to avoid unused variable warning
			}
		}
	}

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
	var ruleConfig domain.RuleConfigProvider

	if _, ok := r.configuration.(domain.RuleConfigProvider); !ok {
		// No rule configuration available, all rules remain active
		return
	}
	// If specific rules are enabled, disable all others
	enabledRules := ruleConfig.EnabledRules()
	if len(enabledRules) > 0 {
		// First, set all rules as inactive
		for name := range r.activeRules {
			r.activeRules[name] = false
		}

		// Then enable only the specified rules
		for _, name := range enabledRules {
			if _, exists := r.rules[name]; exists {
				r.activeRules[name] = true
			}
		}
	} else if len(ruleConfig.DisabledRules()) > 0 {
		// Just disable the specified rules
		for _, name := range ruleConfig.DisabledRules() {
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
// It provides a unified interface for managing validation rules.
type RuleProvider struct {
	registry *RuleRegistry
}

// Ensure RuleProvider implements domain.RuleProvider.
var _ domain.RuleProvider = (*RuleProvider)(nil)

// NewRuleProvider creates a RuleProvider with the given configuration and analyzer.
// This is the main provider for rule creation.
// The config should implement all required configuration interfaces.
func NewRuleProvider(config interface{}, analyzer domain.CommitAnalyzer) *RuleProvider {
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

// GetRuleByName returns a rule by its name.
func (p RuleProvider) GetRuleByName(name string) domain.Rule {
	return p.registry.GetRuleByName(name)
}

// GetAvailableRuleNames returns a list of all available rule names.
// This helps with discovery of supported rules.
func (p RuleProvider) GetAvailableRuleNames() []string {
	names := make([]string, 0, len(standardRuleFactories))
	for name := range standardRuleFactories {
		names = append(names, name)
	}

	return names
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
