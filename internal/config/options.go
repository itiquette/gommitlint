// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package config

// Option is a function that applies changes to a Config and returns a new instance.
type Option func(Config) Config

// NewConfigWithOptions creates a new Config with default values and applies the given options.
func NewConfigWithOptions(opts ...Option) Config {
	config := NewConfig()
	for _, opt := range opts {
		config = opt(config)
	}

	return config
}

// Subject options

// WithSubjectMaxLengthOpt returns an option to set the subject max length.
func WithSubjectMaxLengthOpt(maxLength int) Option {
	return func(c Config) Config {
		return c.WithSubjectMaxLength(maxLength)
	}
}

// WithSubjectCaseOpt returns an option to set the subject case style.
func WithSubjectCaseOpt(caseStyle string) Option {
	return func(c Config) Config {
		return c.WithSubjectCase(caseStyle)
	}
}

// WithSubjectImperativeOpt returns an option to set the subject imperative setting.
func WithSubjectImperativeOpt(imperative bool) Option {
	return func(c Config) Config {
		return c.WithSubjectImperative(imperative)
	}
}

// WithSubjectInvalidSuffixesOpt returns an option to set the subject invalid suffixes.
func WithSubjectInvalidSuffixesOpt(suffixes string) Option {
	return func(c Config) Config {
		return c.WithSubjectInvalidSuffixes(suffixes)
	}
}

// Body options

// WithBodyRequiredOpt returns an option to set the body required setting.
func WithBodyRequiredOpt(required bool) Option {
	return func(c Config) Config {
		return c.WithBodyRequired(required)
	}
}

// WithBodyAllowSignOffOnlyOpt returns an option to set the sign-off only setting.
func WithBodyAllowSignOffOnlyOpt(allow bool) Option {
	return func(c Config) Config {
		return c.WithBodyAllowSignOffOnly(allow)
	}
}

// Conventional options

// WithConventionalTypesOpt returns an option to set the conventional types.
func WithConventionalTypesOpt(types []string) Option {
	return func(c Config) Config {
		return c.WithConventionalTypes(types)
	}
}

// WithConventionalScopesOpt returns an option to set the conventional scopes.
func WithConventionalScopesOpt(scopes []string) Option {
	return func(c Config) Config {
		return c.WithConventionalScopes(scopes)
	}
}

// WithConventionalMaxDescriptionLengthOpt returns an option to set the max description length.
func WithConventionalMaxDescriptionLengthOpt(maxLength int) Option {
	return func(c Config) Config {
		return c.WithConventionalMaxDescriptionLength(maxLength)
	}
}

// WithConventionalRequiredOpt returns an option to set the conventional required setting.
func WithConventionalRequiredOpt(required bool) Option {
	return func(c Config) Config {
		return c.WithConventionalRequired(required)
	}
}

// JIRA options

// WithJiraProjectsOpt returns an option to set the JIRA projects.
func WithJiraProjectsOpt(projects []string) Option {
	return func(c Config) Config {
		return c.WithJiraProjects(projects)
	}
}

// WithJiraBodyRefOpt returns an option to set the JIRA body ref setting.
func WithJiraBodyRefOpt(bodyRef bool) Option {
	return func(c Config) Config {
		return c.WithJiraBodyRef(bodyRef)
	}
}

// WithJiraRequiredOpt returns an option to set the JIRA required setting.
func WithJiraRequiredOpt(required bool) Option {
	return func(c Config) Config {
		return c.WithJiraRequired(required)
	}
}

// WithJiraPatternOpt returns an option to set the JIRA pattern.
func WithJiraPatternOpt(pattern string) Option {
	return func(c Config) Config {
		return c.WithJiraPattern(pattern)
	}
}

// WithJiraStrictOpt returns an option to set the JIRA strict setting.
func WithJiraStrictOpt(strict bool) Option {
	return func(c Config) Config {
		return c.WithJiraStrict(strict)
	}
}

// Security options

// WithSignatureRequiredOpt returns an option to set the signature required setting.
func WithSignatureRequiredOpt(required bool) Option {
	return func(c Config) Config {
		return c.WithSignatureRequired(required)
	}
}

// WithAllowedSignatureTypesOpt returns an option to set the allowed signature types.
func WithAllowedSignatureTypesOpt(types []string) Option {
	return func(c Config) Config {
		return c.WithAllowedSignatureTypes(types)
	}
}

// WithSignOffRequiredOpt returns an option to set the sign-off required setting.
func WithSignOffRequiredOpt(required bool) Option {
	return func(c Config) Config {
		return c.WithSignOffRequired(required)
	}
}

// WithAllowMultipleSignOffsOpt returns an option to set the multiple sign-offs setting.
func WithAllowMultipleSignOffsOpt(allow bool) Option {
	return func(c Config) Config {
		return c.WithAllowMultipleSignOffs(allow)
	}
}

// WithIdentityPublicKeyURIOpt returns an option to set the public key URI.
func WithIdentityPublicKeyURIOpt(uri string) Option {
	return func(c Config) Config {
		return c.WithIdentityPublicKeyURI(uri)
	}
}

// SpellCheck options

// WithSpellLocaleOpt returns an option to set the spell check locale.
func WithSpellLocaleOpt(locale string) Option {
	return func(c Config) Config {
		return c.WithSpellLocale(locale)
	}
}

// WithSpellEnabledOpt returns an option to set the spell check enabled setting.
func WithSpellEnabledOpt(enabled bool) Option {
	return func(c Config) Config {
		return c.WithSpellEnabled(enabled)
	}
}

// WithSpellIgnoreWordsOpt returns an option to set the spell check ignore words.
func WithSpellIgnoreWordsOpt(words []string) Option {
	return func(c Config) Config {
		return c.WithSpellIgnoreWords(words)
	}
}

// WithSpellCustomWordsOpt returns an option to set the spell check custom words.
func WithSpellCustomWordsOpt(words map[string]string) Option {
	return func(c Config) Config {
		return c.WithSpellCustomWords(words)
	}
}

// WithSpellMaxErrorsOpt returns an option to set the spell check max errors.
func WithSpellMaxErrorsOpt(maxErrors int) Option {
	return func(c Config) Config {
		return c.WithSpellMaxErrors(maxErrors)
	}
}

// Repository options

// WithReferenceBranchOpt returns an option to set the reference branch.
func WithReferenceBranchOpt(branch string) Option {
	return func(c Config) Config {
		return c.WithReferenceBranch(branch)
	}
}

// WithIgnoreMergeCommitsOpt returns an option to set the ignore merge commits setting.
func WithIgnoreMergeCommitsOpt(ignore bool) Option {
	return func(c Config) Config {
		return c.WithIgnoreMergeCommits(ignore)
	}
}

// WithMaxCommitsAheadOpt returns an option to set the max commits ahead.
func WithMaxCommitsAheadOpt(maxCommits int) Option {
	return func(c Config) Config {
		return c.WithMaxCommitsAhead(maxCommits)
	}
}

// WithCheckCommitsAheadOpt returns an option to set the check commits ahead setting.
func WithCheckCommitsAheadOpt(check bool) Option {
	return func(c Config) Config {
		return c.WithCheckCommitsAhead(check)
	}
}

// Rule options

// WithEnabledRulesOpt returns an option to set the enabled rules.
func WithEnabledRulesOpt(rules []string) Option {
	return func(c Config) Config {
		return c.WithEnabledRules(rules)
	}
}

// WithDisabledRulesOpt returns an option to set the disabled rules.
func WithDisabledRulesOpt(rules []string) Option {
	return func(c Config) Config {
		return c.WithDisabledRules(rules)
	}
}
