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

// DefaultRuleProvider provides a default implementation of the RuleProvider interface.
type DefaultRuleProvider struct {
	allRules      []domain.Rule
	activeRules   []domain.Rule
	configuration *RuleConfiguration
}

// NewDefaultRuleProvider creates a new DefaultRuleProvider with the given configuration.
func NewDefaultRuleProvider(config *RuleConfiguration) *DefaultRuleProvider {
	provider := &DefaultRuleProvider{
		configuration: config,
	}

	// Initialize rules
	provider.initializeRules()

	return provider
}

// initializeRules initializes all available rules.
func (p *DefaultRuleProvider) initializeRules() {
	// Initialize all rules
	p.allRules = []domain.Rule{
		// Core rule: Subject Length
		rules.NewSubjectLengthRule(p.configuration.MaxSubjectLength),

		// Core rule: Conventional Commit
		rules.NewConventionalCommitRule(
			p.configuration.ConventionalTypes,
			p.configuration.ConventionalScopes,
			p.configuration.MaxDescLength,
		),

		// Core rule: Imperative Verb
		rules.NewImperativeVerbRule(p.configuration.IsConventionalCommit),

		// Core rule: Jira Reference
		func() domain.Rule {
			var opts []rules.JiraReferenceOption

			// Add options based on configuration
			if p.configuration.IsConventionalCommit {
				opts = append(opts, rules.WithConventionalCommit())
			}

			if p.configuration.JiraBodyRef {
				opts = append(opts, rules.WithBodyRefChecking())
			}

			if len(p.configuration.JiraValidProjects) > 0 {
				opts = append(opts, rules.WithValidProjects(p.configuration.JiraValidProjects))
			}

			return rules.NewJiraReferenceRule(opts...)
		}(),

		// Core rule: Signature
		func() domain.Rule {
			var opts []rules.SignatureOption

			// Add options based on configuration
			opts = append(opts, rules.WithRequireSignature(p.configuration.RequireSignature))

			if len(p.configuration.AllowedSignatureTypes) > 0 {
				opts = append(opts, rules.WithAllowedSignatureTypes(p.configuration.AllowedSignatureTypes))
			}

			return rules.NewSignatureRule(opts...)
		}(),

		// Core rule: SignOff
		func() domain.Rule {
			var opts []rules.SignOffOption

			// Add options based on configuration
			opts = append(opts, rules.WithRequireSignOff(p.configuration.RequireSignOff))
			opts = append(opts, rules.WithAllowMultipleSignOffs(p.configuration.AllowMultipleSignOffs))

			return rules.NewSignOffRule(opts...)
		}(),

		// Core rule: Spell
		func() domain.Rule {
			var opts []rules.SpellRuleOption

			// Add options based on configuration
			if p.configuration.SpellLocale != "" {
				opts = append(opts, rules.WithLocale(p.configuration.SpellLocale))
			}

			if len(p.configuration.SpellIgnoreWords) > 0 {
				opts = append(opts, rules.WithIgnoreWords(p.configuration.SpellIgnoreWords))
			}

			if len(p.configuration.SpellCustomWords) > 0 {
				opts = append(opts, rules.WithCustomWords(p.configuration.SpellCustomWords))
			}

			if p.configuration.SpellMaxErrors > 0 {
				opts = append(opts, rules.WithMaxErrors(p.configuration.SpellMaxErrors))
			}

			return rules.NewSpellRule(opts...)
		}(),

		// Core rule: Subject Case
		func() domain.Rule {
			var opts []rules.SubjectCaseOption

			// Add options based on configuration
			if p.configuration.SubjectCaseChoice != "" {
				opts = append(opts, rules.WithCaseChoice(p.configuration.SubjectCaseChoice))
			}

			if p.configuration.IsConventionalCommit {
				opts = append(opts, rules.WithConventionalCommitCase())
			}

			if p.configuration.SubjectCaseAllowNonAlpha {
				opts = append(opts, rules.WithAllowNonAlphaCase())
			}

			return rules.NewSubjectCaseRule(opts...)
		}(),

		// Core rule: Subject Suffix
		func() domain.Rule {
			var opts []rules.SubjectSuffixOption

			// Add options based on configuration
			if p.configuration.SubjectInvalidSuffixes != "" {
				opts = append(opts, rules.WithInvalidSuffixes(p.configuration.SubjectInvalidSuffixes))
			}

			return rules.NewSubjectSuffixRule(opts...)
		}(),

		// Core rule: Commit Body
		func() domain.Rule {
			var opts []rules.CommitBodyOption

			// Add options based on configuration
			opts = append(opts, rules.WithRequireBody(p.configuration.RequireBody))
			opts = append(opts, rules.WithAllowSignOffOnly(p.configuration.AllowSignOffOnly))

			return rules.NewCommitBodyRule(opts...)
		}(),

		// Core rule: Commits Ahead
		func() domain.Rule {
			var opts []rules.CommitsAheadOption

			// Add options based on configuration
			if p.configuration.Reference != "" {
				opts = append(opts, rules.WithReference(p.configuration.Reference))
			}

			if p.configuration.MaxCommitsAhead > 0 {
				opts = append(opts, rules.WithMaxCommitsAhead(p.configuration.MaxCommitsAhead))
			}

			// Add repository getter
			opts = append(opts, rules.WithRepositoryGetter(func() domain.CommitAnalyzer {
				// This will be injected at runtime by the validation service
				return nil
			}))

			return rules.NewCommitsAheadRule(opts...)
		}(),

		// Add more rules here as they are implemented
	}

	// By default, all rules are active
	p.activeRules = p.allRules

	// Apply enabled/disabled rules configuration
	if len(p.configuration.EnabledRules) > 0 {
		p.SetActiveRules(p.configuration.EnabledRules)
	} else if len(p.configuration.DisabledRules) > 0 {
		p.DisableRules(p.configuration.DisabledRules)
	}
}

// GetRules returns all available validation rules.
func (p *DefaultRuleProvider) GetRules() []domain.Rule {
	return p.allRules
}

// GetActiveRules returns all active validation rules based on configuration.
func (p *DefaultRuleProvider) GetActiveRules() []domain.Rule {
	return p.activeRules
}

// SetActiveRules sets which rules are active based on a list of rule names.
func (p *DefaultRuleProvider) SetActiveRules(ruleNames []string) {
	// If the list is empty, enable all rules
	if len(ruleNames) == 0 {
		p.activeRules = p.allRules

		return
	}

	// Otherwise, only enable the specified rules
	p.activeRules = []domain.Rule{}
	for _, rule := range p.allRules {
		for _, name := range ruleNames {
			if rule.Name() == name {
				p.activeRules = append(p.activeRules, rule)

				break
			}
		}
	}
}

// DisableRules disables specific rules by name.
func (p *DefaultRuleProvider) DisableRules(ruleNames []string) {
	// If no rules to disable, do nothing
	if len(ruleNames) == 0 {
		return
	}

	// Create a new active rule list excluding disabled rules
	activeRules := []domain.Rule{}
	for _, rule := range p.activeRules {
		disabled := false

		for _, name := range ruleNames {
			if rule.Name() == name {
				disabled = true

				break
			}
		}

		if !disabled {
			activeRules = append(activeRules, rule)
		}
	}

	p.activeRules = activeRules
}

// GetRuleByName returns a rule by its name.
func (p *DefaultRuleProvider) GetRuleByName(name string) domain.Rule {
	for _, rule := range p.allRules {
		if rule.Name() == name {
			return rule
		}
	}

	return nil
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
