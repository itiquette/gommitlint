// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package types provides the core configuration data structures for gommitlint.
// It is separated from the config package to avoid import cycles.
package types

// Config represents the complete configuration for gommitlint.
// It uses value semantics for immutability.
type Config struct {
	Subject      SubjectConfig      `json:"subject"      yaml:"subject"`
	Body         BodyConfig         `json:"body"         yaml:"body"`
	Conventional ConventionalConfig `json:"conventional" yaml:"conventional"`
	Rules        RulesConfig        `json:"rules"        yaml:"rules"`
	Security     SecurityConfig     `json:"security"     yaml:"security"`
	Repository   RepositoryConfig   `json:"repository"   yaml:"repository"`
	Output       OutputConfig       `json:"output"       yaml:"output"`
	SpellCheck   SpellCheckConfig   `json:"spell_check"  yaml:"spell_check"`
	Jira         JiraConfig         `json:"jira"         yaml:"jira"`
}

// SubjectConfig contains configuration options for commit subject validation.
type SubjectConfig struct {
	Case               string   `json:"case"                yaml:"case"`
	MaxLength          int      `json:"max_length"          yaml:"max_length"`
	Imperative         bool     `json:"imperative"          yaml:"imperative"`
	DisallowedSuffixes []string `json:"disallowed_suffixes" yaml:"disallowed_suffixes"`
}

// BodyConfig contains configuration options for commit body validation.
type BodyConfig struct {
	MinLength        int  `json:"min_length"          yaml:"min_length"`
	MinLines         int  `json:"min_lines"           yaml:"min_lines"`
	AllowSignOffOnly bool `json:"allow_sign_off_only" yaml:"allow_sign_off_only"`
	RequireSignOff   bool `json:"require_sign_off"    yaml:"require_sign_off"`
}

// ConventionalConfig contains configuration options for conventional commit format validation.
type ConventionalConfig struct {
	RequireScope         bool     `json:"require_scope"          yaml:"require_scope"`
	Types                []string `json:"types"                  yaml:"types"`
	Scopes               []string `json:"scopes"                 yaml:"scopes"`
	AllowBreakingChanges bool     `json:"allow_breaking_changes" yaml:"allow_breaking_changes"`
	MaxDescriptionLength int      `json:"max_description_length" yaml:"max_description_length"`
}

// RulesConfig contains configuration for enabled and disabled validation rules.
type RulesConfig struct {
	Enabled  []string `json:"enabled"  yaml:"enabled"`
	Disabled []string `json:"disabled" yaml:"disabled"`
}

// SecurityConfig contains configuration options for security-related validations.
type SecurityConfig struct {
	GPGRequired           bool     `json:"gpg_required"            yaml:"gpg_required"`
	KeyDirectory          string   `json:"key_directory"           yaml:"key_directory"`
	AllowedSignatureTypes []string `json:"allowed_signature_types" yaml:"allowed_signature_types"`
	AllowedKeyrings       []string `json:"allowed_keyrings"        yaml:"allowed_keyrings"`
	AllowedIdentities     []string `json:"allowed_identities"      yaml:"allowed_identities"`
	MultipleSignoffs      bool     `json:"multiple_signoffs"       yaml:"multiple_signoffs"`
}

// RepositoryConfig contains configuration options related to the Git repository.
type RepositoryConfig struct {
	Path               string `json:"path"                 yaml:"path"`
	ReferenceBranch    string `json:"reference_branch"     yaml:"reference_branch"`
	MaxCommitsAhead    int    `json:"max_commits_ahead"    yaml:"max_commits_ahead"`
	MaxHistoryDays     int    `json:"max_history_days"     yaml:"max_history_days"`
	IgnoreMergeCommits bool   `json:"ignore_merge_commits" yaml:"ignore_merge_commits"`
}

// OutputConfig contains configuration options for output formatting.
type OutputConfig struct {
	Format  string `json:"format"  yaml:"format"`
	Verbose bool   `json:"verbose" yaml:"verbose"`
	Quiet   bool   `json:"quiet"   yaml:"quiet"`
	Color   bool   `json:"color"   yaml:"color"`
}

// SpellCheckConfig contains configuration options for spell checking.
type SpellCheckConfig struct {
	Language         string   `json:"language"          yaml:"language"`
	IgnoreCase       bool     `json:"ignore_case"       yaml:"ignore_case"`
	CustomDictionary []string `json:"custom_dictionary" yaml:"custom_dictionary"`
}

// JiraConfig contains configuration options for JIRA ticket reference validation.
type JiraConfig struct {
	Pattern  string   `json:"pattern"  yaml:"pattern"`
	Projects []string `json:"projects" yaml:"projects"`
	BodyRef  bool     `json:"body_ref" yaml:"body_ref"`
}
