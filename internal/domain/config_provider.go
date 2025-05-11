// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package domain

// ValidationConfig defines a domain interface for accessing configuration
// needed by the validation system. This interface decouples validation rules
// from the specific configuration implementation.
type ValidationConfig interface {
	// GetSubjectConfig returns configuration for subject validation
	GetSubjectConfig() SubjectValidationConfig

	// GetBodyConfig returns configuration for body validation
	GetBodyConfig() BodyValidationConfig

	// GetConventionalConfig returns configuration for conventional commit validation
	GetConventionalConfig() ConventionalValidationConfig

	// GetSpellCheckConfig returns configuration for spell checking
	GetSpellCheckConfig() SpellCheckValidationConfig

	// GetSecurityConfig returns configuration for security validation
	GetSecurityConfig() SecurityValidationConfig

	// GetRepositoryConfig returns configuration for repository validation
	GetRepositoryConfig() RepositoryValidationConfig

	// GetRulesConfig returns configuration for rule enablement
	GetRulesConfig() RulesValidationConfig
}

// SubjectValidationConfig defines configuration for subject validation.
type SubjectValidationConfig struct {
	// Case specifies the case that the first word of the description must have
	Case string

	// Imperative enforces the use of imperative verbs
	Imperative bool

	// InvalidSuffixes lists characters that cannot be used at the end of the subject
	InvalidSuffixes string

	// MaxLength is the maximum length of the commit subject
	MaxLength int

	// Jira holds Jira-related validation configuration
	Jira JiraValidationConfig
}

// JiraValidationConfig defines configuration for Jira ticket references.
type JiraValidationConfig struct {
	// Projects specifies the allowed Jira project keys
	Projects []string

	// Required indicates whether a Jira key must be present
	Required bool

	// BodyRef indicates whether a Jira key must be present in body ref
	BodyRef bool

	// Pattern specifies the regex pattern for Jira keys
	Pattern string

	// Strict enables additional validation checks
	Strict bool
}

// BodyValidationConfig defines configuration for body validation.
type BodyValidationConfig struct {
	// Required enforces that the current commit has a body
	Required bool

	// AllowSignOffOnly determines if a body with only sign-off lines is allowed
	AllowSignOffOnly bool
}

// ConventionalValidationConfig defines configuration for conventional commit format.
type ConventionalValidationConfig struct {
	// MaxDescriptionLength specifies the maximum allowed length for the description
	MaxDescriptionLength int

	// Scopes lists the allowed scopes for conventional commits
	Scopes []string

	// Types lists the allowed types for conventional commits
	Types []string

	// Required indicates whether Conventional Commits are required
	Required bool
}

// SpellCheckValidationConfig defines configuration for spell checking.
type SpellCheckValidationConfig struct {
	// Locale specifies the language/locale to use for spell checking
	Locale string

	// Enabled indicates whether spell checking is enabled
	Enabled bool

	// IgnoreWords specifies words to ignore during spell checking
	IgnoreWords []string

	// CustomWords specifies custom word mappings for spell checking
	CustomWords map[string]string

	// MaxErrors specifies the maximum number of spelling errors allowed
	MaxErrors int

	// IgnoreCase indicates whether to ignore case differences when checking spelling
	IgnoreCase bool
}

// SecurityValidationConfig defines configuration for security-related validation.
type SecurityValidationConfig struct {
	// SignatureRequired enforces that the commit has a valid signature
	SignatureRequired bool

	// AllowedSignatureTypes specifies the allowed signature types (gpg, ssh)
	AllowedSignatureTypes []string

	// SignOffRequired enforces that commits are signed off
	SignOffRequired bool

	// AllowMultipleSignOffs determines if multiple sign-offs are allowed
	AllowMultipleSignOffs bool

	// Identity configures identity verification for signatures
	Identity IdentityValidationConfig
}

// IdentityValidationConfig defines configuration for identity verification.
type IdentityValidationConfig struct {
	// PublicKeyURI points to a file containing authorized public keys
	PublicKeyURI string
}

// RepositoryValidationConfig defines configuration for repository-related validation.
type RepositoryValidationConfig struct {
	// Reference branch for comparison (usually main or master)
	Reference string

	// IgnoreMergeCommits indicates whether merge commits should be ignored
	IgnoreMergeCommits bool

	// MaxCommitsAhead specifies the maximum allowed commits ahead of reference branch
	MaxCommitsAhead int

	// CheckCommitsAhead enables checking for commits ahead of reference branch
	CheckCommitsAhead bool
}

// RulesValidationConfig defines configuration for rule enablement/disablement.
type RulesValidationConfig struct {
	// EnabledRules lists rules that are explicitly enabled
	EnabledRules []string

	// DisabledRules lists rules that are explicitly disabled
	DisabledRules []string
}
