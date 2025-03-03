// SPDX-FileCopyrightText: 2025 itiquette/gommitlint
//
// SPDX-License-Identifier: EUPL-1.2

package configuration

type AppConf struct {
	GommitConf *GommitLintConfig `koanf:"gommitlint"`
}

// New loads the gommitlint.yaml file and unmarshals it into a Gommitlint struct.
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
	Signature         *SignatureRule `koanf:"signature"`
	IsSignOffRequired *bool          `koanf:"sign-off"`

	// Misc validation rules
	IsNCommitMax *bool `koanf:"is-n-commit-max"`
}

type SubjectRule struct {
	// Case specifies the case that the first word of the description must have ("upper" or "lower").
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

type ConventionalRule struct {
	// MaxDescriptionLength specifies the maximum allowed length for the description.
	MaxDescriptionLength int `koanf:"max-description-length"`

	// Scopes lists the allowed scopes for conventional commits.
	Scopes []string `koanf:"scopes"`

	// Types lists the allowed types for conventional commits.
	Types []string `koanf:"types"`

	// IsRequired indicates whether Conventional Commits are required.
	IsRequired bool `koanf:"required"`
}

type SpellingRule struct {
	// Locale specifies the language/locale to use for spell checking.
	Locale string `koanf:"locale"`
}

type JiraRule struct {
	// Keys specifies the allowed Jira project keys.
	Keys []string `koanf:"keys"`

	// IsRequired indicates whether a Jira key must be present.
	IsRequired bool `koanf:"required"`
}

type BodyRule struct {
	// IsRequired enforces that the current commit has a body.
	IsRequired bool `koanf:"required"`
}

type SignatureRule struct {
	// Identity configures identity verification for signatures.
	Identity *IdentityRule `koanf:"identity"`

	// IsRequired enforces that the commit has a valid signature.
	IsRequired bool `koanf:"required"`
}

type IdentityRule struct {
	// PublicKeyURI points to a file containing authorized public keys.
	PublicKeyURI string `koanf:"public-key-uri"`
}
