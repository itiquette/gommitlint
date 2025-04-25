// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package config provides configuration management for gommitlint.
package config

// GommitlintConfig is the root configuration structure for the application.
// This preserves the exact original YAML format while using value semantics.
type GommitlintConfig struct {
	// Content validation rules
	Subject            SubjectConfig      `json:"subject,omitempty"            koanf:"subject"`
	Body               BodyConfig         `json:"body,omitempty"               koanf:"body"`
	ConventionalCommit ConventionalConfig `json:"conventionalCommit,omitempty" koanf:"conventional-commit"`
	SpellCheck         SpellCheckConfig   `json:"spellcheck,omitempty"         koanf:"spellcheck"`
	// Security validation rules
	Signature       SignatureConfig `json:"signature,omitempty" koanf:"signature"`
	SignOffRequired bool            `json:"signOff,omitempty"   koanf:"sign-off"`
	// Misc validation rules
	NCommitsAhead      bool   `json:"nCommitsAhead,omitempty"     koanf:"n-commits-ahead"`
	IgnoreMergeCommits bool   `json:"ignoreMergeCommit,omitempty" koanf:"ignore-merge-commit"`
	Reference          string `json:"reference,omitempty"         koanf:"reference"`
}

// Config represents the internal representation with value semantics.
// This is the structure used throughout the application.
type Config struct {
	// Subject configuration
	Subject SubjectConfig

	// Body configuration
	Body BodyConfig

	// Conventional commit configuration
	Conventional ConventionalConfig

	// SpellCheck configuration
	SpellCheck SpellCheckConfig

	// Security validation rules
	Security SecurityConfig

	// Repository configuration
	Repository RepositoryConfig

	// Rules configuration
	Rules RulesConfig
}

// SubjectConfig holds configuration for commit subject validation.
type SubjectConfig struct {
	// Case specifies the case that the first word of the description must have ("upper","lower","ignore").
	Case string `json:"case,omitempty" koanf:"case"`

	// Imperative enforces the use of imperative verbs as the first word of a description.
	Imperative bool `json:"imperative,omitempty" koanf:"imperative"`

	// InvalidSuffixes lists characters that cannot be used at the end of the subject.
	InvalidSuffixes string `json:"invalidSuffixes,omitempty" koanf:"invalid-suffixes"`

	// MaxLength is the maximum length of the commit subject.
	MaxLength int `json:"maxLength,omitempty" koanf:"max-length"`

	// Jira holds Jira-related validation configuration.
	Jira JiraConfig `json:"jira,omitempty" koanf:"jira"`
}

// JiraConfig holds configuration for Jira ticket references.
type JiraConfig struct {
	// Projects specifies the allowed Jira project keys.
	Projects []string `json:"keys,omitempty" koanf:"keys"`

	// Required indicates whether a Jira key must be present.
	Required bool `json:"required,omitempty" koanf:"required"`

	// BodyRef indicates whether a Jira key must be present in body ref.
	BodyRef bool `json:"bodyref,omitempty" koanf:"bodyref"`

	// Pattern specifies the regex pattern for Jira keys.
	Pattern string `json:"pattern,omitempty" koanf:"pattern"`

	// Strict enables additional validation checks.
	Strict bool `json:"strict,omitempty" koanf:"strict"`
}

// BodyConfig holds configuration for commit body validation.
type BodyConfig struct {
	// Required enforces that the current commit has a body.
	Required bool `json:"required,omitempty" koanf:"required"`

	// AllowSignOffOnly determines if a body with only sign-off lines is allowed.
	AllowSignOffOnly bool `json:"allowSignOffOnly,omitempty" koanf:"allow-sign-off-only"`
}

// ConventionalConfig holds configuration for conventional commit format validation.
type ConventionalConfig struct {
	// MaxDescriptionLength specifies the maximum allowed length for the description.
	MaxDescriptionLength int `json:"maxDescriptionLength,omitempty" koanf:"max-description-length"`

	// Scopes lists the allowed scopes for conventional commits.
	Scopes []string `json:"scopes,omitempty" koanf:"scopes"`

	// Types lists the allowed types for conventional commits.
	Types []string `json:"types,omitempty" koanf:"types"`

	// Required indicates whether Conventional Commits are required.
	Required bool `json:"required,omitempty" koanf:"required"`
}

// SpellCheckConfig holds configuration for spell checking.
type SpellCheckConfig struct {
	// Locale specifies the language/locale to use for spell checking.
	Locale string `json:"locale,omitempty" koanf:"locale"`

	// Enabled indicates whether spell checking is enabled.
	Enabled bool `json:"enabled,omitempty" koanf:"enabled"`

	// IgnoreWords specifies words to ignore during spell checking
	IgnoreWords []string `json:"ignoreWords,omitempty" koanf:"ignore-words"`

	// CustomWords specifies custom word mappings for spell checking
	CustomWords map[string]string `json:"customWords,omitempty" koanf:"custom-words"`

	// MaxErrors specifies the maximum number of spelling errors allowed
	MaxErrors int `json:"maxErrors,omitempty" koanf:"max-errors"`
}

// SignatureConfig holds configuration for signature validation.
type SignatureConfig struct {
	// Identity configures identity verification for signatures.
	Identity IdentityConfig `json:"identity,omitempty" koanf:"identity"`

	// Required enforces that the commit has a valid signature.
	Required bool `json:"required,omitempty" koanf:"required"`
}

// SecurityConfig holds configuration for security-related validation.
type SecurityConfig struct {
	// SignatureRequired enforces that the commit has a valid signature.
	SignatureRequired bool

	// AllowedSignatureTypes specifies the allowed signature types (gpg, ssh).
	AllowedSignatureTypes []string

	// SignOffRequired enforces that commits are signed off.
	SignOffRequired bool

	// AllowMultipleSignOffs determines if multiple sign-offs are allowed.
	AllowMultipleSignOffs bool

	// IdentityConfig configures identity verification for signatures.
	Identity IdentityConfig
}

// IdentityConfig holds configuration for identity verification.
type IdentityConfig struct {
	// PublicKeyURI points to a file containing authorized public keys.
	PublicKeyURI string `json:"publicKeyUri,omitempty" koanf:"public-key-uri"`
}

// RepositoryConfig holds configuration for repository-related validation.
type RepositoryConfig struct {
	// Reference branch for comparison (usually main or master).
	Reference string

	// IgnoreMergeCommits indicates whether merge commits should be ignored.
	IgnoreMergeCommits bool

	// MaxCommitsAhead specifies the maximum allowed commits ahead of reference branch.
	MaxCommitsAhead int

	// CheckCommitsAhead enables checking for commits ahead of reference branch.
	CheckCommitsAhead bool
}

// RulesConfig holds configuration for rule enablement/disablement.
type RulesConfig struct {
	// EnabledRules lists rules that are explicitly enabled.
	// If empty, all rules are enabled unless in DisabledRules.
	EnabledRules []string

	// DisabledRules lists rules that are explicitly disabled.
	// Only used if EnabledRules is empty.
	DisabledRules []string
}
