// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package config provides configuration management for gommitlint.
package config

// FromGommitlintConfig converts the user-facing configuration to the internal representation.
// This function preserves the external file format while using value semantics internally.
// For an empty GommitlintConfig, the result should match DefaultConfig().
func FromGommitlintConfig(cfg GommitlintConfig) Config {
	// Start with defaults to ensure all fields are properly initialized
	config := DefaultConfig()

	// Special case for config with only Jira settings
	hasJiraConfig := cfg.Subject.Jira.Required

	// For empty configs (but not those with Jira settings), just return the defaults
	if isEmpty(cfg) && !hasJiraConfig {
		return DefaultConfig()
	}

	// Subject configuration
	if cfg.Subject.MaxLength != 0 {
		config.Subject.MaxLength = cfg.Subject.MaxLength
	}

	if cfg.Subject.Case != "" {
		config.Subject.Case = cfg.Subject.Case
	}
	// Only update if non-default config has been provided
	if !isEmpty(cfg) {
		config.Subject.Imperative = cfg.Subject.Imperative
	}

	if cfg.Subject.InvalidSuffixes != "" {
		config.Subject.InvalidSuffixes = cfg.Subject.InvalidSuffixes
	}

	// Jira configuration
	if len(cfg.Subject.Jira.Projects) > 0 {
		config.Subject.Jira.Projects = cfg.Subject.Jira.Projects
	}

	// Always apply the jira required setting, regardless of other settings
	config.Subject.Jira.Required = cfg.Subject.Jira.Required

	// Apply other Jira settings
	if !isEmpty(cfg) {
		config.Subject.Jira.BodyRef = cfg.Subject.Jira.BodyRef
		config.Subject.Jira.Strict = cfg.Subject.Jira.Strict
	}

	if cfg.Subject.Jira.Pattern != "" {
		config.Subject.Jira.Pattern = cfg.Subject.Jira.Pattern
	}

	// Body configuration
	if !isEmpty(cfg) {
		config.Body.Required = cfg.Body.Required
		config.Body.AllowSignOffOnly = cfg.Body.AllowSignOffOnly
	}

	// Conventional commit configuration
	if cfg.ConventionalCommit.MaxDescriptionLength != 0 {
		config.Conventional.MaxDescriptionLength = cfg.ConventionalCommit.MaxDescriptionLength
	}

	if len(cfg.ConventionalCommit.Scopes) > 0 {
		config.Conventional.Scopes = cfg.ConventionalCommit.Scopes
	}

	if len(cfg.ConventionalCommit.Types) > 0 {
		config.Conventional.Types = cfg.ConventionalCommit.Types
	}

	if !isEmpty(cfg) {
		config.Conventional.Required = cfg.ConventionalCommit.Required
	}

	// Spell check configuration
	if cfg.SpellCheck.Locale != "" {
		config.SpellCheck.Locale = cfg.SpellCheck.Locale
	}

	if !isEmpty(cfg) {
		config.SpellCheck.Enabled = cfg.SpellCheck.Enabled
	}

	if len(cfg.SpellCheck.IgnoreWords) > 0 {
		config.SpellCheck.IgnoreWords = cfg.SpellCheck.IgnoreWords
	}

	if len(cfg.SpellCheck.CustomWords) > 0 {
		config.SpellCheck.CustomWords = cfg.SpellCheck.CustomWords
	}

	if cfg.SpellCheck.MaxErrors > 0 {
		config.SpellCheck.MaxErrors = cfg.SpellCheck.MaxErrors
	}

	// Security configuration
	if !isEmpty(cfg) {
		config.Security.SignatureRequired = cfg.Signature.Required
		config.Security.SignOffRequired = cfg.SignOffRequired
	}

	if cfg.Signature.Identity.PublicKeyURI != "" {
		config.Security.Identity.PublicKeyURI = cfg.Signature.Identity.PublicKeyURI
	}

	// Repository configuration
	if cfg.Reference != "" {
		config.Repository.Reference = cfg.Reference
	}

	if !isEmpty(cfg) {
		config.Repository.IgnoreMergeCommits = cfg.IgnoreMergeCommits
		config.Repository.CheckCommitsAhead = cfg.NCommitsAhead
	}

	return config
}

// isEmpty checks if the given GommitlintConfig is effectively empty.
// This helps distinguish between a zero-value struct and an actual configuration.
func isEmpty(cfg GommitlintConfig) bool {
	// A very simple check: if subject max length and reference are both empty,
	// we consider it an empty config
	return cfg.Subject.MaxLength == 0 &&
		cfg.Subject.Case == "" &&
		cfg.Subject.InvalidSuffixes == "" &&
		cfg.Reference == "" &&
		len(cfg.Subject.Jira.Projects) == 0 &&
		len(cfg.ConventionalCommit.Types) == 0 &&
		len(cfg.ConventionalCommit.Scopes) == 0 &&
		cfg.ConventionalCommit.MaxDescriptionLength == 0 &&
		cfg.SpellCheck.Locale == "" &&
		len(cfg.SpellCheck.IgnoreWords) == 0 &&
		len(cfg.SpellCheck.CustomWords) == 0 &&
		cfg.SpellCheck.MaxErrors == 0 &&
		cfg.Signature.Identity.PublicKeyURI == ""
}

// ToGommitlintConfig converts the internal configuration to the user-facing format.
// This function is used when writing configuration files.
func ToGommitlintConfig(cfg Config) GommitlintConfig {
	// Create a new GommitlintConfig from the internal Config
	gommitCfg := GommitlintConfig{
		Subject:            cfg.Subject,
		Body:               cfg.Body,
		ConventionalCommit: cfg.Conventional,
		SpellCheck:         cfg.SpellCheck,
		Signature: SignatureConfig{
			Required: cfg.Security.SignatureRequired,
			Identity: cfg.Security.Identity,
		},
		SignOffRequired:    cfg.Security.SignOffRequired,
		Reference:          cfg.Repository.Reference,
		IgnoreMergeCommits: cfg.Repository.IgnoreMergeCommits,
		NCommitsAhead:      cfg.Repository.CheckCommitsAhead,
	}

	return gommitCfg
}
