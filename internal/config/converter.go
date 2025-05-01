// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package config

// FromGommitlintConfig converts a GommitlintConfig to a Config.
// This function maps the YAML/JSON configuration structure to our internal Config type.
func FromGommitlintConfig(cfg GommitlintConfig) Config {
	// Start with an empty config, not a default config
	config := Config{
		Subject: SubjectConfig{
			Jira: JiraConfig{
				Projects: []string{},
			},
		},
		Body: BodyConfig{},
		Conventional: ConventionalConfig{
			Scopes: []string{},
			Types:  []string{},
		},
		SpellCheck: SpellCheckConfig{
			IgnoreWords: []string{},
			CustomWords: map[string]string{},
		},
		Security: SecurityConfig{
			AllowedSignatureTypes: []string{"gpg", "ssh"},
			AllowMultipleSignOffs: true,
			Identity:              IdentityConfig{},
		},
		Repository: RepositoryConfig{
			MaxCommitsAhead: 5,
		},
		Rules: RulesConfig{
			EnabledRules:  []string{},
			DisabledRules: []string{},
		},
	}

	// Subject configuration
	config.Subject.MaxLength = cfg.Subject.MaxLength
	config.Subject.Case = cfg.Subject.Case
	config.Subject.Imperative = cfg.Subject.Imperative
	config.Subject.InvalidSuffixes = cfg.Subject.InvalidSuffixes

	// Jira configuration
	if cfg.Subject.Jira.Projects != nil {
		config.Subject.Jira.Projects = deepCopyStringSlice(cfg.Subject.Jira.Projects)
	}

	config.Subject.Jira.Required = cfg.Subject.Jira.Required
	config.Subject.Jira.BodyRef = cfg.Subject.Jira.BodyRef
	config.Subject.Jira.Pattern = cfg.Subject.Jira.Pattern
	config.Subject.Jira.Strict = cfg.Subject.Jira.Strict

	// Body configuration
	config.Body.Required = cfg.Body.Required
	config.Body.AllowSignOffOnly = cfg.Body.AllowSignOffOnly

	// Conventional commit configuration
	if cfg.ConventionalCommit.Types != nil {
		config.Conventional.Types = deepCopyStringSlice(cfg.ConventionalCommit.Types)
	}

	if cfg.ConventionalCommit.Scopes != nil {
		config.Conventional.Scopes = deepCopyStringSlice(cfg.ConventionalCommit.Scopes)
	}

	config.Conventional.MaxDescriptionLength = cfg.ConventionalCommit.MaxDescriptionLength
	config.Conventional.Required = cfg.ConventionalCommit.Required

	// SpellCheck configuration
	config.SpellCheck.Locale = cfg.SpellCheck.Locale
	config.SpellCheck.Enabled = cfg.SpellCheck.Enabled

	if cfg.SpellCheck.IgnoreWords != nil {
		config.SpellCheck.IgnoreWords = deepCopyStringSlice(cfg.SpellCheck.IgnoreWords)
	}

	if cfg.SpellCheck.CustomWords != nil {
		config.SpellCheck.CustomWords = deepCopyStringMap(cfg.SpellCheck.CustomWords)
	}

	config.SpellCheck.MaxErrors = cfg.SpellCheck.MaxErrors

	// Security configuration
	config.Security.SignatureRequired = cfg.Signature.Required
	config.Security.Identity.PublicKeyURI = cfg.Signature.Identity.PublicKeyURI
	config.Security.SignOffRequired = cfg.SignOffRequired

	// Repository configuration
	config.Repository.Reference = cfg.Reference
	config.Repository.IgnoreMergeCommits = cfg.IgnoreMergeCommits
	config.Repository.CheckCommitsAhead = cfg.NCommitsAhead

	return config
}

// FromOldConfig creates a new Config from the old legacy Config structure.
// This is a no-op conversion since we've now replaced the entire configuration
// system with the unified version. It's kept for API compatibility.
func FromOldConfig(cfg Config) Config {
	return cfg
}

// ToOldConfig converts the config to the legacy Config structure.
// This is a no-op conversion since we've now replaced the entire configuration
// system with the unified version. It's kept for API compatibility.
func ToOldConfig(cfg Config) Config {
	return cfg
}

// ToGommitlintConfig converts a Config to a GommitlintConfig.
// This function maps our internal Config type to the YAML/JSON configuration structure.
func (c Config) ToGommitlintConfig() GommitlintConfig {
	gommitCfg := GommitlintConfig{}

	// Subject configuration
	gommitCfg.Subject.MaxLength = c.Subject.MaxLength
	gommitCfg.Subject.Case = c.Subject.Case
	gommitCfg.Subject.Imperative = c.Subject.Imperative
	gommitCfg.Subject.InvalidSuffixes = c.Subject.InvalidSuffixes

	// Jira configuration
	gommitCfg.Subject.Jira.Projects = deepCopyStringSlice(c.Subject.Jira.Projects)
	gommitCfg.Subject.Jira.Required = c.Subject.Jira.Required
	gommitCfg.Subject.Jira.BodyRef = c.Subject.Jira.BodyRef
	gommitCfg.Subject.Jira.Pattern = c.Subject.Jira.Pattern
	gommitCfg.Subject.Jira.Strict = c.Subject.Jira.Strict

	// Body configuration
	gommitCfg.Body.Required = c.Body.Required
	gommitCfg.Body.AllowSignOffOnly = c.Body.AllowSignOffOnly

	// Conventional commit configuration
	gommitCfg.ConventionalCommit.Types = deepCopyStringSlice(c.Conventional.Types)
	gommitCfg.ConventionalCommit.Scopes = deepCopyStringSlice(c.Conventional.Scopes)
	gommitCfg.ConventionalCommit.MaxDescriptionLength = c.Conventional.MaxDescriptionLength
	gommitCfg.ConventionalCommit.Required = c.Conventional.Required

	// SpellCheck configuration
	gommitCfg.SpellCheck.Locale = c.SpellCheck.Locale
	gommitCfg.SpellCheck.Enabled = c.SpellCheck.Enabled
	gommitCfg.SpellCheck.IgnoreWords = deepCopyStringSlice(c.SpellCheck.IgnoreWords)
	gommitCfg.SpellCheck.CustomWords = deepCopyStringMap(c.SpellCheck.CustomWords)
	gommitCfg.SpellCheck.MaxErrors = c.SpellCheck.MaxErrors

	// Security configuration
	gommitCfg.Signature.Required = c.Security.SignatureRequired
	gommitCfg.Signature.Identity.PublicKeyURI = c.Security.Identity.PublicKeyURI
	gommitCfg.SignOffRequired = c.Security.SignOffRequired

	// Repository configuration
	gommitCfg.Reference = c.Repository.Reference
	gommitCfg.IgnoreMergeCommits = c.Repository.IgnoreMergeCommits
	gommitCfg.NCommitsAhead = c.Repository.CheckCommitsAhead

	return gommitCfg
}
