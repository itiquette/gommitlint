// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package types provides the core configuration data structures for gommitlint.
// It is separated from the config package to avoid import cycles.
package types

// Config represents the complete configuration for gommitlint.
// It uses value semantics for immutability.
type Config struct {
	Subject      SubjectConfig
	Body         BodyConfig
	Conventional ConventionalConfig
	Rules        RulesConfig
	Security     SecurityConfig
	Repository   RepositoryConfig
	Output       OutputConfig
	SpellCheck   SpellCheckConfig
	Jira         JiraConfig
}

// SubjectConfig contains configuration options for commit subject validation.
type SubjectConfig struct {
	Case               string
	MaxLength          int
	RequireImperative  bool
	DisallowedSuffixes []string
}

// BodyConfig contains configuration options for commit body validation.
type BodyConfig struct {
	Required         bool
	MinLength        int
	MinimumLines     int
	AllowSignOffOnly bool
}

// ConventionalConfig contains configuration options for conventional commit format validation.
type ConventionalConfig struct {
	Required             bool
	RequireScope         bool
	Types                []string
	Scopes               []string
	AllowBreakingChanges bool
	MaxDescriptionLength int
}

// RulesConfig contains configuration for enabled and disabled validation rules.
type RulesConfig struct {
	EnabledRules  []string
	DisabledRules []string
}

// SecurityConfig contains configuration options for security-related validations.
type SecurityConfig struct {
	SignOffRequired       bool
	GPGRequired           bool
	KeyDirectory          string
	AllowedSignatureTypes []string
	AllowedKeyrings       []string
	AllowedIdentities     []string
	AllowMultipleSignOffs bool
}

// RepositoryConfig contains configuration options related to the Git repository.
type RepositoryConfig struct {
	Path               string
	ReferenceBranch    string
	MaxCommitsAhead    int
	MaxHistoryDays     int
	OutputFormat       string
	IgnoreMergeCommits bool
}

// OutputConfig contains configuration options for output formatting.
type OutputConfig struct {
	Format  string
	Verbose bool
	Quiet   bool
	Color   bool
}

// SpellCheckConfig contains configuration options for spell checking.
type SpellCheckConfig struct {
	Enabled          bool
	Language         string
	IgnoreCase       bool
	CustomDictionary []string
}

// JiraConfig contains configuration options for JIRA ticket reference validation.
type JiraConfig struct {
	Pattern  string
	Projects []string
	BodyRef  bool
}
