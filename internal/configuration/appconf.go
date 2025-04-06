// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package configuration

// AppConf is the root configuration structure for the application.
type AppConf struct {
	GommitConf *GommitLintConfig `koanf:"gommitlint"`
}

// New loads the gommitlint configuration and returns an AppConf instance.
// Returns an error if configuration loading fails.
func New() (*AppConf, error) {
	gommitLintConf, err := DefaultConfigLoader{}.LoadConfiguration()
	if err != nil {
		return nil, err
	}

	return gommitLintConf, nil
}

// GommitLintConfig defines the complete configuration for commit linting rules.
type GommitLintConfig struct {
	// Content validation rules
	Subject            *SubjectRule      `koanf:"subject"`
	Body               *BodyRule         `koanf:"body"`
	ConventionalCommit *ConventionalRule `koanf:"conventional-commit"`
	SpellCheck         *SpellingRule     `koanf:"spellcheck"`
	// Security validation rules
	Signature       *SignatureRule `koanf:"signature"`
	SignOffRequired *bool          `koanf:"sign-off"`
	// Misc validation rules
	NCommitsAhead      *bool `koanf:"n-commits-ahead"`
	IgnoreMergeCommits *bool `koanf:"ignore-merge-commit"`
}

// SubjectRule defines configuration for commit subject validation.
type SubjectRule struct {
	// Case specifies the case that the first word of the description must have ("upper","lower","ignore").
	Case string `koanf:"case"`

	// Imperative enforces the use of imperative verbs as the first word of a description.
	Imperative *bool `koanf:"imperative"`

	// InvalidSuffixes lists characters that cannot be used at the end of the subject.
	InvalidSuffixes string `koanf:"invalid-suffixes"`

	// Jira checks if the subject contains a Jira project key.
	Jira *JiraRule `koanf:"jira"`

	// MaxLength is the maximum length of the commit subject.
	MaxLength int `koanf:"max-length"`
}

// ConventionalRule defines configuration for conventional commit format validation.
type ConventionalRule struct {
	// MaxDescriptionLength specifies the maximum allowed length for the description.
	MaxDescriptionLength int `koanf:"max-description-length"`

	// Scopes lists the allowed scopes for conventional commits.
	Scopes []string `koanf:"scopes"`

	// Types lists the allowed types for conventional commits.
	Types []string `koanf:"types"`

	// Required indicates whether Conventional Commits are required.
	Required bool `koanf:"required"`
}

// SpellingRule defines configuration for spell checking.
type SpellingRule struct {
	// Locale specifies the language/locale to use for spell checking.
	Locale string `koanf:"locale"`
}

// JiraRule defines configuration for Jira key validation.
type JiraRule struct {
	// Keys specifies the allowed Jira project keys.
	Keys []string `koanf:"keys"`

	// Required indicates whether a Jira key must be present.
	Required bool `koanf:"required"`

	// BodyRef indicates whether a Jira key must be present in body ref .
	BodyRef bool `koanf:"bodyref"`
}

// BodyRule defines configuration for commit body validation.
type BodyRule struct {
	// Required enforces that the current commit has a body.
	Required bool `koanf:"required"`
}

// SignatureRule defines configuration for signature validation.
type SignatureRule struct {
	// Identity configures identity verification for signatures.
	Identity *IdentityRule `koanf:"identity"`

	// Required enforces that the commit has a valid signature.
	Required bool `koanf:"required"`
}

// IdentityRule defines configuration for identity verification.
type IdentityRule struct {
	// PublicKeyURI points to a file containing authorized public keys.
	PublicKeyURI string `koanf:"public-key-uri"`
}
