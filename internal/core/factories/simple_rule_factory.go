// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package factories

import (
	"context"
	"strings"

	"github.com/itiquette/gommitlint/internal/config/types"
	"github.com/itiquette/gommitlint/internal/core/rules"
	"github.com/itiquette/gommitlint/internal/domain"
)

// SimpleRuleFactory creates rules with minimal configuration.
type SimpleRuleFactory struct {
	cryptoVerifier   domain.CryptoVerifier
	cryptoRepository domain.CryptoKeyRepository
	config           *types.Config
	priorityService  domain.RulePriorityService
}

// NewSimpleRuleFactory creates a new simple rule factory.
func NewSimpleRuleFactory() SimpleRuleFactory {
	return SimpleRuleFactory{
		priorityService: domain.NewRulePriorityService(domain.GetDefaultDisabledRules()),
	}
}

// WithConfig returns a new factory with the specified configuration.
func (f SimpleRuleFactory) WithConfig(config *types.Config) SimpleRuleFactory {
	newFactory := f
	newFactory.config = config

	return newFactory
}

// Validate ensures all required dependencies are available for the rules this factory will create.
func (f SimpleRuleFactory) Validate() error {
	// For now, crypto dependencies are optional
	// In the future, we could make them required for certain rule sets
	return nil
}

// WithCryptoVerifier returns a new factory with the specified crypto verifier.
func (f SimpleRuleFactory) WithCryptoVerifier(verifier domain.CryptoVerifier) SimpleRuleFactory {
	newFactory := f
	newFactory.cryptoVerifier = verifier

	return newFactory
}

// WithCryptoRepository returns a new factory with the specified crypto repository.
func (f SimpleRuleFactory) WithCryptoRepository(repository domain.CryptoKeyRepository) SimpleRuleFactory {
	newFactory := f
	newFactory.cryptoRepository = repository

	return newFactory
}

// WithPriorityService returns a new factory with the specified priority service.
func (f SimpleRuleFactory) WithPriorityService(service domain.RulePriorityService) SimpleRuleFactory {
	newFactory := f
	newFactory.priorityService = service

	return newFactory
}

// createSubjectLengthRule creates a SubjectLength rule with configuration.
func (f SimpleRuleFactory) createSubjectLengthRule() domain.Rule {
	options := []rules.SubjectLengthOption{}

	if f.config != nil {
		maxLength := f.config.Message.Subject.MaxLength
		if maxLength > 0 {
			options = append(options, rules.WithMaxLength(maxLength))
		}
	}

	return rules.NewSubjectLengthRule(options...)
}

// createSubjectCaseRule creates a SubjectCase rule with configuration.
func (f SimpleRuleFactory) createSubjectCaseRule(ctx context.Context) domain.Rule {
	options := []rules.SubjectCaseOption{}

	if f.config != nil {
		// Get case style configuration
		if caseStyle := f.config.Message.Subject.Case; caseStyle != "" {
			options = append(options, rules.WithCaseChoice(caseStyle))
		}

		// Check if imperative is required - affects allowNonAlpha
		if f.config.Message.Subject.RequireImperative {
			options = append(options, rules.WithAllowNonAlpha(true))
		}

		// Check if conventional commit is enabled
		priorityService := domain.NewRulePriorityService(domain.GetDefaultDisabledRules())
		isConventionalEnabled := priorityService.IsRuleEnabled(
			ctx,
			"Conventional",
			f.config.Rules.Enabled,
			f.config.Rules.Disabled,
		)
		options = append(options, rules.WithSubjectCaseCommitFormat(isConventionalEnabled))
	}

	return rules.NewSubjectCaseRule(options...)
}

// createSubjectSuffixRule creates a SubjectSuffix rule with configuration.
func (f SimpleRuleFactory) createSubjectSuffixRule() domain.Rule {
	options := []rules.SubjectSuffixOption{}

	if f.config != nil {
		// Get disallowed suffixes from config
		disallowedSuffixes := f.config.Message.Subject.ForbidEndings
		if len(disallowedSuffixes) > 0 {
			// Build the string from the slice
			var sb strings.Builder
			for _, suffix := range disallowedSuffixes {
				sb.WriteString(suffix)
			}

			options = append(options, rules.WithInvalidSuffixes(sb.String()))
		}
	}

	return rules.NewSubjectSuffixRule(options...)
}

// createConventionalCommitRule creates a ConventionalCommit rule with configuration.
func (f SimpleRuleFactory) createConventionalCommitRule() domain.Rule {
	options := []rules.ConventionalCommitOption{}

	if f.config != nil {
		if len(f.config.Conventional.Types) > 0 {
			options = append(options, rules.WithAllowedTypes(f.config.Conventional.Types))
		}

		if len(f.config.Conventional.Scopes) > 0 {
			options = append(options, rules.WithAllowedScopes(f.config.Conventional.Scopes))
		}

		if f.config.Conventional.RequireScope {
			options = append(options, rules.WithRequireScope(true))
		}

		if maxDescLen := f.config.Conventional.MaxDescriptionLength; maxDescLen > 0 {
			options = append(options, rules.WithMaxDescLength(maxDescLen))
		} else if maxLen := f.config.Message.Subject.MaxLength; maxLen > 0 {
			// Fall back to subject max length if not specified
			options = append(options, rules.WithMaxDescLength(maxLen))
		}
	}

	return rules.NewConventionalCommitRule(options...)
}

// createImperativeVerbRule creates an ImperativeVerb rule with configuration.
func (f SimpleRuleFactory) createImperativeVerbRule(ctx context.Context) domain.Rule {
	options := []rules.ImperativeVerbOption{}

	if f.config != nil {
		// Check if conventional commit is enabled
		priorityService := domain.NewRulePriorityService(domain.GetDefaultDisabledRules())

		isConventionalEnabled := priorityService.IsRuleEnabled(
			ctx,
			"Conventional",
			f.config.Rules.Enabled,
			f.config.Rules.Disabled,
		)
		if isConventionalEnabled {
			options = append(options, rules.WithImperativeConventionalCommit(true))
		}
	}

	return rules.NewImperativeVerbRule(options...)
}

// createCommitBodyRule creates a CommitBody rule with configuration.
func (f SimpleRuleFactory) createCommitBodyRule() domain.Rule {
	options := []rules.CommitBodyOption{}

	if f.config != nil {
		if minLength := f.config.Message.Body.MinLength; minLength > 0 {
			options = append(options, rules.WithMinLength(minLength))
		}

		if minLines := f.config.Message.Body.MinLines; minLines > 0 {
			options = append(options, rules.WithMinLines(minLines))
		}
		// AllowSignoffOnly is a boolean, always set it
		options = append(options, rules.WithAllowSignOffOnly(f.config.Message.Body.AllowSignoffOnly))
	}

	return rules.NewCommitBodyRule(options...)
}

// createJiraReferenceRule creates a JiraReference rule with configuration.
func (f SimpleRuleFactory) createJiraReferenceRule(ctx context.Context) domain.Rule {
	options := []rules.JiraReferenceOption{}

	if f.config != nil {
		if pattern := f.config.Jira.Pattern; pattern != "" {
			options = append(options, rules.WithJiraPattern(pattern))
		}

		if len(f.config.Jira.Projects) > 0 {
			options = append(options, rules.WithJiraPrefixes(f.config.Jira.Projects))
		}

		if f.config.Jira.CheckBody {
			options = append(options, rules.WithJiraBodySearch(true))
		}

		// Check if conventional commit is enabled
		isConventionalEnabled := f.priorityService.IsRuleEnabled(
			ctx,
			"Conventional",
			f.config.Rules.Enabled,
			f.config.Rules.Disabled,
		)
		if isConventionalEnabled {
			options = append(options, rules.WithConventionalCommit())
		}
	}

	return rules.NewJiraReferenceRule(options...)
}

// createSignOffRule creates a SignOff rule with configuration.
func (f SimpleRuleFactory) createSignOffRule() domain.Rule {
	options := []rules.SignOffOption{}

	if f.config != nil {
		// Apply configuration from body and signing settings
		if f.config.Message.Body.RequireSignoff {
			options = append(options, rules.WithRequireSignOff(true))
		} else {
			options = append(options, rules.WithRequireSignOff(false))
		}

		if f.config.Signing.AllowMultipleSignoffs {
			options = append(options, rules.WithMultipleSignoffs(true))
		}
	}

	return rules.NewSignOffRule(options...)
}

// createSpellRule creates a Spell rule with configuration.
func (f SimpleRuleFactory) createSpellRule() domain.Rule {
	options := []rules.SpellOption{}

	if f.config != nil {
		if len(f.config.Spell.IgnoreWords) > 0 {
			options = append(options, rules.WithCustomDictionary(f.config.Spell.IgnoreWords))
		}
	}

	return rules.NewSpellRule(options...)
}

// CreateBasicRules creates all standard validation rules.
func (f SimpleRuleFactory) CreateBasicRules() map[string]func(context.Context) domain.Rule {
	return map[string]func(context.Context) domain.Rule{
		"SubjectLength": func(_ context.Context) domain.Rule {
			return f.createSubjectLengthRule()
		},
		"SubjectCase": func(ctx context.Context) domain.Rule {
			return f.createSubjectCaseRule(ctx)
		},
		"SubjectSuffix": func(_ context.Context) domain.Rule {
			return f.createSubjectSuffixRule()
		},
		"ConventionalCommit": func(_ context.Context) domain.Rule {
			return f.createConventionalCommitRule()
		},
		"ImperativeVerb": func(ctx context.Context) domain.Rule {
			return f.createImperativeVerbRule(ctx)
		},
		"CommitBody": func(_ context.Context) domain.Rule {
			return f.createCommitBodyRule()
		},
		"JiraReference": func(ctx context.Context) domain.Rule {
			return f.createJiraReferenceRule(ctx)
		},
		"SignOff": func(_ context.Context) domain.Rule {
			return f.createSignOffRule()
		},
		"Signature": func(_ context.Context) domain.Rule {
			return rules.NewSignatureRule()
		},
		"Identity": func(_ context.Context) domain.Rule {
			// Fail fast if required crypto dependencies are missing
			if f.cryptoVerifier == nil || f.cryptoRepository == nil {
				// Return a no-op rule for now, but this should be handled at factory creation
				return rules.NewIdentityRule()
			}

			return rules.NewIdentityRule(
				rules.WithVerifier(f.cryptoVerifier),
				rules.WithKeyRepository(f.cryptoRepository),
			)
		},
		"Spell": func(_ context.Context) domain.Rule {
			return f.createSpellRule()
		},
	}
}

// CreateAnalyzerRules creates rules that require a commit analyzer.
func (f SimpleRuleFactory) CreateAnalyzerRules(analyzer domain.CommitAnalyzer) map[string]func(context.Context) domain.Rule {
	return map[string]func(context.Context) domain.Rule{
		"BranchAhead": func(_ context.Context) domain.Rule {
			options := []rules.BranchAheadOption{
				rules.WithRepositoryGetter(func() domain.CommitAnalyzer {
					return analyzer
				}),
			}
			if f.config != nil {
				if maxCommits := f.config.Repo.MaxCommitsAhead; maxCommits > 0 {
					options = append(options, rules.WithMaxCommitsAhead(maxCommits))
				}
				if branch := f.config.Repo.Branch; branch != "" {
					options = append(options, rules.WithReference(branch))
				}
			}

			return rules.NewBranchAheadRule(options...)
		},
	}
}
