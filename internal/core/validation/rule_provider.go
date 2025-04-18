// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package validation provides core validation components.
package validation

import (
	"github.com/itiquette/gommitlint/internal/core/rules"
	"github.com/itiquette/gommitlint/internal/domain"
)

// RuleConfiguration holds configuration for all rules.
type RuleConfiguration struct {
	// Subject configuration
	MaxSubjectLength int

	// Conventional Commit configuration
	ConventionalTypes    []string
	ConventionalScopes   []string
	MaxDescLength        int
	IsConventionalCommit bool

	// Jira configuration
	JiraValidProjects []string
	JiraBodyRef       bool
	JiraRequired      bool

	// Signature configuration
	RequireSignature      bool
	AllowedSignatureTypes []string

	// SignOff configuration
	RequireSignOff        bool
	AllowMultipleSignOffs bool

	// Spell configuration
	SpellLocale      string
	SpellIgnoreWords []string
	SpellCustomWords map[string]string
	SpellMaxErrors   int

	// SubjectCase configuration
	SubjectCaseChoice        string
	SubjectCaseAllowNonAlpha bool

	// SubjectSuffix configuration
	SubjectInvalidSuffixes string

	// CommitBody configuration
	RequireBody      bool
	AllowSignOffOnly bool

	// CommitsAhead configuration
	Reference       string
	MaxCommitsAhead int

	// Enabled/Disabled rules
	EnabledRules  []string
	DisabledRules []string
}

// RuleRegistry provides a map-based registry for validation rules.
type RuleRegistry struct {
	rules         map[string]domain.Rule
	activeRules   map[string]bool
	configuration *RuleConfiguration
}

// NewRuleRegistry creates a new RuleRegistry with the given configuration.
func NewRuleRegistry(config *RuleConfiguration) *RuleRegistry {
	registry := &RuleRegistry{
		rules:         make(map[string]domain.Rule),
		activeRules:   make(map[string]bool),
		configuration: config,
	}

	// Initialize the registry with rules
	registry.registerRules()

	return registry
}

// registerRules registers all available validation rules.
func (r *RuleRegistry) registerRules() {
	// Register Subject Length rule
	r.registerRule(rules.NewSubjectLengthRule(r.configuration.MaxSubjectLength))

	// Register Conventional Commit rule
	r.registerRule(rules.NewConventionalCommitRule(
		rules.WithAllowedTypes(r.configuration.ConventionalTypes),
		rules.WithAllowedScopes(r.configuration.ConventionalScopes),
		rules.WithMaxDescLength(r.configuration.MaxDescLength),
	))

	// Register Imperative Verb rule
	r.registerRule(rules.NewImperativeVerbRule(
		r.configuration.IsConventionalCommit,
		rules.WithImperativeConventionalCommit(r.configuration.IsConventionalCommit),
	))

	// Register Jira Reference rule only if enabled
	if r.configuration.JiraRequired || !contains(r.configuration.DisabledRules, "JiraReference") {
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
	r.registerRule(createCommitsAheadRule(r.configuration))

	// Apply enabled/disabled rules configuration
	r.applyRuleConfiguration()
}

// Helper functions to create rules with options.
func createJiraRule(config *RuleConfiguration) domain.Rule {
	var opts []rules.JiraReferenceOption

	if config.IsConventionalCommit {
		opts = append(opts, rules.WithConventionalCommit())
	}

	if config.JiraBodyRef {
		opts = append(opts, rules.WithBodyRefChecking())
	}

	if len(config.JiraValidProjects) > 0 {
		opts = append(opts, rules.WithValidProjects(config.JiraValidProjects))
	}

	return rules.NewJiraReferenceRule(opts...)
}

func createSignatureRule(config *RuleConfiguration) domain.Rule {
	var opts []rules.SignatureOption

	opts = append(opts, rules.WithRequireSignature(config.RequireSignature))

	if len(config.AllowedSignatureTypes) > 0 {
		opts = append(opts, rules.WithAllowedSignatureTypes(config.AllowedSignatureTypes))
	}

	return rules.NewSignatureRule(opts...)
}

func createSignOffRule(config *RuleConfiguration) domain.Rule {
	var opts []rules.SignOffOption

	opts = append(opts, rules.WithRequireSignOff(config.RequireSignOff))
	opts = append(opts, rules.WithAllowMultipleSignOffs(config.AllowMultipleSignOffs))

	return rules.NewSignOffRule(opts...)
}

func createSpellRule(config *RuleConfiguration) domain.Rule {
	var opts []rules.SpellRuleOption

	if config.SpellLocale != "" {
		opts = append(opts, rules.WithLocale(config.SpellLocale))
	}

	if len(config.SpellIgnoreWords) > 0 {
		opts = append(opts, rules.WithIgnoreWords(config.SpellIgnoreWords))
	}

	if len(config.SpellCustomWords) > 0 {
		opts = append(opts, rules.WithCustomWords(config.SpellCustomWords))
	}

	if config.SpellMaxErrors > 0 {
		opts = append(opts, rules.WithMaxErrors(config.SpellMaxErrors))
	}

	return rules.NewSpellRule(opts...)
}

func createSubjectCaseRule(config *RuleConfiguration) domain.Rule {
	var opts []rules.SubjectCaseOption

	if config.SubjectCaseChoice != "" {
		opts = append(opts, rules.WithCaseChoice(config.SubjectCaseChoice))
	}

	if config.IsConventionalCommit {
		opts = append(opts, rules.WithSubjectCaseCommitFormat(true))
	}

	if config.SubjectCaseAllowNonAlpha {
		opts = append(opts, rules.WithAllowNonAlpha(true))
	}

	return rules.NewSubjectCaseRule(opts...)
}

func createSubjectSuffixRule(config *RuleConfiguration) domain.Rule {
	var opts []rules.SubjectSuffixOption

	if config.SubjectInvalidSuffixes != "" {
		opts = append(opts, rules.WithInvalidSuffixes(config.SubjectInvalidSuffixes))
	}

	return rules.NewSubjectSuffixRule(opts...)
}

func createCommitBodyRule(config *RuleConfiguration) domain.Rule {
	var opts []rules.CommitBodyOption

	opts = append(opts, rules.WithRequireBody(config.RequireBody))
	opts = append(opts, rules.WithAllowSignOffOnly(config.AllowSignOffOnly))

	return rules.NewCommitBodyRule(opts...)
}

func createCommitsAheadRule(config *RuleConfiguration) domain.Rule {
	var opts []rules.CommitsAheadOption

	if config.Reference != "" {
		opts = append(opts, rules.WithReference(config.Reference))
	}

	if config.MaxCommitsAhead > 0 {
		opts = append(opts, rules.WithMaxCommitsAhead(config.MaxCommitsAhead))
	}

	// Add repository getter
	opts = append(opts, rules.WithRepositoryGetter(func() domain.CommitAnalyzer {
		// This will be injected at runtime by the validation service
		return nil
	}))

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
	if len(r.configuration.EnabledRules) > 0 {
		// First, set all rules as inactive
		for name := range r.activeRules {
			r.activeRules[name] = false
		}

		// Then enable only the specified rules
		for _, name := range r.configuration.EnabledRules {
			if _, exists := r.rules[name]; exists {
				r.activeRules[name] = true
			}
		}
	} else if len(r.configuration.DisabledRules) > 0 {
		// Just disable the specified rules
		for _, name := range r.configuration.DisabledRules {
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

// DefaultRuleProvider is a legacy provider that we'll phase out in favor of the RuleRegistry.
type DefaultRuleProvider struct {
	registry *RuleRegistry
}

// NewDefaultRuleProvider creates a new DefaultRuleProvider with the given configuration.
func NewDefaultRuleProvider(config *RuleConfiguration) *DefaultRuleProvider {
	return &DefaultRuleProvider{
		registry: NewRuleRegistry(config),
	}
}

// GetRules returns all available validation rules.
func (p *DefaultRuleProvider) GetRules() []domain.Rule {
	return p.registry.GetRules()
}

// GetActiveRules returns all active validation rules based on configuration.
func (p *DefaultRuleProvider) GetActiveRules() []domain.Rule {
	return p.registry.GetActiveRules()
}

// SetActiveRules sets which rules are active based on a list of rule names.
func (p *DefaultRuleProvider) SetActiveRules(ruleNames []string) {
	p.registry.SetActiveRules(ruleNames)
}

// DisableRules disables specific rules by name.
func (p *DefaultRuleProvider) DisableRules(ruleNames []string) {
	p.registry.DisableRules(ruleNames)
}

// GetRuleByName returns a rule by its name.
func (p *DefaultRuleProvider) GetRuleByName(name string) domain.Rule {
	return p.registry.GetRuleByName(name)
}

// DefaultConfiguration returns a default rule configuration.
func DefaultConfiguration() *RuleConfiguration {
	return &RuleConfiguration{
		// Default subject length configuration
		MaxSubjectLength: 100,

		// Default conventional commit configuration
		ConventionalTypes: []string{
			"feat", "fix", "docs", "style", "refactor",
			"perf", "test", "build", "ci", "chore", "revert",
		},
		ConventionalScopes: []string{}, // Empty means all scopes are allowed
		MaxDescLength:      72,

		// Default Jira configuration
		JiraValidProjects: []string{}, // Empty means all projects are allowed
		JiraBodyRef:       false,      // Default to checking in subject line
		JiraRequired:      false,      // Default to not requiring Jira references

		// Default Signature configuration
		RequireSignature:      true,                   // By default, require signatures
		AllowedSignatureTypes: []string{"gpg", "ssh"}, // Allow both GPG and SSH signatures by default

		// Default SignOff configuration
		RequireSignOff:        true, // By default, require sign-offs
		AllowMultipleSignOffs: true, // By default, allow multiple sign-offs

		// Default Spell configuration
		SpellLocale:      "US",                // Default to US English spelling
		SpellIgnoreWords: []string{},          // No words ignored by default
		SpellCustomWords: map[string]string{}, // No custom words by default
		SpellMaxErrors:   0,                   // No limit on errors reported by default

		// Default SubjectCase configuration
		SubjectCaseChoice:        "lower", // Default to lowercase first letter
		SubjectCaseAllowNonAlpha: false,   // Don't allow non-alpha first characters by default

		// Default SubjectSuffix configuration
		SubjectInvalidSuffixes: ".,;:!?", // Default invalid suffixes

		// Default CommitBody configuration
		RequireBody:      true,  // By default, require a body
		AllowSignOffOnly: false, // By default, don't allow sign-off only bodies

		// Default CommitsAhead configuration
		Reference:       "main", // Default reference branch
		MaxCommitsAhead: 5,      // Default max commits ahead

		// By default, all rules are enabled
		EnabledRules:  []string{},
		DisabledRules: []string{},
	}
}

// contains checks if a string is present in a slice of strings.
func contains(slice []string, value string) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}

	return false
}
