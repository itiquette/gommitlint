// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package validation provides core validation components.
package validation

import (
	"github.com/itiquette/gommitlint/internal/config"
	"github.com/itiquette/gommitlint/internal/core/rules"
	"github.com/itiquette/gommitlint/internal/domain"
)

// RuleProvider is a function that creates a rule using the configuration.
type RuleProvider func(config config.Config, analyzer domain.CommitAnalyzer) domain.Rule

// ruleFactory represents a factory for creating a rule with configuration.
type ruleFactory struct {
	provider         RuleProvider
	requiresAnalyzer bool
	condition        func(config config.Config) bool // Function that determines if the rule should be created
}

// standardRuleFactories defines factories for all built-in rules using configuration.
var standardRuleFactories = map[string]ruleFactory{
	// SubjectLength rule with configuration
	"SubjectLength": {
		provider: func(config config.Config, _ domain.CommitAnalyzer) domain.Rule {
			return rules.NewSubjectLengthRule(
				rules.WithMaxLength(config.SubjectMaxLength()),
			)
		},
		requiresAnalyzer: false,
		condition:        func(_ config.Config) bool { return true }, // Always create
	},
	// ConventionalCommit rule with configuration
	"ConventionalCommit": {
		provider: func(config config.Config, _ domain.CommitAnalyzer) domain.Rule {
			options := []rules.ConventionalCommitOption{}

			// Apply the allowed types if provided
			if types := config.ConventionalTypes(); len(types) > 0 {
				options = append(options, rules.WithAllowedTypes(types))
			}

			// Apply the allowed scopes if provided
			if scopes := config.ConventionalScopes(); len(scopes) > 0 {
				options = append(options, rules.WithAllowedScopes(scopes))
			}

			// Apply the max description length if provided
			if maxLength := config.ConventionalMaxDescriptionLength(); maxLength > 0 {
				options = append(options, rules.WithMaxDescLength(maxLength))
			}

			return rules.NewConventionalCommitRule(options...)
		},
		requiresAnalyzer: false,
		condition: func(config config.Config) bool {
			// Only create if conventional validation is required
			return config.ConventionalRequired()
		},
	},
	// ImperativeVerb rule with configuration
	"ImperativeVerb": {
		provider: func(config config.Config, _ domain.CommitAnalyzer) domain.Rule {
			isConventional := config.ConventionalRequired()
			options := []rules.ImperativeVerbOption{}

			if isConventional {
				options = append(options, rules.WithImperativeConventionalCommit(true))
			}

			return rules.NewImperativeVerbRule(
				config.SubjectRequireImperative(),
				options...,
			)
		},
		requiresAnalyzer: false,
		condition: func(config config.Config) bool {
			// Only create if subject imperative is required
			return config.SubjectRequireImperative()
		},
	},
	// SubjectCase rule with configuration
	"SubjectCase": {
		provider: func(config config.Config, _ domain.CommitAnalyzer) domain.Rule {
			options := []rules.SubjectCaseOption{}

			if caseChoice := config.SubjectCase(); caseChoice != "" {
				options = append(options, rules.WithCaseChoice(caseChoice))
			}

			if config.ConventionalRequired() {
				options = append(options, rules.WithSubjectCaseCommitFormat(true))
			}

			if config.SubjectRequireImperative() {
				options = append(options, rules.WithAllowNonAlpha(true))
			}

			return rules.NewSubjectCaseRule(options...)
		},
		requiresAnalyzer: false,
		condition: func(config config.Config) bool {
			// Only create if a specific case is required
			caseChoice := config.SubjectCase()

			return caseChoice == "lower" || caseChoice == "upper"
		},
	},
	// SubjectSuffix rule with configuration
	"SubjectSuffix": {
		provider: func(config config.Config, _ domain.CommitAnalyzer) domain.Rule {
			suffixes := config.SubjectInvalidSuffixes()
			if suffixes == "" {
				return rules.NewSubjectSuffixRule()
			}

			return rules.NewSubjectSuffixRule(
				rules.WithInvalidSuffixes(suffixes),
			)
		},
		requiresAnalyzer: false,
		condition:        func(_ config.Config) bool { return true }, // Always create
	},
	// CommitBody rule with configuration
	"CommitBody": {
		provider: func(config config.Config, _ domain.CommitAnalyzer) domain.Rule {
			return rules.NewCommitBodyRule(
				rules.WithRequireBody(config.BodyRequired()),
				rules.WithAllowSignOffOnly(config.BodyAllowSignOffOnly()),
			)
		},
		requiresAnalyzer: false,
		condition: func(config config.Config) bool {
			// Only create if body is required
			return config.BodyRequired()
		},
	},
	// SignOff rule with configuration
	"SignOff": {
		provider: func(config config.Config, _ domain.CommitAnalyzer) domain.Rule {
			options := []rules.SignOffOption{}

			options = append(options, rules.WithRequireSignOff(config.SignOffRequired()))
			options = append(options, rules.WithAllowMultipleSignOffs(config.AllowMultipleSignOffs()))

			return rules.NewSignOffRule(options...)
		},
		requiresAnalyzer: false,
		condition: func(config config.Config) bool {
			// Only create if sign-off is required
			return config.SignOffRequired()
		},
	},
	// Signature rule with configuration
	"Signature": {
		provider: func(config config.Config, _ domain.CommitAnalyzer) domain.Rule {
			options := []rules.SignatureOption{}

			options = append(options, rules.WithRequireSignature(config.SignatureRequired()))

			if types := config.AllowedSignatureTypes(); len(types) > 0 {
				options = append(options, rules.WithAllowedSignatureTypes(types))
			}

			return rules.NewSignatureRule(options...)
		},
		requiresAnalyzer: false,
		condition: func(config config.Config) bool {
			// Only create if signature is required
			return config.SignatureRequired()
		},
	},
	// JiraReference rule with configuration
	"JiraReference": {
		provider: func(config config.Config, _ domain.CommitAnalyzer) domain.Rule {
			options := []rules.JiraReferenceOption{}

			// Check if conventional commit format is required
			if config.ConventionalRequired() {
				options = append(options, rules.WithConventionalCommit())
			}

			// Check if body reference checking is enabled
			if config.JiraBodyRef() {
				options = append(options, rules.WithBodyRefChecking())
			}

			// Add valid projects if provided
			if projects := config.JiraProjects(); len(projects) > 0 {
				options = append(options, rules.WithValidProjects(projects))
			}

			return rules.NewJiraReferenceRule(options...)
		},
		requiresAnalyzer: false,
		condition: func(config config.Config) bool {
			// Only create if Jira is required
			return config.JiraRequired()
		},
	},
	// Spell rule with configuration
	"Spell": {
		provider: func(config config.Config, _ domain.CommitAnalyzer) domain.Rule {
			options := []rules.SpellRuleOption{}

			if locale := config.SpellLocale(); locale != "" {
				options = append(options, rules.WithLocale(locale))
			}

			if maxErrors := config.SpellMaxErrors(); maxErrors > 0 {
				options = append(options, rules.WithMaxErrors(maxErrors))
			}

			if ignoreWords := config.SpellIgnoreWords(); len(ignoreWords) > 0 {
				options = append(options, rules.WithIgnoreWords(ignoreWords))
			}

			if customWords := config.SpellCustomWords(); len(customWords) > 0 {
				options = append(options, rules.WithCustomWords(customWords))
			}

			return rules.NewSpellRule(options...)
		},
		requiresAnalyzer: false,
		condition: func(config config.Config) bool {
			// Only create if spell checking is enabled
			return config.SpellEnabled()
		},
	},
	// CommitsAhead rule with configuration
	"CommitsAhead": {
		provider: func(config config.Config, analyzer domain.CommitAnalyzer) domain.Rule {
			options := []rules.CommitsAheadOption{}

			options = append(options, rules.WithMaxCommitsAhead(config.MaxCommitsAhead()))

			if analyzer != nil {
				options = append(options, rules.WithRepositoryGetter(func() domain.CommitAnalyzer {
					return analyzer
				}))
			}

			return rules.NewCommitsAheadRule(options...)
		},
		requiresAnalyzer: true,
		condition: func(config config.Config) bool {
			// Only create if commits ahead check is enabled
			return true // Always create CommitsAhead rule
		},
	},
	// SignedIdentity rule with configuration
	"SignedIdentity": {
		provider: func(config config.Config, _ domain.CommitAnalyzer) domain.Rule {
			options := []rules.SignedIdentityOption{}

			if uri := config.IdentityPublicKeyURI(); uri != "" {
				options = append(options, rules.WithKeyDirectory(uri))
			}

			return rules.NewSignedIdentityRule(options...)
		},
		requiresAnalyzer: false,
		condition: func(config config.Config) bool {
			// Only create if a public key URI is provided
			return config.IdentityPublicKeyURI() != ""
		},
	},
	// Additional rule factories will be added here as they're migrated
}

// CreateRuleWithConfig creates a rule using the configuration.
func CreateRuleWithConfig(ruleName string, config config.Config, analyzer domain.CommitAnalyzer) domain.Rule {
	factory, exists := standardRuleFactories[ruleName]
	if !exists {
		return nil
	}

	// Check if the rule should be created based on its condition
	if !factory.condition(config) {
		return nil
	}

	// Create the rule
	if factory.requiresAnalyzer {
		return factory.provider(config, analyzer)
	}

	return factory.provider(config, nil)
}

// GetRuleNames returns the names of all available rules.
func GetRuleNames() []string {
	names := make([]string, 0, len(standardRuleFactories))
	for name := range standardRuleFactories {
		names = append(names, name)
	}

	return names
}
