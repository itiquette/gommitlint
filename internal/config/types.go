// SPDX-FileCopyrightText: 2025 itiquette/gommitlint <https://github.com/itiquette/gommitlint>
//
// SPDX-License-Identifier: EUPL-1.2

package config

import (
	"github.com/itiquette/gommitlint/internal/contextx"
)

// GommitlintConfig is the root configuration structure for the application.
// This preserves the exact original YAML format while using value semantics.
type GommitlintConfig struct {
	// Content validation rules
	Subject            SubjectConfig      `json:"subject,omitempty"            koanf:"subject"`
	Body               BodyConfig         `json:"body,omitempty"               koanf:"body"`
	ConventionalCommit ConventionalConfig `json:"conventionalCommit,omitempty" koanf:"conventional-commit"`
	SpellCheck         SpellCheckConfig   `json:"spellcheck,omitempty"         koanf:"spellcheck"`
	// Security validation rules
	Signature       SignatureConfig `json:"signature,omitempty" koanf:"signature"`
	SignOffRequired bool            `json:"signOff,omitempty"   koanf:"sign-off"`
	// Misc validation rules
	NCommitsAhead      bool   `json:"nCommitsAhead,omitempty"     koanf:"n-commits-ahead"`
	IgnoreMergeCommits bool   `json:"ignoreMergeCommit,omitempty" koanf:"ignore-merge-commit"`
	Reference          string `json:"reference,omitempty"         koanf:"reference"`
}

// Config contains all configuration options in a single immutable structure
// organized by logical domains. This simplifies configuration management by
// centralizing all options and providing value-semantic accessor methods.
type Config struct {
	// Subject configuration options
	Subject SubjectConfig

	// Body configuration options
	Body BodyConfig

	// Conventional commit options
	Conventional ConventionalConfig

	// SpellCheck options
	SpellCheck SpellCheckConfig

	// Security options
	Security SecurityConfig

	// Repository options
	Repository RepositoryConfig

	// Rule activation options
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

	// Strict enables additional validation checks.
	Strict bool
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

// SignatureConfig holds configuration for signature validation.
type SignatureConfig struct {
	// Identity configures identity verification for signatures.
	Identity IdentityConfig

	// Required enforces that the commit has a valid signature.
	Required bool
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

	// IdentityConfig configures identity verification for signatures.
	Identity IdentityConfig
}

// IdentityConfig holds configuration for identity verification.
type IdentityConfig struct {
	// PublicKeyURI points to a file containing authorized public keys.
	PublicKeyURI string
}

// RepositoryConfig holds configuration for repository-related validation.
type RepositoryConfig struct {
	// Reference branch for comparison (usually main or master).
	Reference string

	// IgnoreMergeCommits indicates whether merge commits should be ignored.
	IgnoreMergeCommits bool

	// MaxCommitsAhead specifies the maximum allowed commits ahead of reference branch.
	MaxCommitsAhead int

	// CheckCommitsAhead enables checking for commits ahead of reference branch.
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

// NewConfig creates a new Config with default values.
func NewConfig() Config {
	return Config{
		Subject:      DefaultSubjectConfig(),
		Body:         DefaultBodyConfig(),
		Conventional: DefaultConventionalConfig(),
		SpellCheck:   DefaultSpellCheckConfig(),
		Security:     DefaultSecurityConfig(),
		Repository:   DefaultRepositoryConfig(),
		Rules:        DefaultRulesConfig(),
	}
}

// Validate returns validation errors for the configuration.
func (c Config) Validate() []error {
	return ValidateConfig(c)
}

// IsValid returns whether the configuration is valid.
func (c Config) IsValid() bool {
	return len(c.Validate()) == 0
}

// Subject configuration transformer methods - these create new Config instances with updated values
// The accessor methods for domain interfaces are defined in domain_adapter.go

// Subject transformation methods

// WithSubjectMaxLength returns a new config with the specified subject max length.
func (c Config) WithSubjectMaxLength(maxLength int) Config {
	result := c // Create a copy
	result.Subject.MaxLength = maxLength

	return result
}

// WithSubjectCase returns a new config with the specified subject case style.
func (c Config) WithSubjectCase(caseStyle string) Config {
	result := c // Create a copy
	result.Subject.Case = caseStyle

	return result
}

// WithSubjectImperative returns a new config with the subject imperative setting.
func (c Config) WithSubjectImperative(imperative bool) Config {
	result := c // Create a copy
	result.Subject.Imperative = imperative

	return result
}

// WithSubjectInvalidSuffixes returns a new config with the specified invalid suffixes.
func (c Config) WithSubjectInvalidSuffixes(suffixes string) Config {
	result := c // Create a copy
	result.Subject.InvalidSuffixes = suffixes

	return result
}

// Body configuration transformer methods - these create new Config instances with updated values
// The accessor methods for domain interfaces are defined in domain_adapter.go

// Body transformation methods

// WithBodyRequired returns a new config with the body required setting.
func (c Config) WithBodyRequired(required bool) Config {
	result := c // Create a copy
	result.Body.Required = required

	return result
}

// WithBodyAllowSignOffOnly returns a new config with the sign-off only setting.
func (c Config) WithBodyAllowSignOffOnly(allow bool) Config {
	result := c // Create a copy
	result.Body.AllowSignOffOnly = allow

	return result
}

// Conventional configuration transformer methods - these create new Config instances with updated values
// The accessor methods for domain interfaces are defined in domain_adapter.go

// Conventional transformation methods

// WithConventionalTypes returns a new config with the specified conventional types.
func (c Config) WithConventionalTypes(types []string) Config {
	result := c // Create a copy
	result.Conventional.Types = deepCopyStringSlice(types)

	return result
}

// WithConventionalScopes returns a new config with the specified conventional scopes.
func (c Config) WithConventionalScopes(scopes []string) Config {
	result := c // Create a copy
	result.Conventional.Scopes = deepCopyStringSlice(scopes)

	return result
}

// WithConventionalMaxDescriptionLength returns a new config with the max description length.
func (c Config) WithConventionalMaxDescriptionLength(maxLength int) Config {
	result := c // Create a copy
	result.Conventional.MaxDescriptionLength = maxLength

	return result
}

// WithConventionalRequired returns a new config with the conventional required setting.
func (c Config) WithConventionalRequired(required bool) Config {
	result := c // Create a copy
	result.Conventional.Required = required

	return result
}

// JIRA configuration transformer methods - these create new Config instances with updated values
// The accessor methods for domain interfaces are defined in domain_adapter.go

// JIRA transformation methods

// WithJiraProjects returns a new config with the specified JIRA projects.
func (c Config) WithJiraProjects(projects []string) Config {
	result := c // Create a copy
	result.Subject.Jira.Projects = deepCopyStringSlice(projects)

	return result
}

// WithJiraBodyRef returns a new config with the JIRA body ref setting.
func (c Config) WithJiraBodyRef(bodyRef bool) Config {
	result := c // Create a copy
	result.Subject.Jira.BodyRef = bodyRef

	return result
}

// WithJiraRequired returns a new config with the JIRA required setting.
func (c Config) WithJiraRequired(required bool) Config {
	result := c // Create a copy
	result.Subject.Jira.Required = required

	return result
}

// WithJiraPattern returns a new config with the JIRA pattern.
func (c Config) WithJiraPattern(pattern string) Config {
	result := c // Create a copy
	result.Subject.Jira.Pattern = pattern

	return result
}

// WithJiraStrict returns a new config with the JIRA strict setting.
func (c Config) WithJiraStrict(strict bool) Config {
	result := c // Create a copy
	result.Subject.Jira.Strict = strict

	return result
}

// Security configuration transformer methods - these create new Config instances with updated values
// The accessor methods for domain interfaces are defined in domain_adapter.go

// Security transformation methods

// WithSignatureRequired returns a new config with the signature required setting.
func (c Config) WithSignatureRequired(required bool) Config {
	result := c // Create a copy
	result.Security.SignatureRequired = required

	return result
}

// WithAllowedSignatureTypes returns a new config with the allowed signature types.
func (c Config) WithAllowedSignatureTypes(types []string) Config {
	result := c // Create a copy
	result.Security.AllowedSignatureTypes = deepCopyStringSlice(types)

	return result
}

// WithSignOffRequired returns a new config with the sign-off required setting.
func (c Config) WithSignOffRequired(required bool) Config {
	result := c // Create a copy
	result.Security.SignOffRequired = required

	return result
}

// WithAllowMultipleSignOffs returns a new config with the multiple sign-offs setting.
func (c Config) WithAllowMultipleSignOffs(allow bool) Config {
	result := c // Create a copy
	result.Security.AllowMultipleSignOffs = allow

	return result
}

// WithIdentityPublicKeyURI returns a new config with the public key URI.
func (c Config) WithIdentityPublicKeyURI(uri string) Config {
	result := c // Create a copy
	result.Security.Identity.PublicKeyURI = uri

	return result
}

// SpellCheck configuration transformer methods - these create new Config instances with updated values
// The accessor methods for domain interfaces are defined in domain_adapter.go

// SpellCheck transformation methods

// WithSpellLocale returns a new config with the spell check locale.
func (c Config) WithSpellLocale(locale string) Config {
	result := c // Create a copy
	result.SpellCheck.Locale = locale

	return result
}

// WithSpellEnabled returns a new config with the spell check enabled setting.
func (c Config) WithSpellEnabled(enabled bool) Config {
	result := c // Create a copy
	result.SpellCheck.Enabled = enabled

	return result
}

// WithSpellIgnoreWords returns a new config with the spell check ignore words.
func (c Config) WithSpellIgnoreWords(words []string) Config {
	result := c // Create a copy
	result.SpellCheck.IgnoreWords = deepCopyStringSlice(words)

	return result
}

// WithSpellCustomWords returns a new config with the spell check custom words.
func (c Config) WithSpellCustomWords(words map[string]string) Config {
	result := c // Create a copy
	result.SpellCheck.CustomWords = deepCopyStringMap(words)

	return result
}

// WithSpellMaxErrors returns a new config with the spell check max errors.
func (c Config) WithSpellMaxErrors(maxErrors int) Config {
	result := c // Create a copy
	result.SpellCheck.MaxErrors = maxErrors

	return result
}

// Repository configuration transformer methods - these create new Config instances with updated values
// The accessor methods for domain interfaces are defined in domain_adapter.go

// Repository transformation methods

// WithReferenceBranch returns a new config with the reference branch.
func (c Config) WithReferenceBranch(branch string) Config {
	result := c // Create a copy
	result.Repository.Reference = branch

	return result
}

// WithIgnoreMergeCommits returns a new config with the ignore merge commits setting.
func (c Config) WithIgnoreMergeCommits(ignore bool) Config {
	result := c // Create a copy
	result.Repository.IgnoreMergeCommits = ignore

	return result
}

// WithMaxCommitsAhead returns a new config with the max commits ahead.
func (c Config) WithMaxCommitsAhead(maxCommits int) Config {
	result := c // Create a copy
	result.Repository.MaxCommitsAhead = maxCommits

	return result
}

// WithCheckCommitsAhead returns a new config with the check commits ahead setting.
func (c Config) WithCheckCommitsAhead(check bool) Config {
	result := c // Create a copy
	result.Repository.CheckCommitsAhead = check

	return result
}

// Rule configuration transformer methods - these create new Config instances with updated values
// The accessor methods for domain interfaces are defined in domain_adapter.go

// Rule transformation methods

// WithEnabledRules returns a new config with the enabled rules.
func (c Config) WithEnabledRules(rules []string) Config {
	result := c // Create a copy
	result.Rules.EnabledRules = deepCopyStringSlice(rules)

	return result
}

// WithDisabledRules returns a new config with the disabled rules.
func (c Config) WithDisabledRules(rules []string) Config {
	result := c // Create a copy
	result.Rules.DisabledRules = deepCopyStringSlice(rules)

	return result
}

// Helper functions

// deepCopyStringSlice creates a deep copy of a string slice.
func deepCopyStringSlice(src []string) []string {
	return contextx.DeepCopy(src)
}

// deepCopyStringMap creates a deep copy of a string map.
func deepCopyStringMap(src map[string]string) map[string]string {
	if src == nil {
		return nil
	}

	dst := make(map[string]string, len(src))
	for k, v := range src {
		dst[k] = v
	}

	return dst
}
