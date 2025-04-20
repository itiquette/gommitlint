// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package config provides configuration management for gommitlint.
package config

// AppConf is the root configuration structure for the application.
type AppConf struct {
	GommitConf *GommitLintConfig `json:"gommitlint,omitempty" koanf:"gommitlint"`
}

// GommitLintConfig defines the complete configuration for commit linting rules.
type GommitLintConfig struct {
	// Content validation rules
	Subject            *SubjectRule      `json:"subject,omitempty"            koanf:"subject"`
	Body               *BodyRule         `json:"body,omitempty"               koanf:"body"`
	ConventionalCommit *ConventionalRule `json:"conventionalCommit,omitempty" koanf:"conventional-commit"`
	SpellCheck         *SpellingRule     `json:"spellcheck,omitempty"         koanf:"spellcheck"`
	// Security validation rules
	Signature       *SignatureRule `json:"signature,omitempty" koanf:"signature"`
	SignOffRequired *bool          `json:"signOff,omitempty"   koanf:"sign-off"`
	// Misc validation rules
	NCommitsAhead      *bool  `json:"nCommitsAhead,omitempty"     koanf:"n-commits-ahead"`
	IgnoreMergeCommits *bool  `json:"ignoreMergeCommit,omitempty" koanf:"ignore-merge-commit"`
	Reference          string `json:"reference,omitempty"         koanf:"reference"`
}

// SubjectRule defines configuration for commit subject validation.
type SubjectRule struct {
	// Case specifies the case that the first word of the description must have ("upper","lower","ignore").
	Case string `json:"case,omitempty" koanf:"case"`
	// Imperative enforces the use of imperative verbs as the first word of a description.
	Imperative *bool `json:"imperative,omitempty" koanf:"imperative"`
	// InvalidSuffixes lists characters that cannot be used at the end of the subject.
	InvalidSuffixes string `json:"invalidSuffixes,omitempty" koanf:"invalid-suffixes"`
	// Jira checks if the subject contains a Jira project key.
	Jira *JiraRule `json:"jira,omitempty" koanf:"jira"`
	// MaxLength is the maximum length of the commit subject.
	MaxLength int `json:"maxLength,omitempty" koanf:"max-length"`
}

// ConventionalRule defines configuration for conventional commit format validation.
type ConventionalRule struct {
	// MaxDescriptionLength specifies the maximum allowed length for the description.
	MaxDescriptionLength int `json:"maxDescriptionLength,omitempty" koanf:"max-description-length"`
	// Scopes lists the allowed scopes for conventional commits.
	Scopes []string `json:"scopes,omitempty" koanf:"scopes"`
	// Types lists the allowed types for conventional commits.
	Types []string `json:"types,omitempty" koanf:"types"`
	// Required indicates whether Conventional Commits are required.
	Required bool `json:"required,omitempty" koanf:"required"`
}

// SpellingRule defines configuration for spell checking.
type SpellingRule struct {
	// Locale specifies the language/locale to use for spell checking.
	Locale string `json:"locale,omitempty" koanf:"locale"`
	// Enabled indicates whether spell checking is enabled.
	Enabled bool `json:"enabled,omitempty" koanf:"enabled"`
}

// JiraRule defines configuration for Jira key validation.
type JiraRule struct {
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

// BodyRule defines configuration for commit body validation.
type BodyRule struct {
	// Required enforces that the current commit has a body.
	Required bool `json:"required,omitempty" koanf:"required"`
}

// SignatureRule defines configuration for signature validation.
type SignatureRule struct {
	// Identity configures identity verification for signatures.
	Identity *IdentityRule `json:"identity,omitempty" koanf:"identity"`
	// Required enforces that the commit has a valid signature.
	Required bool `json:"required,omitempty" koanf:"required"`
}

// IdentityRule defines configuration for identity verification.
type IdentityRule struct {
	// PublicKeyURI points to a file containing authorized public keys.
	PublicKeyURI string `json:"publicKeyUri,omitempty" koanf:"public-key-uri"`
}
