// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package config provides pure configuration types for gommitlint.
// This package contains only data structures with no behavior.
package config

// Config represents the complete configuration for gommitlint.
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

// SigningConfig contains configuration options for commit signing validation.
type SigningConfig struct {
	RequireSignature    bool     `json:"require_signature"     yaml:"require_signature"`
	RequireVerification bool     `json:"require_verification"  yaml:"require_verification"`
	RequireMultiSignoff bool     `json:"require_multi_signoff" yaml:"require_multi_signoff"`
	KeyDirectory        string   `json:"key_directory"         yaml:"key_directory"`
	AllowedSigners      []string `json:"allowed_signers"       yaml:"allowed_signers"`
}

// RepoConfig contains configuration options for repository-level validation.
type RepoConfig struct {
	MaxCommitsAhead   int    `json:"max_commits_ahead"   yaml:"max_commits_ahead"`
	ReferenceBranch   string `json:"reference_branch"    yaml:"reference_branch"`
	AllowMergeCommits bool   `json:"allow_merge_commits" yaml:"allow_merge_commits"`
}

// JiraConfig contains configuration options for JIRA reference validation.
type JiraConfig struct {
	ProjectPrefixes      []string `json:"project_prefixes"       yaml:"project_prefixes"`
	RequireInBody        bool     `json:"require_in_body"        yaml:"require_in_body"`
	RequireInSubject     bool     `json:"require_in_subject"     yaml:"require_in_subject"`
	IgnoreTicketPatterns []string `json:"ignore_ticket_patterns" yaml:"ignore_ticket_patterns"`
}

// SpellConfig contains configuration options for spell checking.
type SpellConfig struct {
	IgnoreWords []string `json:"ignore_words" yaml:"ignore_words"`
	Locale      string   `json:"locale"       yaml:"locale"`
}

// RulesConfig contains configuration for rule activation.
type RulesConfig struct {
	Enabled  []string `json:"enabled"  yaml:"enabled"`
	Disabled []string `json:"disabled" yaml:"disabled"`
}
