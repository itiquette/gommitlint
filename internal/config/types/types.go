// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package types provides the core configuration data structures for gommitlint.
// It is separated from the config package to avoid import cycles.
package types

// Config represents the complete configuration for gommitlint.
// It uses value semantics for immutability.
type Config struct {
	Message      MessageConfig      `json:"message"      yaml:"message"`
	Conventional ConventionalConfig `json:"conventional" yaml:"conventional"`
	Signing      SigningConfig      `json:"signing"      yaml:"signing"`
	Repo         RepoConfig         `json:"repo"         yaml:"repo"`
	Jira         JiraConfig         `json:"jira"         yaml:"jira"`
	Spell        SpellConfig        `json:"spell"        yaml:"spell"`
	Rules        RulesConfig        `json:"rules"        yaml:"rules"`
	Output       string             `json:"output"       yaml:"output"`
}

// MessageConfig contains configuration for commit message validation.
type MessageConfig struct {
	Subject SubjectConfig `json:"subject" yaml:"subject"`
	Body    BodyConfig    `json:"body"    yaml:"body"`
}

// SubjectConfig contains configuration options for commit subject validation.
type SubjectConfig struct {
	MaxLength         int      `json:"max_length"         yaml:"max_length"`
	Case              string   `json:"case"               yaml:"case"`
	RequireImperative bool     `json:"require_imperative" yaml:"require_imperative"`
	ForbidEndings     []string `json:"forbid_endings"     yaml:"forbid_endings"`
}

// BodyConfig contains configuration options for commit body validation.
type BodyConfig struct {
	MinLength        int  `json:"min_length"         yaml:"min_length"`
	MinLines         int  `json:"min_lines"          yaml:"min_lines"`
	AllowSignoffOnly bool `json:"allow_signoff_only" yaml:"allow_signoff_only"`
	RequireSignoff   bool `json:"require_signoff"    yaml:"require_signoff"`
}

// ConventionalConfig contains configuration options for conventional commit format validation.
type ConventionalConfig struct {
	RequireScope         bool     `json:"require_scope"          yaml:"require_scope"`
	Types                []string `json:"types"                  yaml:"types"`
	Scopes               []string `json:"scopes"                 yaml:"scopes"`
	AllowBreaking        bool     `json:"allow_breaking"         yaml:"allow_breaking"`
	MaxDescriptionLength int      `json:"max_description_length" yaml:"max_description_length"`
}

// SigningConfig contains configuration options for signing-related validations.
type SigningConfig struct {
	RequireGPG            bool     `json:"require_gpg"             yaml:"require_gpg"`
	AllowMultipleSignoffs bool     `json:"allow_multiple_signoffs" yaml:"allow_multiple_signoffs"`
	AllowedSigners        []string `json:"allowed_signers"         yaml:"allowed_signers"`
	KeyDirectory          string   `json:"key_directory"           yaml:"key_directory"`
}

// RepoConfig contains configuration options related to the Git repository.
type RepoConfig struct {
	Path            string `json:"path"              yaml:"path"`
	Branch          string `json:"branch"            yaml:"branch"`
	MaxCommitsAhead int    `json:"max_commits_ahead" yaml:"max_commits_ahead"`
	IgnoreMerges    bool   `json:"ignore_merges"     yaml:"ignore_merges"`
}

// (Removed IntegrationsConfig - Jira and Spell configs are now top-level)

// JiraConfig contains configuration options for JIRA ticket reference validation.
type JiraConfig struct {
	Pattern   string   `json:"pattern"    yaml:"pattern"`
	Projects  []string `json:"projects"   yaml:"projects"`
	CheckBody bool     `json:"check_body" yaml:"check_body"`
}

// SpellConfig contains configuration options for spell checking.
type SpellConfig struct {
	Language    string   `json:"language"     yaml:"language"`
	IgnoreWords []string `json:"ignore_words" yaml:"ignore_words"`
}

// RulesConfig contains configuration for enabled and disabled validation rules.
type RulesConfig struct {
	Enabled  []string `json:"enabled"  yaml:"enabled"`
	Disabled []string `json:"disabled" yaml:"disabled"`
}
