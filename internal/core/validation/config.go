// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

// Package validation provides core validation components.
package validation

// Config contains the configuration needed by the validation engine.
// This is a simplified version of the config.Config structure that contains only
// what the validation engine needs.
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
	Case string

	// Imperative enforces the use of imperative verbs as the first word of a description.
	Imperative bool

	// InvalidSuffixes lists characters that cannot be used at the end of the subject.
	InvalidSuffixes string

	// MaxLength is the maximum length of the commit subject.
	MaxLength int

	// Jira holds Jira-related validation configuration.
	Jira JiraConfig
}

// JiraConfig holds configuration for Jira ticket references.
type JiraConfig struct {
	// Projects specifies the allowed Jira project keys.
	Projects []string

	// Required indicates whether a Jira key must be present.
	Required bool

	// BodyRef indicates whether a Jira key must be present in body ref.
	BodyRef bool

	// Pattern specifies the regex pattern for Jira keys.
	Pattern string
}

// BodyConfig holds configuration for commit body validation.
type BodyConfig struct {
	// Required enforces that the current commit has a body.
	Required bool

	// AllowSignOffOnly determines if a body with only sign-off lines is allowed.
	AllowSignOffOnly bool
}

// ConventionalConfig holds configuration for conventional commit format validation.
type ConventionalConfig struct {
	// MaxDescriptionLength specifies the maximum allowed length for the description.
	MaxDescriptionLength int

	// Scopes lists the allowed scopes for conventional commits.
	Scopes []string

	// Types lists the allowed types for conventional commits.
	Types []string

	// Required indicates whether Conventional Commits are required.
	Required bool
}

// SpellCheckConfig holds configuration for spell checking.
type SpellCheckConfig struct {
	// Locale specifies the language/locale to use for spell checking.
	Locale string

	// Enabled indicates whether spell checking is enabled.
	Enabled bool

	// IgnoreWords specifies words to ignore during spell checking
	IgnoreWords []string

	// CustomWords specifies custom word mappings for spell checking
	CustomWords map[string]string

	// MaxErrors specifies the maximum number of spelling errors allowed
	MaxErrors int
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
}

// RepositoryConfig holds configuration for repository-related validation.
type RepositoryConfig struct {
	// Reference branch for comparison (usually main or master).
	Reference string

	// MaxCommitsAhead specifies the maximum allowed commits ahead of reference branch.
	MaxCommitsAhead int

	// CheckCommitsAhead determines whether to check the number of commits ahead.
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

// DefaultConfig returns the default validation configuration.
func DefaultConfig() Config {
	return Config{
		Subject: SubjectConfig{
			Case:            "lower",
			Imperative:      true,
			InvalidSuffixes: ".,;:!?",
			MaxLength:       100,
			Jira: JiraConfig{
				Required: false,
				BodyRef:  false,
				Pattern:  "[A-Z]+-\\d+",
			},
		},
		Body: BodyConfig{
			Required:         true,
			AllowSignOffOnly: false,
		},
		Conventional: ConventionalConfig{
			MaxDescriptionLength: 72,
			Types: []string{
				"feat", "fix", "docs", "style", "refactor",
				"perf", "test", "build", "ci", "chore", "revert",
			},
			Scopes:   []string{},
			Required: false,
		},
		SpellCheck: SpellCheckConfig{
			Locale:      "US",
			Enabled:     false,
			IgnoreWords: []string{},
			CustomWords: map[string]string{},
			MaxErrors:   0,
		},
		Security: SecurityConfig{
			SignatureRequired:     true,
			AllowedSignatureTypes: []string{"gpg", "ssh"},
			SignOffRequired:       true,
			AllowMultipleSignOffs: true,
		},
		Repository: RepositoryConfig{
			Reference:         "main",
			MaxCommitsAhead:   5,
			CheckCommitsAhead: true,
		},
		Rules: RulesConfig{
			EnabledRules:  []string{},
			DisabledRules: []string{},
		},
	}
}

// FromConfig converts the config.Config to a Config.
func FromConfig(_ interface{}) Config {
	// For now, just return the default config
	// In a real implementation, we would convert the config.Config to our ValidationConfig
	return DefaultConfig()
}
