// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package config provides configuration management for gommitlint.
package config

// Option is a function that modifies a Config instance.
// It follows the functional options pattern.
type Option func(Config) Config

// NewConfig creates a new Config with default values and applies the given options.
func NewConfig(opts ...Option) Config {
	config := DefaultConfig()
	for _, opt := range opts {
		config = opt(config)
	}

	return config
}

// WithSubject sets the subject configuration.
func WithSubject(subject SubjectConfig) Option {
	return func(config Config) Config {
		config.Subject = subject

		return config
	}
}

// WithSubjectCase sets the subject case.
func WithSubjectCase(caseStyle string) Option {
	return func(config Config) Config {
		subject := config.Subject
		subject.Case = caseStyle
		config.Subject = subject

		return config
	}
}

// WithSubjectMaxLength sets the subject maximum length.
func WithSubjectMaxLength(maxLength int) Option {
	return func(config Config) Config {
		subject := config.Subject
		subject.MaxLength = maxLength
		config.Subject = subject

		return config
	}
}

// WithSubjectImperative sets whether imperative mood is required.
func WithSubjectImperative(imperative bool) Option {
	return func(config Config) Config {
		subject := config.Subject
		subject.Imperative = imperative
		config.Subject = subject

		return config
	}
}

// WithSubjectInvalidSuffixes sets the invalid suffixes for subjects.
func WithSubjectInvalidSuffixes(suffixes string) Option {
	return func(config Config) Config {
		subject := config.Subject
		subject.InvalidSuffixes = suffixes
		config.Subject = subject

		return config
	}
}

// WithBody sets the body configuration.
func WithBody(body BodyConfig) Option {
	return func(config Config) Config {
		config.Body = body

		return config
	}
}

// WithBodyRequired sets whether a commit body is required.
func WithBodyRequired(required bool) Option {
	return func(config Config) Config {
		body := config.Body
		body.Required = required
		config.Body = body

		return config
	}
}

// WithBodyAllowSignOffOnly sets whether a body with only sign-off is allowed.
func WithBodyAllowSignOffOnly(allow bool) Option {
	return func(config Config) Config {
		body := config.Body
		body.AllowSignOffOnly = allow
		config.Body = body

		return config
	}
}

// WithConventional sets the conventional commit configuration.
func WithConventional(conventional ConventionalConfig) Option {
	return func(config Config) Config {
		config.Conventional = conventional

		return config
	}
}

// WithConventionalRequired sets whether conventional commits are required.
func WithConventionalRequired(required bool) Option {
	return func(config Config) Config {
		conventional := config.Conventional
		conventional.Required = required
		config.Conventional = conventional

		return config
	}
}

// WithConventionalTypes sets the allowed conventional commit types.
func WithConventionalTypes(types []string) Option {
	return func(config Config) Config {
		conventional := config.Conventional
		conventional.Types = types
		config.Conventional = conventional

		return config
	}
}

// WithConventionalScopes sets the allowed conventional commit scopes.
func WithConventionalScopes(scopes []string) Option {
	return func(config Config) Config {
		conventional := config.Conventional
		conventional.Scopes = scopes
		config.Conventional = conventional

		return config
	}
}

// WithConventionalMaxDescriptionLength sets the max description length.
func WithConventionalMaxDescriptionLength(maxLength int) Option {
	return func(config Config) Config {
		conventional := config.Conventional
		conventional.MaxDescriptionLength = maxLength
		config.Conventional = conventional

		return config
	}
}

// WithSpellCheck sets the spell check configuration.
func WithSpellCheck(spellCheck SpellCheckConfig) Option {
	return func(config Config) Config {
		config.SpellCheck = spellCheck

		return config
	}
}

// WithSpellCheckEnabled sets whether spell checking is enabled.
func WithSpellCheckEnabled(enabled bool) Option {
	return func(config Config) Config {
		spellCheck := config.SpellCheck
		spellCheck.Enabled = enabled
		config.SpellCheck = spellCheck

		return config
	}
}

// WithSpellCheckLocale sets the locale for spell checking.
func WithSpellCheckLocale(locale string) Option {
	return func(config Config) Config {
		spellCheck := config.SpellCheck
		spellCheck.Locale = locale
		config.SpellCheck = spellCheck

		return config
	}
}

// WithSpellCheckIgnoreWords sets words to ignore during spell checking.
func WithSpellCheckIgnoreWords(ignoreWords []string) Option {
	return func(config Config) Config {
		spellCheck := config.SpellCheck
		spellCheck.IgnoreWords = ignoreWords
		config.SpellCheck = spellCheck

		return config
	}
}

// WithSpellCheckCustomWords sets custom word mappings for spell checking.
func WithSpellCheckCustomWords(customWords map[string]string) Option {
	return func(config Config) Config {
		spellCheck := config.SpellCheck
		spellCheck.CustomWords = customWords
		config.SpellCheck = spellCheck

		return config
	}
}

// WithSpellCheckMaxErrors sets the maximum number of spelling errors allowed.
func WithSpellCheckMaxErrors(maxErrors int) Option {
	return func(config Config) Config {
		spellCheck := config.SpellCheck
		spellCheck.MaxErrors = maxErrors
		config.SpellCheck = spellCheck

		return config
	}
}

// WithSecurity sets the security configuration.
func WithSecurity(security SecurityConfig) Option {
	return func(config Config) Config {
		config.Security = security

		return config
	}
}

// WithSignatureRequired sets whether signatures are required.
func WithSignatureRequired(required bool) Option {
	return func(config Config) Config {
		security := config.Security
		security.SignatureRequired = required
		config.Security = security

		return config
	}
}

// WithSignOffRequired sets whether sign-offs are required.
func WithSignOffRequired(required bool) Option {
	return func(config Config) Config {
		security := config.Security
		security.SignOffRequired = required
		config.Security = security

		return config
	}
}

// WithAllowedSignatureTypes sets the allowed signature types.
func WithAllowedSignatureTypes(types []string) Option {
	return func(config Config) Config {
		security := config.Security
		security.AllowedSignatureTypes = types
		config.Security = security

		return config
	}
}

// WithAllowMultipleSignOffs sets whether multiple sign-offs are allowed.
func WithAllowMultipleSignOffs(allow bool) Option {
	return func(config Config) Config {
		security := config.Security
		security.AllowMultipleSignOffs = allow
		config.Security = security

		return config
	}
}

// WithIdentity sets the identity configuration.
func WithIdentity(identity IdentityConfig) Option {
	return func(config Config) Config {
		security := config.Security
		security.Identity = identity
		config.Security = security

		return config
	}
}

// WithPublicKeyURI sets the public key URI for identity verification.
func WithPublicKeyURI(uri string) Option {
	return func(config Config) Config {
		security := config.Security
		identity := security.Identity
		identity.PublicKeyURI = uri
		security.Identity = identity
		config.Security = security

		return config
	}
}

// WithRepository sets the repository configuration.
func WithRepository(repository RepositoryConfig) Option {
	return func(config Config) Config {
		config.Repository = repository

		return config
	}
}

// WithReference sets the reference branch.
func WithReference(reference string) Option {
	return func(config Config) Config {
		repository := config.Repository
		repository.Reference = reference
		config.Repository = repository

		return config
	}
}

// WithIgnoreMergeCommits sets whether merge commits should be ignored.
func WithIgnoreMergeCommits(ignore bool) Option {
	return func(config Config) Config {
		repository := config.Repository
		repository.IgnoreMergeCommits = ignore
		config.Repository = repository

		return config
	}
}

// WithMaxCommitsAhead sets the maximum number of commits ahead.
func WithMaxCommitsAhead(maxCommits int) Option {
	return func(config Config) Config {
		repository := config.Repository
		repository.MaxCommitsAhead = maxCommits
		config.Repository = repository

		return config
	}
}

// WithCheckCommitsAhead sets whether to check for commits ahead.
func WithCheckCommitsAhead(check bool) Option {
	return func(config Config) Config {
		repository := config.Repository
		repository.CheckCommitsAhead = check
		config.Repository = repository

		return config
	}
}

// WithRules sets the rules configuration.
func WithRules(rules RulesConfig) Option {
	return func(config Config) Config {
		config.Rules = rules

		return config
	}
}

// WithEnabledRules sets the enabled rules.
func WithEnabledRules(rules []string) Option {
	return func(config Config) Config {
		rulesConfig := config.Rules
		rulesConfig.EnabledRules = rules
		config.Rules = rulesConfig

		return config
	}
}

// WithDisabledRules sets the disabled rules.
func WithDisabledRules(rules []string) Option {
	return func(config Config) Config {
		rulesConfig := config.Rules
		rulesConfig.DisabledRules = rules
		config.Rules = rulesConfig

		return config
	}
}

// WithJira sets the Jira configuration.
func WithJira(jira JiraConfig) Option {
	return func(config Config) Config {
		subject := config.Subject
		subject.Jira = jira
		config.Subject = subject

		return config
	}
}

// WithJiraRequired sets whether Jira references are required.
func WithJiraRequired(required bool) Option {
	return func(config Config) Config {
		subject := config.Subject
		jira := subject.Jira
		jira.Required = required
		subject.Jira = jira
		config.Subject = subject

		return config
	}
}

// WithJiraProjects sets the allowed Jira projects.
func WithJiraProjects(projects []string) Option {
	return func(config Config) Config {
		subject := config.Subject
		jira := subject.Jira
		jira.Projects = projects
		subject.Jira = jira
		config.Subject = subject

		return config
	}
}

// WithJiraPattern sets the Jira pattern.
func WithJiraPattern(pattern string) Option {
	return func(config Config) Config {
		subject := config.Subject
		jira := subject.Jira
		jira.Pattern = pattern
		subject.Jira = jira
		config.Subject = subject

		return config
	}
}

// WithJiraBodyRef sets whether to check for Jira references in the body.
func WithJiraBodyRef(check bool) Option {
	return func(config Config) Config {
		subject := config.Subject
		jira := subject.Jira
		jira.BodyRef = check
		subject.Jira = jira
		config.Subject = subject

		return config
	}
}

// WithJiraStrict sets whether to use strict Jira validation.
func WithJiraStrict(strict bool) Option {
	return func(config Config) Config {
		subject := config.Subject
		jira := subject.Jira
		jira.Strict = strict
		subject.Jira = jira
		config.Subject = subject

		return config
	}
}
