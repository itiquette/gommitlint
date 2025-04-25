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
	r.registerRule(rules.NewSubjectLengthRule(r.configuration.Subject.MaxLength))

	// Register Conventional Commit rule
	r.registerRule(rules.NewConventionalCommitRule(
		rules.WithAllowedTypes(r.configuration.Conventional.Types),
		rules.WithAllowedScopes(r.configuration.Conventional.Scopes),
		rules.WithMaxDescLength(r.configuration.Conventional.MaxDescriptionLength),
	))

	// Register Imperative Verb rule
	r.registerRule(rules.NewImperativeVerbRule(
		r.configuration.Conventional.Required,
		rules.WithImperativeConventionalCommit(r.configuration.Conventional.Required),
	))

	// Register Jira Reference rule only if explicitly enabled
	if r.configuration.Subject.Jira.Required {
		// Create and register the Jira rule
		r.registerRule(createJiraRule(r.configuration))
	}

	// Register Signature rule
	r.registerRule(createSignatureRule(r.configuration))

	// Register SignOff rule
	r.registerRule(createSignOffRule(r.configuration))

	// Register Spell rule
	r.registerRule(createSpellRule(r.configuration))

	// Register Subject Case rule
	r.registerRule(createSubjectCaseRule(r.configuration))

	// Register Subject Suffix rule
	r.registerRule(createSubjectSuffixRule(r.configuration))

	// Register Commit Body rule
	r.registerRule(createCommitBodyRule(r.configuration))

	// Register Commits Ahead rule
	r.registerRule(createCommitsAheadRule(r.configuration, r.analyzer))

	// Apply enabled/disabled rules configuration
	r.applyRuleConfiguration()
}

// Helper functions to create rules with options.
func createJiraRule(config Config) domain.Rule {
	var opts []rules.JiraReferenceOption

	if config.Conventional.Required {
		opts = append(opts, rules.WithConventionalCommit())
	}

	if config.Subject.Jira.BodyRef {
		opts = append(opts, rules.WithBodyRefChecking())
	}

	if len(config.Subject.Jira.Projects) > 0 {
		opts = append(opts, rules.WithValidProjects(config.Subject.Jira.Projects))
	}

	return rules.NewJiraReferenceRule(opts...)
}

func createSignatureRule(config Config) domain.Rule {
	var opts []rules.SignatureOption

	opts = append(opts, rules.WithRequireSignature(config.Security.SignatureRequired))

	if len(config.Security.AllowedSignatureTypes) > 0 {
		opts = append(opts, rules.WithAllowedSignatureTypes(config.Security.AllowedSignatureTypes))
	}

	return rules.NewSignatureRule(opts...)
}

func createSignOffRule(config Config) domain.Rule {
	var opts []rules.SignOffOption

	opts = append(opts, rules.WithRequireSignOff(config.Security.SignOffRequired))
	opts = append(opts, rules.WithAllowMultipleSignOffs(config.Security.AllowMultipleSignOffs))

	return rules.NewSignOffRule(opts...)
}

func createSpellRule(config Config) domain.Rule {
	var opts []rules.SpellRuleOption

	if config.SpellCheck.Locale != "" {
		opts = append(opts, rules.WithLocale(config.SpellCheck.Locale))
	}

	if len(config.SpellCheck.IgnoreWords) > 0 {
		opts = append(opts, rules.WithIgnoreWords(config.SpellCheck.IgnoreWords))
	}

	if len(config.SpellCheck.CustomWords) > 0 {
		opts = append(opts, rules.WithCustomWords(config.SpellCheck.CustomWords))
	}

	if config.SpellCheck.MaxErrors > 0 {
		opts = append(opts, rules.WithMaxErrors(config.SpellCheck.MaxErrors))
	}

	return rules.NewSpellRule(opts...)
}

func createSubjectCaseRule(config Config) domain.Rule {
	var opts []rules.SubjectCaseOption

	if config.Subject.Case != "" {
		opts = append(opts, rules.WithCaseChoice(config.Subject.Case))
	}

	if config.Conventional.Required {
		opts = append(opts, rules.WithSubjectCaseCommitFormat(true))
	}

	if config.Subject.Imperative {
		opts = append(opts, rules.WithAllowNonAlpha(true))
	}

	return rules.NewSubjectCaseRule(opts...)
}

func createSubjectSuffixRule(config Config) domain.Rule {
	var opts []rules.SubjectSuffixOption

	if config.Subject.InvalidSuffixes != "" {
		opts = append(opts, rules.WithInvalidSuffixes(config.Subject.InvalidSuffixes))
	}

	return rules.NewSubjectSuffixRule(opts...)
}

func createCommitBodyRule(config Config) domain.Rule {
	var opts []rules.CommitBodyOption

	opts = append(opts, rules.WithRequireBody(config.Body.Required))
	opts = append(opts, rules.WithAllowSignOffOnly(config.Body.AllowSignOffOnly))

	return rules.NewCommitBodyRule(opts...)
}

func createCommitsAheadRule(config Config, analyzer domain.CommitAnalyzer) domain.Rule {
	var opts []rules.CommitsAheadOption
	if config.Repository.Reference != "" {
		opts = append(opts, rules.WithReference(config.Repository.Reference))
	}

	if config.Repository.MaxCommitsAhead > 0 {
		opts = append(opts, rules.WithMaxCommitsAhead(config.Repository.MaxCommitsAhead))
	}

	// Use the analyzer passed to the function
	if analyzer != nil {
		opts = append(opts, rules.WithRepositoryGetter(func() domain.CommitAnalyzer {
			return analyzer
		}))
	}

	return rules.NewCommitsAheadRule(opts...)
}

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
