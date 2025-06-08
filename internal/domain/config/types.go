// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package config provides pure configuration types for gommitlint.
// This package contains only data structures with no behavior.
package config

// Config represents the complete configuration for gommitlint.
type Config struct {
	Message      MessageConfig      `json:"message"      toml:"message"      yaml:"message"`
	Conventional ConventionalConfig `json:"conventional" toml:"conventional" yaml:"conventional"`
	Signature    SignatureConfig    `json:"signature"    toml:"signature"    yaml:"signature"`
	Identity     IdentityConfig     `json:"identity"     toml:"identity"     yaml:"identity"`
	Repo         RepoConfig         `json:"repo"         toml:"repo"         yaml:"repo"`
	Jira         JiraConfig         `json:"jira"         toml:"jira"         yaml:"jira"`
	Spell        SpellConfig        `json:"spell"        toml:"spell"        yaml:"spell"`
	Rules        RulesConfig        `json:"rules"        toml:"rules"        yaml:"rules"`
	Output       string             `json:"output"       toml:"output"       yaml:"output"`
}

// MessageConfig contains configuration for commit message validation.
type MessageConfig struct {
	Subject SubjectConfig `json:"subject" toml:"subject" yaml:"subject"`
	Body    BodyConfig    `json:"body"    toml:"body"    yaml:"body"`
}

// SubjectConfig contains configuration options for commit subject validation.
type SubjectConfig struct {
	MaxLength         int      `json:"max_length"         toml:"max_length"         yaml:"max_length"`
	Case              string   `json:"case"               toml:"case"               yaml:"case"`
	RequireImperative bool     `json:"require_imperative" toml:"require_imperative" yaml:"require_imperative"`
	ForbidEndings     []string `json:"forbid_endings"     toml:"forbid_endings"     yaml:"forbid_endings"`
}

// BodyConfig contains configuration options for commit body validation.
type BodyConfig struct {
	Required         bool `json:"required"           toml:"required"           yaml:"required"`
	MinLength        int  `json:"min_length"         toml:"min_length"         yaml:"min_length"`
	AllowSignoffOnly bool `json:"allow_signoff_only" toml:"allow_signoff_only" yaml:"allow_signoff_only"`
	MinSignoffCount  int  `json:"min_signoff_count"  toml:"min_signoff_count"  yaml:"min_signoff_count"`
}

// ConventionalConfig contains configuration options for conventional commit format validation.
type ConventionalConfig struct {
	RequireScope         bool     `json:"require_scope"          toml:"require_scope"          yaml:"require_scope"`
	Types                []string `json:"types"                  toml:"types"                  yaml:"types"`
	Scopes               []string `json:"scopes"                 toml:"scopes"                 yaml:"scopes"`
	AllowBreaking        bool     `json:"allow_breaking"         toml:"allow_breaking"         yaml:"allow_breaking"`
	MaxDescriptionLength int      `json:"max_description_length" toml:"max_description_length" yaml:"max_description_length"`
}

// SignatureConfig contains configuration options for cryptographic signature validation.
type SignatureConfig struct {
	Required       bool     `json:"required"        toml:"required"        yaml:"required"`
	VerifyFormat   bool     `json:"verify_format"   toml:"verify_format"   yaml:"verify_format"`
	KeyDirectory   string   `json:"key_directory"   toml:"key_directory"   yaml:"key_directory"`
	AllowedSigners []string `json:"allowed_signers" toml:"allowed_signers" yaml:"allowed_signers"`
}

// IdentityConfig contains configuration options for commit author identity validation.
type IdentityConfig struct {
	AllowedAuthors []string `json:"allowed_authors" toml:"allowed_authors" yaml:"allowed_authors"`
}

// RepoConfig contains configuration options for repository-level validation.
type RepoConfig struct {
	MaxCommitsAhead   int    `json:"max_commits_ahead"   toml:"max_commits_ahead"   yaml:"max_commits_ahead"`
	ReferenceBranch   string `json:"reference_branch"    toml:"reference_branch"    yaml:"reference_branch"`
	AllowMergeCommits bool   `json:"allow_merge_commits" toml:"allow_merge_commits" yaml:"allow_merge_commits"`
}

// JiraConfig contains configuration options for JIRA reference validation.
type JiraConfig struct {
	ProjectPrefixes      []string `json:"project_prefixes"       toml:"project_prefixes"       yaml:"project_prefixes"`
	RequireInBody        bool     `json:"require_in_body"        toml:"require_in_body"        yaml:"require_in_body"`
	RequireInSubject     bool     `json:"require_in_subject"     toml:"require_in_subject"     yaml:"require_in_subject"`
	IgnoreTicketPatterns []string `json:"ignore_ticket_patterns" toml:"ignore_ticket_patterns" yaml:"ignore_ticket_patterns"`
}

// SpellConfig contains configuration options for spell checking.
type SpellConfig struct {
	IgnoreWords []string `json:"ignore_words" toml:"ignore_words" yaml:"ignore_words"`
	Locale      string   `json:"locale"       toml:"locale"       yaml:"locale"`
}

// RulesConfig contains configuration for rule activation.
type RulesConfig struct {
	Enabled  []string `json:"enabled"  toml:"enabled"  yaml:"enabled"`
	Disabled []string `json:"disabled" toml:"disabled" yaml:"disabled"`
}
